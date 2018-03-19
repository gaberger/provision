package backend

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/VictorLowther/jsonpatch2/utils"
	"github.com/digitalrebar/provision/backend/index"
	"github.com/digitalrebar/provision/models"
	"github.com/digitalrebar/store"
)

var explodeMux = &sync.Mutex{}

// BootEnv encapsulates the machine-agnostic information needed by the
// provisioner to set up a boot environment.
//
// swagger:model
type BootEnv struct {
	*models.BootEnv
	validate
	renderers      renderers
	pathLookaside  func(string) (io.Reader, error)
	installRepo    *Repo
	kernelVerified bool
	bootParamsTmpl *template.Template
	rootTemplate   *template.Template
	tmplMux        sync.Mutex
}

func (b *BootEnv) NetBoot() bool {
	return b.OnlyUnknown || b.Kernel != ""
}

func (obj *BootEnv) SetReadOnly(b bool) {
	obj.ReadOnly = b
}

func (obj *BootEnv) SaveClean() store.KeySaver {
	mod := *obj.BootEnv
	mod.ClearValidation()
	return ModelToBackend(&mod)
}

func (b *BootEnv) Indexes() map[string]index.Maker {
	fix := AsBootEnv
	res := index.MakeBaseIndexes(b)
	res["Name"] = index.Make(
		true,
		"string",
		func(i, j models.Model) bool {
			return fix(i).Name < fix(j).Name
		},
		func(ref models.Model) (gte, gt index.Test) {
			name := fix(ref).Name
			return func(s models.Model) bool {
					return fix(s).Name >= name
				},
				func(s models.Model) bool {
					return fix(s).Name > name
				}
		},
		func(s string) (models.Model, error) {
			res := fix(b.New())
			res.Name = s
			return res, nil
		})
	res["OsName"] = index.Make(
		false,
		"string",
		func(i, j models.Model) bool {
			return fix(i).OS.Name < fix(j).OS.Name
		},
		func(ref models.Model) (gte, gt index.Test) {
			name := fix(ref).OS.Name
			return func(s models.Model) bool {
					return fix(s).OS.Name >= name
				},
				func(s models.Model) bool {
					return fix(s).OS.Name > name
				}
		},
		func(s string) (models.Model, error) {
			res := fix(b.New())
			res.OS.Name = s
			return res, nil
		})
	res["OnlyUnknown"] = index.Make(
		false,
		"boolean",
		func(i, j models.Model) bool {
			return !fix(i).OnlyUnknown && fix(j).OnlyUnknown
		},
		func(ref models.Model) (gte, gt index.Test) {
			unknown := fix(ref).OnlyUnknown
			return func(s models.Model) bool {
					v := fix(s).OnlyUnknown
					return v || (v == unknown)
				},
				func(s models.Model) bool {
					return fix(s).OnlyUnknown && !unknown
				}
		},
		func(s string) (models.Model, error) {
			res := fix(b.New())
			switch s {
			case "true":
				res.OnlyUnknown = true
			case "false":
				res.OnlyUnknown = false
			default:
				return nil, errors.New("OnlyUnknown must be true or false")
			}
			return res, nil
		})
	return res
}

func (b *BootEnv) Backend() store.Store {
	return b.rt.backend(b)
}

func (b *BootEnv) pathFor(f string) string {
	res := b.OS.Name
	if strings.HasSuffix(b.Name, "-install") {
		res = path.Join(res, "install")
	}
	return path.Clean(path.Join("/", res, f))
}

type rt struct {
	io.ReadCloser
	sz int64
}

func (r *rt) Size() int64 {
	return r.sz
}

func (b *BootEnv) fillInstallRepo() {
	if b.Kernel == "" {
		return
	}
	b.rt.Tracef("fillInstallRepo: Looking for global profile")
	o := b.rt.find("profiles", b.rt.dt.GlobalProfileName)
	if o == nil {
		return
	}
	p := AsProfile(o)
	repos := []*Repo{}
	r, ok := b.rt.GetParam(p, "package-repositories", true)
	if !ok || utils.Remarshal(r, &repos) != nil {
		b.rt.Infof("BootEnv %s: No package repositories to use", b.Name)
		return
	}
	for _, r := range repos {
		b.rt.Debugf("BootEnv %s: Considering repo %s", b.Name, r.Tag)
		if !r.InstallSource || len(r.OS) != 1 || r.OS[0] != b.OS.Name {
			continue
		}
		b.rt.Infof("BootEnv %s: Using repo %s as an install source", b.Name, r.Tag)
		b.kernelVerified = true
		b.installRepo = r
		pf := b.pathFor("")
		fileRoot := b.rt.dt.FileRoot
		l := b.rt.Logger
		b.pathLookaside = func(p string) (io.Reader, error) {
			// Always use local copy if available
			if _, err := os.Stat(path.Join(fileRoot, p)); err == nil || b.installRepo == nil {
				return nil, nil
			}
			tgtUri := strings.TrimSuffix(b.installRepo.URL, "/") + strings.TrimPrefix(p, pf)
			l.Debugf("Proxying %s to %s", p, tgtUri)
			resp, err := http.Get(tgtUri)
			if err != nil {
				return nil, err
			}
			if resp.ContentLength < 0 {
				return resp.Body, nil
			}
			return &rt{resp.Body, resp.ContentLength}, nil
		}
		return
	}
}

