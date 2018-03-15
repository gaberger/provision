package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/digitalrebar/logger"
	"github.com/digitalrebar/provision/backend/index"
	"github.com/digitalrebar/provision/models"
	"github.com/digitalrebar/store"
)

func BasicContent() store.Store {
	var (
		ignoreBoot = &models.BootEnv{
			Name:        `ignore`,
			Description: "The boot environment you should use to have unknown machines boot off their local hard drive",
			OS: models.OsInfo{
				Name: `ignore`,
			},
			OnlyUnknown: true,
			Templates: []models.TemplateInfo{
				{
					Name: "pxelinux",
					Path: `pxelinux.cfg/default`,
					Contents: `DEFAULT local
PROMPT 0
TIMEOUT 10
LABEL local
localboot 0
`,
				},
				{
					Name: `ipxe`,
					Path: `default.ipxe`,
					Contents: `#!ipxe
chain {{.ProvisionerURL}}/${netX/mac}.ipxe && exit || goto chainip
:chainip
chain tftp://{{.ProvisionerAddress}}/${netX/ip}.ipxe || exit
`,
				},
			},
			Meta: map[string]string{
				"feature-flags": "change-stage-v2",
				"icon":          "circle thin",
				"color":         "green",
				"title":         "Digital Rebar Provision",
			},
		}

		localBoot = &models.BootEnv{
			Name:        "local",
			Description: "The boot environment you should use to have known machines boot off their local hard drive",
			OS: models.OsInfo{
				Name: "local",
			},
			OnlyUnknown: false,
			Templates: []models.TemplateInfo{
				{
					Name: "pxelinux",
					Path: "pxelinux.cfg/{{.Machine.HexAddress}}",
					Contents: `DEFAULT local
PROMPT 0
TIMEOUT 10
LABEL local
localboot 0
`,
				},
				{
					Name: "ipxe",
					Path: "{{.Machine.Address}}.ipxe",
					Contents: `#!ipxe
exit
`,
				},
				{
					Name: "pxelinux-mac",
					Path: "pxelinux.cfg/{{.Machine.MacAddr \"pxelinux\"}}",
					Contents: `DEFAULT local
PROMPT 0
TIMEOUT 10
LABEL local
localboot 0
`,
				},
				{
					Name: "ipxe-mac",
					Path: "{{.Machine.MacAddr \"ipxe\"}}.ipxe",
					Contents: `#!ipxe
exit
`,
				},
			},
			Meta: map[string]string{
				"feature-flags": "change-stage-v2",
				"icon":          "radio",
				"color":         "green",
				"title":         "Digital Rebar Provision",
			},
		}
		noneStage = &models.Stage{
			Name:        "none",
			Description: "Noop / Nothing stage",
			Meta: map[string]string{
				"icon":  "circle thin",
				"color": "green",
				"title": "Digital Rebar Provision",
			},
		}
		localStage = &models.Stage{
			Name:        "local",
			BootEnv:     "local",
			Description: "Stage to boot into the local BootEnv.",
			Meta: map[string]string{
				"icon":  "radio",
				"color": "green",
				"title": "Digital Rebar Provision",
			},
		}
	)
	res, _ := store.Open("memory:///")
	bootEnvs, _ := res.MakeSub("bootenvs")
	stages, _ := res.MakeSub("stages")
	localBoot.ClearValidation()
	ignoreBoot.ClearValidation()
	noneStage.ClearValidation()
	localStage.ClearValidation()
	localBoot.Fill()
	ignoreBoot.Fill()
	noneStage.Fill()
	localStage.Fill()
	bootEnvs.Save("local", localBoot)
	bootEnvs.Save("ignore", ignoreBoot)
	stages.Save("none", noneStage)
	stages.Save("local", localStage)
	res.(*store.Memory).SetMetaData(map[string]string{
		"Name":        "BasicStore",
		"Description": "Default objects that must be present",
		"Version":     "Unversioned",
		"Type":        "default",
	})
	return res
}

