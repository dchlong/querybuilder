package repository

import (
	"context"

	"gorm.io/gorm"
)

type Operator string

// Enum values for Operator
const (
	OperatorEqual              Operator = "="
	OperatorNotEqual           Operator = "!="
	OperatorLessThan           Operator = "<"
	OperatorLessThanOrEqual    Operator = "<="
	OperatorGreaterThan        Operator = ">"
	OperatorGreaterThanOrEqual Operator = ">="
	OperatorLike               Operator = "LIKE"
	OperatorNotLike            Operator = "NOT_LIKE"
	OperatorIsNull             Operator = "IS_NULL"
	OperatorIsNotNull          Operator = "IS_NOT_NULL"
	OperatorIn                 Operator = "IN"
	OperatorNotIn              Operator = "NOT_IN"
)

type Filter struct {
	Field    string
	Operator Operator
	Value    interface{}
}

type SortField struct {
	Field     string
	Direction string
}

type OptionFunc interface {
	Apply(*Options)
}

type functionOption struct {
	f func(*Options)
}

func (fo *functionOption) Apply(opts *Options) {
	fo.f(opts)
}

type Options struct {
	Limit      *int
	Offset     *int
	SortFields []*SortField
}

func WithLimit(limit int) OptionFunc {
	return &functionOption{
		f: func(o *Options) {
			o.Limit = &limit
		},
	}
}

func WithOffset(offset int) OptionFunc {
	return &functionOption{
		f: func(o *Options) {
			o.Offset = &offset
		},
	}
}

type EntityFilter interface {
	ListFilters() []*Filter
}

type EntityUpdater interface {
	GetChangeSet() map[string]interface{}
}

// Repository defines the contract for a generic repository
// that provides CRUD operations and query capabilities
type Repository[Entity any, Filter EntityFilter, Updater EntityUpdater] interface {
	// Create creates one or more records
	Create(ctx context.Context, records ...*Entity) error

	// FindOneByID finds a single record by its ID
	FindOneByID(ctx context.Context, id int64) (*Entity, bool, error)

	// FindOne finds a single record matching the filter
	FindOne(ctx context.Context, filter Filter, options ...OptionFunc) (*Entity, bool, error)

	// FindAll finds all records matching the filter
	FindAll(ctx context.Context, filter Filter, options ...OptionFunc) ([]*Entity, error)

	// Update updates a single record using the updater
	Update(ctx context.Context, record *Entity, updater Updater) error

	// Delete deletes a single record
	Delete(ctx context.Context, record *Entity) error

	// DeleteByID deletes a single record by its ID
	DeleteByID(ctx context.Context, id int64) error

	// CreateInBatches creates records in batches
	CreateInBatches(ctx context.Context, batchSize int, records ...*Entity) error

	// UpdateWithFilter updates all records matching the filter
	UpdateWithFilter(ctx context.Context, filter Filter, updater Updater) (int64, error)

	// DeleteWithFilter deletes all records matching the filter
	DeleteWithFilter(ctx context.Context, filter Filter) (int64, error)

	// Count counts records matching the filter
	Count(ctx context.Context, filter Filter) (int64, error)

	// Exists checks if any records match the filter
	Exists(ctx context.Context, filter Filter) (bool, error)

	// GetDB returns the underlying GORM database instance
	GetDB() *gorm.DB

	// Health performs a health check on the database connection
	Health(ctx context.Context) error
}
