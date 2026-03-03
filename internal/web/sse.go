package web

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ReactorEvent represents a classified event from the reactor SSE stream.
type ReactorEvent struct {
	ID          string          `json:"id"`
	EventType   string          `json:"event_type"`
	SourceDB    string          `json:"source_db"`
	SourceTable string          `json:"source_table"`
	SubjectID   string          `json:"subject_id"`
	Actor       string          `json:"actor"`
	Summary     string          `json:"summary"`
	Timestamp   string          `json:"timestamp"`
	Payload     json.RawMessage `json:"payload"`
}

// EventHub subscribes to the reactor SSE stream and fans out events to
// browser clients via Server-Sent Events.
type EventHub struct {
	reactorURL string

	mu      sync.RWMutex
	clients map[chan ReactorEvent]struct{}
	recent  []ReactorEvent // last N events for replay to new clients
}

// NewEventHub creates a hub that connects to the reactor SSE endpoint.
func NewEventHub(reactorURL string) *EventHub {
	return &EventHub{
		reactorURL: reactorURL,
		clients:    make(map[chan ReactorEvent]struct{}),
	}
}

// Start begins the background reactor SSE subscription loop.
func (h *EventHub) Start(ctx context.Context) {
	go h.connectLoop(ctx)
}

func (h *EventHub) connectLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := h.connect(ctx); err != nil {
			log.Printf("sse: reactor connection error: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (h *EventHub) connect(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", h.reactorURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 0} // no timeout for SSE
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reactor returned %d", resp.StatusCode)
	}

	log.Printf("sse: connected to reactor at %s", h.reactorURL)

	scanner := bufio.NewScanner(resp.Body)
	var eventType, data, id string

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "event: "):
			eventType = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimPrefix(line, "data: ")
		case strings.HasPrefix(line, "id: "):
			id = strings.TrimPrefix(line, "id: ")
		case line == "":
			if eventType == "reactor" && data != "" {
				var evt ReactorEvent
				if err := json.Unmarshal([]byte(data), &evt); err == nil {
					if evt.ID == "" {
						evt.ID = id
					}
					h.broadcast(evt)
				}
			}
			eventType, data, id = "", "", ""
		}
	}

	return scanner.Err()
}

func (h *EventHub) broadcast(evt ReactorEvent) {
	h.mu.Lock()
	h.recent = append(h.recent, evt)
	if len(h.recent) > 50 {
		h.recent = h.recent[len(h.recent)-50:]
	}
	clients := make([]chan ReactorEvent, 0, len(h.clients))
	for ch := range h.clients {
		clients = append(clients, ch)
	}
	h.mu.Unlock()

	for _, ch := range clients {
		select {
		case ch <- evt:
		default: // drop for slow clients
		}
	}
}

// Subscribe returns a channel of events and the recent event buffer.
func (h *EventHub) Subscribe() (chan ReactorEvent, []ReactorEvent) {
	ch := make(chan ReactorEvent, 50)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	recent := make([]ReactorEvent, len(h.recent))
	copy(recent, h.recent)
	h.mu.Unlock()
	return ch, recent
}

// Unsubscribe removes a client channel.
func (h *EventHub) Unsubscribe(ch chan ReactorEvent) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

// handleEventStream serves SSE to browser clients.
func (s *Server) handleEventStream(w http.ResponseWriter, r *http.Request) {
	if s.hub == nil {
		http.Error(w, "SSE not configured (no reactor URL)", http.StatusServiceUnavailable)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	typeFilter := r.URL.Query().Get("type")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, recent := s.hub.Subscribe()
	defer s.hub.Unsubscribe(ch)

	// Replay recent events
	for _, evt := range recent {
		if typeFilter != "" && !strings.HasPrefix(evt.EventType, typeFilter) {
			continue
		}
		writeSSE(w, flusher, evt)
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case evt := <-ch:
			if typeFilter != "" && !strings.HasPrefix(evt.EventType, typeFilter) {
				continue
			}
			writeSSE(w, flusher, evt)
		}
	}
}

func writeSSE(w http.ResponseWriter, f http.Flusher, evt ReactorEvent) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "id: %s\nevent: reactor\ndata: %s\n\n", evt.ID, data)
	f.Flush()
}