type followUpSaver interface {
	followUpSave()
}

type followUpDeleter interface {
	followUpDelete()
}

type AuthSaver interface {
	AuthKey() string
}

// dtobjs is an in-memory cache of all the objects we could
// reference. The implementation of this may need to change from
// storing a slice of things to a more elaborate datastructure at some
// point in time.  Since that point in time is when the slices are
// forced out of CPU cache, I am not terribly concerned for now.
// Until that point is reached, sorting and searching slices is
// fantastically efficient.
type Store struct {
	sync.Mutex
	index.Index
	backingStore store.Store
}

func (s *Store) getBackend(obj models.Model) store.Store {
	return s.backingStore
}

type dtSetter interface {
	models.Model
	setDT(*DataTracker)
}

func Fill(t store.KeySaver) {
	switch obj := t.(type) {
	case *Stage:
		if obj.Stage == nil {
			obj.Stage = &models.Stage{}
		}
	case *BootEnv:
		if obj.BootEnv == nil {
			obj.BootEnv = &models.BootEnv{}
		}
	case *Job:
		if obj.Job == nil {
			obj.Job = &models.Job{}
		}
	case *Lease:
		if obj.Lease == nil {
			obj.Lease = &models.Lease{}
		}
	case *Machine:
		if obj.Machine == nil {
			obj.Machine = &models.Machine{}
		}
	case *Param:
		if obj.Param == nil {
			obj.Param = &models.Param{}
		}
	case *Plugin:
		if obj.Plugin == nil {
			obj.Plugin = &models.Plugin{}
		}
	case *Pref:
		if obj.Pref == nil {
			obj.Pref = &models.Pref{}
		}
	case *Profile:
		if obj.Profile == nil {
			obj.Profile = &models.Profile{}
		}
	case *Reservation:
		if obj.Reservation == nil {
			obj.Reservation = &models.Reservation{}
		}
	case *Subnet:
		if obj.Subnet == nil {
			obj.Subnet = &models.Subnet{}
		}
	case *Task:
		if obj.Task == nil {
			obj.Task = &models.Task{}
		}
	case *Template:
		if obj.Template == nil {
			obj.Template = &models.Template{}
		}
	case *User:
		if obj.User == nil {
			obj.User = &models.User{}
		}
	case *Workflow:
		if obj.Workflow == nil {
			obj.Workflow = &models.Workflow{}
		}
	default:
		panic(fmt.Sprintf("Unknown backend model %T", t))
	}
}

func ModelToBackend(m models.Model) store.KeySaver {
	switch obj := m.(type) {
	case store.KeySaver:
		return obj
	case *models.Stage:
		return &Stage{Stage: obj}
	case *models.BootEnv:
		return &BootEnv{BootEnv: obj}
	case *models.Job:
		return &Job{Job: obj}
	case *models.Lease:
		return &Lease{Lease: obj}
	case *models.Machine:
		return &Machine{Machine: obj}
	case *models.Param:
		return &Param{Param: obj}
	case *models.Plugin:
		return &Plugin{Plugin: obj}
	case *models.Pref:
		return &Pref{Pref: obj}
	case *models.Profile:
		return &Profile{Profile: obj}
	case *models.Reservation:
		return &Reservation{Reservation: obj}
	case *models.Subnet:
		return &Subnet{Subnet: obj}
	case *models.Task:
		return &Task{Task: obj}
	case *models.Template:
		return &Template{Template: obj}
	case *models.User:
		return &User{User: obj}
	case *models.Workflow:
		return &Workflow{Workflow: obj}
	default:
		panic(fmt.Sprintf("Unknown model %T", m))
	}
}