func (b *BootEnv) AddDynamicTree() {
	if b.pathLookaside != nil {
		b.rt.dt.FS.AddDynamicTree(b.pathFor(""), b.pathLookaside)
	}
}

func (b *BootEnv) localPathFor(f string) string {
	return path.Join(b.rt.dt.FileRoot, b.pathFor(f))
}

func (b *BootEnv) genRoot(commonRoot *template.Template, e models.ErrorAdder) *template.Template {
	res := models.MergeTemplates(commonRoot, b.Templates, e)
	for i, tmpl := range b.Templates {
		if tmpl.Path == "" {
			e.Errorf("Template[%d] needs a Path", i)
		}
	}
	if b.BootParams != "" {
		tmpl, err := template.New("machine").Parse(b.BootParams)
		if err != nil {
			e.Errorf("Error compiling boot parameter template: %v", err)
		} else {
			b.bootParamsTmpl = tmpl.Option("missingkey=error")
		}
	}
	if b.HasError() != nil {
		return nil
	}
	return res
}

func explodeISO(p *DataTracker, envName, osName, fileRoot, isoFile, dest, shaSum string) {
	explodeMux.Lock()
	defer explodeMux.Unlock()
	res := &models.Error{
		Model: "bootenvs",
		Key:   envName,
	}

	// Only check the has if we have one.
	if shaSum != "" {
		f, err := os.Open(isoFile)
		if err != nil {
			res.Errorf("Explode ISO: failed to open iso file %s: %v", p.reportPath(isoFile), err)
		} else {
			defer f.Close()
			hasher := sha256.New()
			if _, err := io.Copy(hasher, f); err != nil {
				res.Errorf("Explode ISO: failed to read iso file %s: %v", p.reportPath(isoFile), err)
			} else {
				hash := hex.EncodeToString(hasher.Sum(nil))
				if hash != shaSum {
					res.Errorf("Explode ISO: SHA256 bad. actual: %v expected: %v", hash, shaSum)
				}
			}
		}
	}
	if !res.ContainsError() {
		// Call extract script
		// /explode_iso.sh b.OS.Name fileRoot isoPath path.Dir(canaryPath)
		cmdName := path.Join(fileRoot, "explode_iso.sh")
		cmdArgs := []string{osName, fileRoot, isoFile, dest, shaSum}
		out, err := exec.Command(cmdName, cmdArgs...).CombinedOutput()
		if err != nil {
			res.Errorf("Explode ISO: explode_iso.sh failed for %s: %s", envName, err)
			res.Errorf("Command output:\n%s", string(out))
		}
	}
	ref := &BootEnv{}
	rt := p.Request(p.Logger, ref.Locks("update")...)
	rt.Do(func(d Stores) {
		b := d("bootenvs").Find(envName)
		if b == nil {
			// Bootenv vanished
			return
		}
		ref = AsBootEnv(b)
		if ref.Available {
			// Bootenv must have vanished
			return
		}
		if res.ContainsError() {
			ref.AddError(res)
		} else {
			ref.Available = true
			rt.Save(ref)
		}
	})
}

