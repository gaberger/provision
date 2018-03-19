package backend

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/digitalrebar/provision/backend/index"
	"github.com/digitalrebar/provision/models"
	"github.com/digitalrebar/store"
	"github.com/pborman/uuid"
)

// Job represents a task that is running (or has run) on a machine.
// The job create workflow I envision works like this:
//
// * POST to api/v3/jobs with a body containing {"Machine":
//   "a-machine-uuid"} If there is no current job, or the current job
//   is "failed", a new job is created for the Task indexed by
//   CurrentTask. If the current job is "finished", the machine
//   CurrentTask is incremented.  If that causes CurrentTask to go
//   past the end of the Tasks list for the machine, no job is created
//   and the API returns a 204. If the current job is in the imcomplete
//   state, that job is returned with a 202.  Otherwise a new job is
//   created and is returned with a 201. If there is a current job that is neither
//   "incomplete", "failed", nor "finished", the POST fails.  The new job will be
//   created with its Previous value set to the machine's CurrentJob,
//   and the machine's CurrentJob is updated with the UUID of the new
//   job.
//
// * When a new Job is created, it makes a RenderData for the
//   templates contained in the Task the job was created against.  The
//   client will be able to retrieve the rendered templates via GET
//   from api/v3/jobs/:job-id/templates.
//
// * The client will place or execute the templates based on whether
//   there is a Path associated with the expanded Template in the
//   order that the jobs/:id/templates API endpoint returns them in.
//   As it does so, it will log its progress via POST to jobs/:id/log.
//
// * If any job operation fails, the client will update the job status to "failed".
//
// * If all job operations succeed, the client will update the job status to "finished"
//
// * On provisioner startup, all machine CurrentJobs are set to "failed" if they are not "finished"
//
// swagger:model
type Job struct {
	*models.Job
	validate
	oldState string
}

func (obj *Job) SetReadOnly(b bool) {
	obj.ReadOnly = b
}

func (j *Job) LogPath(rt *RequestTracker) string {
	if j.rt == nil {
		j.setRT(rt)
		defer j.clearRT()
	}
	return filepath.Join(j.rt.dt.LogRoot, j.Uuid.String())
}

func (obj *Job) SaveClean() store.KeySaver {
	mod := *obj.Job
	mod.ClearValidation()
	return toBackend(&mod, obj.rt)
}
func AsJob(o models.Model) *Job {
	return o.(*Job)
}

func AsJobs(o []models.Model) []*Job {
	res := make([]*Job, len(o))
	for i := range o {
		res[i] = AsJob(o[i])
	}
	return res
}

func (j *Job) New() store.KeySaver {
	res := &Job{Job: &models.Job{}}
	if j.Job != nil && j.ChangeForced() {
		res.ForceChange()
	}
	res.rt = j.rt
	return res
}

func (j *Job) UUID() string {
	return j.Uuid.String()
}

