package ipmi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fake branche le faux ipmitool de testdata et renvoie le journal des appels.
func fake(t *testing.T) func() string {
	t.Helper()
	bin, _ := filepath.Abs("testdata/ipmitool")
	old := Bin
	Bin = bin
	t.Cleanup(func() { Bin = old })
	dir := t.TempDir()
	log := filepath.Join(dir, "log")
	t.Setenv("FAKE_LOG", log)
	t.Setenv("FAKE_STATE", filepath.Join(dir, "state"))
	return func() string { b, _ := os.ReadFile(log); return string(b) }
}

func TestPasswordNotOnCommandLine(t *testing.T) {
	calls := fake(t)
	tg := Target{Host: "10.0.0.10", Password: "s3cret"}
	if _, err := tg.Power("status"); err != nil {
		t.Fatal(err)
	}
	got := calls()
	want := "-I lanplus -H 10.0.0.10 -p 623 -U root -E chassis power status | pw=s3cret"
	if strings.TrimSpace(got) != want {
		t.Fatalf("appel = %q, attendu %q", got, want)
	}
}

func TestPowerOnAndReboot(t *testing.T) {
	calls := fake(t)
	tg := Target{Host: "h", User: "ADMIN", Port: 1623}
	if _, err := tg.Power("off"); err != nil {
		t.Fatal(err)
	}
	if on, _ := tg.PowerOn(); on {
		t.Fatal("devrait être éteint")
	}
	// Éteint : Reboot doit allumer, pas faire de cycle.
	if err := tg.Reboot(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(calls(), "chassis power on") || strings.Contains(calls(), "cycle") {
		t.Fatalf("appels inattendus :\n%s", calls())
	}
	if !strings.Contains(calls(), "-U ADMIN") || !strings.Contains(calls(), "-p 1623") {
		t.Fatalf("user/port non transmis :\n%s", calls())
	}
}

func TestSetBoot(t *testing.T) {
	calls := fake(t)
	tg := Target{Host: "h"}
	dropped, err := tg.SetBoot("pxe", BootOptions{Persistent: true})
	if err != nil || dropped {
		t.Fatalf("dropped=%v err=%v", dropped, err)
	}
	if !strings.Contains(calls(), "bootdev pxe options=persistent,efiboot") {
		t.Fatalf("options absentes :\n%s", calls())
	}

	t.Setenv("FAKE_NO_OPTS", "1")
	dropped, err = tg.SetBoot("disk", BootOptions{})
	if err != nil || !dropped {
		t.Fatalf("repli sans options attendu : dropped=%v err=%v", dropped, err)
	}

	dropped, err = tg.SetBoot("disk", BootOptions{Legacy: true})
	if err != nil || dropped {
		t.Fatalf("legacy sans options : dropped=%v err=%v", dropped, err)
	}
}

func TestAuthErrorHint(t *testing.T) {
	fake(t)
	_, err := Target{Host: "h", Password: "mauvais"}.Power("status")
	if err == nil || !strings.Contains(err.Error(), "identifiants refusés") {
		t.Fatalf("err = %v", err)
	}
}
