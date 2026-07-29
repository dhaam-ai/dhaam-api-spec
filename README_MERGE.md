# OpenAPI Specification Merger

This tool merges separate OpenAPI YAML specification files into a single consolidated specification.

## Features

- ✨ Combines multiple service specifications into one file
- 🔧 Automatically updates internal `$ref` references
- 🏷️ Adds service prefixes to avoid naming conflicts
- 📊 Provides summary statistics after merge
- 🎯 Configurable service paths and prefixes

## Prerequisites

- Go 1.21 or higher
- The `specs/` directory containing your individual YAML files:
  - `catalog.yml`
  - `order.yml`
  - `user_profile.yml`

## Installation

1. Initialize Go modules and download dependencies:

```bash
go mod download
```

## Usage

### Run the merger

```bash
go run merge_specs.go
```

This will:
1. Read all YAML files from the `specs/` directory
2. Merge them into a consolidated specification
3. Output the result to `consolidated-openapi.yml`

### Build and run as executable

```bash
# Build the executable
go build -o merge-specs merge_specs.go

# Run it
./merge-specs
```

## What it does

The merger performs the following operations:

1. **Loads each service specification** from the `specs/` directory
2. **Adds service prefixes** to schemas and parameters to avoid conflicts:
   - `Category` → `Catalog_Category`
   - `Order` → `Order_Order`
   - `UserDetails` → `UserProfile_UserDetails`
3. **Updates all internal references** (`$ref`) to use the new prefixed names
4. **Combines paths** with service-specific URL prefixes:
   - `/category` → `/catalog/v1/category`
   - `/quote` → `/order/v1/quote`
   - `/health` → `/profile/v1/health`
5. **Merges tags** with service name prefixes
6. **Outputs a single consolidated YAML file**

## Output

The script generates `consolidated-openapi.yml` containing:
- All schemas from all services (with prefixes)
- All API endpoints with proper paths
- Combined tags and parameters
- Unified security schemes
- Common response schemas

## Example Output

```
🔄 Starting OpenAPI specification merge...
📄 Processing Catalog service from catalog.yml...
✅ Merged Catalog service
📄 Processing Order service from order.yml...
✅ Merged Order service
📄 Processing UserProfile service from user_profile.yml...
✅ Merged UserProfile service

💾 Writing consolidated spec to consolidated-openapi.yml...

============================================================
✨ Merge completed successfully!
============================================================
📊 Summary:
   - Schemas: 156
   - Parameters: 15
   - Paths: 45
   - Tags: 12

📁 Output file: consolidated-openapi.yml
============================================================
```

## Customization

To modify which services are merged or change the URL prefixes, edit the `services` array in `merge_specs.go`:

```go
services := []ServiceConfig{
    {File: "catalog.yml", Name: "Catalog", Prefix: "catalog"},
    {File: "order.yml", Name: "Order", Prefix: "order"},
    {File: "user_profile.yml", Name: "UserProfile", Prefix: "profile"},
}
```

## Troubleshooting

### Missing YAML dependency error

If you see an error about missing `gopkg.in/yaml.v3`, run:

```bash
go get gopkg.in/yaml.v3
```

### Specs directory not found

Ensure you're running the script from the project root directory where the `specs/` folder exists.

### YAML parsing errors

Verify that your individual YAML files are valid OpenAPI 3.0 specifications using an online validator or tool like `swagger-cli`.

## Notes

- The original files in `specs/` are not modified
- The consolidated file can be used with OpenAPI tools like Swagger UI, Postman, or code generators
- All internal references are preserved and updated automatically
