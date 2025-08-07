package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GormRepository provides a complete GORM-based repository implementation
// that integrates seamlessly with the existing filter and updater system
type GormRepository[Entity any, Filter EntityFilter, Updater EntityUpdater] struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based repository
func NewGormRepository[Entity any, Filter EntityFilter, Updater EntityUpdater](
	db *gorm.DB,
) *GormRepository[Entity, Filter, Updater] {
	return &GormRepository[Entity, Filter, Updater]{
		db: db,
	}
}

// Create implements efficient record creation
func (r *GormRepository[Entity, Filter, Updater]) Create(ctx context.Context, records ...*Entity) error {
	if len(records) == 0 {
		return ErrNoRecordsProvided
	}

	err := r.db.WithContext(ctx).Create(records).Error
	if err != nil {
		return fmt.Errorf("create records: %w", err)
	}

	return nil
}

// FindOneByID implements single record lookup by ID
func (r *GormRepository[Entity, Filter, Updater]) FindOneByID(
	ctx context.Context,
	id int64,
) (*Entity, bool, error) {
	var result Entity
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("find record by ID %d: %w", id, err)
	}

	return &result, true, nil
}

// FindOne implements single record lookup with filters
func (r *GormRepository[Entity, Filter, Updater]) FindOne(
	ctx context.Context,
	filter Filter,
	options ...OptionFunc,
) (*Entity, bool, error) {
	var result Entity
	query, err := r.buildQuery(r.db.WithContext(ctx), filter)
	if err != nil {
		return nil, false, fmt.Errorf("FindOne build query: %w", err)
	}

	query = r.applyOptions(query, options...)

	err = query.Take(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("find one record: %w", err)
	}

	return &result, true, nil
}

// FindAll implements multiple record lookup with filters
func (r *GormRepository[Entity, Filter, Updater]) FindAll(
	ctx context.Context,
	filter Filter,
	options ...OptionFunc,
) ([]*Entity, error) {
	var result []*Entity
	query, err := r.buildQuery(r.db.WithContext(ctx), filter)
	if err != nil {
		return nil, fmt.Errorf("FindAll build query: %w", err)
	}

	query = r.applyOptions(query, options...)

	err = query.Find(&result).Error
	if err != nil {
		return nil, fmt.Errorf("find all records: %w", err)
	}

	return result, nil
}

// Update implements record updates using updaters
func (r *GormRepository[Entity, Filter, Updater]) Update(
	ctx context.Context,
	record *Entity,
	updater Updater,
) error {
	changeSet := updater.GetChangeSet()
	if len(changeSet) == 0 {
		return nil // No changes to apply
	}

	result := r.db.WithContext(ctx).Model(record).Updates(changeSet)
	if result.Error != nil {
		return fmt.Errorf("update record: %w", result.Error)
	}

	return nil
}

// WithTransaction executes a function within a database transaction
func (r *GormRepository[Entity, Filter, Updater]) WithTransaction(
	ctx context.Context,
	fn func(*GormRepository[Entity, Filter, Updater]) error,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &GormRepository[Entity, Filter, Updater]{
			db: tx,
		}
		return fn(txRepo)
	})
}

// CreateInBatches implements batch creation
func (r *GormRepository[Entity, Filter, Updater]) CreateInBatches(
	ctx context.Context,
	batchSize int,
	records ...*Entity,
) error {
	if len(records) == 0 {
		return ErrNoRecordsProvided
	}

	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	result := r.db.WithContext(ctx).CreateInBatches(records, batchSize)
	if result.Error != nil {
		return fmt.Errorf("create records in batches: %w", result.Error)
	}

	return nil
}