func toBackend(m models.Model, rt *RequestTracker) store.KeySaver {
	if res, ok := m.(store.KeySaver); ok {
		if v, ok := res.(validator); ok {
			v.setRT(rt)
		}
		return res
	}
	var ours store.KeySaver
	if rt != nil {
		backend := rt.stores(m.Prefix())
		if backend == nil {
			rt.Panicf("No store for %T", m)
		}
		if this := backend.Find(m.Key()); this != nil {
			ours = this.(store.KeySaver)
		}
	}

	switch obj := m.(type) {
	case *models.Stage:
		var res Stage
		if ours != nil {
			res = *ours.(*Stage)
		} else {
			res = Stage{}
		}
		res.Stage = obj
		res.rt = rt
		return &res
	case *models.BootEnv:
		var res BootEnv
		if ours != nil {
			res = *ours.(*BootEnv)
		} else {
			res = BootEnv{}
		}
		res.BootEnv = obj
		res.rt = rt
		return &res
	case *models.Job:
		var res Job
		if ours != nil {
			res = *ours.(*Job)
		} else {
			res = Job{}
		}
		res.Job = obj
		res.rt = rt
		return &res
	case *models.Lease:
		var res Lease
		if ours != nil {
			res = *ours.(*Lease)
		} else {
			res = Lease{}
		}
		res.Lease = obj
		res.rt = rt
		return &res
	case *models.Machine:
		var res Machine
		if ours != nil {
			res = *ours.(*Machine)
		} else {
			res = Machine{}
		}
		res.Machine = obj
		res.rt = rt
		return &res
	case *models.Param:
		var res Param
		if ours != nil {
			res = *ours.(*Param)
		} else {
			res = Param{}
		}
		res.Param = obj
		res.rt = rt
		return &res
	case *models.Plugin:
		var res Plugin
		if ours != nil {
			res = *ours.(*Plugin)
		} else {
			res = Plugin{}
		}
		res.Plugin = obj
		res.rt = rt
		return &res
	case *models.Pref:
		var res Pref
		if ours != nil {
			res = *ours.(*Pref)
		} else {
			res = Pref{}
		}
		res.Pref = obj
		res.rt = rt
		return &res
	case *models.Profile:
		var res Profile
		if ours != nil {
			res = *ours.(*Profile)
		} else {
			res = Profile{}
		}
		res.Profile = obj
		res.rt = rt
		return &res
	case *models.Reservation:
		var res Reservation
		if ours != nil {
			res = *ours.(*Reservation)
		} else {
			res = Reservation{}
		}
		res.Reservation = obj
		res.rt = rt
		return &res
	case *models.Subnet:
		var res Subnet
		if ours != nil {
			res = *ours.(*Subnet)
		} else {
			res = Subnet{}
		}
		res.Subnet = obj
		res.rt = rt
		return &res
	case *models.Task:
		var res Task
		if ours != nil {
			res = *ours.(*Task)
		} else {
			res = Task{}
		}
		res.Task = obj
		res.rt = rt
		return &res
	case *models.Template:
		var res Template
		if ours != nil {
			res = *ours.(*Template)
		} else {
			res = Template{}
		}
		res.Template = obj
		res.rt = rt
		return &res

	case *models.User:
		var res User
		if ours != nil {
			res = *ours.(*User)
		} else {
			res = User{}
		}
		res.User = obj
		res.rt = rt
		return &res
	case *models.Workflow:
		var res Workflow
		if ours != nil {
			res = *ours.(*Workflow)
		} else {
			res = Workflow{}
		}
		res.Workflow = obj
		res.rt = rt
		return &res

	default:
		log.Panicf("Unknown model %T", m)
	}
	return nil
}

