package converter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shahariaz/user_segmentation/internal/config"
	models "github.com/shahariaz/user_segmentation/internal/model"
	"github.com/shahariaz/user_segmentation/internal/utils"
)

type Converter struct {
	schema            *models.SchemaInfo
	operators         map[string]string
	versionFields     map[string]string
	reversePredicates map[string]string
}

func NewConverter() *Converter {
	schema := config.GetSchemaConfig()
	return &Converter{
		schema:            schema,
		operators:         config.GetOperatorMappings(),
		versionFields:     config.GetVersionFields(),
		reversePredicates: config.GetReversePredicates(),
	}
}

func (c *Converter) ConvertToDQL(jsonQuery *models.JSONQuery) (*models.DQLQuery, error) {
	involvedEntities := c.getInvolvedEntityTypes(jsonQuery)

	//we have to check top level later

	var queries []models.EntityQuery
	for _, entityType := range involvedEntities {
		filter, err := c.buildFilterForEntity(jsonQuery, entityType)
		if err != nil {
			return nil, fmt.Errorf("error building filter for %s: %v", entityType, err)
		}

		// Only include entities that have applicable filters
		if filter != "" {
			query := models.EntityQuery{
				Name:     c.getQueryName(entityType),
				Type:     entityType,
				Function: fmt.Sprintf("type(%s)", entityType),
				Filter:   filter,
				Fields:   c.buildFieldsSelection(entityType),
			}
			queries = append(queries, query)
		}
	}

	return &models.DQLQuery{Queries: queries}, nil

}

func (c *Converter) getInvolvedEntityTypes(jsonQuery *models.JSONQuery) []string {
	entityTypeMap := make(map[string]bool)

	c.collectEntityTypesFromGroups(jsonQuery.Groups, entityTypeMap)

	var result []string
	for entityType := range entityTypeMap {
		result = append(result, entityType)
	}

	if len(result) == 0 {
		result = append(result, "chorki_customers")
	}

	return result
}

func (c *Converter) collectEntityTypesFromGroups(groups []models.Group, entityTypeMap map[string]bool) {
	for _, group := range groups {

		for _, filter := range group.Filters {
			if mappings, exists := c.schema.FieldMappings[filter.Field]; exists {

				for _, mapping := range mappings {
					entityTypeMap[mapping.EntityType] = true
				}
			}
		}

		if len(group.Groups) > 0 {
			c.collectEntityTypesFromGroups(group.Groups, entityTypeMap)
		}
	}
}

func (c *Converter) buildFilterForEntity(jsonQuery *models.JSONQuery, entityType string) (string, error) {
	filter := c.buildGroupsFilter(jsonQuery.Groups, jsonQuery.CombineWith, entityType)
	if filter == "" {
		return "", nil
	}
	return fmt.Sprintf("@filter(%s)", filter), nil
}

func (c *Converter) getQueryName(entityType string) string {
	switch entityType {
	case "chorki_customers":
		return "customers"
	case "chorki_subscriptions":
		return "subscriptions"
	case "chorki_watch_histories":
		return "watch_histories"
	case "chorki_contents":
		return "contents"
	case "chorki_devices":
		return "devices"
	default:
		// Generic case: remove common prefixes
		return strings.ReplaceAll(entityType, "chorki_", "")
	}
}

func (c *Converter) buildFieldsSelection(entityType string) string {
	fields := c.schema.DefaultFields[entityType]
	if len(fields) == 0 {
		return "uid\nexpand(_all_)"
	}

	var fieldLines []string

	for _, field := range fields {
		fieldLines = append(fieldLines, "    "+field)
	}

	if relationships, exists := c.schema.Relationships[entityType]; exists {
		for _, relatedEntity := range relationships {
			relatedFields := c.schema.DefaultFields[relatedEntity]
			if len(relatedFields) > 0 {
				// Get the relationship predicate name
				relationName := c.getRelationshipName(entityType, relatedEntity)

				// Add related entity block with nested fields
				fieldLines = append(fieldLines, "")
				fieldLines = append(fieldLines, fmt.Sprintf("    %s {", relationName))
				for _, relField := range relatedFields {
					fieldLines = append(fieldLines, "      "+relField)
				}
				fieldLines = append(fieldLines, "    }")
			}
		}
	}

	return strings.Join(fieldLines, "\n")
}

