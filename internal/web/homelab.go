package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type homelabData struct {
	GeneratedAt time.Time
	Targets     []promTarget
	Alerts      []promAlert
	UpCount     int
	DownCount   int
	TotalCount  int
	Silenced    int
	Firing      int
	Err         string
}

type promTarget struct {
	Name     string
	Job      string
	Instance string
	Up       bool
}

type promAlert struct {
	Name        string
	Severity    string
	State       string
	Summary     string
	Description string
	Instance    string
	Silenced    bool
	ActiveAt    time.Time
}

type promClient struct {
	baseURL  string
	user     string
	passFile string

	mu       sync.Mutex
	password string
	passRead bool
}

func newPromClient() *promClient {
	return &promClient{
		baseURL:  "http://monitoring.lan:9090",
		user:     "aegis",
		passFile: "/etc/tapestry/.prometheus_password",
	}
}

func (p *promClient) getPassword() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.passRead {
		return p.password, nil
	}
	data, err := os.ReadFile(p.passFile)
	if err != nil {
		return "", fmt.Errorf("read prometheus password: %w", err)
	}
	p.password = strings.TrimSpace(string(data))
	p.passRead = true
	return p.password, nil
}

func (p *promClient) query(ctx context.Context, path string) ([]byte, error) {
	pass, err := p.getPassword()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.user, pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("prometheus: %s", resp.Status)
	}

	var buf [512 * 1024]byte
	n := 0
	for n < len(buf) {
		nn, err := resp.Body.Read(buf[n:])
		n += nn
		if err != nil {
			break
		}
	}
	return buf[:n], nil
}

func (p *promClient) queryTargets(ctx context.Context) ([]promTarget, error) {
	data, err := p.query(ctx, "/api/v1/query?query=up")
	if err != nil {
		return nil, err
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  [2]interface{}    `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	var targets []promTarget
	for _, r := range result.Data.Result {
		val := "0"
		if len(r.Value) > 1 {
			if s, ok := r.Value[1].(string); ok {
				val = s
			}
		}
		name := r.Metric["hostname"]
		if name == "" {
			name = r.Metric["instance"]
		}
		if svc := r.Metric["service"]; svc != "" {
			name = svc
		}
		targets = append(targets, promTarget{
			Name:     name,
			Job:      r.Metric["job"],
			Instance: r.Metric["instance"],
			Up:       val == "1",
		})
	}

	sort.Slice(targets, func(i, j int) bool {
		if targets[i].Up != targets[j].Up {
			return !targets[i].Up // down first
		}
		return targets[i].Name < targets[j].Name
	})

	return targets, nil
}

func (p *promClient) queryAlerts(ctx context.Context) ([]promAlert, error) {
	data, err := p.query(ctx, "/api/v1/alerts")
	if err != nil {
		return nil, err
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Alerts []struct {
				State       string            `json:"state"`
				Labels      map[string]string  `json:"labels"`
				Annotations map[string]string  `json:"annotations"`
				ActiveAt    time.Time          `json:"activeAt"`
			} `json:"alerts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// Also check alertmanager for silenced status
	silenced := make(map[string]bool)
	amData, err := p.queryAlertmanager(ctx)
	if err == nil {
		for _, a := range amData {
			if len(a.Status.SilencedBy) > 0 {
				key := a.Labels["alertname"] + "/" + a.Labels["instance"]
				silenced[key] = true
			}
		}
	}

	var alerts []promAlert
	for _, a := range result.Data.Alerts {
		name := a.Labels["alertname"]
		key := name + "/" + a.Labels["instance"]
		alerts = append(alerts, promAlert{
			Name:        name,
			Severity:    a.Labels["severity"],
			State:       a.State,
			Summary:     a.Annotations["summary"],
			Description: a.Annotations["description"],
			Instance:    a.Labels["hostname"],
			Silenced:    silenced[key],
			ActiveAt:    a.ActiveAt,
		})
	}

	sort.Slice(alerts, func(i, j int) bool {
		return sevRank(alerts[i].Severity) < sevRank(alerts[j].Severity)
	})

	return alerts, nil
}

type amAlert struct {
	Labels map[string]string `json:"labels"`
	Status struct {
		SilencedBy []string `json:"silencedBy"`
	} `json:"status"`
}

func (p *promClient) queryAlertmanager(ctx context.Context) ([]amAlert, error) {
	pass, err := p.getPassword()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "http://monitoring.lan:9093/api/v2/alerts", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.user, pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("alertmanager: %s", resp.Status)
	}

	var alerts []amAlert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}

func sevRank(s string) int {
	switch s {
	case "critical":
		return 0
	case "warning":
		return 1
	case "review":
		return 2
	case "info":
		return 3
	default:
		return 4
	}
}

func (s *Server) handleHomelab(w http.ResponseWriter, r *http.Request) {
	data := homelabData{GeneratedAt: time.Now()}

	if s.prom == nil {
		data.Err = "Prometheus client not configured"
		s.render(w, r, "homelab", data)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	targets, err := s.prom.queryTargets(ctx)
	if err != nil {
		log.Printf("homelab: targets: %v", err)
		data.Err = fmt.Sprintf("Failed to query Prometheus: %v", err)
		s.render(w, r, "homelab", data)
		return
	}
	data.Targets = targets
	data.TotalCount = len(targets)
	for _, t := range targets {
		if t.Up {
			data.UpCount++
		} else {
			data.DownCount++
		}
	}

	alerts, err := s.prom.queryAlerts(ctx)
	if err != nil {
		log.Printf("homelab: alerts: %v", err)
	} else {
		data.Alerts = alerts
		for _, a := range alerts {
			if a.Silenced {
				data.Silenced++
			} else if a.State == "firing" {
				data.Firing++
			}
		}
	}

	s.render(w, r, "homelab", data)
}
