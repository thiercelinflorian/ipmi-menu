package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thiercelinflorian/ipmi-menu/internal/ipmi"
)

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "hosts.yaml")
	c, err := LoadFile(path)
	if err != nil || len(c.Hosts) != 0 {
		t.Fatalf("fichier absent : c=%v err=%v", c, err)
	}

	c.Put(ipmi.Target{Name: "srv2", Host: "10.0.0.2"})
	c.Put(ipmi.Target{Name: "srv1", Host: "10.0.0.1", User: "ADMIN", Password: "pw"})
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	if st, _ := os.Stat(path); st.Mode().Perm() != 0o600 {
		t.Fatalf("permissions = %v", st.Mode().Perm())
	}

	c, err = LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Names(); len(got) != 2 || got[0] != "srv1" {
		t.Fatalf("noms = %v", got)
	}
	s1, ok := c.Get("srv1")
	if !ok || s1.Name != "srv1" || s1.User != "ADMIN" || s1.Password != "pw" {
		t.Fatalf("srv1 = %+v", s1)
	}
	c.Delete("srv1")
	if _, ok := c.Get("srv1"); ok {
		t.Fatal("srv1 devrait être supprimé")
	}
}
