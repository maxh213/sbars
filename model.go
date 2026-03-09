package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	modeDisplay mode = iota
	modeInput
	modeHistory
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type model struct {
	mode          mode
	values        [NeedCount]int
	inputIndex    int
	inputBuf      string
	inputValues   [NeedCount]int
	inputErr      string
	history       History
	historyPath   string
	lastUpdate    time.Time
	historyScroll int
	width         int
	height        int
}

func NewModel(path string) model {
	h, _ := Load(path)
	var values [NeedCount]int
	if len(h.Entries) > 0 {
		values = h.Entries[len(h.Entries)-1].Values
	}
	return model{
		mode:        modeDisplay,
		values:      values,
		history:     h,
		historyPath: path,
		lastUpdate:  time.Now(),
		width:       80,
		height:      24,
	}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		elapsed := time.Since(m.lastUpdate)
		if m.mode == modeDisplay && elapsed >= time.Hour {
			m.mode = modeInput
			m.inputIndex = 0
			m.inputBuf = ""
			m.inputErr = ""
			m.inputValues = [NeedCount]int{}
		}
		return m, tickCmd()

	case tea.KeyMsg:
		switch m.mode {
		case modeDisplay:
			return m.updateDisplay(msg)
		case modeInput:
			return m.updateInput(msg)
		case modeHistory:
			return m.updateHistory(msg)
		}
	}
	return m, nil
}

func (m model) updateDisplay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		m.mode = modeInput
		m.inputIndex = 0
		m.inputBuf = ""
		m.inputErr = ""
		m.inputValues = [NeedCount]int{}
		return m, nil
	case "h":
		m.mode = modeHistory
		m.historyScroll = 0
		return m, nil
	}
	return m, nil
}

func (m model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeDisplay
		return m, nil
	case "enter":
		val, err := strconv.Atoi(m.inputBuf)
		if err != nil || val < 1 || val > 10 {
			m.inputErr = "Enter a number 1-10"
			return m, nil
		}
		m.inputValues[m.inputIndex] = val
		m.inputErr = ""
		m.inputBuf = ""
		m.inputIndex++
		if m.inputIndex >= int(NeedCount) {
			entry := Entry{
				Timestamp: time.Now(),
				Values:    m.inputValues,
			}
			m.history = AppendEntry(m.history, entry)
			m.values = m.inputValues
			m.lastUpdate = time.Now()
			_ = Save(m.historyPath, m.history)
			m.mode = modeDisplay
		}
		return m, nil
	case "backspace":
		if len(m.inputBuf) > 0 {
			m.inputBuf = m.inputBuf[:len(m.inputBuf)-1]
		}
		return m, nil
	default:
		if len(msg.String()) == 1 && msg.String()[0] >= '0' && msg.String()[0] <= '9' {
			m.inputBuf += msg.String()
		}
		return m, nil
	}
}

func (m model) updateHistory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "h", "q":
		m.mode = modeDisplay
		return m, nil
	case "up", "k":
		if m.historyScroll > 0 {
			m.historyScroll--
		}
		return m, nil
	case "down", "j":
		maxScroll := len(m.history.Entries)
		if m.historyScroll < maxScroll {
			m.historyScroll++
		}
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	switch m.mode {
	case modeInput:
		return m.viewInput()
	case modeHistory:
		return m.viewHistory()
	default:
		return m.viewDisplay()
	}
}

func (m model) viewDisplay() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#0E1550")).Width(m.width).Align(lipgloss.Center).Render("Needs")
	grid := RenderGrid(m.values, 20)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A5080"))
	updated := dimStyle.Render(fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("2006-01-02 15:04")))
	help := dimStyle.Render("[r] record  [h] history  [q] quit")

	content := title + "\n\n" + grid + "\n\n" + updated + "\n\n" + help
	bg := SimsBlueBackground(m.width, m.height)
	return bg.Render(content)
}

func (m model) viewInput() string {
	name := NeedName(Need(m.inputIndex))
	prompt := fmt.Sprintf("%s (1-10): %s", name, m.inputBuf)

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("  Record Needs"))
	lines = append(lines, "")

	// Show already entered values
	for i := 0; i < m.inputIndex; i++ {
		lines = append(lines, fmt.Sprintf("  %s: %d ✓", NeedName(Need(i)), m.inputValues[i]))
	}
	lines = append(lines, "  "+prompt+"▌")
	if m.inputErr != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("  "+m.inputErr))
	}
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#4A5080")).Render("  [enter] confirm  [esc] cancel"))

	bg := SimsBlueBackground(m.width, m.height)
	return bg.Render(strings.Join(lines, "\n"))
}

func (m model) viewHistory() string {
	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("  History"))
	lines = append(lines, "")

	// Header
	header := fmt.Sprintf("  %-20s", "Date")
	for _, n := range NeedNames() {
		header += fmt.Sprintf(" %4s", n[:3])
	}
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#0E1550")).Render(header))

	// Entries (newest first)
	entries := m.history.Entries
	for i := len(entries) - 1 - m.historyScroll; i >= 0; i-- {
		e := entries[i]
		row := fmt.Sprintf("  %-20s", e.Timestamp.Format("2006-01-02 15:04"))
		for _, v := range e.Values {
			row += fmt.Sprintf(" %4d", v)
		}
		lines = append(lines, row)
	}

	if len(entries) == 0 {
		lines = append(lines, "  No entries yet. Press [r] to record.")
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#4A5080")).Render("  [↑/↓] scroll  [esc] back"))

	bg := SimsBlueBackground(m.width, m.height)
	return bg.Render(strings.Join(lines, "\n"))
}
