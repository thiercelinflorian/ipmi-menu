package ipmi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultUser      = "root"
	DefaultPort      = 623
	DefaultInterface = "lanplus"

	Fast   = 10 * time.Second
	Normal = 30 * time.Second
	Slow   = 90 * time.Second
)

var Bin = "ipmitool"

type Target struct {
	Name      string `yaml:"-"`
	Host      string `yaml:"host"`
	User      string `yaml:"user,omitempty"`
	Password  string `yaml:"password,omitempty"`
	Port      int    `yaml:"port,omitempty"`
	Interface string `yaml:"interface,omitempty"`
}

func (t Target) Label() string {
	if t.Name != "" {
		return t.Name
	}
	return t.Host
}

func Check() error {
	if _, err := exec.LookPath(Bin); err != nil {
		return errors.New("ipmitool est introuvable (brew install ipmitool / apt install ipmitool)")
	}
	return nil
}

func (t Target) args(extra ...string) []string {
	user, port, iface := t.User, t.Port, t.Interface
	if user == "" {
		user = DefaultUser
	}
	if port == 0 {
		port = DefaultPort
	}
	if iface == "" {
		iface = DefaultInterface
	}
	base := []string{"-I", iface, "-H", t.Host, "-p", strconv.Itoa(port), "-U", user, "-E"}
	return append(base, extra...)
}

func (t Target) Command(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, Bin, t.args(args...)...)
	cmd.Env = append(os.Environ(), "IPMI_PASSWORD="+t.Password)
	return cmd
}

func (t Target) Run(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := t.Command(ctx, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())

	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("%s ne répond pas (délai de %s dépassé)", t.Host, timeout)
	}
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return out, errors.New(msg + hint(msg))
	}
	return out, nil
}

func hint(msg string) string {
	m := strings.ToLower(msg)
	switch {
	case strings.Contains(m, "rakp"), strings.Contains(m, "unauthorized"),
		strings.Contains(m, "invalid user"), strings.Contains(m, "password"):
		return "\n→ identifiants refusés par le BMC"
	case strings.Contains(m, "unable to establish"), strings.Contains(m, "no response"):
		return "\n→ BMC injoignable (adresse, réseau, port 623/udp ?)"
	}
	return ""
}

var PowerActions = []string{"status", "on", "off", "soft", "cycle", "reset"}

func (t Target) Power(action string) (string, error) {
	return t.Run(Normal, "chassis", "power", action)
}

func (t Target) PowerOn() (bool, error) {
	out, err := t.Run(Fast, "chassis", "power", "status")
	if err != nil {
		return false, err
	}
	return strings.HasSuffix(strings.ToLower(out), " on"), nil
}

func (t Target) Reboot() error {
	on, err := t.PowerOn()
	if err != nil {
		return err
	}
	action := "on"
	if on {
		action = "cycle"
	}
	_, err = t.Power(action)
	return err
}

var BootDevices = []string{"pxe", "disk", "cdrom", "bios"}

type BootOptions struct {
	Legacy     bool
	Persistent bool
}

func (t Target) SetBoot(dev string, o BootOptions) (dropped bool, err error) {
	args := []string{"chassis", "bootdev", dev}
	var opts []string
	if o.Persistent {
		opts = append(opts, "persistent")
	}
	if !o.Legacy {
		opts = append(opts, "efiboot")
	}
	if len(opts) == 0 {
		_, err = t.Run(Normal, args...)
		return false, err
	}
	if _, err = t.Run(Normal, append(args, "options="+strings.Join(opts, ","))...); err == nil {
		return false, nil
	}
	if _, err2 := t.Run(Normal, args...); err2 == nil {
		return true, nil
	}
	return false, err
}

func (t Target) SOL() *exec.Cmd {
	t.Run(Fast, "sol", "deactivate")
	return t.Command(context.Background(), "sol", "activate")
}

func (t Target) Sensors() (string, error) { return t.Run(Slow, "sdr", "list") }

func (t Target) Hardware() (string, error) {
	mc, err := t.Run(Normal, "mc", "info")
	if err != nil {
		return "", err
	}
	fru, err := t.Run(Slow, "fru", "print")
	if err != nil {
		return mc, nil
	}
	return mc + "\n\n" + fru, nil
}

func (t Target) LAN() (string, error) {
	out, err := t.Run(Normal, "lan", "print")
	if err == nil {
		return out, nil
	}
	return t.Run(Normal, "lan", "print", "1")
}
