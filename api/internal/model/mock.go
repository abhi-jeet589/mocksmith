package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Mock struct {
	ID          bson.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	Method      string            `bson:"method"        json:"method"`
	Path        string            `bson:"path"          json:"path"`
	StatusCode  int               `bson:"statusCode"    json:"statusCode"`
	ContentType string            `bson:"contentType,omitempty" json:"contentType,omitempty"`
	Headers     map[string]string `bson:"headers,omitempty"     json:"headers,omitempty"`
	Body        string            `bson:"body"          json:"body"`
	CreatedAt   time.Time         `bson:"createdAt"     json:"createdAt"`
	UpdatedAt   time.Time         `bson:"updatedAt"     json:"updatedAt"`
}
