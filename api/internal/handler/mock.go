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

	mock, err := m.Repo.FindByRoute(ctx, r.Method, r.URL.Path)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "no mock registered for "+r.Method+" "+r.URL.Path)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for k, v := range mock.Headers {
		w.Header().Set(k, v)
	}
	if mock.ContentType != "" {
		w.Header().Set("Content-Type", mock.ContentType)
	}

	status := mock.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write([]byte(mock.Body))
}
