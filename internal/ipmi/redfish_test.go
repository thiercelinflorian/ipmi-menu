package ipmi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// fakeBMC simule un service Redfish. Avec actions=false, le CD virtuel est
// rangé sous Systems et ne se pilote que par PATCH (anciens BMC).
type fakeBMC struct {
	mu       sync.Mutex
	actions  bool
	inserted bool
	image    string
	calls    []string
}

func (b *fakeBMC) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if u, p, _ := r.BasicAuth(); u != "ADMIN" || p != "pw" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	body, _ := io.ReadAll(r.Body)
	b.calls = append(b.calls, r.Method+" "+r.URL.Path+" "+string(body))

	cd := "/redfish/v1/Managers/1/VirtualMedia/CD"
	if !b.actions {
		cd = "/redfish/v1/Systems/1/VirtualMedia/CD"
	}
	send := func(v any) { json.NewEncoder(w).Encode(v) }
	members := func(ids ...string) map[string]any {
		var m []map[string]string
		for _, id := range ids {
			m = append(m, map[string]string{"@odata.id": id})
		}
		return map[string]any{"Members": m}
	}

	switch {
	case r.Method == "GET" && r.URL.Path == "/redfish/v1/Managers":
		send(members("/redfish/v1/Managers/1"))
	case r.Method == "GET" && r.URL.Path == "/redfish/v1/Systems":
		send(members("/redfish/v1/Systems/1"))
	case r.Method == "GET" && r.URL.Path == "/redfish/v1/Managers/1/VirtualMedia" && b.actions:
		send(members("/redfish/v1/Managers/1/VirtualMedia/Floppy", cd))
	case r.Method == "GET" && r.URL.Path == "/redfish/v1/Systems/1/VirtualMedia" && !b.actions:
		send(members(cd))
	case r.Method == "GET" && r.URL.Path == "/redfish/v1/Managers/1/VirtualMedia/Floppy":
		send(map[string]any{"@odata.id": r.URL.Path, "MediaTypes": []string{"Floppy", "USBStick"}})
	case r.Method == "GET" && r.URL.Path == cd:
		v := map[string]any{"@odata.id": cd, "MediaTypes": []string{"CD", "DVD"}, "Inserted": b.inserted, "Image": b.image}
		if b.actions {
			v["Actions"] = map[string]any{
				"#VirtualMedia.InsertMedia": map[string]string{"target": cd + "/Actions/VirtualMedia.InsertMedia"},
				"#VirtualMedia.EjectMedia":  map[string]string{"target": cd + "/Actions/VirtualMedia.EjectMedia"},
			}
		}
		send(v)
	case r.Method == "POST" && r.URL.Path == cd+"/Actions/VirtualMedia.InsertMedia":
		var req map[string]any
		json.Unmarshal(body, &req)
		b.image, b.inserted = req["Image"].(string), true
		w.WriteHeader(http.StatusNoContent)
	case r.Method == "POST" && r.URL.Path == cd+"/Actions/VirtualMedia.EjectMedia":
		b.image, b.inserted = "", false
		w.WriteHeader(http.StatusNoContent)
	case r.Method == "PATCH" && r.URL.Path == cd && !b.actions:
		var req struct {
			Image    *string
			Inserted bool
		}
		json.Unmarshal(body, &req)
		b.inserted, b.image = req.Inserted, ""
		if req.Image != nil {
			b.image = *req.Image
		}
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusNotFound)
		send(map[string]any{"error": map[string]any{"message": "introuvable : " + r.URL.Path}})
	}
}

func startBMC(t *testing.T, b *fakeBMC) {
	t.Helper()
	srv := httptest.NewTLSServer(b)
	t.Cleanup(srv.Close)
	old, oldDelay := redfishBase, mediaDelay
	redfishBase, mediaDelay = func(Target) string { return srv.URL }, 0
	t.Cleanup(func() { redfishBase, mediaDelay = old, oldDelay })
}

func TestNetbootWithActions(t *testing.T) {
	calls := fake(t)
	bmc := &fakeBMC{actions: true, inserted: true, image: "http://ancienne.iso"}
	startBMC(t, bmc)
	tg := Target{Host: "h", User: "ADMIN", Password: "pw"}

	if _, err := tg.Netboot("", false, true); err != nil {
		t.Fatal(err)
	}
	if !bmc.inserted || bmc.image != NetbootISO {
		t.Fatalf("CD virtuel : inserted=%v image=%q", bmc.inserted, bmc.image)
	}
	all := strings.Join(bmc.calls, "\n")
	if !strings.Contains(all, "EjectMedia") {
		t.Fatalf("l'ancienne image aurait dû être éjectée :\n%s", all)
	}
	if !strings.Contains(calls(), "bootdev cdrom options=efiboot") || !strings.Contains(calls(), "chassis power cycle") {
		t.Fatalf("appels ipmitool :\n%s", calls())
	}

	if err := tg.EjectISO(); err != nil || bmc.inserted {
		t.Fatalf("éjection : err=%v inserted=%v", err, bmc.inserted)
	}
}

func TestNetbootPatchOnly(t *testing.T) {
	fake(t)
	bmc := &fakeBMC{}
	startBMC(t, bmc)
	tg := Target{Host: "h", User: "ADMIN", Password: "pw"}

	if _, err := tg.Netboot("http://miroir.local/netboot.xyz.iso", true, false); err != nil {
		t.Fatal(err)
	}
	if !bmc.inserted || bmc.image != "http://miroir.local/netboot.xyz.iso" {
		t.Fatalf("CD virtuel : inserted=%v image=%q", bmc.inserted, bmc.image)
	}
	if err := tg.EjectISO(); err != nil || bmc.inserted {
		t.Fatalf("éjection : err=%v inserted=%v", err, bmc.inserted)
	}
}

func TestRedfishBadCredentials(t *testing.T) {
	startBMC(t, &fakeBMC{actions: true})
	err := Target{Host: "h", User: "ADMIN", Password: "faux"}.InsertISO(NetbootISO)
	if err == nil || !strings.Contains(err.Error(), "identifiants refusés") {
		t.Fatalf("err = %v", err)
	}
}