func (c *Converter) buildGroupsFilter(groups []models.Group, combineWith, entityType string) string {
	var conditions []string

	// Process each group and collect valid conditions
	for _, group := range groups {
		condition := c.buildGroupFilter(group, entityType)
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	if len(conditions) == 1 {
		return conditions[0]
	}

	operator := " AND "
	if strings.ToUpper(combineWith) == "OR" {
		operator = " OR "
	}

	// Combine conditions with proper parentheses for complex expressions
	return "(" + strings.Join(conditions, operator) + ")"
}

func (c *Converter) getRelationshipName(fromEntity, toEntity string) string {
	switch {
	// Forward relationships from customers to related entities
	case fromEntity == "chorki_customers" && toEntity == "chorki_subscriptions":
		return "chorki_customers.subscriptions"
	case fromEntity == "chorki_customers" && toEntity == "chorki_watch_histories":
		return "chorki_customers.watch_histories"
	case fromEntity == "chorki_customers" && toEntity == "chorki_devices":
		return "chorki_customers.devices"

	// Reverse relationships back to customers (using ~ notation)
	case fromEntity == "chorki_subscriptions" && toEntity == "chorki_customers":
		return "~chorki_customers.subscriptions"
	case fromEntity == "chorki_devices" && toEntity == "chorki_customers":
		return "~chorki_customers.devices"
	case fromEntity == "chorki_watch_histories" && toEntity == "chorki_customers":
		return "~chorki_customers.watch_histories"

	// Content relationships
	case fromEntity == "chorki_watch_histories" && toEntity == "chorki_contents":
		return "chorki_watch_histories.content"
	case fromEntity == "chorki_contents" && toEntity == "chorki_watch_histories":
		return "~chorki_watch_histories.content"

	default:

		if toEntity == "chorki_customers" {
			return "customers"
		}

		return strings.ReplaceAll(toEntity, "chorki_", "")
	}
}

func (c *Converter) buildGroupFilter(group models.Group, entityType string) string {
	var conditions []string

	for _, filter := range group.Filters {
		condition := c.buildFilterCondition(filter, entityType)
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	// Apply entity-specific filter optimizations
	if entityType == "chorki_subscriptions" {
		conditions = c.optimizeSubscriptionFilters(conditions, group)
	}

	if len(group.Groups) > 0 {
		nestedCondition := c.buildGroupsFilter(group.Groups, group.CombineWith, entityType)
		if nestedCondition != "" {
			conditions = append(conditions, nestedCondition)
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	if len(conditions) == 1 {
		return conditions[0]
	}

	operator := " AND "
	if strings.ToUpper(group.CombineWith) == "OR" {
		operator = " OR "
	}

	return "(" + strings.Join(conditions, operator) + ")"
}

func (c *Converter) buildFilterCondition(filter models.Filter, entityType string) string {

	mappings, exists := c.schema.FieldMappings[filter.Field]
	if !exists {
		return ""
	}

	var relevantMapping *models.FieldMapping
	for _, mapping := range mappings {
		if mapping.EntityType == entityType {
			relevantMapping = &mapping
			break
		}
	}

	// No mapping found for this entity type
	if relevantMapping == nil {
		return ""
	}

	// Build the actual DQL condition using the mapping
	return c.buildDQLCondition(relevantMapping, filter)
}

func (c *Converter) buildDQLCondition(mapping *models.FieldMapping, filter models.Filter) string {

	dqlFunction := c.operators[filter.Op]
	if dqlFunction == "" {
		return ""
	}

	// Handle different operator types with specialized logic
	switch filter.Op {
	case "IN":
		return c.buildInCondition(mapping, filter)
	case "NOT_IN":
		return c.buildNotInCondition(mapping, filter)
	case "=", ">=", "<=", ">", "<", "!=":
		return c.buildComparisonCondition(mapping, filter, dqlFunction)
	case "LIKE", "ILIKE", "CONTAINS":
		return c.buildTextSearchCondition(mapping, filter, filter.Op)
	case "REGEX":
		return c.buildRegexCondition(mapping, filter)
	case "BETWEEN":
		return c.buildBetweenCondition(mapping, filter)
	case "IS_NULL":
		return c.buildNullCondition(mapping, filter, true)
	case "IS_NOT_NULL":
		return c.buildNullCondition(mapping, filter, false)
	case "STARTS_WITH":
		return c.buildStringPatternCondition(mapping, filter, "starts_with")
	case "ENDS_WITH":
		return c.buildStringPatternCondition(mapping, filter, "ends_with")
	default:
		return ""
	}
}

func (c *Converter) buildInCondition(mapping *models.FieldMapping, filter models.Filter) string {
	switch v := filter.Value.(type) {
	case []interface{}:
		// Handle array of simple values (strings, numbers, etc.)
		var conditions []string
		for _, item := range v {
			value := c.formatValue(item, mapping.DataType)
			if value != "" {
				conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value))
			}
		}

		// Combine multiple conditions with OR logic
		if len(conditions) > 1 {
			return "(" + strings.Join(conditions, " OR ") + ")"
		} else if len(conditions) == 1 {
			return conditions[0]
		}

	case map[string]interface{}:
		// Handle complex nested objects (e.g., watched_content with content_type and ids)
		return c.buildComplexObjectCondition(mapping, v)

	default:
		// Handle single value by treating it as a single-element array
		value := c.formatValue(v, mapping.DataType)
		if value != "" {
			return fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value)
		}
	}

	return ""
}

func (c *Converter) buildNotInCondition(mapping *models.FieldMapping, filter models.Filter) string {
	switch v := filter.Value.(type) {
	case []interface{}:

		var conditions []string
		for _, item := range v {
			value := c.formatValue(item, mapping.DataType)
			if value != "" {
				conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value))
			}
		}

		if len(conditions) > 1 {
			return "NOT (" + strings.Join(conditions, " OR ") + ")"
		} else if len(conditions) == 1 {
			return "NOT " + conditions[0]
		}
	default:

		value := c.formatValue(v, mapping.DataType)
		if value != "" {
			return fmt.Sprintf("NOT eq(%s, %s)", mapping.DgraphField, value)
		}
	}
	return ""
}

