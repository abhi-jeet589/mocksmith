package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/abhi-jeet589/mocksmith/internal/model"
	"github.com/abhi-jeet589/mocksmith/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Admin struct {
	Repo Store
}

func (a *Admin) Routes(r chi.Router) {
	r.Post("/", a.create)
	r.Get("/", a.list)
	r.Get("/{id}", a.get)
	r.Put("/{id}", a.update)
	r.Delete("/{id}", a.delete)
}

type mockInput struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	StatusCode  int               `json:"statusCode"`
	ContentType string            `json:"contentType,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body"`
}

// Paths reserved by mocksmith itself. Users cannot register mocks at these
// paths because the server's own routes (or the frontend's static files) would
// shadow them.
var reservedPathPrefixes = []string{"/admin/"}
var reservedExactPaths = []string{
	"/", "/admin", "/health",
	"/index.html", "/app.js", "/styles.css",
}

func (in mockInput) validate() error {
	if strings.TrimSpace(in.Method) == "" {
		return errors.New("method is required")
	}
	if strings.TrimSpace(in.Path) == "" {
		return errors.New("path is required")
	}
	if in.StatusCode < 100 || in.StatusCode > 599 {
		return errors.New("statusCode must be between 100 and 599")
	}
	normalized := normalizePath(in.Path)
	for _, p := range reservedExactPaths {
		if normalized == p {
			return errors.New("path " + normalized + " is reserved")
		}
	}
	for _, prefix := range reservedPathPrefixes {
		if strings.HasPrefix(normalized, prefix) {
			return errors.New("path " + normalized + " is reserved")
		}
	}
	return nil
}

func (in mockInput) toModel() *model.Mock {
	return &model.Mock{
		Method:      strings.ToUpper(strings.TrimSpace(in.Method)),
		Path:        normalizePath(in.Path),
		StatusCode:  in.StatusCode,
		ContentType: in.ContentType,
		Headers:     in.Headers,
		Body:        in.Body,
	}
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func (a *Admin) create(w http.ResponseWriter, r *http.Request) {
	var in mockInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := in.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := withTimeout(r.Context())
	defer cancel()

	m := in.toModel()
	if err := a.Repo.Create(ctx, m); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (a *Admin) list(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := withTimeout(r.Context())
	defer cancel()

	mocks, err := a.Repo.List(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, mocks)
}

func (a *Admin) get(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	ctx, cancel := withTimeout(r.Context())
	defer cancel()

	m, err := a.Repo.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "mock not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (a *Admin) update(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var in mockInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := in.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := withTimeout(r.Context())
	defer cancel()

	m := in.toModel()
	if err := a.Repo.Update(ctx, id, m); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			writeError(w, http.StatusNotFound, "mock not found")
		case errors.Is(err, repository.ErrConflict):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	m.ID = id
	writeJSON(w, http.StatusOK, m)
}

func (a *Admin) delete(w http.ResponseWriter, r *http.Request) {
	id, err := bson.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	ctx, cancel := withTimeout(r.Context())
	defer cancel()

	if err := a.Repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "mock not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
