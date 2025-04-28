package anytype

import (
	"context"
)

// PropertyClient provides operations on properties of an object
type PropertyClient interface {
	// Get retrieves a specific property value
	Get(ctx context.Context, key string) (*Property, error)

	// Set sets a property value
	Set(ctx context.Context, key string, value interface{}) error

	// Delete removes a property
	Delete(ctx context.Context, key string) error

	// List returns all properties of the object
	List(ctx context.Context) ([]Property, error)
}

// Property represents an object property
type Property struct {
	ID          string
	Format      string
	Text        string
	Number      float64
	Checkbox    bool
	Date        string
	URL         string
	Email       string
	Phone       string
	File        []string
	Select      *Tag     `json:"select"`
	MultiSelect []Tag    `json:"multi_select"`
	ObjectLinks []string `json:"object"`
	Relations   []Relation
	Key         string
	Name        string
	Required    bool
}

// Tag represents a select or multi-select value option
type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Relation represents a relation within a property
type Relation struct {
	ID       string
	Type     string
	Format   string
	ObjectID string `json:"object_id"`
}