func (p *DataTracker) logCheck(prefName, prefVal string) (logName, logTarget string, lvl logger.Level, err error) {
	logName = "default"
	logTarget = "warn"
	lvl = logger.Warn

	switch prefVal {
	case "trace", "debug", "info", "warn", "error", "panic", "fatal":
		logTarget = prefVal
	case "0":
		logTarget = "warn"
	case "1":
		logTarget = "info"
	case "2":
		logTarget = "debug"
	default:
		err = fmt.Errorf("Invalid log value %s for %s,  Ignoring change", prefVal, prefName)
		return
	}
	if logTarget != prefVal {
		p.Logger.Errorf("Pref %s log level %s is obsolete.  Please migrate to using %s",
			prefName, prefVal, logTarget)
	}
	switch prefName {
	case "debugDhcp":
		logName = "dhcp"
	case "debugRenderer":
		logName = "render"
	case "debugBootEnv":
		logName = "bootenv"
	case "debugFrontend":
		logName = "frontend"
	case "debugPlugins":
		logName = "plugin"
	case "logLevel":
		logName = "default"
	default:
		err = fmt.Errorf("Invalid logging preference %s, ignoring change", prefName)
		return
	}
	lvl, err = logger.ParseLevel(logTarget)
	return
}

// DataTracker represents everything there is to know about acting as
// a dataTracker.
type DataTracker struct {
	logger.Logger
	FileRoot            string
	LogRoot             string
	OurAddress          string
	ForceOurAddress     bool
	StaticPort, ApiPort int
	FS                  *FileSystem
	Backend             store.Store
	objs                map[string]*Store
	defaultPrefs        map[string]string
	runningPrefs        map[string]string
	prefMux             *sync.Mutex
	allMux              *sync.RWMutex
	GlobalProfileName   string
	tokenManager        *JwtManager
	rootTemplate        *template.Template
	tmplMux             *sync.Mutex
	thunks              []func()
	thunkMux            *sync.Mutex
	publishers          *Publishers
	macAddrMap          map[string]string
	macAddrMux          *sync.RWMutex
}

func (p *DataTracker) LogFor(s string) logger.Logger {
	return p.Logger.Buffer().Log(s)
}

func (dt *DataTracker) reportPath(s string) string {
	return strings.TrimPrefix(s, dt.FileRoot)
}

type Stores func(string) *Store

func allKeySavers() []models.Model {
	return []models.Model{
		&Pref{},
		&Param{},
		&User{},
		&Template{},
		&Task{},
		&Profile{},
		&BootEnv{},
		&Stage{},
		&Machine{},
		&Subnet{},
		&Reservation{},
		&Lease{},
		&Plugin{},
		&Job{},
		&Workflow{},
	}
}

// LockEnts grabs the requested Store locks a consistent order.
// It returns a function to get an Index that was requested, and
// a function that unlocks the taken locks in the right order.
func (p *DataTracker) lockEnts(ents ...string) (stores Stores, unlocker func()) {
	p.allMux.RLock()
	sortedEnts := make([]string, len(ents))
	copy(sortedEnts, ents)
	s := sort.StringSlice(sortedEnts)
	sort.Sort(sort.Reverse(s))
	sortedRes := map[string]*Store{}
	for _, ent := range s {
		objs, ok := p.objs[ent]
		if !ok {
			log.Panicf("Tried to reference nonexistent object type '%s'", ent)
		}
		sortedRes[ent] = objs
	}
	for _, ent := range s {
		sortedRes[ent].Lock()
	}
	srMux := &sync.Mutex{}
	return func(ref string) *Store {
			srMux.Lock()
			idx, ok := sortedRes[ref]
			srMux.Unlock()
			if !ok {
				log.Panicf("Tried to access unlocked resource %s", ref)
			}
			return idx
		},
		func() {
			srMux.Lock()
			for i := len(s) - 1; i >= 0; i-- {
				sortedRes[s[i]].Unlock()
				delete(sortedRes, s[i])
			}
			srMux.Unlock()
			p.allMux.RUnlock()
		}
}

func (p *DataTracker) lockAll() (stores Stores, unlocker func()) {
	p.allMux.Lock()
	return func(ref string) *Store {
			return p.objs[ref]
		},
		func() {
			p.allMux.Unlock()
		}
}

