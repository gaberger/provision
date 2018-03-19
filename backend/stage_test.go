package backend

import (
	"testing"

	"github.com/digitalrebar/provision/models"
	"github.com/pborman/uuid"
)

func TestStageCrud(t *testing.T) {
	dt := mkDT(nil)
	rt := dt.Request(dt.Logger, "stages", "bootenvs", "templates", "tasks", "machines", "profiles", "workflows")
	tmpl := &models.Template{ID: "ok", Contents: "{{ .Env.Name }}"}
	var ok bool
	var err error
	rt.Do(func(d Stores) {
		ok, err = rt.Create(tmpl)
	})
	if !ok {
		t.Errorf("Failed to create test OK template: %#v: %#v", tmpl, err)
		return
	}

	tests := []crudTest{
		{"Create Stage with nonexistent Name", rt.Create, &models.Stage{}, false},
		{"Create Stage with no BootEnv", rt.Create, &models.Stage{Name: "nobootenv"}, true},
		{"Create Stage with bad name /", rt.Create, &models.Stage{Name: "no/bootenv"}, false},
		{"Create Stage with bad name \\", rt.Create, &models.Stage{Name: "no\\bootenv"}, false},
		{"Create Stage with nonexistent BootEnv", rt.Create, &models.Stage{Name: "missingbootenv", BootEnv: "missingbootenv"}, true},
		{"Create Stage with missing Task", rt.Create, &models.Stage{Name: "missingtask", BootEnv: "local", Tasks: []string{"jj"}}, true},
		{"Create Stage with missing profile", rt.Create, &models.Stage{Name: "missingprofile", BootEnv: "local", Profiles: []string{"jj"}}, true},
		{"Create Stage with invalid models.TemplateInfo (missing Name)", rt.Create, &models.Stage{Name: "test 3", BootEnv: "local", Templates: []models.TemplateInfo{{Path: "{{ .Env.Name }}", ID: "ok"}}}, false},
		{"Create Stage with invalid models.TemplateInfo (missing ID)", rt.Create, &models.Stage{Name: "test 3", BootEnv: "local", Templates: []models.TemplateInfo{{Name: "test 3", Path: "{{ .Env.Name }}"}}}, false},
		{"Create Stage with invalid models.TemplateInfo (missing Path)", rt.Create, &models.Stage{Name: "test 3", BootEnv: "local", Templates: []models.TemplateInfo{{Name: "test 3", ID: "ok"}}}, false},
		{"Create Stage with invalid models.TemplateInfo (invalid ID)", rt.Create, &models.Stage{Name: "invalidTemplateID", BootEnv: "local", Templates: []models.TemplateInfo{{Name: "test 3", Path: "{{ .Env.Name }}", ID: "okp"}}}, true},
		{"Create Stage with invalid models.TemplateInfo (invalid Path)", rt.Create, &models.Stage{Name: "test 3", BootEnv: "local", Templates: []models.TemplateInfo{{Name: "test 3", Path: "{{ .Env.Name }", ID: "ok"}}}, false},
		{"Create Stage with valid models.TemplateInfo (not available}", rt.Create, &models.Stage{Name: "test 1", BootEnv: "local", Templates: []models.TemplateInfo{{Name: "unavailable", Path: "{{ .Env.Name }}", ID: "ok"}}}, true},
		{"Create Stage with valid models.TemplateInfo (available)", rt.Create, &models.Stage{Name: "available", BootEnv: "local", Templates: []models.TemplateInfo{{Name: "ipxe", Path: "{{ .Env.Name }}", ID: "ok"}}}, true},
	}

	for _, test := range tests {
		test.Test(t, rt)
	}

	// List test.
	rt.Do(func(d Stores) {
		bes := d("stages").Items()
		if bes != nil {
			if len(bes) != 9 {
				t.Errorf("List function should have returned: 9, but got %d\n", len(bes))
			}
		} else {
			t.Errorf("List function returned nil!!")
		}

		// We need a Machine that refers to one of our Stage to
		// test proper delete restrictions
		machine := &models.Machine{Name: "test 1", Stage: "available", Uuid: uuid.NewRandom()}
		ok, err = rt.Create(machine)
	})
	if !ok {
		t.Errorf("Failed to create test machine: %v", err)
		return
	}
	rmTests := []crudTest{
		{"Remove Stage that is not in use", rt.Remove, &models.Stage{Name: "test 1"}, true},
		{"Remove nonexistent Stage", rt.Remove, &models.Stage{Name: "test 1"}, false},
		{"Remove Stage that is in use", rt.Remove, &models.Stage{Name: "available"}, false},
	}
	for _, test := range rmTests {
		test.Test(t, rt)
	}
}
