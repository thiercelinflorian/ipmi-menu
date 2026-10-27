package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thiercelinflorian/ipmi-menu/internal/ipmi"
)

type item struct {
	label   string
	hint    string
	key     string
	sep     bool
	value   func(m *Model) string
	confirm func(m *Model) string
	action  func(m *Model) tea.Cmd
}

type menu struct {
	title  string
	items  []item
	cursor int
}

func (m *Model) showHosts() {
	var items []item
	for _, name := range m.cfg.Names() {
		t, _ := m.cfg.Get(name)
		items = append(items, item{label: name, hint: t.Host, key: name,
			action: func(m *Model) tea.Cmd { return m.connect(t) }})
	}
	items = append(items, item{label: "+ Nouveau serveur", sep: len(items) > 0,
		action: func(m *Model) tea.Cmd { return m.openForm(ipmi.Target{}) }})

	cursor := 0
	if m.hosts != nil {
		cursor = min(m.hosts.cursor, len(items)-1)
	}
	m.hosts = &menu{items: items, cursor: cursor}
	m.screen, m.status = hostsScreen, ""
}

func (m *Model) hostKey(k, name string) tea.Cmd {
	if name == "" {
		return nil
	}
	if k == "e" {
		t, _ := m.cfg.Get(name)
		return m.openForm(t)
	}
	m.ask("Supprimer "+name+" ?", item{action: func(m *Model) tea.Cmd {
		m.cfg.Delete(name)
		if err := m.cfg.Save(); err != nil {
			m.setStatus(err.Error(), true)
			return nil
		}
		m.showHosts()
		m.setStatus(name+" supprimé", false)
		return nil
	}})
	return nil
}

func submenu(build func() *menu) func(*Model) tea.Cmd {
	return func(m *Model) tea.Cmd {
		m.stack, m.status = append(m.stack, build()), ""
		return nil
	}
}

func mainMenu() *menu {
	return &menu{title: "Menu", items: []item{
		{label: "Alimentation", hint: "›", action: submenu(powerMenu)},
		{label: "Démarrage", hint: "›", action: submenu(bootMenu)},
		{label: "Console SOL", hint: "~. pour quitter", action: (*Model).sol},
		{label: "Capteurs", action: show("Capteurs", ipmi.Target.Sensors)},
		{label: "Infos matériel", action: show("Infos matériel", ipmi.Target.Hardware)},
		{label: "Réseau BMC", action: show("Réseau BMC", ipmi.Target.LAN)},
		{label: "Changer de serveur", sep: true, action: func(m *Model) tea.Cmd { m.showHosts(); return nil }},
		{label: "Quitter", action: func(*Model) tea.Cmd { return tea.Quit }},
	}}
}

func powerMenu() *menu {
	return &menu{title: "Alimentation", items: []item{
		{label: "Allumer", action: power("on", "Allumage demandé")},
		{label: "Arrêt propre", hint: "ACPI", action: power("soft", "Arrêt propre demandé")},
		{label: "Arrêt forcé", confirm: ask("Couper l'alimentation de %s ?"), action: power("off", "Alimentation coupée")},
		{label: "Redémarrer", hint: "power cycle", confirm: ask("Redémarrer %s ?"), action: power("cycle", "Redémarrage envoyé")},
		{label: "Reset matériel", confirm: ask("Reset matériel de %s ?"), action: power("reset", "Reset envoyé")},
	}}
}

func bootMenu() *menu {
	type setter func(t ipmi.Target, opts ipmi.BootOptions, iso string) (dropped bool, err error)
	entry := func(label, hint string, set setter) item {
		return item{
			label: label,
			hint:  hint,
			confirm: func(m *Model) string {
				if !m.rebootAfter {
					return ""
				}
				return fmt.Sprintf("Redémarrer %s sur %s ?", m.target.Label(), label)
			},
			action: func(m *Model) tea.Cmd {
				t, opts, iso, reboot := m.target, m.boot, m.cfg.NetbootISO, m.rebootAfter
				return m.run("Configuration du démarrage…", func() tea.Msg {
					dropped, err := set(t, opts, iso)
					if err != nil {
						return result{err: err}
					}
					text := "Prochain démarrage : " + label
					if dropped {
						text += " (options UEFI/persistant refusées par le BMC)"
					}
					if !reboot {
						return result{text: text}
					}
					if err := t.Reboot(); err != nil {
						return result{err: err}
					}
					on, err := t.PowerOn()
					return result{text: text + ", serveur redémarré", power: onOff(on, err)}
				})
			},
		}
	}
	dev := func(label, dev string) item {
		return entry(label, "", func(t ipmi.Target, opts ipmi.BootOptions, _ string) (bool, error) {
			return t.SetBoot(dev, opts)
		})
	}
	return &menu{title: "Démarrage", items: []item{
		dev("PXE", "pxe"),
		dev("Disque", "disk"),
		dev("CD/DVD · ISO virtuel", "cdrom"),
		dev("Setup BIOS", "bios"),
		entry("netboot.xyz", "ISO montée via Redfish", func(t ipmi.Target, opts ipmi.BootOptions, iso string) (bool, error) {
			return t.Netboot(iso, opts.Legacy, false)
		}),
		{label: "Mode", sep: true,
			value:  func(m *Model) string { return pick(m.boot.Legacy, "Legacy", "UEFI") },
			action: func(m *Model) tea.Cmd { m.boot.Legacy = !m.boot.Legacy; return nil }},
		{label: "Persistant",
			value:  func(m *Model) string { return pick(m.boot.Persistent, "oui", "non (une seule fois)") },
			action: func(m *Model) tea.Cmd { m.boot.Persistent = !m.boot.Persistent; return nil }},
		{label: "Redémarrer après",
			value:  func(m *Model) string { return pick(m.rebootAfter, "oui", "non") },
			action: func(m *Model) tea.Cmd { m.rebootAfter = !m.rebootAfter; return nil }},
		{label: "Éjecter l'ISO virtuelle", sep: true, action: func(m *Model) tea.Cmd {
			t := m.target
			return m.run("Éjection…", func() tea.Msg {
				return result{text: "ISO éjectée", err: t.EjectISO()}
			})
		}},
	}}
}

