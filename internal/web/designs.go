package web

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type designsListData struct {
	Designs []designEntry
	Total   int
	Err     string
}

type designEntry struct {
	Name     string
	Title    string
	Size     int
	Path     string
	Modified time.Time
}

type designViewData struct {
	Name    string
	Title   string
	Content template.HTML
	Raw     string
	GitURL  string
	Err     string
}

type forgejoClient struct {
	baseURL string

	mu      sync.Mutex
	cache   []designEntry
	cacheAt time.Time
}

const designsCacheTTL = 2 * time.Minute
const designsRepo = "stiwi/aegis"
const designsPath = "docs/designs"

func newForgejoClient() *forgejoClient {
	return &forgejoClient{
		baseURL: "http://git.svc",
	}
}

type forgejoFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int    `json:"size"`
	Content string `json:"content"`
}

func (f *forgejoClient) listDesigns(ctx context.Context) ([]designEntry, error) {
	f.mu.Lock()
	if f.cache != nil && time.Since(f.cacheAt) < designsCacheTTL {
		result := f.cache
		f.mu.Unlock()
		return result, nil
	}
	f.mu.Unlock()

	url := fmt.Sprintf("%s/api/v1/repos/%s/contents/%s", f.baseURL, designsRepo, designsPath)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("forgejo API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("forgejo API: %s", resp.Status)
	}

	var files []forgejoFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	var designs []designEntry
	for _, file := range files {
		if !strings.HasSuffix(file.Name, ".md") {
			continue
		}
		name := strings.TrimSuffix(file.Name, ".md")
		title := strings.ReplaceAll(name, "-", " ")
		// Title case the first word
		if len(title) > 0 {
			title = strings.ToUpper(title[:1]) + title[1:]
		}
		designs = append(designs, designEntry{
			Name:  name,
			Title: title,
			Size:  file.Size,
			Path:  file.Path,
		})
	}

	sort.Slice(designs, func(i, j int) bool {
		return designs[i].Name < designs[j].Name
	})

	f.mu.Lock()
	f.cache = designs
	f.cacheAt = time.Now()
	f.mu.Unlock()

	return designs, nil
}

func (f *forgejoClient) getDesign(ctx context.Context, name string) (string, error) {
	filename := name + ".md"
	url := fmt.Sprintf("%s/api/v1/repos/%s/contents/%s/%s", f.baseURL, designsRepo, designsPath, filename)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("forgejo API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("design doc not found: %s", name)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("forgejo API: %s", resp.Status)
	}

	var file forgejoFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return "", err
	}

	content, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return "", fmt.Errorf("decode content: %w", err)
	}

	return string(content), nil
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

func renderMarkdown(source string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

func (s *Server) handleDesignsList(w http.ResponseWriter, r *http.Request) {
	data := designsListData{}

	if s.forgejo == nil {
		data.Err = "Forgejo client not configured"
		s.render(w, r, "designs", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	designs, err := s.forgejo.listDesigns(ctx)
	if err != nil {
		log.Printf("designs: list: %v", err)
		data.Err = fmt.Sprintf("Failed to list designs: %v", err)
		s.render(w, r, "designs", data)
		return
	}

	data.Designs = designs
	data.Total = len(designs)
	s.render(w, r, "designs", data)
}

func (s *Server) handleDesignView(w http.ResponseWriter, r *http.Request, name string) {
	// Sanitize: only allow alphanumeric, hyphens, underscores
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			http.NotFound(w, r)
			return
		}
	}

	data := designViewData{
		Name:   name,
		Title:  strings.ReplaceAll(name, "-", " "),
		GitURL: fmt.Sprintf("http://git.svc/%s/src/branch/main/%s/%s.md", designsRepo, designsPath, name),
	}

	if s.forgejo == nil {
		data.Err = "Forgejo client not configured"
		s.render(w, r, "design", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	content, err := s.forgejo.getDesign(ctx, name)
	if err != nil {
		log.Printf("designs: view %s: %v", name, err)
		if strings.Contains(err.Error(), "not found") {
			http.NotFound(w, r)
			return
		}
		data.Err = fmt.Sprintf("Failed to load design: %v", err)
		s.render(w, r, "design", data)
		return
	}

	data.Raw = content
	rendered, err := renderMarkdown(content)
	if err != nil {
		data.Err = fmt.Sprintf("Failed to render markdown: %v", err)
	} else {
		data.Content = rendered
	}

	// Extract title from first heading
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "# ") {
			data.Title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	s.render(w, r, "design", data)
}
