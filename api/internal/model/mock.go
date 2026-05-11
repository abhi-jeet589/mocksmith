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

	// IsPattern is set by the repository when Path contains "{name}" tokens.
	// It lets pattern lookup filter to only the relevant rows.
	IsPattern bool `bson:"isPattern" json:"isPattern,omitempty"`

	// PathParams maps each "{name}" in Path to a declared type
	// (e.g. "string", "int", "uuid"). Used to build the matching regex.
	PathParams map[string]string `bson:"pathParams,omitempty" json:"pathParams,omitempty"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}
