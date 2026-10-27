package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thiercelinflorian/ipmi-menu/internal/config"
	"github.com/thiercelinflorian/ipmi-menu/internal/ipmi"
)

type screen int

const (
	hostsScreen screen = iota
	formScreen
	menuScreen
	outputScreen
)

type Model struct {
	cfg    *config.Config
	target ipmi.Target
	screen screen

	hosts *menu
	stack []*menu
	form  form

	out      viewport.Model
	outTitle string

	spin     spinner.Model
	busy     string
	status   string
	isErr    bool
	pending  *item
	question string

	power       string
	boot        ipmi.BootOptions
	rebootAfter bool

	width, height int
	init          tea.Cmd
}

type (
	result struct {
		title, text string
		power       string
		err         error
	}
	solReady struct{ cmd *exec.Cmd }
	solDone  struct{ err error }
)

func Run(cfg *config.Config, target *ipmi.Target) error {
	m := &Model{cfg: cfg, rebootAfter: true, width: 80, height: 24}
	m.spin = spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(accent))
	if target != nil {
		m.init = m.connect(*target)
	} else {
		m.showHosts()
	}
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (m *Model) Init() tea.Cmd { return m.init }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.out.Width, m.out.Height = m.width-4, m.outputHeight()
		return m, nil

	case spinner.TickMsg:
		if m.busy == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case result:
		m.busy = ""
		if msg.power != "" {
			m.power = msg.power
		}
		switch {
		case msg.err != nil:
			m.setStatus(msg.err.Error(), true)
		case msg.title != "":
			m.openOutput(msg.title, msg.text)
		default:
			m.setStatus(msg.text, false)
		}
		return m, nil

	case solReady:
		m.busy = ""
		return m, tea.ExecProcess(msg.cmd, func(err error) tea.Msg { return solDone{err} })

	case solDone:
		if msg.err != nil {
			m.setStatus("Console SOL : "+msg.err.Error(), true)
		} else {
			m.setStatus("Console SOL fermée", false)
		}
		return m, nil

	case tea.KeyMsg:
		return m, m.key(msg)
	}

	if m.screen == formScreen {
		return m, m.form.update(msg)
	}
	return m, nil
}

func (m *Model) key(k tea.KeyMsg) tea.Cmd {
	s := k.String()
	if s == "ctrl+c" {
		return tea.Quit
	}
	if m.busy != "" {
		return nil
	}
	if m.pending != nil {
		it := m.pending
		m.pending = nil
		if s == "o" || s == "y" || s == "O" || s == "Y" {
			return it.action(m)
		}
		m.setStatus("Annulé", false)
		return nil
	}

	switch m.screen {
	case formScreen:
		return m.formKey(k)
	case outputScreen:
		switch s {
		case "esc", "q", "backspace", "left", "h":
			m.screen = menuScreen
			return nil
		}
		var cmd tea.Cmd
		m.out, cmd = m.out.Update(k)
		return cmd
	}

	mn := m.current()
	n := len(mn.items)
	switch s {
	case "q":
		return tea.Quit
	case "up", "k", "shift+tab":
		mn.cursor = (mn.cursor - 1 + n) % n
	case "down", "j", "tab":
		mn.cursor = (mn.cursor + 1) % n
	case "home", "g":
		mn.cursor = 0
	case "end", "G":
		mn.cursor = n - 1
	case "enter", " ", "right", "l":
		return m.activate(mn.items[mn.cursor])
	case "esc", "backspace", "left", "h":
		m.back()
	case "r":
		if m.screen == menuScreen {
			return m.refresh("Lecture de l'état…")
		}
	case "e", "d":
		if m.screen == hostsScreen {
			return m.hostKey(s, mn.items[mn.cursor].key)
		}
	default:
		if len(s) == 1 && s >= "1" && s <= "9" {
			if i := int(s[0] - '1'); i < n {
				mn.cursor = i
				return m.activate(mn.items[i])
			}
		}
	}
	return nil
}

func (m *Model) activate(it item) tea.Cmd {
	if it.confirm != nil {
		if q := it.confirm(m); q != "" {
			m.ask(q, it)
			return nil
		}
	}
	return it.action(m)
}

func (m *Model) ask(question string, it item) {
	m.pending, m.question = &it, question
}

func (m *Model) back() {
	m.status = ""
	switch {
	case m.screen == hostsScreen:
	case len(m.stack) > 1:
		m.stack = m.stack[:len(m.stack)-1]
	default:
		m.showHosts()
	}
}

func (m *Model) current() *menu {
	if m.screen == hostsScreen {
		return m.hosts
	}
	return m.stack[len(m.stack)-1]
}

