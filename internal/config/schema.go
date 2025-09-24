package config

import models "github.com/shahariaz/user_segmentation/internal/model"

// GetSchemaConfig returns the predefined schema configuration for Chorki platform
func GetSchemaConfig() *models.SchemaInfo {
	return &models.SchemaInfo{
		EntityTypes: []string{
			"chorki_customers",
			"chorki_subscriptions",
			"chorki_watch_histories",
			"chorki_contents",
			"chorki_devices",
		},
		FieldMappings: getFieldMappings(),
		Relationships: getRelationships(),
		DefaultFields: getDefaultFields(),
	}
}

func getFieldMappings() map[string][]models.FieldMapping {
	return map[string][]models.FieldMapping{

		"age": {
			{JSONField: "age", DgraphField: "chorki_customers.age", EntityType: "chorki_customers", DataType: "int"},
		},
		"country": {
			{JSONField: "country", DgraphField: "chorki_customers.country", EntityType: "chorki_customers", DataType: "string"},
		},
		"device": {
			{JSONField: "device", DgraphField: "chorki_customers.device", EntityType: "chorki_customers", DataType: "string"},
			{JSONField: "device", DgraphField: "chorki_devices.device_type", EntityType: "chorki_devices", DataType: "string"},
		},
		"app_version": {
			{JSONField: "app_version", DgraphField: "chorki_customers.app_version", EntityType: "chorki_customers", DataType: "string"},
			{JSONField: "app_version", DgraphField: "chorki_devices.app_version", EntityType: "chorki_devices", DataType: "string"},
		},
		"last_login_days": {
			{JSONField: "last_login_days", DgraphField: "chorki_customers.last_login_days", EntityType: "chorki_customers", DataType: "int"},
		},
		"email": {
			{JSONField: "email", DgraphField: "chorki_customers.email", EntityType: "chorki_customers", DataType: "string"},
		},
		"name": {
			{JSONField: "name", DgraphField: "chorki_customers.name", EntityType: "chorki_customers", DataType: "string"},
		},
		"is_active": {
			{JSONField: "is_active", DgraphField: "chorki_customers.is_active", EntityType: "chorki_customers", DataType: "bool"},
		},
		"city": {
			{JSONField: "city", DgraphField: "chorki_customers.city", EntityType: "chorki_customers", DataType: "string"},
		},

		// Subscription fields
		"subscription_status": {
			{JSONField: "subscription_status", DgraphField: "chorki_subscriptions.status", EntityType: "chorki_subscriptions", DataType: "string"},
		},
		"subscribed_package": {
			{JSONField: "subscribed_package", DgraphField: "chorki_subscriptions.package", EntityType: "chorki_subscriptions", DataType: "string"},
		},
		"package": {
			{JSONField: "package", DgraphField: "chorki_subscriptions.package", EntityType: "chorki_subscriptions", DataType: "string"},
		},
		"status": {
			{JSONField: "status", DgraphField: "chorki_subscriptions.status", EntityType: "chorki_subscriptions", DataType: "string"},
		},
		"price": {
			{JSONField: "price", DgraphField: "chorki_subscriptions.price", EntityType: "chorki_subscriptions", DataType: "float"},
		},
		"currency": {
			{JSONField: "currency", DgraphField: "chorki_subscriptions.currency", EntityType: "chorki_subscriptions", DataType: "string"},
		},
		"payment_method": {
			{JSONField: "payment_method", DgraphField: "chorki_subscriptions.payment_method", EntityType: "chorki_subscriptions", DataType: "string"},
		},
		"auto_renewal": {
			{JSONField: "auto_renewal", DgraphField: "chorki_subscriptions.auto_renewal", EntityType: "chorki_subscriptions", DataType: "bool"},
		},
		"trial_period": {
			{JSONField: "trial_period", DgraphField: "chorki_subscriptions.trial_period", EntityType: "chorki_subscriptions", DataType: "bool"},
		},

		// Watch history fields
		"watched_content": {
			{JSONField: "watched_content", DgraphField: "chorki_watch_histories.content_id", EntityType: "chorki_watch_histories", DataType: "complex"},
		},
		"favorite_genres": {
			{JSONField: "favorite_genres", DgraphField: "chorki_watch_histories.genre", EntityType: "chorki_watch_histories", DataType: "array"},
		},
		"content_type": {
			{JSONField: "content_type", DgraphField: "chorki_watch_histories.type", EntityType: "chorki_watch_histories", DataType: "string"},
			{JSONField: "content_type", DgraphField: "chorki_contents.type", EntityType: "chorki_contents", DataType: "string"},
		},

		// Content fields
		"genre": {
			{JSONField: "genre", DgraphField: "chorki_contents.genre", EntityType: "chorki_contents", DataType: "array"},
		},
		"title": {
			{JSONField: "title", DgraphField: "chorki_contents.title", EntityType: "chorki_contents", DataType: "string"},
		},

		// Device fields
		"device_type": {
			{JSONField: "device_type", DgraphField: "chorki_devices.device_type", EntityType: "chorki_devices", DataType: "string"},
		},
		"os_version": {
			{JSONField: "os_version", DgraphField: "chorki_devices.os_version", EntityType: "chorki_devices", DataType: "string"},
		},
	}
}

