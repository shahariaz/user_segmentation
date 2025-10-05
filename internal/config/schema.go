package config

import models "github.com/shahariaz/user_segmentation/internal/model"

// GetSchemaConfig returns the predefined schema configuration for the platform
func GetSchemaConfig() *models.SchemaInfo {
	return &models.SchemaInfo{
		EntityTypes:   []string{"customers", "subscriptions", "watch_histories", "contents", "devices", "purchases"},
		FieldMappings: getFieldMappings(),
		Relationships: getRelationships(),
		DefaultFields: getDefaultFields(),
	}
}

func getFieldMappings() map[string][]models.FieldMapping {
	return map[string][]models.FieldMapping{
		// Customer fields
		"age": {
			{JSONField: "age", DgraphField: "customers.age", EntityType: "customers", DataType: "int"},
		},
		"country": {
			{JSONField: "country", DgraphField: "customers.country", EntityType: "customers", DataType: "string"},
		},
		"device": {
			{JSONField: "device", DgraphField: "devices.device_type", EntityType: "devices", DataType: "string"},
			{JSONField: "device", DgraphField: "customers.device", EntityType: "customers", DataType: "string"},
		},
		"app_version": {
			{JSONField: "app_version", DgraphField: "customers.app_version", EntityType: "customers", DataType: "string"},
			{JSONField: "app_version", DgraphField: "devices.app_version", EntityType: "devices", DataType: "string"},
		},
		"last_login_days": {
			{JSONField: "last_login_days", DgraphField: "customers.last_login_days", EntityType: "customers", DataType: "int"},
		},
		"email": {
			{JSONField: "email", DgraphField: "customers.email", EntityType: "customers", DataType: "string"},
		},
		"name": {
			{JSONField: "name", DgraphField: "customers.name", EntityType: "customers", DataType: "string"},
		},
		"is_active": {
			{JSONField: "is_active", DgraphField: "customers.is_active", EntityType: "customers", DataType: "bool"},
		},
		"city": {
			{JSONField: "city", DgraphField: "customers.city", EntityType: "customers", DataType: "string"},
		},

		// Subscription fields
		"subscription_status": {
			{JSONField: "subscription_status", DgraphField: "subscriptions.status", EntityType: "subscriptions", DataType: "string"},
		},
		"subscribed_package": {
			{JSONField: "subscribed_package", DgraphField: "subscriptions.package", EntityType: "subscriptions", DataType: "string"},
		},
		"package": {
			{JSONField: "package", DgraphField: "subscriptions.package", EntityType: "subscriptions", DataType: "string"},
		},
		"status": {
			{JSONField: "status", DgraphField: "subscriptions.status", EntityType: "subscriptions", DataType: "string"},
		},
		"price": {
			{JSONField: "price", DgraphField: "subscriptions.price", EntityType: "subscriptions", DataType: "float"},
		},
		"currency": {
			{JSONField: "currency", DgraphField: "subscriptions.currency", EntityType: "subscriptions", DataType: "string"},
		},
		"payment_method": {
			{JSONField: "payment_method", DgraphField: "subscriptions.payment_method", EntityType: "subscriptions", DataType: "string"},
		},
		"auto_renewal": {
			{JSONField: "auto_renewal", DgraphField: "subscriptions.auto_renewal", EntityType: "subscriptions", DataType: "bool"},
		},
		"trial_period": {
			{JSONField: "trial_period", DgraphField: "subscriptions.trial_period", EntityType: "subscriptions", DataType: "bool"},
		},

		// Watch history fields
		"watched_content": {
			{JSONField: "watched_content", DgraphField: "watch_histories.content_id", EntityType: "watch_histories", DataType: "complex"},
		},
		"favorite_genres": {
			{JSONField: "favorite_genres", DgraphField: "watch_histories.genre", EntityType: "watch_histories", DataType: "array"},
		},
		"content_type": {
			{JSONField: "content_type", DgraphField: "watch_histories.type", EntityType: "watch_histories", DataType: "string"},
			{JSONField: "content_type", DgraphField: "contents.type", EntityType: "contents", DataType: "string"},
		},

		// Content fields
		"genre": {
			{JSONField: "genre", DgraphField: "contents.genre", EntityType: "contents", DataType: "array"},
		},
		"title": {
			{JSONField: "title", DgraphField: "contents.title", EntityType: "contents", DataType: "string"},
		},

		// Device fields
		"device_type": {
			{JSONField: "device_type", DgraphField: "devices.device_type", EntityType: "devices", DataType: "string"},
		},
		"os_version": {
			{JSONField: "os_version", DgraphField: "devices.os_version", EntityType: "devices", DataType: "string"},
		},

		// Purchase fields
		"purchasable_id": {
			{JSONField: "purchasable_id", DgraphField: "purchases.purchasable_id", EntityType: "purchases", DataType: "string"},
		},
		"purchase_status": {
			{JSONField: "purchase_status", DgraphField: "purchases.status", EntityType: "purchases", DataType: "string"},
		},

		// Datetime fields for customers
		"created_at": {
			{JSONField: "created_at", DgraphField: "customers.created_at", EntityType: "customers", DataType: "datetime"},
		},
		"updated_at": {
			{JSONField: "updated_at", DgraphField: "customers.updated_at", EntityType: "customers", DataType: "datetime"},
		},
		"last_login_date": {
			{JSONField: "last_login_date", DgraphField: "customers.last_login_date", EntityType: "customers", DataType: "datetime"},
		},
		"registration_date": {
			{JSONField: "registration_date", DgraphField: "customers.created_at", EntityType: "customers", DataType: "datetime"},
		},

		// Datetime fields for subscriptions
		"subscription_start_date": {
			{JSONField: "subscription_start_date", DgraphField: "subscriptions.start_date", EntityType: "subscriptions", DataType: "datetime"},
		},
		"subscription_end_date": {
			{JSONField: "subscription_end_date", DgraphField: "subscriptions.end_date", EntityType: "subscriptions", DataType: "datetime"},
		},
		"subscription_created_at": {
			{JSONField: "subscription_created_at", DgraphField: "subscriptions.created_at", EntityType: "subscriptions", DataType: "datetime"},
		},
		"start_date": {
			{JSONField: "start_date", DgraphField: "subscriptions.start_date", EntityType: "subscriptions", DataType: "datetime"},
		},
		"end_date": {
			{JSONField: "end_date", DgraphField: "subscriptions.end_date", EntityType: "subscriptions", DataType: "datetime"},
		},

		// Datetime fields for watch history
		"watch_date": {
			{JSONField: "watch_date", DgraphField: "watch_histories.watch_date", EntityType: "watch_histories", DataType: "datetime"},
		},
		"watched_at": {
			{JSONField: "watched_at", DgraphField: "watch_histories.watch_date", EntityType: "watch_histories", DataType: "datetime"},
		},
		"watch_history_created_at": {
			{JSONField: "watch_history_created_at", DgraphField: "watch_histories.created_at", EntityType: "watch_histories", DataType: "datetime"},
		},

		// Datetime fields for content
		"content_release_date": {
			{JSONField: "content_release_date", DgraphField: "contents.release_date", EntityType: "contents", DataType: "datetime"},
		},
		"content_created_at": {
			{JSONField: "content_created_at", DgraphField: "contents.created_at", EntityType: "contents", DataType: "datetime"},
		},
		"release_date": {
			{JSONField: "release_date", DgraphField: "contents.release_date", EntityType: "contents", DataType: "datetime"},
		},

		// Datetime fields for devices
		"device_last_seen": {
			{JSONField: "device_last_seen", DgraphField: "devices.last_seen", EntityType: "devices", DataType: "datetime"},
		},
		"device_created_at": {
			{JSONField: "device_created_at", DgraphField: "devices.created_at", EntityType: "devices", DataType: "datetime"},
		},
		"last_used": {
			{JSONField: "last_used", DgraphField: "devices.last_used", EntityType: "devices", DataType: "datetime"},
		},
	}
}