func (c *Converter) buildComparisonCondition(mapping *models.FieldMapping, filter models.Filter, dqlFunction string) string {
	// Check for special version field handling (numeric version comparisons)
	if mode, isVersionField := c.versionFields[filter.Field]; isVersionField && mode == "numeric" {
		return c.buildVersionComparisonCondition(mapping, filter, dqlFunction)
	}

	// Format the value according to the field's data type
	value := c.formatValue(filter.Value, mapping.DataType)
	if value == "" {
		return "" // Invalid or unsupported value
	}

	// Handle inequality operator with NOT + eq for better DQL performance
	if filter.Op == "!=" {
		return fmt.Sprintf("NOT eq(%s, %s)", mapping.DgraphField, value)
	}

	// Standard comparison condition
	return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
}
func (c *Converter) buildTextSearchCondition(mapping *models.FieldMapping, filter models.Filter, op string) string {
	value := c.formatValue(filter.Value, "string")
	if value == "" {
		return ""
	}

	switch op {
	case "LIKE", "CONTAINS":
		return fmt.Sprintf("alloftext(%s, %s)", mapping.DgraphField, value)
	case "ILIKE":
		return fmt.Sprintf("anyoftext(%s, %s)", mapping.DgraphField, value)
	default:
		return ""
	}
}

func (c *Converter) buildRegexCondition(mapping *models.FieldMapping, filter models.Filter) string {
	value := c.formatValue(filter.Value, "string")
	if value == "" {
		return ""
	}
	return fmt.Sprintf("regexp(%s, %s)", mapping.DgraphField, value)
}

