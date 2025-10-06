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
	const mainEntityType = "customers"

	var variables []models.VariableBlock
	varCounter := 0

	var groupExpressions []string

	for _, group := range jsonQuery.Groups {
		groupExpr, vars, counter := c.processGroup(group, mainEntityType, varCounter)
		if groupExpr != "" {

			groupExpressions = append(groupExpressions, groupExpr)
			variables = append(variables, vars...)
			varCounter = counter

		}

	}

	var mainFilter string
	if len(groupExpressions) > 0 {
		combiner := " AND "
		if strings.ToUpper(jsonQuery.Operation) == "OR" {
			combiner = " OR "
		}

		if len(groupExpressions) == 1 {
			mainFilter = fmt.Sprintf("@filter(%s)", groupExpressions[0])
		} else {
			combinedFilter := strings.Join(groupExpressions, combiner)

			mainFilter = fmt.Sprintf("@filter(%s)", combinedFilter)
		}
	}

	limit := jsonQuery.Limit
	offset := jsonQuery.Offset
	if limit == 0 {
		limit = 100
	}
	pagination := fmt.Sprintf("first: %d, offset: %d", limit, offset)

	mainQuery := models.MainQuery{
		Name:       "customers",
		Type:       mainEntityType,
		Function:   fmt.Sprintf("type(%s)", mainEntityType),
		Filter:     mainFilter,
		Fields:     c.buildFieldsSelection(mainEntityType),
		Pagination: pagination,
	}

	return &models.DQLQuery{
		Variables: variables,
		MainQuery: mainQuery,
	}, nil
}

// processGroup processes a single group and returns the filter expression, variables, and updated counter
func (c *Converter) processGroup(group models.Group, mainEntityType string, varCounter int) (string, []models.VariableBlock, int) {
	var variables []models.VariableBlock
	var filterExpressions []string

	isOrGroup := strings.ToUpper(group.Operation) == "OR"

	for _, filter := range group.Filters {
		expr, vars, counter := c.processFilter(filter, mainEntityType, varCounter)

		if expr != "" {
			filterExpressions = append(filterExpressions, expr)
			variables = append(variables, vars...)
			varCounter = counter
		}
	}

	for _, nestedGroup := range group.Groups {
		expr, vars, counter := c.processGroup(nestedGroup, mainEntityType, varCounter)
		if expr != "" {
			filterExpressions = append(filterExpressions, expr)
			variables = append(variables, vars...)
			varCounter = counter
		}
	}

	if len(filterExpressions) == 0 {
		return "", variables, varCounter
	}

	// Combine expressions based on group's operation
	combiner := " AND "
	if isOrGroup {
		combiner = " OR "
	}

	var groupExpression string
	if len(filterExpressions) == 1 {
		groupExpression = filterExpressions[0]
	} else {
		groupExpression = "(" + strings.Join(filterExpressions, combiner) + ")"
	}

	return groupExpression, variables, varCounter
}

func (c *Converter) processFilter(filter models.Filter, mainEntityType string, varCounter int) (string, []models.VariableBlock, int) {
	var variables []models.VariableBlock

	mappings, exists := c.schema.FieldMappings[filter.Field]
	if !exists {
		return "", variables, varCounter
	}

	var mainEntityMapping *models.FieldMapping
	var crossEntityMapping *models.FieldMapping

	for _, mapping := range mappings {
		if mapping.EntityType == mainEntityType {
			mainEntityMapping = &mapping
			break
		}
	}

	if mainEntityMapping == nil && len(mappings) > 0 {
		crossEntityMapping = &mappings[0]
	}

	if mainEntityMapping != nil {

		condition := c.buildDQLCondition(mainEntityMapping, filter)
		if condition != "" {
			return condition, variables, varCounter
		}
	}

	// Handle cross-entity filters
	if c != nil {
		varName := fmt.Sprintf("var%d", varCounter)
		varCounter++

		filterCondition := c.buildDQLCondition(crossEntityMapping, filter)
		if filterCondition == "" {
			return "", variables, varCounter
		}

		forwardPredicate := c.getForwardPredicate(crossEntityMapping.EntityType)
		if forwardPredicate == "" {
			return "", variables, varCounter
		}

		variable := models.VariableBlock{
			Name:   varName,
			Type:   mainEntityType,
			Filter: "",
			Fields: fmt.Sprintf("    %s @filter(%s)", forwardPredicate, filterCondition),
		}

		variables = append(variables, variable)
		return fmt.Sprintf("uid(%s)", varName), variables, varCounter
	}

	return "", variables, varCounter
}

func (c *Converter) getForwardPredicate(entityType string) string {
	switch entityType {
	case "subscriptions":
		return "customers.subscriptions"
	case "devices":
		return "customers.devices"
	case "watch_histories":
		return "customers.watch_histories"
	case "purchases":
		return "customers.purchases"
	default:
		return ""
	}
}

func (c *Converter) buildFieldsSelection(entityType string) string {
	fields := c.schema.DefaultFields[entityType]
	if len(fields) == 0 {
		return "    uid\n    expand(_all_)"
	}

	var fieldLines []string
	for _, field := range fields {
		fieldLines = append(fieldLines, "    "+field)
	}

	return strings.Join(fieldLines, "\n")
}

