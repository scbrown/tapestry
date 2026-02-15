package web

import (
	"html/template"
	"strings"
	"testing"
)

func TestPriorityLabel(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "P0"},
		{1, "P1"},
		{2, "P2"},
		{3, "P3"},
		{5, "P5"},
	}
	for _, tt := range tests {
		got := priorityLabel(tt.in)
		if got != tt.want {
			t.Errorf("priorityLabel(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestStatusBadge(t *testing.T) {
	tests := []struct {
		in    string
		color string
	}{
		{"open", "#3b82f6"},
		{"in_progress", "#f59e0b"},
		{"closed", "#22c55e"},
		{"unknown", "gray"},
	}
	for _, tt := range tests {
		got := string(statusBadge(tt.in))
		if !strings.Contains(got, tt.color) {
			t.Errorf("statusBadge(%q) = %q, want color %q", tt.in, got, tt.color)
		}
	}
}

func TestTemplatesParse(t *testing.T) {
	funcMap := template.FuncMap{
		"priorityLabel": priorityLabel,
		"statusBadge":   statusBadge,
		"progressPct":   progressPct,
		"payloadString": func(s string) string { return s },
		"timeAgo":       timeAgo,
		"shortActor":    shortActor,
		"fmtDuration":   fmtDuration,
		"rigName":       func(s string) string { return strings.TrimPrefix(s, "beads_") },
		"nl":            func(s string) string { return strings.ReplaceAll(s, `\n`, "\n") },
	}

	for _, name := range []string{"monthly.html", "bead.html", "beads.html", "epic.html"} {
		_, err := template.New(name).Funcs(funcMap).ParseFS(templateFS,
			"templates/layout.html", "templates/"+name)
		if err != nil {
			t.Errorf("parse %s: %v", name, err)
		}
	}
}
