package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"uuid"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{clock: clock, tasks: make(map[string]Task)}
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrInvalidTitle
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New().String()
	t := Task{
		ID:        id,
		Title:     title,
		UpdatedAt: r.clock.Now(),
	}

	r.tasks[id] = t

	return t, nil
}
func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[id]
	return t, ok
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]Task, 0)

	for _, t := range r.tasks {
		if t.Done == done {
			res = append(res, t)
		}
	}
	slices.SortFunc(res, func(t1 Task, t2 Task) int {
		cmp := t2.UpdatedAt.Compare(t1.UpdatedAt)
		if cmp == 0 {
			return strings.Compare(t1.ID, t2.ID)
		}
		return cmp
	})
	return res
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if t.Done != done {
		t.Done = done
		t.UpdatedAt = r.clock.Now()
	}
	r.tasks[id] = t

	return t, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{repo: repo}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tasks" {
		switch r.Method {
		case "POST":
			h.handleCreate(w, r)
		case "GET":
			h.handleList(w, r)
		default:
			w.WriteHeader(405)
		}
		return
	}

	if id, ok := strings.CutPrefix(r.URL.Path, "/tasks/"); ok {
		if id == "" || strings.Contains(id, "/") {
			w.WriteHeader(404)
			return
		}

		switch r.Method {
		case "GET":
			h.handleGet(w, r, id)
		case "PATCH":
			h.handlePatch(w, r, id)
		default:
			w.WriteHeader(405)
		}
		return
	}
	w.WriteHeader(404)
}

func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var input struct {
		Title *string `json:"title"`
	}
	err := decodeStrictJSON(r.Body, &input)
	if err != nil || input.Title == nil {
		w.WriteHeader(400)
		return
	}

	t, err := h.repo.Create(*input.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	writeJSON(w, 201, t)
}

func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	t, ok := h.repo.Get(id)
	if !ok {
		w.WriteHeader(404)
		return
	}
	writeJSON(w, 200, t)
}

func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	var done bool

	if vals, has := queryParams["done"]; has {
		if len(vals) != 1 || (vals[0] != "true" && vals[0] != "false") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		done, _ = strconv.ParseBool(vals[0])
	}
	res := h.repo.List(done)
	writeJSON(w, 200, res)
}

func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()

	var input struct {
		Done *bool `json:"done"`
	}

	err := decodeStrictJSON(r.Body, &input)
	if err != nil || input.Done == nil {
		w.WriteHeader(400)
		return
	}

	t, err := h.repo.SetDone(id, *input.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	writeJSON(w, 200, t)
}

func decodeStrictJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)

	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	if string(raw) == "null" || string(raw) == "{}" {
		return errors.New("Body cannot be null")
	}

	var garbage any
	if err := decoder.Decode(&garbage); !errors.Is(err, io.EOF) {
		return errors.New("Trailing data in request body")
	}

	strictDecoder := json.NewDecoder(bytes.NewReader(raw))
	strictDecoder.DisallowUnknownFields()
	if err := strictDecoder.Decode(v); err != nil {
		return err
	}

	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return
	}
}