func (c *Converter) buildDQLCondition(mapping *models.FieldMapping, filter models.Filter) string {
	dqlFunction := c.operators[filter.Op]
	if dqlFunction == "" {
		return ""
	}

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
		return c.buildNullCondition(mapping, true)
	case "IS_NOT_NULL":
		return c.buildNullCondition(mapping, false)
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
		var conditions []string
		for _, item := range v {
			value := c.formatValue(item, mapping.DataType)
			if value != "" {
				conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, value))
			}
		}

		if len(conditions) > 1 {
			return "(" + strings.Join(conditions, " OR ") + ")"
		} else if len(conditions) == 1 {
			return conditions[0]
		}

	case map[string]interface{}:
		return c.buildComplexObjectCondition(mapping, v)

	default:
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

	if mode, isVersionField := c.versionFields[filter.Field]; isVersionField && mode == "numeric" {
		return c.buildVersionComparisonCondition(mapping, filter, dqlFunction)
	}

	value := c.formatValue(filter.Value, mapping.DataType)
	if value == "" {
		return ""
	}

	if filter.Op == "!=" {
		return fmt.Sprintf("NOT eq(%s, %s)", mapping.DgraphField, value)
	}

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

func (c *Converter) buildBetweenCondition(mapping *models.FieldMapping, filter models.Filter) string {
	switch v := filter.Value.(type) {
	case []interface{}:
		if len(v) == 2 {
			min := c.formatValue(v[0], mapping.DataType)
			max := c.formatValue(v[1], mapping.DataType)
			if min != "" && max != "" {
				return fmt.Sprintf("(ge(%s, %s) AND le(%s, %s))",
					mapping.DgraphField, min, mapping.DgraphField, max)
			}
		}
	case map[string]interface{}:
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

func (c *Converter) buildNullCondition(mapping *models.FieldMapping, isNull bool) string {
	if isNull {
		return fmt.Sprintf("NOT has(%s)", mapping.DgraphField)
	}
	return fmt.Sprintf("has(%s)", mapping.DgraphField)
}

func (c *Converter) buildStringPatternCondition(mapping *models.FieldMapping, filter models.Filter, pattern string) string {
	value := c.formatValue(filter.Value, "string")
	if value == "" {
		return ""
	}

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

	case "datetime":

		if str, ok := value.(string); ok {
			return fmt.Sprintf(`"%s"`, str)
		}
		return fmt.Sprintf(`"%v"`, value)

	default:
		return fmt.Sprintf(`"%v"`, value)
	}
}

func (c *Converter) buildVersionComparisonCondition(mapping *models.FieldMapping, filter models.Filter, dqlFunction string) string {
	versionStr, ok := filter.Value.(string)
	if !ok {
		value := c.formatValue(filter.Value, mapping.DataType)
		return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
	}

	numericVersion, err := utils.ConvertVersionToNumeric(versionStr)
	if err != nil {
		value := c.formatValue(filter.Value, mapping.DataType)
		return fmt.Sprintf("%s(%s, %s)", dqlFunction, mapping.DgraphField, value)
	}

	numericField := mapping.DgraphField + "_numeric"
	return fmt.Sprintf("%s(%s, %d)", dqlFunction, numericField, numericVersion)
}

func (c *Converter) buildComplexObjectCondition(mapping *models.FieldMapping, obj map[string]interface{}) string {
	if mapping.JSONField == "watched_content" {
		if contentType, exists := obj["content_type"]; exists {
			if ids, idsExist := obj["ids"]; idsExist {
				if idArray, ok := ids.([]interface{}); ok {
					var conditions []string

					if ctMappings, ctExists := c.schema.FieldMappings["content_type"]; ctExists {
						for _, ctMapping := range ctMappings {
							if ctMapping.EntityType == mapping.EntityType {
								typeValue := c.formatValue(contentType, "string")
								conditions = append(conditions, fmt.Sprintf("eq(%s, %s)", ctMapping.DgraphField, typeValue))
								break
							}
						}
					}

					var idConditions []string
					for _, id := range idArray {
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
							idConditions = append(idConditions, fmt.Sprintf("eq(%s, %s)", mapping.DgraphField, idValue))
						}
					}

					if len(idConditions) > 1 {
						conditions = append(conditions, "("+strings.Join(idConditions, " OR ")+")")
					} else if len(idConditions) == 1 {
						conditions = append(conditions, idConditions[0])
					}

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
	var blocks []string

	for _, variable := range dqlQuery.Variables {
		var block string
		if variable.Filter != "" {
			block = fmt.Sprintf("  %s as var(func: type(%s)) %s {\n%s\n  }",
				variable.Name,
				variable.Type,
				variable.Filter,
				variable.Fields,
			)
		} else {
			block = fmt.Sprintf("  %s as var(func: type(%s)) {\n%s\n  }",
				variable.Name,
				variable.Type,
				variable.Fields,
			)
		}
		blocks = append(blocks, block)
	}

	mainBlock := fmt.Sprintf("  %s(func: %s, %s) %s {\n%s\n  }",
		dqlQuery.MainQuery.Name,
		dqlQuery.MainQuery.Function,
		dqlQuery.MainQuery.Pagination,
		dqlQuery.MainQuery.Filter,
		dqlQuery.MainQuery.Fields,
	)
	blocks = append(blocks, mainBlock)

	return "query {\n" + strings.Join(blocks, "\n") + "\n}"
}
