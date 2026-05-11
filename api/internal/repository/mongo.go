package repository

import (
	"context"
	"errors"
	"time"

	"github.com/abhi-jeet589/mocksmith/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrNotFound = errors.New("mock not found")
	ErrConflict = errors.New("mock with this method and path already exists")
)

type Repository struct {
	client *mongo.Client
	mocks  *mongo.Collection
}

func New(ctx context.Context, uri, db string) (*Repository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	coll := client.Database(db).Collection("mocks")

	idxCtx, cancelIdx := context.WithTimeout(ctx, 10*time.Second)
	defer cancelIdx()
	_, err = coll.Indexes().CreateOne(idxCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "method", Value: 1}, {Key: "path", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return &Repository{client: client, mocks: coll}, nil
}

func (r *Repository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

func (r *Repository) Create(ctx context.Context, m *model.Mock) error {
	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now
	m.IsPattern = HasPatternSyntax(m.Path)

	res, err := r.mocks.InsertOne(ctx, m)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrConflict
		}
		return err
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		m.ID = oid
	}
	return nil
}

func (r *Repository) List(ctx context.Context) ([]model.Mock, error) {
	cur, err := r.mocks.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := []model.Mock{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) Get(ctx context.Context, id bson.ObjectID) (*model.Mock, error) {
	var m model.Mock
	err := r.mocks.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) FindByRoute(ctx context.Context, method, path string) (*model.Mock, error) {
	// 1. Exact match (literal path) — fast path, uses the unique index.
	var m model.Mock
	err := r.mocks.FindOne(ctx, bson.D{
		{Key: "method", Value: method},
		{Key: "path", Value: path},
	}).Decode(&m)
	if err == nil {
		return &m, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	// 2. Pattern fallback — iterate mocks flagged as patterns for this method
	// and return the first whose compiled regex matches.
	cur, err := r.mocks.Find(ctx, bson.D{
		{Key: "method", Value: method},
		{Key: "isPattern", Value: true},
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var candidate model.Mock
		if err := cur.Decode(&candidate); err != nil {
			continue
		}
		re := PatternToRegex(candidate.Path, candidate.PathParams)
		if re == nil {
			continue
		}
		if re.MatchString(path) {
			return &candidate, nil
		}
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return nil, ErrNotFound
}

func (r *Repository) Update(ctx context.Context, id bson.ObjectID, m *model.Mock) error {
	m.UpdatedAt = time.Now().UTC()
	m.IsPattern = HasPatternSyntax(m.Path)
	res, err := r.mocks.UpdateOne(ctx,
		bson.D{{Key: "_id", Value: id}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "method", Value: m.Method},
			{Key: "path", Value: m.Path},
			{Key: "statusCode", Value: m.StatusCode},
			{Key: "contentType", Value: m.ContentType},
			{Key: "headers", Value: m.Headers},
			{Key: "body", Value: m.Body},
			{Key: "isPattern", Value: m.IsPattern},
			{Key: "pathParams", Value: m.PathParams},
			{Key: "updatedAt", Value: m.UpdatedAt},
		}}},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrConflict
		}
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id bson.ObjectID) error {
	res, err := r.mocks.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
