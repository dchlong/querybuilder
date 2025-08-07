package generation

import (
	"fmt"
	"strings"

	"github.com/dchlong/querybuilder/domain"
	"github.com/dchlong/querybuilder/repository"
)

// MethodFactory creates methods for querybuilder generation
type MethodFactory struct {
	operatorNames  map[repository.Operator]string
	methodSuffixes map[repository.Operator]string
}

// NewMethodFactory creates a new method factory
func NewMethodFactory() *MethodFactory {
	return &MethodFactory{
		operatorNames: map[repository.Operator]string{
			repository.OperatorEqual:              "OperatorEqual",
			repository.OperatorNotEqual:           "OperatorNotEqual",
			repository.OperatorLessThan:           "OperatorLessThan",
			repository.OperatorLessThanOrEqual:    "OperatorLessThanOrEqual",
			repository.OperatorGreaterThan:        "OperatorGreaterThan",
			repository.OperatorGreaterThanOrEqual: "OperatorGreaterThanOrEqual",
			repository.OperatorLike:               "OperatorLike",
			repository.OperatorNotLike:            "OperatorNotLike",
			repository.OperatorIsNull:             "OperatorIsNull",
			repository.OperatorIsNotNull:          "OperatorIsNotNull",
			repository.OperatorIn:                 "OperatorIn",
			repository.OperatorNotIn:              "OperatorNotIn",
		},
		methodSuffixes: map[repository.Operator]string{
			repository.OperatorEqual:              "Eq",
			repository.OperatorNotEqual:           "Ne",
			repository.OperatorLessThan:           "Lt",
			repository.OperatorLessThanOrEqual:    "Lte",
			repository.OperatorGreaterThan:        "Gt",
			repository.OperatorGreaterThanOrEqual: "Gte",
			repository.OperatorLike:               "Like",
			repository.OperatorNotLike:            "NotLike",
			repository.OperatorIsNull:             "IsNull",
			repository.OperatorIsNotNull:          "IsNotNull",
			repository.OperatorIn:                 "In",
			repository.OperatorNotIn:              "NotIn",
		},
	}
}

// CreateFilterMethod creates a filter method for a field and operator
func (f *MethodFactory) CreateFilterMethod(structName string, field domain.Field, op repository.Operator) domain.Method {
	methodName := field.Name + f.methodSuffixes[op]
	filterTypeName := structName + "Filters"
	receiverName := strings.ToLower(string(filterTypeName[0]))

	if f.isUnaryOperator(op) {
		return f.createUnaryFilterMethod(methodName, filterTypeName, receiverName, structName, field, op)
	}

	if f.isVariadicOperator(op) {
		return f.createVariadicFilterMethod(methodName, filterTypeName, receiverName, structName, field, op)
	}

	return f.createBinaryFilterMethod(methodName, filterTypeName, receiverName, structName, field, op)
}

// createBinaryFilterMethod creates a method that takes one parameter
func (f *MethodFactory) createBinaryFilterMethod(methodName, filterTypeName, receiverName, structName string, field domain.Field, op repository.Operator) domain.Method {
	paramName := f.fieldNameToParamName(field.Name)

	return domain.Method{
		Name:       methodName,
		Receiver:   fmt.Sprintf("%s *%s", receiverName, filterTypeName),
		Parameters: fmt.Sprintf("%s %s", paramName, field.TypeName),
		ReturnType: "*" + filterTypeName,
		Body: fmt.Sprintf(`%s.filters[%sDBSchema.%s] = append(%s.filters[%sDBSchema.%s], 
	&repository.Filter{
		Field:    string(%sDBSchema.%s),
		Operator: repository.%s,
		Value:    %s,
	})
return %s`,
			receiverName, structName, field.Name,
			receiverName, structName, field.Name,
			structName, field.Name,
			f.operatorNames[op], paramName, receiverName),
		Documentation: fmt.Sprintf("%s filters by %s %s", methodName, field.Name, strings.ToLower(f.methodSuffixes[op])),
	}
}

