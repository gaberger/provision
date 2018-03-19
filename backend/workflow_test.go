package backend

import (
	"testing"

	"github.com/digitalrebar/provision/models"
)

func TestWorkflowCrud(t *testing.T) {
	dt := mkDT(nil)
	rt := dt.Request(dt.Logger, "stages", "bootenvs", "templates", "tasks", "machines", "profiles", "workflows")
	tests := []crudTest{
		{"Create Workflow with nonexistent Name", rt.Create, &models.Workflow{}, false},
		{"Create Workflow with no Stages", rt.Create, &models.Workflow{Name: "nostage"}, true},
		{"Create Workflow with bad name /", rt.Create, &models.Workflow{Name: "no/stage"}, false},
		{"Create Workflow with bad name \\", rt.Create, &models.Workflow{Name: "no\\stage"}, false},
		{"Create Workflow with nonexistent Stage", rt.Create, &models.Workflow{Name: "missingstage", Stages: []string{"missingstage"}}, true},
		{"Create Stage with no BootEnv", rt.Create, &models.Stage{Name: "nobootenv"}, true},
		{"Create Workflow with stage that exists", rt.Create, &models.Workflow{Name: "havestage", Stages: []string{"nobootenv"}}, true},
	}
	for _, test := range tests {
		test.Test(t, rt)
	}
}