func (b *BootEnv) explodeIso() {
	// Only work on things that are requested.
	if b.OS.IsoFile == "" {
		b.rt.Infof("Explode ISO: Skipping %s becausing no iso image specified\n", b.Name)
		return
	}
	b.kernelVerified = false
	// Have we already exploded this?  If file exists, then good!
	canaryPath := b.localPathFor("." + strings.Replace(b.OS.Name, "/", "_", -1) + ".rebar_canary")
	buf, err := ioutil.ReadFile(canaryPath)
	if err == nil && string(bytes.TrimSpace(buf)) == b.OS.IsoSha256 {
		b.rt.Infof("Explode ISO: canary file %s, in place and has proper SHA256\n", b.rt.dt.reportPath(canaryPath))
		return
	}
	isoPath := filepath.Join(b.rt.dt.FileRoot, "isos", b.OS.IsoFile)
	if _, err := os.Stat(isoPath); os.IsNotExist(err) {
		if b.installRepo != nil {
			b.rt.Infof("BootEnv: Explode ISO: ISO does not exist, falling back to install repo at %s", b.installRepo.URL)
			b.kernelVerified = true
			return
		}
		b.Errorf("Explode ISO: iso does not exist: %s\n", b.rt.dt.reportPath(isoPath))
		if b.OS.IsoUrl != "" {
			b.Errorf("You can download the required ISO from %s", b.OS.IsoUrl)
		}
		return
	}
	b.Errorf("Exploding ISO: %s", b.rt.dt.reportPath(isoPath))
	go explodeISO(b.rt.dt, b.Name, b.OS.Name, b.rt.dt.FileRoot, isoPath, b.localPathFor(""), b.OS.IsoSha256)
}

func (b *BootEnv) Validate() {
	b.renderers = renderers{}
	b.BootEnv.Validate()
	// First, the stuff that must be correct in order for
	b.AddError(index.CheckUnique(b, b.rt.stores("bootenvs").Items()))
	// If our basic templates do not parse, it is game over for us
	b.rt.dt.tmplMux.Lock()
	b.tmplMux.Lock()
	root := b.genRoot(b.rt.dt.rootTemplate, b)
	b.rt.dt.tmplMux.Unlock()
	if root != nil {
		b.rootTemplate = root
	}
	b.tmplMux.Unlock()
	if !b.SetValid() {
		// If we have not been validated at this point, return.
		return
	}
	b.fillInstallRepo()
	// OK, we are sane, if not useable.  Check to see if we are useable
	seenPxeLinux := false
	seenIPXE := false
	for _, template := range b.Templates {
		if template.Name == "pxelinux" || template.Name == "pxelinux-mac" {
			seenPxeLinux = true
		}
		if template.Name == "ipxe" || template.Name == "ipxe-mac" {
			seenIPXE = true
		}
	}
	if !(seenPxeLinux || seenIPXE) && b.Kernel != "" {
		b.Errorf("bootenv: Missing elilo or pxelinux template")
	}
	// Make sure the ISO for this bootenv has been exploded locally so that
	// the boot env can use its contents.
	b.explodeIso()
	if !b.kernelVerified {
		// If we have a non-empty Kernel, make sure it points at something kernel-ish.
		if b.Kernel != "" {
			kPath := b.localPathFor(b.Kernel)
			kernelStat, err := os.Stat(kPath)
			if err != nil {
				b.Errorf("bootenv: %s: missing kernel %s (%s)",
					b.Name,
					b.Kernel,
					b.rt.dt.reportPath(kPath))
			} else if !kernelStat.Mode().IsRegular() {
				b.Errorf("bootenv: %s: invalid kernel %s (%s)",
					b.Name,
					b.Kernel,
					b.rt.dt.reportPath(kPath))
			}
		}
		// Ditto for all the initrds.
		if len(b.Initrds) > 0 {
			for _, initrd := range b.Initrds {
				iPath := b.localPathFor(initrd)
				initrdStat, err := os.Stat(iPath)
				if err != nil {
					b.Errorf("bootenv: %s: missing initrd %s (%s)",
						b.Name,
						initrd,
						b.rt.dt.reportPath(iPath))
					continue
				}
				if !initrdStat.Mode().IsRegular() {
					b.Errorf("bootenv: %s: invalid initrd %s (%s)",
						b.Name,
						initrd,
						b.rt.dt.reportPath(iPath))
				}
			}
		}
	}
	if b.OnlyUnknown {
		b.renderers = append(b.renderers, b.Render(b.rt, nil, b)...)
	} else {
		machines := b.rt.stores("machines")
		if machines != nil {
			for _, i := range machines.Items() {
				machine := AsMachine(i)
				if machine.BootEnv != b.Name {
					continue
				}
				b.renderers = append(b.renderers, b.Render(b.rt, machine, b)...)
			}
		}
	}
	b.SetAvailable()
	stages := b.rt.stores("stages")
	if stages != nil {
		for _, i := range stages.Items() {
			stage := AsStage(i)
			if stage.BootEnv != b.Name {
				continue
			}
			func() {
				stage.rt = b.rt
				defer func() { stage.rt = nil }()
				stage.ClearValidation()
				stage.Validate()
			}()
		}
	}
}