func (p *DataTracker) LocalIP(remote net.IP) string {
	// If we are behind a NAT, always use Our Address
	if p.ForceOurAddress && p.OurAddress != "" {
		p.Debugf("addrCache: Forced to use static address %s", p.OurAddress)
		return p.OurAddress
	}
	if localIP := LocalFor(p.Logger, remote); localIP != nil {
		return localIP.String()
	}
	// Determining what this is needs to be made smarter, probably by
	// firguing out which interface the default route goes over for ipv4
	// then ipv6, and then figurig out the appropriate address on that
	// interface
	if p.OurAddress != "" {
		return p.OurAddress
	}
	gwIp := DefaultIP(p.Logger)
	if gwIp == nil {
		p.Warnf("Failed to find appropriate local IP to use for %s", remote)
		p.Warnf("No --static-ip and no default gateway to use in its place")
		p.Warnf("Please set --static-ip ")
		return ""
	}
	p.Infof("Falling back to local address %s as default target for remote %s", gwIp, remote)
	return gwIp.String()
}

func (p *DataTracker) rebuildCache(loadRT *RequestTracker) (hard, soft *models.Error) {
	hard = &models.Error{Code: 500, Type: "Failed to load backing objects from cache"}
	soft = &models.Error{Code: 422, Type: ValidationError}
	p.objs = map[string]*Store{}
	objs := allKeySavers()
	for _, obj := range objs {
		prefix := obj.Prefix()
		bk := p.Backend.GetSub(prefix)
		p.objs[prefix] = &Store{backingStore: bk}
		storeObjs, err := store.List(bk, toBackend(obj, loadRT))
		if err != nil {
			// Make fake index to keep others from failing and exploding.
			res := make([]models.Model, 0)
			p.objs[prefix].Index = *index.Create(res)
			hard.Errorf("Unable to load %s: %v", prefix, err)
			continue
		}
		res := make([]models.Model, len(storeObjs))
		for i := range storeObjs {
			res[i] = models.Model(storeObjs[i])
			if v, ok := res[i].(models.Validator); ok && v.Useable() {
				soft.AddError(v.HasError())
			}
		}
		if prefix == "tasks" {
			stack, ok := bk.(*store.StackedStore)
			if ok {
				subStore := stack.Subs()[prefix]
				if subStore != nil {
					sub := stack.Subs()[prefix].(*store.StackedStore)
					for i := range res {
						obj := AsTask(res[i])
						key := obj.Key()
						meta := sub.MetaFor(key)
						if flagStr, ok := meta["feature-flags"]; ok && len(flagStr) > 0 {
							obj.MergeFeatures(strings.Split(flagStr, ","))
						}
						if obj.HasFeature("original-exit-codes") {
							obj.RemoveFeature("sane-exit-codes")
						}
						if !obj.HasFeature("sane-exit-codes") {
							obj.AddFeature("original-exit-codes")
						}
						res[i] = obj
					}
				}
			}
		}

		p.objs[prefix].Index = *index.Create(res)
		if prefix == "bootenvs" {
			for _, thing := range p.objs[prefix].Items() {
				benv := AsBootEnv(thing)
				benv.rt = loadRT
				benv.AddDynamicTree()
				benv.rt = nil
			}
		}

		if prefix == "templates" {
			buf := &bytes.Buffer{}
			for _, thing := range p.objs[prefix].Items() {
				tmpl := AsTemplate(thing)
				fmt.Fprintf(buf, `{{define "%s"}}%s{{end}}`, tmpl.ID, tmpl.Contents)
			}
			root, err := template.New("").Parse(buf.String())
			if err != nil {
				hard.Errorf("Unable to load root templates: %v", err)
				return
			}
			p.rootTemplate = root
			p.rootTemplate.Option("missingkey=error")
		}
	}
	return
}