// createVariadicFilterMethod creates a method that takes variadic parameters (for IN/NOT IN)
func (f *MethodFactory) createVariadicFilterMethod(methodName, filterTypeName, receiverName, structName string, field domain.Field, op repository.Operator) domain.Method {
	paramName := f.fieldNameToParamName(field.Name) + "s"

	return domain.Method{
		Name:       methodName,
		Receiver:   fmt.Sprintf("%s *%s", receiverName, filterTypeName),
		Parameters: fmt.Sprintf("%s ...%s", paramName, field.TypeName),
		ReturnType: "*" + filterTypeName,
		Body: fmt.Sprintf(`%s.filters[%sDBSchema.%s] = append(%s.filters[%sDBSchema.%s], 
	&repository.Filter{
		Field:    string(%sDBSchema.%s),
		Operator: repository.%s,
		Value:    %s,
	})
return %s`,
			receiverName, structName, field.Name,
			receiverName, structName, field.Name,
			structName, field.Name,
			f.operatorNames[op], paramName, receiverName),
		Documentation: fmt.Sprintf("%s filters by %s in list", methodName, field.Name),
	}
}

// createUnaryFilterMethod creates a method that takes no parameters (for IS NULL/IS NOT NULL)
func (f *MethodFactory) createUnaryFilterMethod(methodName, filterTypeName, receiverName, structName string, field domain.Field, op repository.Operator) domain.Method {
	return domain.Method{
		Name:       methodName,
		Receiver:   fmt.Sprintf("%s *%s", receiverName, filterTypeName),
		Parameters: "",
		ReturnType: "*" + filterTypeName,
		Body: fmt.Sprintf(`%s.filters[%sDBSchema.%s] = append(%s.filters[%sDBSchema.%s], 
	&repository.Filter{
		Field:    string(%sDBSchema.%s),
		Operator: repository.%s,
		Value:    nil,
	})
return %s`,
			receiverName, structName, field.Name,
			receiverName, structName, field.Name,
			structName, field.Name,
			f.operatorNames[op], receiverName),
		Documentation: fmt.Sprintf("%s filters by %s is null check", methodName, field.Name),
	}
}

// CreateUpdaterMethod creates an updater setter method
func (f *MethodFactory) CreateUpdaterMethod(structName string, field domain.Field) domain.Method {
	methodName := "Set" + field.Name
	updaterTypeName := structName + "Updater"
	receiverName := strings.ToLower(string(updaterTypeName[0]))
	paramName := f.fieldNameToParamName(field.Name)

	return domain.Method{
		Name:       methodName,
		Receiver:   fmt.Sprintf("%s *%s", receiverName, updaterTypeName),
		Parameters: fmt.Sprintf("%s %s", paramName, field.TypeName),
		ReturnType: "*" + updaterTypeName,
		Body: fmt.Sprintf(`%s.fields[string(%sDBSchema.%s)] = %s
return %s`, receiverName, structName, field.Name, paramName, receiverName),
		Documentation: fmt.Sprintf("%s sets the %s field for update", methodName, field.Name),
	}
}

// CreateOrderMethod creates an ordering method
func (f *MethodFactory) CreateOrderMethod(structName string, field domain.Field, ascending bool) domain.Method {
	direction := "Desc"
	directionLower := "desc"
	if ascending {
		direction = "Asc"
		directionLower = "asc"
	}

	methodName := "OrderBy" + field.Name + direction
	optionsTypeName := structName + "Options"
	receiverName := strings.ToLower(string(optionsTypeName[0]))

	return domain.Method{
		Name:       methodName,
		Receiver:   fmt.Sprintf("%s *%s", receiverName, optionsTypeName),
		Parameters: "",
		ReturnType: "*" + optionsTypeName,
		Body: fmt.Sprintf(`%s.options = append(%s.options, func(options *repository.Options) {
	options.SortFields = append(options.SortFields, &repository.SortField{
		Field:     string(%sDBSchema.%s),
		Direction: "%s",
	})
})
return %s`, receiverName, receiverName, structName, field.Name, directionLower, receiverName),
		Documentation: fmt.Sprintf("%s orders results by %s %s", methodName, field.Name, directionLower),
	}
}

// Helper methods

func (f *MethodFactory) isUnaryOperator(op repository.Operator) bool {
	return op == repository.OperatorIsNull || op == repository.OperatorIsNotNull
}

func (f *MethodFactory) isVariadicOperator(op repository.Operator) bool {
	return op == repository.OperatorIn || op == repository.OperatorNotIn
}

func (f *MethodFactory) fieldNameToParamName(fieldName string) string {
	if len(fieldName) == 0 {
		return "value"
	}

	// Convert first character to lowercase
	runes := []rune(fieldName)
	runes[0] = runes[0] + ('a' - 'A')
	paramName := string(runes)

	// Check if it's a Go keyword and append "Value" if needed
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}

	if keywords[paramName] {
		return paramName + "Value"
	}

	return paramName
}
