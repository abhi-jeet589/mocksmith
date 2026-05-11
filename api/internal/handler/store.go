package handler

import (
	"context"

	"github.com/abhi-jeet589/mocksmith/internal/model"
	"github.com/abhi-jeet589/mocksmith/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Store is the data-access surface that the HTTP handlers depend on.
// Declared here (consumer-side) so handlers can be unit-tested with a fake.
type Store interface {
	Create(ctx context.Context, m *model.Mock) error
	List(ctx context.Context) ([]model.Mock, error)
	Get(ctx context.Context, id bson.ObjectID) (*model.Mock, error)
	Update(ctx context.Context, id bson.ObjectID, m *model.Mock) error
	Delete(ctx context.Context, id bson.ObjectID) error
	FindByRoute(ctx context.Context, method, path string) (*repository.MatchResult, error)
}