// This must be locked with ALL locks on the source datatracker from the caller.
func ValidateDataTrackerStore(backend store.Store, logger logger.Logger) (hard, soft error) {
	res := &DataTracker{
		Backend:           backend,
		FileRoot:          "baddir",
		LogRoot:           "baddir",
		StaticPort:        1,
		ApiPort:           2,
		Logger:            logger,
		defaultPrefs:      map[string]string{},
		runningPrefs:      map[string]string{},
		tokenManager:      NewJwtManager([]byte{}, JwtConfig{Method: jwt.SigningMethodHS256}),
		prefMux:           &sync.Mutex{},
		allMux:            &sync.RWMutex{},
		FS:                NewFS(".", logger),
		tmplMux:           &sync.Mutex{},
		GlobalProfileName: "global",
		thunks:            make([]func(), 0),
		thunkMux:          &sync.Mutex{},
		publishers:        &Publishers{},
		macAddrMap:        map[string]string{},
		macAddrMux:        &sync.RWMutex{},
	}

	// Load stores.
	rt := res.Request(logger)
	rt.AllLocked(func(d Stores) {
		a, b := res.rebuildCache(rt)
		hard, soft = a.HasError(), b.HasError()
	})
	return
}

// Create a new DataTracker that will use passed store to save all operational data
func NewDataTracker(backend store.Store,
	fileRoot, logRoot, addr string, forceAddr bool,
	staticPort, apiPort int,
	logger logger.Logger,
	defaultPrefs map[string]string,
	publishers *Publishers) *DataTracker {
	res := &DataTracker{
		Backend:           backend,
		FileRoot:          fileRoot,
		LogRoot:           logRoot,
		StaticPort:        staticPort,
		ApiPort:           apiPort,
		OurAddress:        addr,
		ForceOurAddress:   forceAddr,
		Logger:            logger,
		defaultPrefs:      defaultPrefs,
		runningPrefs:      map[string]string{},
		tokenManager:      NewJwtManager([]byte{}, JwtConfig{Method: jwt.SigningMethodHS256}),
		prefMux:           &sync.Mutex{},
		allMux:            &sync.RWMutex{},
		FS:                NewFS(fileRoot, logger),
		tmplMux:           &sync.Mutex{},
		GlobalProfileName: "global",
		thunks:            make([]func(), 0),
		thunkMux:          &sync.Mutex{},
		publishers:        publishers,
		macAddrMap:        map[string]string{},
		macAddrMux:        &sync.RWMutex{},
	}

	// Make sure incoming writable backend has all stores created
	loadRT := res.Request(logger)
	loadRT.AllLocked(func(d Stores) {
		objs := allKeySavers()
		for _, obj := range objs {
			prefix := obj.Prefix()
			_, err := backend.MakeSub(prefix)
			if err != nil {
				loadRT.Fatalf("dataTracker: Error creating substore %s: %v", prefix, err)
			}
		}
		// Load stores.
		hard, _ := res.rebuildCache(loadRT)
		if hard.HasError() != nil {
			loadRT.Fatalf("dataTracker: Error loading data: %v", hard.HasError())
		}
	})
	// Create minimal content.
	rt := res.Request(res.Logger, "stages", "bootenvs", "preferences", "users", "machines", "profiles", "params")
	rt.Do(func(d Stores) {
		// Load the prefs - overriding defaults.
		savePrefs := false
		for _, prefIsh := range d("preferences").Items() {
			pref := AsPref(prefIsh)
			res.runningPrefs[pref.Name] = pref.Val
		}

		// Set systemGrantorSecret and baseTokenSecret if unset and save it to backing store.
		prefs := res.Prefs()
		for _, pref := range []string{"systemGrantorSecret", "baseTokenSecret"} {
			if val, ok := prefs[pref]; !ok || val == "" {
				prefs[pref] = randString(32)
				savePrefs = true
			}
		}
		// Migrate any number-based logging preferences
		for _, name := range []string{"debugDhcp",
			"debugRenderer",
			"debugBootEnv",
			"debugFrontend",
			"debugPlugins",
			"logLevel",
		} {
			val := prefs[name]
			if val == "" {
				val = "warn"
			}
			logName, logTarget, logLevel, lErr := res.logCheck(name, val)
			if lErr != nil {
				res.Logger.Fatalf("dataTracker: Invalid log level %v", lErr)
			}
			if val != logTarget {
				savePrefs = true
			}
			prefs[name] = logTarget
			// GREG: SetLevel in the logger codebase should track forks and update appropriately.
			res.LogFor(logName).SetLevel(logLevel)
		}
		if savePrefs {
			res.SetPrefs(rt, prefs)
		}
		res.tokenManager.updateKey([]byte(res.pref("baseTokenSecret")))

		if d("profiles").Find(res.GlobalProfileName) == nil {
			res.Infof("Creating %s profile", res.GlobalProfileName)
			rt.Create(&models.Profile{
				Name:        res.GlobalProfileName,
				Description: "Global profile attached automatically to all machines.",
				Params:      map[string]interface{}{},
				Meta: map[string]string{
					"icon":  "world",
					"color": "blue",
					"title": "Digital Rebar Provision",
				},
			})
		}
		users := d("users")
		if users.Count() == 0 {
			res.Infof("Creating rocketskates user")
			user := &User{}
			Fill(user)
			user.Name = "rocketskates"
			if err := user.ChangePassword(rt, "r0cketsk8ts"); err != nil {
				logger.Fatalf("Failed to create rocketskates user: %v", err)
			}
			rt.Create(user)
		}
		machines := d("machines")
		for _, obj := range machines.Items() {
			machine := AsMachine(obj)
			bootEnv := d("bootenvs").Find(machine.BootEnv)
			if bootEnv == nil {
				continue
			}
			err := &models.Error{}
			AsBootEnv(bootEnv).Render(rt, machine, err).register(res.FS)
			if err.ContainsError() {
				logger.Errorf("Error rendering machine %s at startup: %v", machine.UUID(), err)
			}
		}
		if err := res.RenderUnknown(rt); err != nil {
			logger.Fatalf("Failed to render unknown bootenv: %v", err)
		}
	})
	return res
}

