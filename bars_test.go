package main

import (
	"regexp"
	"strings"
	"testing"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func TestBarColor_AlwaysGreen(t *testing.T) {
	for _, v := range []int{0, 1, 5, 10} {
		got := BarColor(v)
		if string(got) != "#30B020" {
			t.Errorf("BarColor(%d) = %s, want Sims green", v, got)
		}
	}
}

func TestEmptyBarColor_Adaptive(t *testing.T) {
	tests := []struct {
		value int
		want  string
	}{
		{0, "#8B1A1A"},
		{4, "#8B1A1A"},
		{5, "#B08B00"},
		{6, "#B08B00"},
		{7, "#1A6B10"},
		{9, "#1A6B10"},
	}
	for _, tt := range tests {
		got := string(EmptyBarColor(tt.value))
		if got != tt.want {
			t.Errorf("EmptyBarColor(%d) = %s, want %s", tt.value, got, tt.want)
		}
	}
}

func TestRenderBar_Width(t *testing.T) {
	for _, w := range []int{10, 20, 30} {
		bar := RenderBar(5, w)
		stripped := stripAnsi(bar)
		got := len([]rune(stripped))
		if got != w {
			t.Errorf("RenderBar(5, %d) stripped width = %d, want %d; stripped=%q", w, got, w, stripped)
		}
	}
}

func TestRenderBar_FilledAndEmpty(t *testing.T) {
	// value=10, width=10 → all filled
	bar := stripAnsi(RenderBar(10, 10))
	if strings.Count(bar, "█") != 10 {
		t.Errorf("value=10: bar=%q, want all █", bar)
	}

	// value=0, width=10 → all empty
	bar = stripAnsi(RenderBar(0, 10))
	if strings.Count(bar, "░") != 10 {
		t.Errorf("value=0: bar=%q, want all ░", bar)
	}

	// value=5, width=20 → 10 filled + 10 empty
	bar = stripAnsi(RenderBar(5, 20))
	if strings.Count(bar, "█") != 10 {
		t.Errorf("value=5,w=20: filled=%d, want 10; bar=%q", strings.Count(bar, "█"), bar)
	}
	if strings.Count(bar, "░") != 10 {
		t.Errorf("value=5,w=20: empty=%d, want 10; bar=%q", strings.Count(bar, "░"), bar)
	}
}

func TestRenderGrid_HasAllLabels(t *testing.T) {
	values := [NeedCount]int{5, 5, 5, 5, 5, 5, 5, 5}
	grid := stripAnsi(RenderGrid(values, 20))

	for _, name := range NeedNames() {
		if !strings.Contains(grid, name) {
			t.Errorf("grid missing label %q", name)
		}
	}
}

func TestRenderGrid_TwoColumns(t *testing.T) {
	values := [NeedCount]int{5, 5, 5, 5, 5, 5, 5, 5}
	grid := RenderGrid(values, 20)
	// 4 rows separated by blank lines = 7 lines total
	lines := strings.Split(grid, "\n")
	if len(lines) != 7 {
		t.Errorf("grid has %d lines, want 7", len(lines))
	}
}
