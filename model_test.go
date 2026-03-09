package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func testModel(t *testing.T) model {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.json")
	return NewModel(path)
}

func testModelWithHistory(t *testing.T) model {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.json")
	h := History{Entries: []Entry{
		{Timestamp: time.Now(), Values: [NeedCount]int{3, 4, 5, 6, 7, 8, 9, 10}},
	}}
	Save(path, h)
	return NewModel(path)
}

func sendKey(m tea.Model, key string) tea.Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated
}

func sendSpecialKey(m tea.Model, keyType tea.KeyType) tea.Model {
	updated, _ := m.Update(tea.KeyMsg{Type: keyType})
	return updated
}

func TestNewModel_EmptyHistory(t *testing.T) {
	m := testModel(t)
	if m.mode != modeDisplay {
		t.Errorf("mode = %d, want display", m.mode)
	}
	for i, v := range m.values {
		if v != 0 {
			t.Errorf("values[%d] = %d, want 0", i, v)
		}
	}
}

func TestNewModel_WithHistory(t *testing.T) {
	m := testModelWithHistory(t)
	expected := [NeedCount]int{3, 4, 5, 6, 7, 8, 9, 10}
	if m.values != expected {
		t.Errorf("values = %v, want %v", m.values, expected)
	}
}

func TestUpdate_Quit(t *testing.T) {
	m := testModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	// tea.Quit returns a special message
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected QuitMsg, got %T", msg)
	}
}

func TestUpdate_RKey_SwitchesToInput(t *testing.T) {
	m := testModel(t)
	result := sendKey(m, "r")
	rm := result.(model)
	if rm.mode != modeInput {
		t.Errorf("mode = %d, want input", rm.mode)
	}
	if rm.inputIndex != 0 {
		t.Errorf("inputIndex = %d, want 0", rm.inputIndex)
	}
}

func TestUpdate_HKey_SwitchesToHistory(t *testing.T) {
	m := testModel(t)
	result := sendKey(m, "h")
	rm := result.(model)
	if rm.mode != modeHistory {
		t.Errorf("mode = %d, want history", rm.mode)
	}
}

func TestUpdate_TickNoHourPassed(t *testing.T) {
	m := testModel(t)
	m.lastUpdate = time.Now()
	result, _ := m.Update(tickMsg(time.Now()))
	rm := result.(model)
	if rm.mode != modeDisplay {
		t.Errorf("mode = %d, want display", rm.mode)
	}
}

func TestUpdate_TickHourPassed(t *testing.T) {
	m := testModel(t)
	m.lastUpdate = time.Now().Add(-61 * time.Minute)
	result, _ := m.Update(tickMsg(time.Now()))
	rm := result.(model)
	if rm.mode != modeInput {
		t.Errorf("mode = %d, want input", rm.mode)
	}
}

func TestUpdate_InputDigit(t *testing.T) {
	m := testModel(t)
	m.mode = modeInput
	result := sendKey(m, "5")
	rm := result.(model)
	if rm.inputBuf != "5" {
		t.Errorf("inputBuf = %q, want %q", rm.inputBuf, "5")
	}
}

func TestUpdate_InputEnterValid(t *testing.T) {
	m := testModel(t)
	m.mode = modeInput
	m.inputBuf = "7"
	result := sendSpecialKey(m, tea.KeyEnter)
	rm := result.(model)
	if rm.inputValues[0] != 7 {
		t.Errorf("inputValues[0] = %d, want 7", rm.inputValues[0])
	}
	if rm.inputIndex != 1 {
		t.Errorf("inputIndex = %d, want 1", rm.inputIndex)
	}
}

func TestUpdate_InputEnterInvalid(t *testing.T) {
	tests := []struct {
		name string
		buf  string
	}{
		{"zero", "0"},
		{"eleven", "11"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel(t)
			m.mode = modeInput
			m.inputBuf = tt.buf
			result := sendSpecialKey(m, tea.KeyEnter)
			rm := result.(model)
			if rm.inputErr == "" {
				t.Error("expected error message")
			}
			if rm.inputIndex != 0 {
				t.Errorf("inputIndex = %d, want 0", rm.inputIndex)
			}
		})
	}
}