func pick(b bool, yes, no string) string {
	if b {
		return yes
	}
	return no
}

func ask(format string) func(*Model) string {
	return func(m *Model) string { return fmt.Sprintf(format, m.target.Label()) }
}

func power(action, done string) func(*Model) tea.Cmd {
	return func(m *Model) tea.Cmd {
		t := m.target
		return m.run("Envoi de la commande…", func() tea.Msg {
			if _, err := t.Power(action); err != nil {
				return result{err: err}
			}
			on, err := t.PowerOn()
			return result{text: done, power: onOff(on, err)}
		})
	}
}

func show(title string, f func(ipmi.Target) (string, error)) func(*Model) tea.Cmd {
	return func(m *Model) tea.Cmd {
		t := m.target
		return m.run("Lecture…", func() tea.Msg {
			out, err := f(t)
			return result{title: title, text: out, err: err}
		})
	}
}

func (m *Model) sol() tea.Cmd {
	t := m.target
	return m.run("Ouverture de la console…", func() tea.Msg { return solReady{t.SOL()} })
}

var formLabels = []string{"Nom", "Adresse BMC", "Utilisateur", "Mot de passe"}

type form struct {
	inputs []textinput.Model
	focus  int
	orig   ipmi.Target
	err    string
}

func (m *Model) openForm(t ipmi.Target) tea.Cmd {
	in := func(placeholder, val string) textinput.Model {
		ti := textinput.New()
		ti.Prompt, ti.Placeholder, ti.CharLimit = "", placeholder, 128
		ti.SetValue(val)
		return ti
	}
	f := form{orig: t, inputs: []textinput.Model{
		in("vide = connexion sans enregistrer", t.Name),
		in("10.0.0.10 ou bmc.exemple.fr", t.Host),
		in(ipmi.DefaultUser, t.User),
		in("", t.Password),
	}}
	f.inputs[3].EchoMode, f.inputs[3].EchoCharacter = textinput.EchoPassword, '•'
	m.form, m.screen = f, formScreen
	return m.form.inputs[0].Focus()
}

func (f *form) move(d int) tea.Cmd {
	f.inputs[f.focus].Blur()
	f.focus = (f.focus + d + len(f.inputs)) % len(f.inputs)
	return f.inputs[f.focus].Focus()
}

func (f *form) update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.inputs[f.focus], cmd = f.inputs[f.focus].Update(msg)
	return cmd
}

func (m *Model) formKey(k tea.KeyMsg) tea.Cmd {
	f := &m.form
	switch k.String() {
	case "esc":
		m.showHosts()
		return nil
	case "tab", "down":
		return f.move(1)
	case "shift+tab", "up":
		return f.move(-1)
	case "enter":
		if f.focus < len(f.inputs)-1 {
			return f.move(1)
		}
		return m.submitForm()
	}
	return f.update(k)
}

func (m *Model) submitForm() tea.Cmd {
	f := &m.form
	val := func(i int) string { return strings.TrimSpace(f.inputs[i].Value()) }
	t := f.orig
	t.Name, t.Host, t.User, t.Password = val(0), val(1), val(2), f.inputs[3].Value()
	if t.Host == "" {
		f.err = "L'adresse du BMC est obligatoire."
		f.inputs[f.focus].Blur()
		f.focus = 1
		return f.inputs[1].Focus()
	}
	if t.Name == "" {
		return m.connect(t)
	}
	if f.orig.Name != "" && f.orig.Name != t.Name {
		m.cfg.Delete(f.orig.Name)
	}
	m.cfg.Put(t)
	if err := m.cfg.Save(); err != nil {
		f.err = err.Error()
		return nil
	}
	if f.orig.Name != "" {
		m.showHosts()
		m.setStatus(t.Name+" enregistré", false)
		return nil
	}
	return m.connect(t)
}

func (f *form) view() string {
	heading := "Nouveau serveur"
	if f.orig.Name != "" {
		heading = "Modifier " + f.orig.Name
	}
	s := title.Render(heading) + "\n"
	for i, in := range f.inputs {
		label := dim.Render(fmt.Sprintf("%-14s", formLabels[i]))
		if i == f.focus {
			label = accent.Render(fmt.Sprintf("%-14s", formLabels[i]))
		}
		s += label + in.View() + "\n"
	}
	if f.err != "" {
		s += "\n" + red.Render(f.err)
	}
	return s
}
