package web

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type probeEntry struct {
	Name     string
	Category string // subdirectory name or "general"
	Title    string // first H1 from the file
	Date     time.Time
	Summary  string // first non-empty paragraph after the title
	Path     string // relative path for linking
}

type probesData struct {
	Categories []probeCategory
	Total      int
	Filter     string
	Err        string
}

type probeCategory struct {
	Name    string
	Probes  []probeEntry
	Count   int
	Latest  time.Time
}

func (s *Server) handleProbes(w http.ResponseWriter, r *http.Request) {
	if s.workspacePath == "" {
		s.render(w, r, "probes", probesData{Err: "No workspace configured"})
		return
	}

	filter := r.URL.Query().Get("category")

	// Search for docs/probes in the workspace path and sibling directories
	probesDir := findProbesDir(s.workspacePath)
	if probesDir == "" {
		s.render(w, r, "probes", probesData{Err: "Probes directory not found in any workspace"})
		return
	}

	catMap := map[string][]probeEntry{}

	// Scan top-level markdown files (general category)
	topFiles, _ := filepath.Glob(filepath.Join(probesDir, "*.md"))
	for _, f := range topFiles {
		if e := parseProbeFile(f, "general"); e != nil {
			catMap["general"] = append(catMap["general"], *e)
		}
	}

	// Scan subdirectories
	entries, _ := os.ReadDir(probesDir)
	for _, d := range entries {
		if !d.IsDir() {
			continue
		}
		subFiles, _ := filepath.Glob(filepath.Join(probesDir, d.Name(), "*.md"))
		for _, f := range subFiles {
			if e := parseProbeFile(f, d.Name()); e != nil {
				catMap[d.Name()] = append(catMap[d.Name()], *e)
			}
		}
	}

	var categories []probeCategory
	for name, probes := range catMap {
		if filter != "" && name != filter {
			continue
		}
		sort.Slice(probes, func(i, j int) bool {
			return probes[i].Date.After(probes[j].Date)
		})
		var latest time.Time
		if len(probes) > 0 {
			latest = probes[0].Date
		}
		categories = append(categories, probeCategory{
			Name:   name,
			Probes: probes,
			Count:  len(probes),
			Latest: latest,
		})
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Latest.After(categories[j].Latest)
	})

	total := 0
	for _, c := range categories {
		total += c.Count
	}

	s.render(w, r, "probes", probesData{
		Categories: categories,
		Total:      total,
		Filter:     filter,
	})
}

// parseProbeFile reads a markdown file and extracts title, date, and summary.
func parseProbeFile(path, category string) *probeEntry {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("probes: open %s: %v", path, err)
		return nil
	}
	defer f.Close()

	base := filepath.Base(path)
	entry := &probeEntry{
		Name:     base,
		Category: category,
		Path:     base,
	}

	// Try to parse date from filename (YYYY-MM-DD pattern)
	entry.Date = parseDateFromFilename(base)

	scanner := bufio.NewScanner(f)
	var foundTitle bool
	var summaryLines []string
	for scanner.Scan() {
		line := scanner.Text()

		// Extract title from first H1
		if !foundTitle && strings.HasPrefix(line, "# ") {
			entry.Title = strings.TrimPrefix(line, "# ")
			foundTitle = true
			continue
		}

		// Extract date from **Date**: line if not found in filename
		if entry.Date.IsZero() && strings.Contains(line, "**Date**:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				entry.Date = tryParseDate(strings.TrimSpace(parts[1]))
			}
		}

		// Collect first non-empty paragraph after title as summary
		if foundTitle && len(summaryLines) < 3 {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" && len(summaryLines) > 0 {
				break // end of first paragraph
			}
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "**") {
				summaryLines = append(summaryLines, trimmed)
			}
		}
	}

	if !foundTitle {
		// Use filename as title fallback
		entry.Title = strings.TrimSuffix(base, ".md")
	}

	entry.Summary = strings.Join(summaryLines, " ")
	if len(entry.Summary) > 200 {
		entry.Summary = entry.Summary[:200] + "..."
	}

	return entry
}

func parseDateFromFilename(name string) time.Time {
	// Try patterns like "2026-03-13" anywhere in the filename
	for i := 0; i+10 <= len(name); i++ {
		if t, err := time.Parse("2006-01-02", name[i:i+10]); err == nil {
			return t
		}
	}
	return time.Time{}
}

// findProbesDir searches for a docs/probes directory in the given workspace
// path and its sibling directories (other repos in the same parent).
func findProbesDir(workspacePath string) string {
	// Try the workspace directly
	candidate := filepath.Join(workspacePath, "docs", "probes")
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return candidate
	}

	// Try sibling directories (e.g., /opt/tapestry/repos/aegis when workspace is /opt/tapestry/repos/goldblum)
	parent := filepath.Dir(workspacePath)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		candidate = filepath.Join(parent, e.Name(), "docs", "probes")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

func tryParseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"Jan 2, 2006",
		"January 2, 2006",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
