package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestYAMLConfigLoader_Load(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		wantErr  bool
		validate func(*testing.T, *Config)
	}{
		{
			name: "valid config",
			config: `
users:
  path_pattern: "/api/users/*"
  body_id_paths:
    - "data.user.id"
    - "user.id"
    - "id"
  header_id_name: "X-User-ID"
`,
			wantErr: false,
			validate: func(t *testing.T, config *Config) {
				section, exists := config.Sections["users"]
				if !exists {
					t.Fatal("Expected 'users' section to exist")
				}

				if section.PathPattern != "/api/users/*" {
					t.Errorf("Expected path pattern '/api/users/*', got '%s'", section.PathPattern)
				}

				if len(section.BodyIDPaths) != 3 {
					t.Errorf("Expected 3 body ID paths, got %d", len(section.BodyIDPaths))
				}

				if section.HeaderIDName != "X-User-ID" {
					t.Errorf("Expected header ID name 'X-User-ID', got '%s'", section.HeaderIDName)
				}
			},
		},
		{
			name: "multiple sections",
			config: `
users:
  path_pattern: "/api/users/*"
  body_id_paths:
    - "data.user.id"
  header_id_name: "X-User-ID"
orders:
  path_pattern: "/api/orders/*"
  body_id_paths:
    - "data.order.id"
  header_id_name: "X-Order-ID"
`,
			wantErr: false,
			validate: func(t *testing.T, config *Config) {
				if len(config.Sections) != 2 {
					t.Errorf("Expected 2 sections, got %d", len(config.Sections))
				}

				_, usersExists := config.Sections["users"]
				if !usersExists {
					t.Error("Expected 'users' section to exist")
				}

				_, ordersExists := config.Sections["orders"]
				if !ordersExists {
					t.Error("Expected 'orders' section to exist")
				}
			},
		},
		{
			name: "invalid yaml",
			config: `
users:
  path_pattern: "/api/users/*"
  body_id_paths:
    - "data.user.id"
  header_id_name: "X-User-ID"
  invalid_field: true
`,
			wantErr:  true,
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary test config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "test_config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.config), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Test loading
			loader := &YAMLConfigLoader{}
			config, err := loader.Load(configPath)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

func TestConfig_MatchPath(t *testing.T) {
	config := &Config{
		Sections: map[string]Section{
			"users": {
				PathPattern:  "/api/users/*",
				BodyIDPaths:  []string{"data.user.id"},
				HeaderIDName: "X-User-ID",
			},
			"orders": {
				PathPattern:  "/api/orders/*",
				BodyIDPaths:  []string{"data.order.id"},
				HeaderIDName: "X-Order-ID",
			},
			"deep": {
				PathPattern:  "/api/users/*/orders/*",
				BodyIDPaths:  []string{"data.order.id"},
				HeaderIDName: "X-Order-ID",
			},
		},
	}

	tests := []struct {
		name            string
		path            string
		wantSection     string
		wantErr         bool
		validateSection func(*testing.T, *Section)
	}{
		{
			name:        "match users path",
			path:        "/api/users/123",
			wantSection: "users",
			wantErr:     false,
			validateSection: func(t *testing.T, section *Section) {
				if section.HeaderIDName != "X-User-ID" {
					t.Errorf("Expected header ID name 'X-User-ID', got '%s'", section.HeaderIDName)
				}
			},
		},
		{
			name:        "match orders path",
			path:        "/api/orders/456",
			wantSection: "orders",
			wantErr:     false,
			validateSection: func(t *testing.T, section *Section) {
				if section.HeaderIDName != "X-Order-ID" {
					t.Errorf("Expected header ID name 'X-Order-ID', got '%s'", section.HeaderIDName)
				}
			},
		},
		{
			name:        "match deep path",
			path:        "/api/users/123/orders/456",
			wantSection: "deep",
			wantErr:     false,
			validateSection: func(t *testing.T, section *Section) {
				if section.HeaderIDName != "X-Order-ID" {
					t.Errorf("Expected header ID name 'X-Order-ID', got '%s'", section.HeaderIDName)
				}
			},
		},
		{
			name:            "no match",
			path:            "/api/products/789",
			wantSection:     "",
			wantErr:         false,
			validateSection: nil,
		},
		{
			name:        "trailing slash",
			path:        "/api/users/123/",
			wantSection: "users",
			wantErr:     false,
			validateSection: func(t *testing.T, section *Section) {
				if section.HeaderIDName != "X-User-ID" {
					t.Errorf("Expected header ID name 'X-User-ID', got '%s'", section.HeaderIDName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sectionName, section, err := config.MatchPath(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if sectionName != tt.wantSection {
				t.Errorf("Expected section name '%s', got '%s'", tt.wantSection, sectionName)
			}

			if tt.validateSection != nil {
				if section == nil {
					t.Fatal("Expected section to be non-nil")
				}
				tt.validateSection(t, section)
			}
		})
	}
}