func getRelationships() map[string][]string {
	return map[string][]string{
		"customers": {
			"subscriptions",
			"watch_histories",
			"devices",
			"purchases",
		},
		"subscriptions": {
			"customers",
		},
		"watch_histories": {
			"customers",
			"contents",
		},
		"devices": {
			"customers",
		},
		"contents": {
			"watch_histories",
		},
		"purchases": {
			"customers",
		},
	}
}

func getDefaultFields() map[string][]string {
	return map[string][]string{
		"customers": {
			"uid",
			"customers.id",
			"customers.name",
			"customers.email",
			"customers.age",
			"customers.country",
			"customers.city",
			"customers.device",
			"customers.app_version",
			"customers.last_login_days",
			"customers.is_active",
			"customers.created_at",
		},
		"subscriptions": {
			"uid",
			"subscriptions.id",
			"subscriptions.package",
			"subscriptions.status",
			"subscriptions.start_date",
			"subscriptions.end_date",
		},
		"watch_histories": {
			"uid",
			"watch_histories.id",
			"watch_histories.content_id",
			"watch_histories.content_title",
			"watch_histories.type",
			"watch_histories.genre",
			"watch_histories.watch_date",
		},
		"contents": {
			"uid",
			"contents.id",
			"contents.title",
			"contents.type",
			"contents.genre",
			"contents.duration",
			"contents.rating",
		},
		"devices": {
			"uid",
			"devices.id",
			"devices.device_type",
			"devices.device_model",
			"devices.app_version",
			"devices.is_active",
		},
		"purchases": {
			"uid",
			"purchases.id",
			"purchases.purchasable_id",
			"purchases.status",
			"purchases.amount",
			"purchases.created_at",
		},
	}
}

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

func GetVersionFields() map[string]string {
	return map[string]string{
		"app_version": "numeric",
		"os_version":  "numeric",
		"version":     "numeric",
	}
}

func GetReversePredicates() map[string]string {
	return map[string]string{
		"subscriptions":   "~customers.subscriptions",
		"devices":         "~customers.devices",
		"watch_histories": "~customers.watch_histories",
		"contents":        "~watch_histories.content",
		"purchases":       "~customers.purchases",
	}
}

// GetFilterOptimizations returns common filter optimization patterns
func GetFilterOptimizations() map[string][]string {
	return map[string][]string{
		"subscription_status_premium": {
			"eq(subscriptions.package, \"Premium\")",
			"(eq(subscriptions.status, \"active\") OR eq(subscriptions.status, \"trial\"))",
		},
		"subscription_status_basic": {
			"eq(subscriptions.package, \"Basic\")",
			"eq(subscriptions.status, \"active\")",
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