// getRelationships returns the relationships between entity types
func getRelationships() map[string][]string {
	return map[string][]string{
		"chorki_customers": {
			"chorki_subscriptions",
			"chorki_watch_histories",
			"chorki_devices",
		},
		"chorki_subscriptions": {
			"chorki_customers",
		},
		"chorki_watch_histories": {
			"chorki_customers",
			"chorki_contents",
		},
		"chorki_devices": {
			"chorki_customers",
		},
		"chorki_contents": {
			"chorki_watch_histories",
		},
	}
}

// getDefaultFields returns the default fields to select for each entity type
func getDefaultFields() map[string][]string {
	return map[string][]string{
		"chorki_customers": {
			"uid",
			"chorki_customers.id",
			"chorki_customers.name",
			"chorki_customers.email",
			"chorki_customers.age",
			"chorki_customers.country",
			"chorki_customers.device",
			"chorki_customers.app_version",
		},
		"chorki_subscriptions": {
			"uid",
			"chorki_subscriptions.id",
			"chorki_subscriptions.package",
			"chorki_subscriptions.status",
			"chorki_subscriptions.start_date",
			"chorki_subscriptions.end_date",
		},
		"chorki_watch_histories": {
			"uid",
			"chorki_watch_histories.id",
			"chorki_watch_histories.content_id",
			"chorki_watch_histories.content_title",
			"chorki_watch_histories.type",
			"chorki_watch_histories.genre",
			"chorki_watch_histories.watch_date",
		},
		"chorki_contents": {
			"uid",
			"chorki_contents.id",
			"chorki_contents.title",
			"chorki_contents.type",
			"chorki_contents.genre",
			"chorki_contents.duration",
			"chorki_contents.rating",
		},
		"chorki_devices": {
			"uid",
			"chorki_devices.id",
			"chorki_devices.device_type",
			"chorki_devices.device_model",
			"chorki_devices.app_version",
			"chorki_devices.is_active",
		},
	}
}

// GetOperatorMappings returns the mapping between JSON operators and DQL functions
func GetOperatorMappings() map[string]string {
	return map[string]string{
		"=":           "eq",
		">=":          "ge",
		"<=":          "le",
		">":           "gt",
		"<":           "lt",
		"IN":          "eq",
		"NOT_IN":      "not",
		"!=":          "not",
		"LIKE":        "alloftext",
		"ILIKE":       "anyoftext",
		"REGEX":       "regexp",
		"BETWEEN":     "between",
		"IS_NULL":     "eq",
		"IS_NOT_NULL": "has",
		"STARTS_WITH": "alloftext",
		"ENDS_WITH":   "alloftext",
		"CONTAINS":    "alloftext",
	}
}

// GetVersionFields returns fields that should be treated as version fields
func GetVersionFields() map[string]string {
	return map[string]string{
		"app_version": "numeric",
		"os_version":  "numeric",
		"version":     "numeric",
	}
}

// GetReversePredicates returns the reverse predicate mappings
func GetReversePredicates() map[string]string {
	return map[string]string{
		"customers":                      "~chorki_customers.subscriptions",
		"customers_from_devices":         "~chorki_customers.devices",
		"customers_from_watch_histories": "~chorki_customers.watch_histories",
	}
}

// GetFilterOptimizations returns suggested filter optimizations for common patterns
func GetFilterOptimizations() map[string][]string {
	return map[string][]string{
		"subscription_status_premium": {
			"eq(chorki_subscriptions.package, \"Premium\")",
			"(eq(chorki_subscriptions.status, \"active\") OR eq(chorki_subscriptions.status, \"trial\"))",
		},
		"subscription_status_basic": {
			"eq(chorki_subscriptions.package, \"Basic\")",
			"eq(chorki_subscriptions.status, \"active\")",
		},
	}
}

// GetPaginationConfig returns default pagination settings
func GetPaginationConfig() map[string]int {
	return map[string]int{
		"default_limit":  100,
		"max_limit":      1000,
		"default_offset": 0,
	}
}