func (p *DataTracker) Prefs() map[string]string {
	vals := map[string]string{}
	p.prefMux.Lock()
	for k, v := range p.defaultPrefs {
		vals[k] = v
	}
	for k, v := range p.runningPrefs {
		vals[k] = v
	}
	p.prefMux.Unlock()
	return vals
}

func (p *DataTracker) Pref(name string) (string, error) {
	res, ok := p.Prefs()[name]
	if !ok {
		return "", fmt.Errorf("No such preference %s", name)
	}
	return res, nil
}

func (p *DataTracker) pref(name string) string {
	return p.Prefs()[name]
}

func (p *DataTracker) SetPrefs(rt *RequestTracker, prefs map[string]string) error {
	err := &models.Error{}
	bootenvs := rt.d("bootenvs")
	stages := rt.d("stages")
	lenCheck := func(name, val string) bool {
		if len(val) != 32 {
			err.Errorf("%s: Must be a string of length 32: %s", name, val)
			return false
		}
		return true
	}
	benvCheck := func(name, val string) *BootEnv {
		be := bootenvs.Find(val)
		if be == nil {
			err.Errorf("%s: Bootenv %s does not exist", name, val)
			return nil
		}
		return AsBootEnv(be)
	}
	stageCheck := func(name, val string) bool {
		stage := stages.Find(val)
		if stage == nil {
			err.Errorf("%s: Stage %s does not exist", name, val)
			return false
		}
		return true
	}
	intCheck := func(name, val string) bool {
		_, e := strconv.Atoi(val)
		if e == nil {
			return true
		}
		err.Errorf("%s: %s", name, e.Error())
		return false
	}

	savePref := func(name, val string) bool {
		p.prefMux.Lock()
		defer p.prefMux.Unlock()
		pref := &models.Pref{}
		pref.Name = name
		pref.Val = val
		if _, saveErr := rt.Save(pref); saveErr != nil {
			err.Errorf("%s: Failed to save %s: %v", name, val, saveErr)
			return false
		}
		p.runningPrefs[name] = val
		return true
	}
	for name, val := range prefs {
		switch name {
		case "systemGrantorSecret":
			savePref(name, val)
		case "baseTokenSecret":
			if lenCheck(name, val) && savePref(name, val) {
				p.tokenManager.updateKey([]byte(val))
			}
		case "defaultBootEnv":
			be := benvCheck(name, val)
			if be != nil && !be.OnlyUnknown {
				savePref(name, val)
			}
		case "defaultStage":
			if stageCheck(name, val) {
				savePref(name, val)
			}
		case "unknownBootEnv":
			if benvCheck(name, val) != nil && savePref(name, val) {
				err.AddError(p.RenderUnknown(rt))
			}
		case "unknownTokenTimeout",
			"knownTokenTimeout":
			if intCheck(name, val) {
				savePref(name, val)
			}
		case "debugDhcp",
			"debugRenderer",
			"debugBootEnv",
			"debugFrontend",
			"debugPlugins",
			"logLevel":
			logName, logTarget, logLevel, lErr := p.logCheck(name, val)
			if lErr != nil {
				err.AddError(lErr)
			} else {
				savePref(name, logTarget)
				p.LogFor(logName).SetLevel(logLevel)
			}
		default:
			err.Errorf("Unknown preference %s", name)
		}
	}
	return err.HasError()
}

