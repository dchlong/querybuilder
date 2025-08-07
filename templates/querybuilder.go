package templates

import (
	"text/template"
)

// QueryBuilderTemplates contains all code generation templates
type QueryBuilderTemplates struct {
	Main *template.Template
}

// NewQueryBuilderTemplates creates a new template set
func NewQueryBuilderTemplates() *QueryBuilderTemplates {
	main := template.Must(template.New("querybuilder").Parse(mainTemplate))

	return &QueryBuilderTemplates{
		Main: main,
	}
}

const mainTemplate = `
{{- range .Structs }}
{{- $structName := .Name }}
{{- $filterTypeName := printf "%sFilters" .Name }}
{{- $updaterTypeName := printf "%sUpdater" .Name }}
{{- $optionsTypeName := printf "%sOptions" .Name }}
{{- $schemaTypeName := printf "%sDBSchemaField" .Name }}

// {{ $filterTypeName }} provides filtering capabilities for {{ .Name }}
type {{ $filterTypeName }} struct {
	filters map[{{ $schemaTypeName }}][]*repository.Filter
}

// New{{ $filterTypeName }} creates a new filter instance
func New{{ $filterTypeName }}() *{{ $filterTypeName }} {
	return &{{ $filterTypeName }}{
		filters: make(map[{{ $schemaTypeName }}][]*repository.Filter),
	}
}

// ListFilters returns all configured filters
func (f *{{ $filterTypeName }}) ListFilters() []*repository.Filter {
	var result []*repository.Filter
	for _, filterList := range f.filters {
		result = append(result, filterList...)
	} 
	return result
}

{{- range .FilterMethods }}

// {{ .Documentation }}
func ({{ .Receiver }}) {{ .Name }}({{ .Parameters }}) {{ .ReturnType }} {
	{{ .Body }}
}
{{- end }}

// {{ $updaterTypeName }} provides update capabilities for {{ .Name }}
type {{ $updaterTypeName }} struct {
	fields map[string]interface{}
}

// New{{ $updaterTypeName }} creates a new updater instance
func New{{ $updaterTypeName }}() *{{ $updaterTypeName }} {
	return &{{ $updaterTypeName }}{
		fields: make(map[string]interface{}),
	}
}

// GetChangeSet returns the fields to update
func (u *{{ $updaterTypeName }}) GetChangeSet() map[string]interface{} {
	return u.fields
}

{{- range .UpdaterMethods }}

// {{ .Documentation }}
func ({{ .Receiver }}) {{ .Name }}({{ .Parameters }}) {{ .ReturnType }} {
	{{ .Body }}
}
{{- end }}

// {{ $optionsTypeName }} provides query options for {{ .Name }}
type {{ $optionsTypeName }} struct {
	options []func(*repository.Options)
}

// New{{ $optionsTypeName }} creates a new options instance
func New{{ $optionsTypeName }}() *{{ $optionsTypeName }} {
	return &{{ $optionsTypeName }}{}
}

// Apply applies all configured options to repository options
func (o *{{ $optionsTypeName }}) Apply(repoOpts *repository.Options) {
	for _, option := range o.options {
		option(repoOpts)
	}
}

{{- range .OrderMethods }}

// {{ .Documentation }}  
func ({{ .Receiver }}) {{ .Name }}({{ .Parameters }}) {{ .ReturnType }} {
	{{ .Body }}
}
{{- end }}

// {{ $schemaTypeName }} represents database field names
type {{ $schemaTypeName }} string

// String returns the string representation of the field
func (f {{ $schemaTypeName }}) String() string {
	return string(f)
}

// {{ .Name }}DBSchema contains database field mappings for {{ .Name }}
var {{ .Name }}DBSchema = struct {
{{- range .Fields }}
	{{ .Name }} {{ $schemaTypeName }}
{{- end }}
}{
{{- range .Fields }}
	{{ .Name }}: {{ $schemaTypeName }}("{{ .DBName }}"),
{{- end }}
}

{{- end }}
`
