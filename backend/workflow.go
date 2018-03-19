package backend

import (
	"github.com/digitalrebar/provision/backend/index"
	"github.com/digitalrebar/provision/models"
	"github.com/digitalrebar/store"
)

type Workflow struct {
	*models.Workflow
	validate
}

func (w *Workflow) SetReadOnly(b bool) {
	w.ReadOnly = b
}

func (w *Workflow) SaveClean() store.KeySaver {
	mod := *w.Workflow
	mod.ClearValidation()
	return toBackend(&mod, w.rt)
}

func AsWorkflow(o models.Model) *Workflow {
	return o.(*Workflow)
}

func AsWorkflows(o []models.Model) []*Workflow {
	res := make([]*Workflow, len(o))
	for i := range o {
		res[i] = AsWorkflow(o[i])
	}
	return res
}

func (w *Workflow) New() store.KeySaver {
	res := &Workflow{Workflow: &models.Workflow{}}
	if w.Workflow != nil && w.ChangeForced() {
		res.ForceChange()
	}
	res.rt = w.rt
	res.Fill()
	return res
}

func (w *Workflow) Indexes() map[string]index.Maker {
	fix := AsWorkflow
	res := index.MakeBaseIndexes(w)
	res["Name"] = index.Make(
		true,
		"string",
		func(i, j models.Model) bool {
			return fix(i).Name < fix(j).Name
		},
		func(ref models.Model) (gte, gt index.Test) {
			name := fix(ref).Name
			return func(ss models.Model) bool {
					return fix(ss).Name >= name
				},
				func(ss models.Model) bool {
					return fix(ss).Name > name
				}
		},
		func(ss string) (models.Model, error) {
			res := fix(w.New())
			res.Name = ss
			return res, nil
		})
	return res
}

func (w *Workflow) Validate() {
	w.Workflow.Validate()
	w.AddError(index.CheckUnique(w, w.rt.stores("workflows").Items()))
	if !w.SetValid() {
		return
	}
	for _, stageName := range w.Stages {
		if stage := w.rt.find("stages", stageName); stage == nil {
			w.Errorf("Stage %s does not exist", stageName)
		} else if !stage.(*Stage).Available {
			w.Errorf("Stage %s is not available", stageName)
		}
	}
	w.SetAvailable()
}

func (w *Workflow) BeforeSave() error {
	w.Fill()
	w.Validate()
	if !w.Validated {
		return w.MakeError(422, ValidationError, w)
	}
	return nil
}

func (w *Workflow) OnLoad() error {
	defer func() { w.rt = nil }()
	w.Fill()
	return w.BeforeSave()
}

var workflowLockMap = map[string][]string{
	"get":     []string{"workflows"},
	"create":  []string{"stages", "bootenvs", "machines", "tasks", "templates", "profiles", "workflows"},
	"update":  []string{"stages", "bootenvs", "machines", "tasks", "templates", "profiles", "workflows"},
	"patch":   []string{"stages", "bootenvs", "machines", "tasks", "templates", "profiles", "workflows"},
	"delete":  []string{"stages", "bootenvs", "machines", "tasks", "templates", "profiles", "workflows"},
	"actions": []string{"workflows", "stages", "profiles", "params"},
}

func (w *Workflow) Locks(action string) []string {
	return workflowLockMap[action]
}
