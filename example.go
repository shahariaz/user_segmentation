package querybuilder

import (
	"encoding/json"
	"fmt"
	"strings"

	models "github.com/shahariaz/user_segmentation/internal/model"
)

// QueryFilter represents a single filter in the JSON query
type QueryFilter struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

// QueryGroup represents a group of filters with a combine_with operator
type QueryGroup struct {
	CombineWith string        `json:"combine_with"`
	Filters     []QueryFilter `json:"filters"`
}

// Query represents the top-level JSON query structure
type Query struct {
	CombineWith string       `json:"combine_with"`
	Groups      []QueryGroup `json:"groups"`
}

// SchemaInfo is from the provided schema config
type SchemaInfo = models.SchemaInfo

// BuildDQLQuery builds a DQL query from the JSON query and schema
func BuildDQLQuery(queryJSON string, schema *SchemaInfo, entityType string, limit, offset int) (string, error) {
	var query Query
	if err := json.Unmarshal([]byte(queryJSON), &query); err != nil {
		return "", fmt.Errorf("failed to parse query JSON: %w", err)
	}

	// Get operator mappings and reverse predicates from schema
	opMappings := schema.GetOperatorMappings()
	reversePreds := models.GetReversePredicates()

	// Track var blocks and their names for multi-level filters
	var varBlocks []string
	var varNames []string
	varCount := 0

	// Build filter strings for each group
	var groupFilters []string
	for _, group := range query.Groups {
		var filterStrs []string
		for _, filter := range group.Filters {
			// Find the Dgraph field and entity type from schema
			mappings, ok := schema.FieldMappings[filter.Field]
			if !ok {
				return "", fmt.Errorf("field %s not found in schema", filter.Field)
			}

			// Assume the first mapping is relevant; handle ambiguity if needed
			mapping := mappings[0]
			dgraphField := mapping.DgraphField
			dataType := mapping.DataType
			targetEntity := mapping.EntityType

			// Format the value based on data type
			var valueStr string
			switch v := filter.Value.(type) {
			case string:
				valueStr = fmt.Sprintf(`"%s"`, v)
			case float64, int:
				valueStr = fmt.Sprintf("%v", v)
			case bool:
				valueStr = fmt.Sprintf("%v", v)
			case []interface{}:
				if filter.Op != "IN" {
					return "", fmt.Errorf("array value only supported for IN operator")
				}
				var inFilters []string
				for _, val := range v {
					if s, ok := val.(string); ok {
						inFilters = append(inFilters, fmt.Sprintf(`eq(%s, "%s")`, dgraphField, s))
					} else {
						inFilters = append(inFilters, fmt.Sprintf("eq(%s, %v)", dgraphField, val))
					}
				}
				valueStr = fmt.Sprintf("(%s)", strings.Join(inFilters, " OR "))
			default:
				return "", fmt.Errorf("unsupported value type for field %s", filter.Field)
			}

			// Map JSON operator to DQL operator
			dqlOp, ok := opMappings[filter.Op]
			if !ok {
				return "", fmt.Errorf("unsupported operator %s", filter.Op)
			}

			// If filter applies to a different entity, create a var block
			if targetEntity != entityType {
				varName := fmt.Sprintf("var%d", varCount)
				varCount++

				// Find the relationship path (e.g., chorki_customers -> chorki_subscriptions)
				var edge string
				for _, rel := range schema.Relationships[entityType] {
					if rel == targetEntity {
						// Direct edge, e.g., chorki_subscriptions
						edge = strings.Replace(targetEntity, "chorki_", "chorki_customers.", 1)
						break
					}
				}
				if edge == "" {
					// Try reverse edge
					for pred, reverse := range reversePreds {
						if reverse == fmt.Sprintf("~%s", targetEntity) {
							edge = pred
							break
						}
					}
				}
				if edge == "" {
					return "", fmt.Errorf("no relationship found from %s to %s", entityType, targetEntity)
				}

				// Build var block
				varBlock := fmt.Sprintf(`
  %s as var(func: type(%s)) @filter(%s(%s, %s)) {
    %s
  }`, varName, targetEntity, dqlOp, dgraphField, valueStr, edge)
				varBlocks = append(varBlocks, varBlock)
				varNames = append(varNames, varName)
				filterStrs = append(filterStrs, fmt.Sprintf("uid(%s)", varName))
			} else {
				// Direct filter on the main entity
				if filter.Op == "IN" {
					filterStrs = append(filterStrs, valueStr)
				} else {
					filterStrs = append(filterStrs, fmt.Sprintf("%s(%s, %s)", dqlOp, dgraphField, valueStr))
				}
			}
		}

		// Combine filters within the group
		if len(filterStrs) > 0 {
			combined := strings.Join(filterStrs, fmt.Sprintf(" %s ", strings.ToUpper(group.CombineWith)))
			if len(filterStrs) > 1 {
				combined = fmt.Sprintf("(%s)", combined)
			}
			groupFilters = append(groupFilters, combined)
		}
	}

	// Combine groups
	var filterDirective string
	if len(groupFilters) > 0 {
		combinedGroups := strings.Join(groupFilters, fmt.Sprintf(" %s ", strings.ToUpper(query.CombineWith)))
		if len(groupFilters) > 1 || len(varBlocks) > 0 {
			combinedGroups = fmt.Sprintf("(%s)", combinedGroups)
		}
		filterDirective = fmt.Sprintf("@filter%s", combinedGroups)
	}

	// Get default fields for the entity
	defaultFields, ok := schema.DefaultFields[entityType]
	if !ok {
		return "", fmt.Errorf("no default fields for entity %s", entityType)
	}
	fieldsStr := strings.Join(defaultFields, "\n    ")

	// Build final DQL query
	dqlQuery := fmt.Sprintf(`
query {
%s
  %ss(func: type(%s)) %s {
    %s
  }
}`, strings.Join(varBlocks, "\n"), entityType, entityType, filterDirective, fieldsStr)

	// Apply pagination
	if limit > 0 {
		dqlQuery = strings.Replace(dqlQuery, " {", fmt.Sprintf("(first: %d, offset: %d) {", limit, offset), 1)
	}

	return dqlQuery, nil
}
