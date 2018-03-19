package backend

import (
	"testing"

	"github.com/digitalrebar/provision/models"
	"github.com/pborman/uuid"
)

func TestBootEnvCrud(t *testing.T) {
	dt := mkDT(nil)
	rt := dt.Request(dt.Logger, "stages", "bootenvs", "templates", "tasks", "machines", "profiles", "workflows")
	tmpl := &models.Template{ID: "ok", Contents: "{{ .Env.Name }}"}
	var ok bool
	var err error
	rt.Do(func(d Stores) { ok, err = rt.Create(tmpl) })
	if !ok {
		t.Errorf("Failed to create test OK template: %#v: %#v", tmpl, err)
		return
	}

	crudTest{"Create Bootenv with nonexistent Name", rt.Create, &models.BootEnv{}, false}.Test(t, rt)
	crudTest{"Create Bootenv with no templates", rt.Create, &models.BootEnv{Name: "test 1"}, true}.Test(t, rt)
	crudTest{"Create Bootenv with invalid Name /", rt.Create, &models.BootEnv{Name: "test/greg"}, false}.Test(t, rt)
	crudTest{"Create Bootenv with invalid Name \\", rt.Create, &models.BootEnv{Name: "test\\greg"}, false}.Test(t, rt)
	crudTest{"Create Bootenv with invalid BootParams tmpl", rt.Create, &models.BootEnv{Name: "test 2", BootParams: "{{ }"}, false}.Test(t, rt)
	crudTest{"Create Bootenv with valid BootParams tmpl", rt.Create, &models.BootEnv{Name: "test 2", BootParams: "{{ .Env.Name }}"}, true}.Test(t, rt)
	crudTest{"Create Bootenv with invalid models.TemplateInfo (missing Name)", rt.Create, &models.BootEnv{Name: "test 3", Templates: []models.TemplateInfo{{Path: "{{ .Env.Name }}", ID: "ok"}}}, false}.Test(t, rt)
	crudTest{"Create Bootenv with invalid models.TemplateInfo (missing ID)", rt.Create, &models.BootEnv{Name: "test 3", Templates: []models.TemplateInfo{{Name: "test 3", Path: "{{ .Env.Name }}"}}}, false}.Test(t, rt)
	crudTest{"Create Bootenv with invalid models.TemplateInfo (missing Path)", rt.Create, &models.BootEnv{Name: "test 3", Templates: []models.TemplateInfo{{Name: "test 3", ID: "ok"}}}, false}.Test(t, rt)
	crudTest{"Create Bootenv with invalid models.TemplateInfo (invalid ID)", rt.Create, &models.BootEnv{Name: "test 3", Templates: []models.TemplateInfo{{Name: "test 3", Path: "{{ .Env.Name }}", ID: "okp"}}}, false}.Test(t, rt)
	crudTest{"Create Bootenv with invalid models.TemplateInfo (invalid Path)", rt.Create, &models.BootEnv{Name: "test 3", Templates: []models.TemplateInfo{{Name: "test 3", Path: "{{ .Env.Name }", ID: "ok"}}}, false}.Test(t, rt)
	crudTest{"Create Bootenv with valid models.TemplateInfo (not available}", rt.Create, &models.BootEnv{Name: "test 3", Templates: []models.TemplateInfo{{Name: "unavailable", Path: "{{ .Env.Name }}", ID: "ok"}}}, true}.Test(t, rt)
	crudTest{"Create Bootenv with valid models.TemplateInfo (available)", rt.Create, &models.BootEnv{Name: "available", Templates: []models.TemplateInfo{{Name: "ipxe", Path: "{{ .Env.Name }}", ID: "ok"}}}, true}.Test(t, rt)

	// List test.
	rt.Do(func(d Stores) {
		bes := d("bootenvs").Items()
		if bes != nil {
			if len(bes) != 6 {
				t.Errorf("List function should have returned: 6, but got %d\n", len(bes))
			}
		} else {
			t.Errorf("List function returned nil!!")
		}
	})
	// We need a Machine that refers to one of our BootEnvs to
	// test proper delete restrictions
	rt.Do(func(d Stores) {
		machine := &models.Machine{Name: "test 1", BootEnv: "available", Uuid: uuid.NewRandom()}
		ok, err = rt.Create(machine)
	})
	if !ok {
		t.Errorf("Failed to create test machine: %v", err)
		return
	}
	crudTest{"Remove BootEnv that is not in use", rt.Remove, &models.BootEnv{Name: "test 1"}, true}.Test(t, rt)
	crudTest{"Remove nonexistent BootEnv", rt.Remove, &models.BootEnv{Name: "test 1"}, false}.Test(t, rt)
	crudTest{"Remove BootEnv that is in use", rt.Remove, &models.BootEnv{Name: "available"}, false}.Test(t, rt)
}
