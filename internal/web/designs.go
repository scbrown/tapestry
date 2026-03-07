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
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/scbrown/tapestry/internal/dolt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type designsListData struct {
	Designs []designEntry
	Total   int
	Filter  string
	Err     string
}

type designEntry struct {
	Name       string
	Title      string
	Size       int
	Path       string
	Modified   time.Time
	BeadID     string
	BeadStatus string
	Priority   int
}

type designViewData struct {
	Name     string
	Title    string
	Content  template.HTML
	Raw      string
	GitURL   string
	Err      string
	BeadID   string
	BeadDB   string
	Comments []dolt.Comment
	Feedback string
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
	Name           string `json:"name"`
	Path           string `json:"path"`
	Size           int    `json:"size"`
	Content        string `json:"content"`
	LastCommitWhen string `json:"last_commit_when"`
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
		modified, _ := time.Parse(time.RFC3339, file.LastCommitWhen)
		designs = append(designs, designEntry{
			Name:     name,
			Title:    title,
			Size:     file.Size,
			Path:     file.Path,
			Modified: modified,
		})
	}

	sort.Slice(designs, func(i, j int) bool {
		return designs[i].Modified.After(designs[j].Modified)
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
	data.Filter = r.URL.Query().Get("filter")

	if s.forgejo == nil {
		data.Err = "Forgejo client not configured"
		s.render(w, r, "designs", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	designs, err := s.forgejo.listDesigns(ctx)
	if err != nil {
		log.Printf("designs: list: %v", err)
		data.Err = fmt.Sprintf("Failed to list designs: %v", err)
		s.render(w, r, "designs", data)
		return
	}

	// Enrich designs with bead metadata
	if s.ds != nil {
		s.enrichDesignsWithBeads(ctx, designs)
	}

	// Apply filter
	if data.Filter != "" {
		var filtered []designEntry
		for _, d := range designs {
			switch data.Filter {
			case "review":
				if d.BeadStatus == "open" || d.BeadStatus == "" {
					filtered = append(filtered, d)
				}
			case "progress":
				if d.BeadStatus == "in_progress" {
					filtered = append(filtered, d)
				}
			case "done":
				if d.BeadStatus == "closed" {
					filtered = append(filtered, d)
				}
			}
		}
		designs = filtered
	}

	data.Designs = designs
	data.Total = len(designs)
	s.render(w, r, "designs", data)
}

func (s *Server) enrichDesignsWithBeads(ctx context.Context, designs []designEntry) {
	dbs, err := s.databases(ctx)
	if err != nil {
		return
	}

	// Build a map of all beads across DBs for matching
	type beadInfo struct {
		ID       string
		Status   string
		Priority int
	}
	allBeads := map[string]beadInfo{} // lowercase title fragment → bead info

	for _, db := range dbs {
		issues, err := s.ds.Issues(ctx, db.Name, dolt.IssueFilter{Limit: 5000})
		if err != nil {
			continue
		}
		for _, iss := range issues {
			titleLower := strings.ToLower(iss.Title)
			allBeads[titleLower] = beadInfo{ID: iss.ID, Status: iss.Status, Priority: iss.Priority}
		}
	}

	// Match each design to a bead by checking if design name appears in bead title
	for i := range designs {
		nameLower := strings.ToLower(designs[i].Name)
		for title, info := range allBeads {
			if strings.Contains(title, nameLower) || strings.Contains(title, strings.ReplaceAll(nameLower, "-", " ")) {
				designs[i].BeadID = info.ID
				designs[i].BeadStatus = info.Status
				designs[i].Priority = info.Priority
				break
			}
		}
	}
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

	switch r.URL.Query().Get("feedback") {
	case "ok":
		data.Feedback = "Comment added."
	case "approved":
		data.Feedback = "Design approved — GO."
	case "missing":
		data.Feedback = "All fields are required."
	case "error":
		data.Feedback = "Failed to save comment."
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

	// Parse bead link from markdown: <!-- bead: aegis-xxxx --> or <!-- bead: aegis/aegis-xxxx -->
	if beadID, beadDB := parseBeadLink(content); beadID != "" {
		data.BeadID = beadID
		data.BeadDB = beadDB
	}

	// Fallback: search for a bead whose title matches the design doc name
	if data.BeadID == "" && s.ds != nil {
		dbs, _ := s.databases(ctx)
		for _, db := range dbs {
			results, err := s.ds.SearchIssues(ctx, db.Name, name, 1)
			if err == nil && len(results) > 0 {
				data.BeadID = results[0].ID
				data.BeadDB = db.Name
				break
			}
		}
	}

	// Load comments for linked bead
	if data.BeadID != "" && s.ds != nil {
		comments, err := s.ds.Comments(ctx, data.BeadDB, data.BeadID)
		if err != nil {
			log.Printf("designs: comments for %s: %v", data.BeadID, err)
		} else {
			data.Comments = comments
		}
	}

	s.render(w, r, "design", data)
}

var mentionRe = regexp.MustCompile(`@([a-zA-Z][a-zA-Z0-9_-]*)`)

func notifyDesignFeedback(design, beadID, author, body string) {
	mentions := parseMentions(body)

	msg := fmt.Sprintf("[%s] %s on /designs/%s:\n%s", beadID, author, design, body)
	if len(msg) > 500 {
		msg = msg[:500]
	}

	title := fmt.Sprintf("Design feedback: %s", design)
	if len(mentions) > 0 {
		atMentions := make([]string, len(mentions))
		for i, m := range mentions {
			atMentions[i] = "@" + m
		}
		title = fmt.Sprintf("Design feedback: %s (cc %s)", design, strings.Join(atMentions, " "))
	}

	// Always notify the design-feedback topic (routing agent watches this)
	sendNtfy("design-feedback", title, msg, "memo,tapestry", "3")

	// For each @mention, send a targeted notification
	for _, m := range mentions {
		mentionMsg := fmt.Sprintf("%s mentioned you in feedback on /designs/%s [%s]:\n%s", author, design, beadID, body)
		if len(mentionMsg) > 500 {
			mentionMsg = mentionMsg[:500]
		}
		sendNtfy("crew-"+m, fmt.Sprintf("Design feedback mention: %s", design), mentionMsg, "speech_balloon,tapestry", "4")
	}
}

func parseMentions(body string) []string {
	matches := mentionRe.FindAllStringSubmatch(body, -1)
	seen := map[string]bool{}
	var result []string
	for _, m := range matches {
		name := strings.ToLower(m[1])
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	return result
}

func sendNtfy(topic, title, msg, tags, priority string) {
	req, err := http.NewRequest("POST", "http://ntfy.svc/"+topic, strings.NewReader(msg))
	if err != nil {
		log.Printf("designs: ntfy request: %v", err)
		return
	}
	req.Header.Set("Title", title)
	req.Header.Set("Tags", tags)
	req.Header.Set("Priority", priority)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("designs: ntfy send to %s: %v", topic, err)
		return
	}
	resp.Body.Close()
}

var beadLinkRe = regexp.MustCompile(`<!--\s*bead:\s*(?:(\w+)/)?(\S+)\s*-->`)

func parseBeadLink(content string) (beadID, database string) {
	m := beadLinkRe.FindStringSubmatch(content)
	if m == nil {
		return "", ""
	}
	database = "aegis"
	if m[1] != "" {
		database = m[1]
	}
	return m[2], database
}

func (s *Server) handleDesignComment(w http.ResponseWriter, r *http.Request, name string) {
	// Sanitize name
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			http.NotFound(w, r)
			return
		}
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	beadID := strings.TrimSpace(r.FormValue("bead_id"))
	beadDB := strings.TrimSpace(r.FormValue("bead_db"))
	author := strings.TrimSpace(r.FormValue("author"))
	body := strings.TrimSpace(r.FormValue("body"))

	if beadID == "" || beadDB == "" || author == "" || body == "" {
		http.Redirect(w, r, "/designs/"+name+"?feedback=missing", http.StatusSeeOther)
		return
	}

	// Sanitize author: alphanumeric, hyphens, underscores, dots only
	for _, c := range author {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			http.Redirect(w, r, "/designs/"+name+"?feedback=invalid", http.StatusSeeOther)
			return
		}
	}

	if s.ds == nil {
		http.Redirect(w, r, "/designs/"+name+"?feedback=nodb", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.ds.AddComment(ctx, beadDB, beadID, author, body); err != nil {
		log.Printf("designs: add comment on %s: %v", beadID, err)
		http.Redirect(w, r, "/designs/"+name+"?feedback=error", http.StatusSeeOther)
		return
	}

	// Notify via ntfy so feedback is visible immediately
	go notifyDesignFeedback(name, beadID, author, body)

	http.Redirect(w, r, "/designs/"+name+"?feedback=ok", http.StatusSeeOther)
}

// forgejoCommit represents a commit returned by the Forgejo API.
type forgejoCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	HTMLURL string `json:"html_url"`
}

// beadCommit is a simplified commit for display on bead pages.
type beadCommit struct {
	SHA       string
	ShortSHA  string
	Subject   string
	Author    string
	Timestamp time.Time
	CommitURL string
	RepoName  string
}

// searchRepos is the list of repos to search for bead-linked commits.
var searchRepos = []string{
	"stiwi/aegis",
	"stiwi/gastown",
	"stiwi/beads",
	"stiwi/bobbin",
	"stiwi/tapestry",
}

// searchCommitsForBead searches Forgejo repos for commits mentioning a bead ID.
func (f *forgejoClient) searchCommitsForBead(ctx context.Context, beadID string) []beadCommit {
	var results []beadCommit
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, repo := range searchRepos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			url := fmt.Sprintf("%s/api/v1/repos/%s/git/commits?sha=main&keyword=%s&limit=20",
				f.baseURL, repo, beadID)
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return
			}
			var commits []forgejoCommit
			if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
				return
			}
			// Extract repo short name
			parts := strings.Split(repo, "/")
			repoName := parts[len(parts)-1]

			mu.Lock()
			defer mu.Unlock()
			for _, c := range commits {
				// Verify bead ID actually appears in message (keyword search may be fuzzy)
				if !strings.Contains(c.Commit.Message, beadID) {
					continue
				}
				subject := c.Commit.Message
				if idx := strings.IndexByte(subject, '\n'); idx > 0 {
					subject = subject[:idx]
				}
				shortSHA := c.SHA
				if len(shortSHA) > 7 {
					shortSHA = shortSHA[:7]
				}
				ts, _ := time.Parse(time.RFC3339, c.Commit.Author.Date)
				results = append(results, beadCommit{
					SHA:       c.SHA,
					ShortSHA:  shortSHA,
					Subject:   subject,
					Author:    c.Commit.Author.Name,
					Timestamp: ts,
					CommitURL: c.HTMLURL,
					RepoName:  repoName,
				})
			}
		}(repo)
	}
	wg.Wait()

	// Sort by timestamp descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})
	return results
}