func (j *Job) Indexes() map[string]index.Maker {
	fix := AsJob
	res := index.MakeBaseIndexes(j)
	res["Uuid"] = index.Make(
		true,
		"UUID string",
		func(i, j models.Model) bool { return fix(i).Uuid.String() < fix(j).Uuid.String() },
		func(ref models.Model) (gte, gt index.Test) {
			refUuid := fix(ref).Uuid.String()
			return func(s models.Model) bool {
					return fix(s).Uuid.String() >= refUuid
				},
				func(s models.Model) bool {
					return fix(s).Uuid.String() > refUuid
				}
		},
		func(s string) (models.Model, error) {
			id := uuid.Parse(s)
			if id == nil {
				return nil, fmt.Errorf("Invalid UUID: %s", s)
			}
			job := fix(j.New())
			job.Uuid = id
			return job, nil
		})
	res["Previous"] = index.Make(
		false,
		"UUID string",
		func(i, j models.Model) bool { return fix(i).Previous.String() < fix(j).Previous.String() },
		func(ref models.Model) (gte, gt index.Test) {
			refUuid := fix(ref).Previous.String()
			return func(s models.Model) bool {
					return fix(s).Previous.String() >= refUuid
				},
				func(s models.Model) bool {
					return fix(s).Previous.String() > refUuid
				}
		},
		func(s string) (models.Model, error) {
			id := uuid.Parse(s)
			if id == nil {
				return nil, fmt.Errorf("Invalid UUID: %s", s)
			}
			job := fix(j.New())
			job.Previous = id
			return job, nil
		})
	res["Stage"] = index.Make(
		false,
		"string",
		func(i, j models.Model) bool { return fix(i).Stage < fix(j).Stage },
		func(ref models.Model) (gte, gt index.Test) {
			refStage := fix(ref).Stage
			return func(s models.Model) bool {
					return fix(s).Stage >= refStage
				},
				func(s models.Model) bool {
					return fix(s).Stage > refStage
				}
		},
		func(s string) (models.Model, error) {
			job := fix(j.New())
			job.Stage = s
			return job, nil
		})
	res["Task"] = index.Make(
		false,
		"string",
		func(i, j models.Model) bool { return fix(i).Task < fix(j).Task },
		func(ref models.Model) (gte, gt index.Test) {
			refTask := fix(ref).Task
			return func(s models.Model) bool {
					return fix(s).Task >= refTask
				},
				func(s models.Model) bool {
					return fix(s).Task > refTask
				}
		},
		func(s string) (models.Model, error) {
			job := fix(j.New())
			job.Task = s
			return job, nil
		})
	res["State"] = index.Make(
		false,
		"string",
		func(i, j models.Model) bool { return fix(i).State < fix(j).State },
		func(ref models.Model) (gte, gt index.Test) {
			refState := fix(ref).State
			return func(s models.Model) bool {
					return fix(s).State >= refState
				},
				func(s models.Model) bool {
					return fix(s).State > refState
				}
		},
		func(s string) (models.Model, error) {
			job := fix(j.New())
			job.State = s
			return job, nil
		})
	res["Machine"] = index.Make(
		true,
		"UUID string",
		func(i, j models.Model) bool { return fix(i).Machine.String() < fix(j).Machine.String() },
		func(ref models.Model) (gte, gt index.Test) {
			refMachine := fix(ref).Machine.String()
			return func(s models.Model) bool {
					return fix(s).Machine.String() >= refMachine
				},
				func(s models.Model) bool {
					return fix(s).Machine.String() > refMachine
				}
		},
		func(s string) (models.Model, error) {
			id := uuid.Parse(s)
			if id == nil {
				return nil, fmt.Errorf("Invalid UUID: %s", s)
			}
			job := fix(j.New())
			job.Machine = id
			return job, nil
		})
	res["Archived"] = index.Make(
		false,
		"boolean",
		func(i, j models.Model) bool {
			return (!fix(i).Archived) && fix(j).Archived
		},
		func(ref models.Model) (gte, gt index.Test) {
			avail := fix(ref).Archived
			return func(s models.Model) bool {
					v := fix(s).Archived
					return v || (v == avail)
				},
				func(s models.Model) bool {
					return fix(s).Archived && !avail
				}
		},
		func(s string) (models.Model, error) {
			res := fix(j.New())
			switch s {
			case "true":
				res.Archived = true
			case "false":
				res.Archived = false
			default:
				return nil, errors.New("Archived must be true or false")
			}
			return res, nil
		})
	res["StartTime"] = index.Make(
		false,
		"dateTime",
		func(i, j models.Model) bool {
			return fix(i).StartTime.Before(fix(j).StartTime)
		},
		func(ref models.Model) (gte, gt index.Test) {
			refTime := fix(ref).StartTime
			return func(s models.Model) bool {
					cmpTime := fix(s).StartTime
					return refTime.Equal(cmpTime) || cmpTime.After(refTime)
				},
				func(s models.Model) bool {
					return fix(s).StartTime.After(refTime)
				}
		},
		func(s string) (models.Model, error) {
			parsedTime, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return nil, err
			}
			job := fix(j.New())
			job.StartTime = parsedTime
			return job, nil
		})
	res["EndTime"] = index.Make(
		false,
		"dateTime",
		func(i, j models.Model) bool {
			return fix(i).EndTime.Before(fix(j).EndTime)
		},
		func(ref models.Model) (gte, gt index.Test) {
			refTime := fix(ref).EndTime
			return func(s models.Model) bool {
					cmpTime := fix(s).EndTime
					return refTime.Equal(cmpTime) || cmpTime.After(refTime)
				},
				func(s models.Model) bool {
					return fix(s).EndTime.After(refTime)
				}
		},
		func(s string) (models.Model, error) {
			parsedTime, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return nil, err
			}
			job := fix(j.New())
			job.EndTime = parsedTime
			return job, nil
		})
	res["Current"] = index.Make(
		false,
		"boolean",
		func(i, j models.Model) bool {
			j1, j2 := fix(i), fix(j)
			return (!j1.Current) && j2.Current
		},
		func(ref models.Model) (gte, gt index.Test) {
			current := fix(ref).Current
			return func(s models.Model) bool {
					v := fix(s).Current
					return v || (v == current)
				},
				func(s models.Model) bool {
					return fix(s).Current && !current
				}
		},
		func(s string) (models.Model, error) {
			res := fix(j.New())
			switch s {
			case "true":
				res.Current = true
			case "false":
				res.Current = false
			default:
				return nil, errors.New("Current must be true or false")
			}
			return res, nil
		})
	return res
}

