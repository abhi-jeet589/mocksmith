package handler

import (
	"errors"
	"net/http"

	"github.com/abhi-jeet589/mocksmith/internal/repository"
)

type Mock struct {
	Repo Store
}

func (m *Mock) Serve(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := withTimeout(r.Context())
	defer cancel()

	result, err := m.Repo.FindByRoute(ctx, r.Method, r.URL.Path)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "no mock registered for "+r.Method+" "+r.URL.Path)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mock := result.Mock
	params := result.Params // nil for exact-match mocks

	for k, v := range mock.Headers {
		w.Header().Set(k, repository.ApplyTemplate(v, params))
	}
	if mock.ContentType != "" {
		w.Header().Set("Content-Type", mock.ContentType)
	}

	status := mock.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write([]byte(repository.ApplyTemplate(mock.Body, params)))
}
