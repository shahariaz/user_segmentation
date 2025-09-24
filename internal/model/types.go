package models

// JSONQuery represents the root structure of the incoming JSON query
type JSONQuery struct {
	CombineWith string  `json:"combine_with" binding:"required"` // "AND" or "OR"
	Groups      []Group `json:"groups" binding:"required"`
}

type Group struct {
	CombineWith string   `json:"combine_with" binding:"required"` // "AND" or "OR"
	Filters     []Filter `json:"filters,omitempty"`
	Groups      []Group  `json:"groups,omitempty"` // Nested groups for complex queries
}

// Filter represents a single filter condition
type Filter struct {
	Field string      `json:"field" binding:"required"`
	Op    string      `json:"op" binding:"required"`    // "=", ">=", "<=", ">", "<", "IN"
	Value interface{} `json:"value" binding:"required"` // Can be string, int, float, array, or complex object
}

type DQLQuery struct {
	Queries []EntityQuery `json:"queries"`
}

type EntityQuery struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Function string `json:"function"`
	Filter   string `json:"filter"`
	Fields   string `json:"fields"`
}

type FieldMapping struct {
	JSONField      string `json:"json_field"`
	DgraphField    string `json:"dgraph_field"`
	EntityType     string `json:"entity_type"`
	DataType       string `json:"data_type"`
	IsRelationship bool   `json:"is_relationship"`
}

type OperatorMapping struct {
	JSONOperator string `json:"json_operator"`
	DQLFunction  string `json:"dql_function"`
}

type SchemaInfo struct {
	EntityTypes   []string                  `json:"entity_types"`
	FieldMappings map[string][]FieldMapping `json:"field_mappings"`
	Relationships map[string][]string       `json:"relationships"`
	DefaultFields map[string][]string       `json:"default_fields"`
}
