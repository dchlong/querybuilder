package repository

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