func (b *BootEnv) OnLoad() error {
	defer func() { b.rt = nil }()
	b.Fill()
	return b.BeforeSave()
}

func (b *BootEnv) New() store.KeySaver {
	res := &BootEnv{BootEnv: &models.BootEnv{}}
	if b.BootEnv != nil && b.ChangeForced() {
		res.ForceChange()
	}
	res.rt = b.rt
	return res
}

func (b *BootEnv) BeforeSave() error {
	b.Validate()
	if !b.Validated {
		return b.MakeError(422, ValidationError, b)
	}
	return nil
}

func (b *BootEnv) BeforeDelete() error {
	e := &models.Error{Code: 409, Type: StillInUseError, Model: b.Prefix(), Key: b.Key()}
	machines := b.rt.stores("machines")
	stages := b.rt.stores("stages")
	prefToFind := ""
	if b.OnlyUnknown {
		prefToFind = "unknownBootEnv"
	} else {
		prefToFind = "defaultBootEnv"
	}
	if b.rt.dt.pref(prefToFind) == b.Name {
		e.Errorf("BootEnv %s is the active %s, cannot remove it", b.Name, prefToFind)
	}
	if !b.OnlyUnknown {
		for _, i := range machines.Items() {
			machine := AsMachine(i)
			if machine.BootEnv != b.Name {
				continue
			}
			e.Errorf("Bootenv %s in use by Machine %s", b.Name, machine.Name)
		}
		for _, i := range stages.Items() {
			stage := AsStage(i)
			if stage.BootEnv != b.Name {
				continue
			}
			e.Errorf("Bootenv %s in use by Stage %s", b.Name, stage.Name)
		}
	}
	return e.HasError()
}

func (b *BootEnv) AfterDelete() {
	if b.OnlyUnknown {
		err := &models.Error{Object: b}
		rts := b.Render(b.rt, nil, err)
		if err.ContainsError() {
			b.Errors = err.Messages
		} else {
			rts.deregister(b.rt.dt.FS)
		}
		idx, idxerr := index.All(
			index.Sort(b.Indexes()["OsName"]),
			index.Eq(b.OS.Name))(&(b.rt.stores("bootenvs").Index))
		if idxerr == nil && idx.Count() == 0 {
			b.rt.dt.FS.DelDynamicTree(b.pathFor(""))
		}
	}
}

func AsBootEnv(o models.Model) *BootEnv {
	return o.(*BootEnv)
}

func AsBootEnvs(o []models.Model) []*BootEnv {
	res := make([]*BootEnv, len(o))
	for i := range o {
		res[i] = AsBootEnv(o[i])
	}
	return res
}

func (b *BootEnv) renderInfo() ([]models.TemplateInfo, []string) {
	return b.Templates, b.RequiredParams
}

func (b *BootEnv) templates() *template.Template {
	return b.rootTemplate
}

func (b *BootEnv) Render(rt *RequestTracker, m *Machine, e models.ErrorAdder) renderers {
	if len(b.RequiredParams) > 0 && m == nil {
		e.Errorf("Machine is nil or does not have params")
		return nil
	}
	r := newRenderData(rt, m, b)
	if m == nil {
		return r.makeRenderers(e)
	}
	res := renderers([]renderer{})
	toRender := r.validateRequiredParams(e)
	for i := range toRender {
		switch toRender[i].Name {
		case "ipxe-mac", "pxelinux-mac":
			for _, mac := range m.HardwareAddrs {
				r.Machine.currMac = mac
				res = r.addRenderer(e, &toRender[i], res)
			}
		default:
			res = r.addRenderer(e, &toRender[i], res)
		}
	}
	return res
}

func (b *BootEnv) AfterSave() {
	if b.Available && b.renderers != nil {
		b.renderers.register(b.rt.dt.FS)
	}
	b.AddDynamicTree()
	b.renderers = nil
}

var bootEnvLockMap = map[string][]string{
	"get":     []string{"bootenvs"},
	"create":  []string{"stages", "bootenvs", "machines", "tasks", "templates", "profiles", "params", "workflows"},
	"update":  []string{"stages", "bootenvs", "machines", "tasks", "templates", "profiles", "params", "workflows"},
	"patch":   []string{"stages", "bootenvs", "machines", "tasks", "templates", "profiles", "params", "workflows"},
	"delete":  []string{"stages", "bootenvs", "machines", "tasks", "templates", "profiles", "params", "workflows"},
	"actions": []string{"stages", "bootenvs", "machines", "tasks", "templates", "profiles", "params"},
}

func (b *BootEnv) Locks(action string) []string {
	return bootEnvLockMap[action]
}
