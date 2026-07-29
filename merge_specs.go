package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ServiceConfig defines the configuration for each service to merge
type ServiceConfig struct {
	File   string
	Name   string
	Prefix string
}

// OpenAPISpec represents the structure of an OpenAPI specification
type OpenAPISpec struct {
	OpenAPI    string                 `yaml:"openapi"`
	Info       map[string]interface{} `yaml:"info"`
	Servers    []interface{}          `yaml:"servers,omitempty"`
	Security   []interface{}          `yaml:"security,omitempty"`
	Components map[string]interface{} `yaml:"components,omitempty"`
	Paths      map[string]interface{} `yaml:"paths,omitempty"`
	Tags       []interface{}          `yaml:"tags,omitempty"`
}

func main() {
	specsDir := "specs"
	outputFile := "consolidated-openapi.yml"

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		log.Fatalf("❌ Error: specs directory not found at %s", specsDir)
	}

	fmt.Println("🔄 Starting OpenAPI specification merge...")

	// Merge specs
	if err := mergeSpecs(specsDir, outputFile); err != nil {
		log.Fatalf("❌ Error during merge: %v", err)
	}
}

func mergeSpecs(specsDir, outputFile string) error {
	// Initialize consolidated spec
	consolidated := &OpenAPISpec{
		OpenAPI: "3.0.1",
		Info: map[string]interface{}{
			"title": "Dhaam Platform API - Consolidated",
			"description": `Unified OpenAPI 3.0 specification for Dhaam Platform.

This specification consolidates multiple microservices:
- **Catalog Service**: Category and item management (products, modifiers, variants, bundles)
- **Order Service**: Order and quotation management
- **User Profile Service**: User profiles, customers, merchants, stores, locations (regions, geofences, outlets)

All specifications have been merged into a single file for easier consumption.`,
			"version": "2.0.0",
			"contact": map[string]interface{}{
				"name":  "Dhaam API Support",
				"email": "support@dhaam.ai",
			},
		},
		Servers: []interface{}{
			map[string]interface{}{
				"url":         "https://api.dhaam.ai",
				"description": "Production server",
			},
			map[string]interface{}{
				"url":         "https://staging-api.dhaam.ai",
				"description": "Staging server",
			},
			map[string]interface{}{
				"url":         "https://dev-nexus.dhaamai.com/api/v1",
				"description": "Development server",
			},
		},
		Security: []interface{}{
			map[string]interface{}{"BearerAuth": []interface{}{}},
			map[string]interface{}{"ApiKeyAuth": []interface{}{}},
		},
		Components: map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
					"description":  "JWT Bearer token authentication",
				},
				"ApiKeyAuth": map[string]interface{}{
					"type":        "apiKey",
					"in":          "header",
					"name":        "X-API-Key",
					"description": "API Key authentication",
				},
			},
			"schemas":    map[string]interface{}{},
			"parameters": map[string]interface{}{},
		},
		Paths: map[string]interface{}{},
		Tags:  []interface{}{},
	}

	// Define services to merge
	services := []ServiceConfig{
		{File: "catalog.yml", Name: "Catalog", Prefix: "catalog"},
		{File: "order.yml", Name: "Order", Prefix: "order"},
		{File: "user_profile.yml", Name: "UserProfile", Prefix: "profile"},
	}

	// Merge each service
	for _, service := range services {
		servicePath := filepath.Join(specsDir, service.File)

		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			fmt.Printf("⚠️  Warning: %s not found, skipping...\n", service.File)
			continue
		}

		fmt.Printf("📄 Processing %s service from %s...\n", service.Name, service.File)

		// Load service spec
		serviceSpec, err := loadYAML(servicePath)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", service.File, err)
		}

		// Update internal references
		updateRefs(serviceSpec, service.Name)

		// Merge components
		mergeComponents(consolidated, serviceSpec, service.Name)

		// Merge paths
		mergePaths(consolidated, serviceSpec, service.Prefix)

		// Merge tags
		mergeTags(consolidated, serviceSpec, service.Name)

		fmt.Printf("✅ Merged %s service\n", service.Name)
	}

	// Add common response schemas
	addCommonSchemas(consolidated)

	// Write consolidated spec
	fmt.Printf("\n💾 Writing consolidated spec to %s...\n", outputFile)

	if err := writeYAML(outputFile, consolidated); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Print summary
	printSummary(consolidated, outputFile)

	return nil
}