func (m *Model) run(busy string, f func() tea.Msg) tea.Cmd {
	m.busy, m.status = busy, ""
	return tea.Batch(m.spin.Tick, f)
}

func (m *Model) setStatus(s string, isErr bool) { m.status, m.isErr = s, isErr }

func (m *Model) connect(t ipmi.Target) tea.Cmd {
	m.target, m.power = t, ""
	m.stack = []*menu{mainMenu()}
	m.screen = menuScreen
	return m.refresh("Connexion à " + t.Label() + "…")
}

func (m *Model) refresh(busy string) tea.Cmd {
	t := m.target
	return m.run(busy, func() tea.Msg {
		on, err := t.PowerOn()
		return result{power: onOff(on, err), err: err}
	})
}

func onOff(on bool, err error) string {
	switch {
	case err != nil:
		return ""
	case on:
		return "on"
	}
	return "off"
}

func (m *Model) openOutput(title, text string) {
	if strings.TrimSpace(text) == "" {
		text = "(aucune donnée)"
	}
	m.outTitle = title
	m.out = viewport.New(m.width-4, m.outputHeight())
	m.out.SetContent(text)
	m.screen = outputScreen
}

func (m *Model) outputHeight() int { return max(m.height-9, 3) }

func (m *Model) View() string {
	var b strings.Builder
	b.WriteString(m.header() + "\n\n")
	switch m.screen {
	case formScreen:
		b.WriteString(m.form.view())
	case outputScreen:
		b.WriteString(title.Render(m.outTitle) + "\n" + m.out.View())
	default:
		b.WriteString(m.menuView(m.current()))
	}
	b.WriteString("\n\n" + m.statusView() + "\n" + dim.Render(m.help()))
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m *Model) header() string {
	left := brand.Render("ipmi-menu")
	right := ""
	if m.screen != hostsScreen && m.screen != formScreen && m.target.Host != "" {
		left += dim.Render("  " + m.target.Label())
		if m.target.Name != "" {
			left += dim.Render(" · " + m.target.Host)
		}
		switch m.power {
		case "on":
			right = green.Render("● allumé")
		case "off":
			right = red.Render("○ éteint")
		default:
			right = dim.Render("? état inconnu")
		}
	}
	gap := max(m.width-4-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) menuView(mn *menu) string {
	var b strings.Builder
	if m.screen == hostsScreen {
		b.WriteString(title.Render("Serveurs") + "\n")
		if len(m.cfg.Hosts) == 0 {
			b.WriteString(dim.Render("Aucun serveur enregistré.") + "\n")
		}
	} else {
		crumbs := make([]string, len(m.stack))
		for i, s := range m.stack {
			crumbs[i] = s.title
		}
		b.WriteString(title.Render(strings.Join(crumbs, " › ")) + "\n")
	}
	for i, it := range mn.items {
		if it.sep {
			b.WriteString("\n")
		}
		cursor, label := "  ", it.label
		if i == mn.cursor {
			cursor, label = accent.Render("▸ "), selected.Render(label)
		}
		num := "  "
		if i < 9 {
			num = dim.Render(fmt.Sprintf("%d ", i+1))
		}
		line := cursor + num + label
		if it.value != nil {
			line += "  " + value.Render(it.value(m))
		}
		if it.hint != "" {
			line += "  " + dim.Render(it.hint)
		}
		b.WriteString(line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m *Model) statusView() string {
	switch {
	case m.pending != nil:
		return warn.Render("⚠ "+m.question) + dim.Render("  [o/N]")
	case m.busy != "":
		return m.spin.View() + " " + m.busy
	case m.isErr:
		return red.Render("✗ " + m.status)
	case m.status != "":
		return green.Render("✓ ") + m.status
	}
	return ""
}

func (m *Model) help() string {
	switch m.screen {
	case hostsScreen:
		return "↑↓ choisir · ⏎ connecter · e modifier · d supprimer · q quitter"
	case formScreen:
		return "tab champ suivant · ⏎ valider · esc annuler"
	case outputScreen:
		return "↑↓ pgup pgdn défiler · esc retour"
	}
	return "↑↓ choisir · ⏎ valider · esc retour · r rafraîchir l'état · q quitter"
}

var (
	accentColor = lipgloss.AdaptiveColor{Light: "#5B3FD9", Dark: "#A78BFA"}
	accent      = lipgloss.NewStyle().Foreground(accentColor)
	brand       = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	title       = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	selected    = lipgloss.NewStyle().Bold(true)
	dim         = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8A8A8A", Dark: "#6E6E6E"})
	value       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0E7490", Dark: "#67E8F9"})
	green       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#15803D", Dark: "#4ADE80"})
	red         = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#B91C1C", Dark: "#F87171"})
	warn        = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FBBF24"}).Bold(true)
)