func (p *DataTracker) setDT(s models.Model) {
	if tgt, ok := s.(dtSetter); ok {
		tgt.setDT(p)
	}
}

func (p *DataTracker) RenderUnknown(rt *RequestTracker) error {
	pref, e := p.Pref("unknownBootEnv")
	if e != nil {
		return e
	}
	envIsh := rt.d("bootenvs").Find(pref)
	if envIsh == nil {
		return fmt.Errorf("No such BootEnv: %s", pref)
	}
	env := AsBootEnv(envIsh)
	err := &models.Error{Object: env, Model: env.Prefix(), Key: env.Key(), Type: "StartupError"}
	if !env.Validated {
		err.AddError(env)
		return err
	}
	if !env.OnlyUnknown {
		err.Errorf("BootEnv %s cannot be used for the unknownBootEnv", env.Name)
		return err
	}
	env.Render(rt, nil, err).register(p.FS)
	return err.HasError()
}

func (p *DataTracker) getBackend(t models.Model) store.Store {
	res, ok := p.objs[t.Prefix()]
	if !ok {
		p.Logger.Fatalf("%s: No registered storage backend!", t.Prefix())
	}
	return res.backingStore
}

func (p *DataTracker) GetToken(tokenString string) (*DrpCustomClaims, error) {
	return p.tokenManager.get(tokenString)
}

func (p *DataTracker) SealClaims(claims *DrpCustomClaims) (string, error) {
	return claims.Seal(p.tokenManager)
}

func (p *DataTracker) Backup() ([]byte, error) {
	keys := make([]string, len(p.objs))
	for k := range p.objs {
		keys = append(keys, k)
	}
	_, unlocker := p.lockEnts(keys...)
	defer unlocker()
	res := map[string][]models.Model{}
	for _, k := range keys {
		res[k] = p.objs[k].Items()
	}
	return json.Marshal(res)
}

// Assumes that all locks are held
func (p *DataTracker) ReplaceBackend(rt *RequestTracker, st store.Store) (hard, soft error) {
	p.Backend = st
	return p.rebuildCache(rt)
}