func loadYAML(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func writeYAML(path string, data interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	defer encoder.Close()

	return encoder.Encode(data)
}

func updateRefs(obj interface{}, serviceName string) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if key == "$ref" {
				if str, ok := value.(string); ok {
					if strings.HasPrefix(str, "#/components/schemas/") {
						schemaName := strings.TrimPrefix(str, "#/components/schemas/")
						v[key] = fmt.Sprintf("#/components/schemas/%s_%s", serviceName, schemaName)
					} else if strings.HasPrefix(str, "#/components/parameters/") {
						paramName := strings.TrimPrefix(str, "#/components/parameters/")
						v[key] = fmt.Sprintf("#/components/parameters/%s_%s", serviceName, paramName)
					}
				}
			} else {
				updateRefs(value, serviceName)
			}
		}
	case []interface{}:
		for _, item := range v {
			updateRefs(item, serviceName)
		}
	}
}

func mergeComponents(target *OpenAPISpec, source map[string]interface{}, serviceName string) {
	sourceComponents, ok := source["components"].(map[string]interface{})
	if !ok {
		return
	}

	targetComponents := target.Components

	// Merge schemas
	if schemas, ok := sourceComponents["schemas"].(map[string]interface{}); ok {
		targetSchemas := targetComponents["schemas"].(map[string]interface{})
		for schemaName, schemaDef := range schemas {
			prefixedName := fmt.Sprintf("%s_%s", serviceName, schemaName)
			targetSchemas[prefixedName] = schemaDef
		}
	}

	// Merge parameters
	if params, ok := sourceComponents["parameters"].(map[string]interface{}); ok {
		targetParams := targetComponents["parameters"].(map[string]interface{})
		for paramName, paramDef := range params {
			prefixedName := fmt.Sprintf("%s_%s", serviceName, paramName)
			targetParams[prefixedName] = paramDef
		}
	}

	// Merge security schemes (without duplicates)
	if secSchemes, ok := sourceComponents["securitySchemes"].(map[string]interface{}); ok {
		targetSecSchemes := targetComponents["securitySchemes"].(map[string]interface{})
		for schemeName, schemeDef := range secSchemes {
			if _, exists := targetSecSchemes[schemeName]; !exists {
				targetSecSchemes[schemeName] = schemeDef
			}
		}
	}
}

func mergePaths(target *OpenAPISpec, source map[string]interface{}, prefix string) {
	sourcePaths, ok := source["paths"].(map[string]interface{})
	if !ok {
		return
	}

	for path, pathDef := range sourcePaths {
		fullPath := fmt.Sprintf("/%s/v1%s", prefix, path)
		target.Paths[fullPath] = pathDef
	}
}

func mergeTags(target *OpenAPISpec, source map[string]interface{}, serviceName string) {
	sourceTags, ok := source["tags"].([]interface{})
	if !ok {
		return
	}

	for _, tag := range sourceTags {
		if tagMap, ok := tag.(map[string]interface{}); ok {
			prefixedTag := map[string]interface{}{
				"name": fmt.Sprintf("%s - %s", serviceName, tagMap["name"]),
			}
			if desc, exists := tagMap["description"]; exists {
				prefixedTag["description"] = desc
			}
			target.Tags = append(target.Tags, prefixedTag)
		}
	}
}

func addCommonSchemas(target *OpenAPISpec) {
	schemas := target.Components["schemas"].(map[string]interface{})

	schemas["SuccessResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":    "boolean",
				"example": true,
			},
			"data": map[string]interface{}{
				"type":        "object",
				"description": "Response payload (varies by endpoint)",
			},
			"meta": map[string]interface{}{
				"type":        "object",
				"description": "Metadata (pagination, timestamps, etc.)",
			},
		},
	}

	schemas["ErrorResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":    "boolean",
				"example": false,
			},
			"error": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Error code",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Human-readable error message",
					},
				},
			},
		},
	}
}

func printSummary(spec *OpenAPISpec, outputFile string) {
	schemas := spec.Components["schemas"].(map[string]interface{})
	params := spec.Components["parameters"].(map[string]interface{})

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✨ Merge completed successfully!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 Summary:")
	fmt.Printf("   - Schemas: %d\n", len(schemas))
	fmt.Printf("   - Parameters: %d\n", len(params))
	fmt.Printf("   - Paths: %d\n", len(spec.Paths))
	fmt.Printf("   - Tags: %d\n", len(spec.Tags))
	fmt.Printf("\n📁 Output file: %s\n", outputFile)
	fmt.Println(strings.Repeat("=", 60))
}
