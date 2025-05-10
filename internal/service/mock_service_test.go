package service

import (
	"github.com/bmcszk/unimock/pkg/config"
)

// createTestConfig creates a config for testing
func createTestConfig() *config.MockConfig {
	return &config.MockConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern: "/api/users/*",
				BodyIDPaths: []string{"/id"},
			},
			"user_items": {
				PathPattern: "/api/users/*/items",
				BodyIDPaths: []string{"/id"},
			},
			"items_collection": {
				PathPattern: "/api/items/collection/*",
				BodyIDPaths: []string{"/id"},
			},
			"collections": {
				PathPattern: "/api/collections/*",
				BodyIDPaths: []string{"/id"},
			},
			"section_collections": {
				PathPattern: "/api/collections/section/*",
				BodyIDPaths: []string{"/id"},
			},
			"orders": {
				PathPattern: "/api/orders/*",
			},
			"products": {
				PathPattern: "/api/products/*",
				BodyIDPaths: []string{"/id"},
			},
		},
	}
}