func (j *Job) OnCreate() error {
	j.Current = true
	if _, err := os.Stat(j.LogPath(j.rt)); err != nil {
		if f, err := os.Create(j.LogPath(j.rt)); err != nil {
			j.AddError(err)
		} else {
			f.Close()
		}
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "Log for Job: %s\n", j.Uuid.String())
		j.AddError(j.Log(j.rt, buf))
	}
	return j.HasError()
}

func (j *Job) OnLoad() error {
	defer func() { j.rt = nil }()
	j.oldState = j.State
	j.Fill()
	j.SetValid()
	j.Validate()
	return nil
}

func (j *Job) OnChange(oldThing store.KeySaver) error {
	ot := AsJob(oldThing)
	j.Current = ot.Current
	j.oldState = ot.State
	return nil
}

func (j *Job) Validate() {
	if j.Uuid == nil {
		j.Errorf("Job %#v was not assigned a uuid!", j)
	}
	if j.Previous == nil {
		j.Errorf("Job %s does not have a Previous job", j.UUID())
	}
	if j.State == "finished" || j.State == "failed" {
		if j.oldState != j.State {
			j.EndTime = time.Now()
		}
		j.SetValid()
	}
	if !j.Current {
		j.SetValid()
		j.SetAvailable()
		return
	}
	j.Job.Validate()
	objs := j.rt.d
	tasks := objs("tasks")
	stages := objs("stages")
	machines := objs("machines")

	var m *Machine
	if om := machines.Find(j.Machine.String()); om == nil {
		j.Errorf("Machine %s does not exist", j.Machine.String())
	} else {
		m = AsMachine(om)
		if j.oldState != j.State {
			if j.State == "failed" {
				m.Runnable = false
				_, e2 := j.rt.Save(m)
				j.AddError(e2)
			}
		}
	}

	if tasks.Find(j.Task) == nil {
		j.Errorf("Task %s does not exist", j.Task)
	}

	var env *Stage
	if nbFound := stages.Find(j.Stage); nbFound == nil {
		j.Errorf("Stage %s does not exist", j.Stage)
	} else {
		env = AsStage(nbFound)
	}
	if env != nil && !env.Available {
		j.Errorf("Stage %s is not available", j.Stage)
	}

	j.SetValid()
	j.SetAvailable()
	if !j.Available {
		j.State = "failed"
	} else if j.oldState != j.State && j.State == "running" {
		j.StartTime = time.Now()
	}
}

func (j *Job) BeforeSave() error {
	j.Validate()
	if !j.Validated {
		return j.MakeError(422, ValidationError, j)
	}
	return nil
}