func (s *Server) handleDesignApprove(w http.ResponseWriter, r *http.Request, name string) {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			http.NotFound(w, r)
			return
		}
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	beadID := strings.TrimSpace(r.FormValue("bead_id"))
	beadDB := strings.TrimSpace(r.FormValue("bead_db"))

	if beadID == "" || beadDB == "" || s.ds == nil {
		http.Redirect(w, r, "/designs/"+name+"?feedback=error", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Update status to closed (approved)
	if err := s.ds.UpdateStatus(ctx, beadDB, beadID, "closed"); err != nil {
		log.Printf("designs: approve %s: update status: %v", beadID, err)
		http.Redirect(w, r, "/designs/"+name+"?feedback=error", http.StatusSeeOther)
		return
	}

	// Add approved label
	if err := s.ds.AddLabel(ctx, beadDB, beadID, "approved"); err != nil {
		log.Printf("designs: approve %s: add label: %v", beadID, err)
	}

	// Add approval comment
	if err := s.ds.AddComment(ctx, beadDB, beadID, "stiwi", "Approved — GO"); err != nil {
		log.Printf("designs: approve %s: add comment: %v", beadID, err)
	}

	// Notify
	go sendNtfy("design-feedback",
		fmt.Sprintf("APPROVED: %s", name),
		fmt.Sprintf("Design %s (%s) approved by Stiwi — GO. Begin implementation.", name, beadID),
		"white_check_mark,tapestry", "4")

	http.Redirect(w, r, "/designs/"+name+"?feedback=approved", http.StatusSeeOther)
}