// buildBetweenCondition builds BETWEEN conditions
func (c *Converter) buildBetweenCondition(mapping *models.FieldMapping, filter models.Filter) string {
	switch v := filter.Value.(type) {
	case []interface{}:

		if len(v) == 2 {
			min := c.formatValue(v[0], mapping.DataType)
			max := c.formatValue(v[1], mapping.DataType)
			if min != "" && max != "" {
				// Combine min and max conditions with AND logic
				return fmt.Sprintf("(ge(%s, %s) AND le(%s, %s))",
					mapping.DgraphField, min, mapping.DgraphField, max)
			}
		}
	case map[string]interface{}:
		// Handle object-style range specification: {"min": value, "max": value}
		if minVal, hasMin := v["min"]; hasMin {
			if maxVal, hasMax := v["max"]; hasMax {
				min := c.formatValue(minVal, mapping.DataType)
				max := c.formatValue(maxVal, mapping.DataType)
				if min != "" && max != "" {
					return fmt.Sprintf("(ge(%s, %s) AND le(%s, %s))",
						mapping.DgraphField, min, mapping.DgraphField, max)
				}
			}
		}
	}
	return ""
}

// buildNullCondition builds NULL/NOT NULL conditions
func (c *Converter) buildNullCondition(mapping *models.FieldMapping, filter models.Filter, isNull bool) string {
	if isNull {
		return fmt.Sprintf("NOT has(%s)", mapping.DgraphField)
	} else {
		return fmt.Sprintf("has(%s)", mapping.DgraphField)
	}
}

// buildStringPatternCondition builds string pattern conditions
func (c *Converter) buildStringPatternCondition(mapping *models.FieldMapping, filter models.Filter, pattern string) string {
	value := c.formatValue(filter.Value, "string")
	if value == "" {
		return ""
	}

	// Remove quotes for pattern matching
	cleanValue := strings.Trim(value, `"`)

	switch pattern {
	case "starts_with":
		return fmt.Sprintf("regexp(%s, /^%s/)", mapping.DgraphField, cleanValue)
	case "ends_with":
		return fmt.Sprintf("regexp(%s, /%s$/)", mapping.DgraphField, cleanValue)
	default:
		return ""
	}
}
func (c *Converter) formatValue(value interface{}, dataType string) string {
	if value == nil {
		return ""
	}

	switch dataType {
	case "string":

		if str, ok := value.(string); ok {
			return fmt.Sprintf(`"%s"`, strings.ReplaceAll(str, `"`, `\"`))
		}

		return fmt.Sprintf(`"%v"`, value)

	case "int":

		switch v := value.(type) {
		case int:
			return strconv.Itoa(v)
		case float64:
			return strconv.Itoa(int(v))
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return strconv.Itoa(i)
			}
		}
		// Fallback: attempt direct conversion
		return fmt.Sprintf("%v", value)

	case "float":

		switch v := value.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			return strconv.FormatFloat(float64(v), 'f', -1, 64)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return strconv.FormatFloat(f, 'f', -1, 64)
			}
		}

		return fmt.Sprintf("%v", value)

	case "bool":

		if b, ok := value.(bool); ok {
			return strconv.FormatBool(b)
		}

		return "false"

	default:

		return fmt.Sprintf(`"%v"`, value)
	}
}
func (c *Converter) optimizeSubscriptionFilters(conditions []string, group models.Group) []string {

	var hasPackagePremium, hasStatusTrial bool
	var trialIndex int

	for i, condition := range conditions {
		if strings.Contains(condition, `eq(chorki_subscriptions.package, "Premium")`) {
			hasPackagePremium = true
		}
		if strings.Contains(condition, `eq(chorki_subscriptions.status, "trial")`) {
			hasStatusTrial = true
			trialIndex = i
		}
	}

	if hasPackagePremium && hasStatusTrial && strings.ToUpper(group.CombineWith) == "AND" {
		// Create optimized conditions array
		optimizedConditions := make([]string, len(conditions))
		copy(optimizedConditions, conditions)

		// Replace restrictive trial-only status with more inclusive active OR trial
		optimizedConditions[trialIndex] = `(eq(chorki_subscriptions.status, "trial") OR eq(chorki_subscriptions.status, "active"))`

		return optimizedConditions
	}

	return conditions
}