func (j *Job) AfterSave() {
	if !j.Current {
		return
	}
	oldJ := j.rt.d("jobs").Find(j.Previous.String())
	if oldJ == nil {
		return
	}
	oj := oldJ.(*Job)
	oj.Current = false
	j.rt.Save(oj)
}

func (j *Job) BeforeDelete() error {
	e := &models.Error{Code: 422, Type: ValidationError, Model: j.Prefix(), Key: j.Key()}
	if j.State == "finished" || j.State == "failed" {
		return nil
	}
	machines := j.rt.stores("machines")
	if machines == nil || machines.Find(j.Machine.String()) == nil {
		// Machine has vanished.  We can delete this job
		return nil
	}
	jobs := j.rt.stores("jobs")
	// Test to see if there is a job that follows this one.  If there is, we can delete this one.
	if jobs != nil {
		prevMaker := j.Indexes()["Previous"]
		prevIdx, err := index.Sort(prevMaker)(&jobs.Index)
		if err != nil || prevIdx.Find(j.Uuid.String()) != nil {
			return nil
		}
	}
	e.Errorf("Jobs %s is not in a deletable state: %v", j.UUID(), j.State)

	return e.HasError()
}

func (j *Job) RenderActions(rt *RequestTracker) ([]*models.JobAction, error) {
	var rds renderers
	var addr net.IP
	var err *models.Error
	rt.Do(func(d Stores) {
		machines := d("machines")
		tasks := d("tasks")

		// This should not happen, but we treat task in the job as soft.
		var to models.Model
		if to = tasks.Find(j.Task); to == nil {
			err = &models.Error{Code: http.StatusUnprocessableEntity, Type: ValidationError,
				Messages: []string{fmt.Sprintf("Task %s does not exist", j.Task)}}
			return
		}
		t := AsTask(to)

		// This should not happen, but we treat machine in the job as soft.
		var mo models.Model
		if mo = machines.Find(j.Machine.String()); mo == nil {
			err = &models.Error{Code: http.StatusUnprocessableEntity, Type: ValidationError,
				Messages: []string{fmt.Sprintf("Machine %s does not exist", j.Machine.String())}}
			return
		}
		m := AsMachine(mo)

		err = &models.Error{Code: http.StatusUnprocessableEntity, Type: ValidationError}
		renderers := t.Render(rt, m, err)
		if err.HasError() != nil {
			return
		}
		rds, addr, err = renderers, m.Address, nil
	})
	if err != nil {
		return nil, err
	}
	err = &models.Error{}
	actions := []*models.JobAction{}
	for _, r := range rds {
		rr, err1 := r.write(addr)
		if err1 != nil {
			err.AddError(err1)
		} else {
			b, err2 := ioutil.ReadAll(rr)
			if err2 != nil {
				err.AddError(err2)
			} else {
				na := &models.JobAction{Name: r.name, Path: r.path, Content: string(b)}
				actions = append(actions, na)
			}
		}
	}

	return actions, err.HasError()
}

func (j *Job) Log(rt *RequestTracker, src io.Reader) error {
	if j.rt == nil {
		j.setRT(rt)
		defer j.clearRT()
	}
	f, err := os.OpenFile(j.LogPath(rt), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Umm err: %v\n", err)
		return err
	}
	_, err = io.Copy(f, src)
	if err != nil {
		fmt.Printf("Umm write err: %v\n", err)
		return err
	}
	return nil
}

var jobLockMap = map[string][]string{
	"get":     []string{"jobs"},
	"create":  []string{"stages", "bootenvs", "jobs", "machines", "tasks", "profiles"},
	"update":  []string{"stages", "bootenvs", "jobs", "machines", "tasks", "profiles"},
	"patch":   []string{"stages", "bootenvs", "jobs", "machines", "tasks", "profiles"},
	"delete":  []string{"machines", "jobs"},
	"actions": []string{"stages", "jobs", "machines", "tasks", "profiles", "bootenvs", "params"},
}

func (j *Job) Locks(action string) []string {
	return jobLockMap[action]
}