func TestUpdate_InputComplete(t *testing.T) {
	m := testModel(t)
	m.mode = modeInput
	// Fill in all 8 values
	for i := 0; i < int(NeedCount); i++ {
		m.inputBuf = "5"
		result := sendSpecialKey(m, tea.KeyEnter)
		m = result.(model)
	}
	if m.mode != modeDisplay {
		t.Errorf("mode = %d, want display", m.mode)
	}
	if len(m.history.Entries) != 1 {
		t.Errorf("history entries = %d, want 1", len(m.history.Entries))
	}
	for i, v := range m.values {
		if v != 5 {
			t.Errorf("values[%d] = %d, want 5", i, v)
		}
	}
}

func TestUpdate_InputEscape(t *testing.T) {
	m := testModel(t)
	m.mode = modeInput
	m.inputBuf = "5"
	result := sendSpecialKey(m, tea.KeyEscape)
	rm := result.(model)
	if rm.mode != modeDisplay {
		t.Errorf("mode = %d, want display", rm.mode)
	}
	if len(rm.history.Entries) != 0 {
		t.Error("history should not be modified on escape")
	}
}

func TestUpdate_HistoryEscape(t *testing.T) {
	m := testModel(t)
	m.mode = modeHistory
	result := sendSpecialKey(m, tea.KeyEscape)
	rm := result.(model)
	if rm.mode != modeDisplay {
		t.Errorf("mode = %d, want display", rm.mode)
	}
}

func TestUpdate_HistoryScroll(t *testing.T) {
	m := testModelWithHistory(t)
	m.mode = modeHistory
	m.historyScroll = 0

	// Scroll down
	result := sendSpecialKey(m, tea.KeyDown)
	rm := result.(model)
	if rm.historyScroll != 1 {
		t.Errorf("scroll = %d, want 1", rm.historyScroll)
	}

	// Scroll up
	result = sendSpecialKey(rm, tea.KeyUp)
	rm = result.(model)
	if rm.historyScroll != 0 {
		t.Errorf("scroll = %d, want 0", rm.historyScroll)
	}

	// Can't scroll below 0
	result = sendSpecialKey(rm, tea.KeyUp)
	rm = result.(model)
	if rm.historyScroll != 0 {
		t.Errorf("scroll = %d, want 0", rm.historyScroll)
	}
}

func TestView_DisplayContainsNeedNames(t *testing.T) {
	m := testModel(t)
	view := m.View()
	for _, name := range NeedNames() {
		if !strings.Contains(view, name) {
			t.Errorf("display view missing %q", name)
		}
	}
}

func TestView_DisplayShowsLastUpdated(t *testing.T) {
	m := testModel(t)
	view := m.View()
	if !strings.Contains(view, "Last updated") {
		t.Error("display view missing last updated timestamp")
	}
	if !strings.Contains(view, m.lastUpdate.Format("2006-01-02")) {
		t.Error("display view missing formatted date")
	}
}

func TestView_InputShowsPrompt(t *testing.T) {
	m := testModel(t)
	m.mode = modeInput
	m.inputIndex = 0
	view := m.View()
	if !strings.Contains(view, "Hunger") {
		t.Error("input view missing need name")
	}
	if !strings.Contains(view, "(1-10)") {
		t.Error("input view missing range prompt")
	}
}

func TestView_HistoryShowsTable(t *testing.T) {
	m := testModelWithHistory(t)
	m.mode = modeHistory
	view := m.View()
	if !strings.Contains(view, "History") {
		t.Error("history view missing header")
	}
	if !strings.Contains(view, "Date") {
		t.Error("history view missing Date column")
	}
}
