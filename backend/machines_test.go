package backend

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/VictorLowther/jsonpatch2"
	"github.com/digitalrebar/provision/models"
	"github.com/pborman/uuid"
)

type patchTest struct {
	desc  string
	pass  bool
	loc   int
	patch string
}

func (p *patchTest) test(t *testing.T, target models.Model) {
	t.Logf("Testing %s", p.desc)
	buf, err := json.Marshal(target)
	if err != nil {
		t.Errorf("Unable to marshal %s: %v", target.Key(), err)
		return
	}
	patch, err := jsonpatch2.NewPatch([]byte(p.patch))
	if err != nil {
		t.Errorf("Patch %s is not valid: %v", p.patch, err)
		return
	}
	_, err, loc := patch.Apply(buf)
	if !p.pass && err != nil {
		if loc != p.loc {
			t.Errorf("Expected patch to fail at loc %d, not %d", p.loc, loc)
		} else {
			t.Logf("Failed at expected loc %d", loc)
		}
		t.Logf("Error: %v", err)
	} else if p.pass && err == nil {
		t.Logf("Patch succeeded")
	} else if err == nil {
		t.Errorf("Patch was expected to fail, but succeeded!")
	} else {
		t.Errorf("Patch failed at %d: %v", loc, err)
	}
}

func TestMachineCrud(t *testing.T) {
	dt := mkDT(nil)
	rt := dt.Request(dt.Logger, "stages", "templates", "machines", "tasks", "bootenvs", "profiles", "jobs", "workflows")
	okUUID := uuid.NewRandom()
	tests := []crudTest{
		{"Create known-good Template", rt.Create, &models.Template{ID: "default"}, true},
		{"Create known-good Bootenv", rt.Create, &models.BootEnv{Name: "default", Templates: []models.TemplateInfo{{Name: "ipxe", Path: "{{ .Env.Name }}", ID: "default"}}}, true},
		{"Create known-unavailable Bootenv", rt.Create, &models.BootEnv{Name: "unavailable"}, true},
		{"Create empty machine", rt.Create, &models.Machine{}, false},
		{"Create machine with bad Name /", rt.Create, &models.Machine{Name: "greg/greg"}, false},
		{"Create machine with bad Name \\", rt.Create, &models.Machine{Name: "greg\\greg"}, false},
		{"Create unnamed machine", rt.Create, &models.Machine{Uuid: okUUID}, false},
		{"Create named machine", rt.Create, &models.Machine{Uuid: okUUID, Name: "default.fqdn"}, true},
		{"Create new machine with same UUID", rt.Create, &models.Machine{Uuid: okUUID, Name: "other.fqdn"}, false},
		{"Create new machine with same name", rt.Create, &models.Machine{Uuid: uuid.NewRandom(), Name: "default.fqdn"}, false},
		{"Create new machine with invalid bootenv", rt.Create, &models.Machine{Uuid: uuid.NewRandom(), Name: "badenv.fqdn", BootEnv: "blargh"}, false},
		{"Create new machine with bad address", rt.Create, &models.Machine{Uuid: uuid.NewRandom(), Name: "badaddr.fqdn", BootEnv: "default", Address: net.ParseIP("127.0.0.1")}, false},
		{"Create another known-good bootenv", rt.Create, &models.BootEnv{Name: "new", Templates: []models.TemplateInfo{{Name: "ipxe", Path: "{{ .Env.Name }}", ID: "default"}}}, true},
		{"Update node with different bootenv", rt.Update, &models.Machine{Uuid: okUUID, Name: "default.fqdn", BootEnv: "new"}, true},
		{"Update node with unavailable bootenv", rt.Update, &models.Machine{Uuid: okUUID, Name: "default.fqdn", BootEnv: "unavailable"}, true},
		{"Remove machine that does not exist", rt.Remove, &models.Machine{Uuid: uuid.NewRandom()}, false},
		{"Remove machine that does exist", rt.Remove, &models.Machine{Uuid: okUUID, BootEnv: "new"}, true},
		{"Create named machine for patch", rt.Create, &models.Machine{Uuid: okUUID, Name: "default.fqdn"}, true},
	}
	for _, test := range tests {
		test.Test(t, rt)
	}
	rt.Do(func(d Stores) {
		machine := d("machines").Find(okUUID.String())
		patchTests := []patchTest{
			{"force replace name pass", true, 0, `[{"op":"replace","path":"/Name","value":"default2"}]`},
			{"replace name pass", true, 0, `[
{"op":"test","path":"/Name","value":"default.fqdn"},
{"op":"replace","path":"/Name","value":"default2"}
]`},
			{"replace name fail", false, 0, `[
{"op":"test","path":"/Name","value":"default2"},
{"op":"replace","path":"/Name","value":"default2"}
]`},
		}
		for _, test := range patchTests {
			test.test(t, machine)
		}
		// List test.
		bes := d("machines").Items()
		if bes != nil {
			if len(bes) != 1 {
				t.Errorf("List function should have returned: 1, but got %d\n", len(bes))
			}
		} else {
			t.Errorf("List function returned nil!!")
		}
	})
}
