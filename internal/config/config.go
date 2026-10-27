// Package config gère la liste des serveurs enregistrés.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/thiercelinflorian/ipmi-menu/internal/ipmi"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Hosts map[string]ipmi.Target `yaml:"hosts"`
	path  string
}

// Path renvoie ~/.config/ipmi-menu/hosts.yaml (ou $XDG_CONFIG_HOME).
func Path() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "ipmi-menu", "hosts.yaml")
}

func Load() (*Config, error) { return LoadFile(Path()) }

func LoadFile(path string) (*Config, error) {
	c := &Config{path: path}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("%s : %w", path, err)
	}
	if c.Hosts == nil {
		c.Hosts = map[string]ipmi.Target{}
	}
	return c, nil
}

// Save écrit le fichier en 0600 : il contient des mots de passe en clair.
func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.WriteFile(c.path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(c.path, 0o600)
}

func (c *Config) Names() []string {
	names := make([]string, 0, len(c.Hosts))
	for n := range c.Hosts {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (c *Config) Get(name string) (ipmi.Target, bool) {
	t, ok := c.Hosts[name]
	t.Name = name
	return t, ok
}

func (c *Config) Put(t ipmi.Target) { c.Hosts[t.Name] = t }

func (c *Config) Delete(name string) { delete(c.Hosts, name) }