// UpdateWithFilter implements batch updates using filters
func (r *GormRepository[Entity, Filter, Updater]) UpdateWithFilter(
	ctx context.Context,
	filter Filter,
	updater Updater,
) (int64, error) {
	changeSet := updater.GetChangeSet()
	if len(changeSet) == 0 {
		return 0, nil
	}

	query, err := r.buildQuery(r.db.WithContext(ctx), filter)
	if err != nil {
		return 0, fmt.Errorf("UpdateWithFilter build query: %w", err)
	}

	result := query.Model(new(Entity)).Updates(changeSet)
	if result.Error != nil {
		return 0, fmt.Errorf("update records with filter: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// DeleteWithFilter implements batch deletion using filters
func (r *GormRepository[Entity, Filter, Updater]) DeleteWithFilter(
	ctx context.Context,
	filter Filter,
) (int64, error) {
	query, err := r.buildQuery(r.db.WithContext(ctx), filter)
	if err != nil {
		return 0, fmt.Errorf("DeleteWithFilter build query: %w", err)
	}

	result := query.Delete(new(Entity))
	if result.Error != nil {
		return 0, fmt.Errorf("delete records with filter: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// Count implements record counting
func (r *GormRepository[Entity, Filter, Updater]) Count(
	ctx context.Context,
	filter Filter,
) (int64, error) {
	query, err := r.buildQuery(r.db.WithContext(ctx), filter)
	if err != nil {
		return 0, fmt.Errorf("count build query: %w", err)
	}

	var count int64
	err = query.Model(new(Entity)).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count records: %w", err)
	}

	return count, nil
}

// Exists checks if any records match the filter efficiently
func (r *GormRepository[Entity, Filter, Updater]) Exists(
	ctx context.Context,
	filter Filter,
) (bool, error) {
	count, err := r.Count(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("exists check: %w", err)
	}
	return count > 0, nil
}

// applyOptions applies query options
func (r *GormRepository[Entity, Filter, Updater]) applyOptions(query *gorm.DB, options ...OptionFunc) *gorm.DB {
	opts := &Options{}
	for _, opt := range options {
		opt.Apply(opts)
	}

	if opts.Limit != nil {
		query = query.Limit(*opts.Limit)
	}

	if opts.Offset != nil {
		query = query.Offset(*opts.Offset)
	}

	for _, field := range opts.SortFields {
		quotedField := query.Statement.Quote(field.Field)
		query = query.Order(fmt.Sprintf("%s %s", quotedField, field.Direction))
	}

	return query
}

// buildQuery builds a GORM query from filters
func (r *GormRepository[Entity, Filter, Updater]) buildQuery(db *gorm.DB, filter Filter) (*gorm.DB, error) {
	for _, repositoryFilter := range filter.ListFilters() {
		if repositoryFilter.Field == "" {
			return nil, ErrEmptyFieldName
		}

		quotedField := db.Statement.Quote(repositoryFilter.Field)

		switch repositoryFilter.Operator {
		case OperatorEqual:
			db = db.Where(quotedField+" = ?", repositoryFilter.Value)
		case OperatorNotEqual:
			db = db.Where(quotedField+" != ?", repositoryFilter.Value)
		case OperatorLessThan:
			db = db.Where(quotedField+" < ?", repositoryFilter.Value)
		case OperatorLessThanOrEqual:
			db = db.Where(quotedField+" <= ?", repositoryFilter.Value)
		case OperatorGreaterThan:
			db = db.Where(quotedField+" > ?", repositoryFilter.Value)
		case OperatorGreaterThanOrEqual:
			db = db.Where(quotedField+" >= ?", repositoryFilter.Value)
		case OperatorLike:
			db = db.Where(quotedField+" LIKE ?", repositoryFilter.Value)
		case OperatorNotLike:
			db = db.Where(quotedField+" NOT LIKE ?", repositoryFilter.Value)
		case OperatorIsNull:
			db = db.Where(quotedField + " IS NULL")
		case OperatorIsNotNull:
			db = db.Where(quotedField + " IS NOT NULL")
		case OperatorIn:
			db = db.Where(quotedField+" IN (?)", repositoryFilter.Value)
		case OperatorNotIn:
			db = db.Where(quotedField+" NOT IN (?)", repositoryFilter.Value)
		default:
			return nil, fmt.Errorf("unknown operator %s: %w", repositoryFilter.Operator, ErrUnknownOperator)
		}
	}

	return db, nil
}

// GetDB returns the underlying GORM database instance for advanced operations
func (r *GormRepository[Entity, Filter, Updater]) GetDB() *gorm.DB {
	return r.db
}

// Health performs a health check on the database connection
func (r *GormRepository[Entity, Filter, Updater]) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