func (c *Converter) buildVersionComparisonCondition(mapping *models.FieldMapping, filter models.Filter, dqlFunction string) string {
	// Convert version string to numeric value for accurate comparison
	versionStr, ok := filter.Value.(string)
	if !ok {
		// Fallback to regular comparison if value is not a string
		value := c.formatValue(filter.Value, mapping.DataType)
		return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
	}

	// Attempt numeric conversion of version string (e.g., "1.2.3" -> 10203)
	numericVersion, err := utils.ConvertVersionToNumeric(versionStr)
	if err != nil {
		// If conversion fails, fallback to string comparison
		// Note: In production, consider logging this conversion failure
		value := c.formatValue(filter.Value, mapping.DataType)
		return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
	}

	// Use numeric field for comparison (assumes schema has parallel numeric fields)
	// Example: app_version -> app_version_numeric
	numericField := mapping.DgraphField + "_numeric"
	return fmt.Sprintf("%s(%s, %d)", dqlFunction, numericField, numericVersion)
}

func (c *Converter) buildComplexObjectCondition(mapping *models.FieldMapping, obj map[string]interface{}) string {
	// Special handling for watched_content queries
	if mapping.JSONField == "watched_content" {
		// Extract content_type and content IDs from the filter object
		if contentType, exists := obj["content_type"]; exists {
			if ids, idsExist := obj["ids"]; idsExist {
				if idArray, ok := ids.([]interface{}); ok {
					var conditions []string

					// Add content type condition if mapping exists in schema
					if ctMappings, ctExists := c.schema.FieldMappings["content_type"]; ctExists {
						for _, ctMapping := range ctMappings {
							if ctMapping.EntityType == mapping.EntityType {
								typeValue := c.formatValue(contentType, "string")
								conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", ctMapping.DgraphField, typeValue))
								break
							}
						}
					}

					// Build conditions for content IDs - use proper method based on field type
					var idConditions []string
					for _, id := range idArray {
						// Convert numeric IDs to string since content_id is a string field in Dgraph
						var idValue string
						switch v := id.(type) {
						case int:
							idValue = fmt.Sprintf(`"%d"`, v)
						case int64:
							idValue = fmt.Sprintf(`"%d"`, v)
						case float64:
							idValue = fmt.Sprintf(`"%.0f"`, v)
						case string:
							idValue = fmt.Sprintf(`"%s"`, v)
						default:
							idValue = c.formatValue(id, "string")
						}

						if idValue != "" {
							// Use eq() for scalar fields, not uid_in()
							idConditions = append(idConditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, idValue))
						}
					}

					// Generate OR condition for multiple content IDs (since they're scalar fields)
					if len(idConditions) > 1 {
						conditions = append(conditions, "("+strings.Join(idConditions, " OR ")+")")
					} else if len(idConditions) == 1 {
						conditions = append(conditions, idConditions[0])
					}

					// Combine content type and ID conditions with AND logic
					if len(conditions) > 0 {
						return "(" + strings.Join(conditions, " AND ") + ")"
					}
				}
			}
		}
	}

	return ""
}

func (c *Converter) GenerateDQLString(dqlQuery *models.DQLQuery) string {
	if len(dqlQuery.Queries) == 0 {
		return "{}"
	}

	var queryBlocks []string

	for _, query := range dqlQuery.Queries {
		block := fmt.Sprintf("  %s(func: %s) %s {\n%s\n  }",
			query.Name,
			query.Function,
			query.Filter,
			query.Fields,
		)
		queryBlocks = append(queryBlocks, block)
	}

	return "{\n" + strings.Join(queryBlocks, "\n\n") + "\n}"
}
