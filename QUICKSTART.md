# Quick Start Guide

## Merge OpenAPI Specs in 3 Steps

### 1. Install Dependencies

```bash
go mod download
```

### 2. Run the Merger

Choose one of these methods:

**Option A: Using Make (recommended)**
```bash
make merge
```

**Option B: Using Go directly**
```bash
go run merge_specs.go
```

**Option C: Build and run**
```bash
go build -o merge-specs merge_specs.go
./merge-specs
```

### 3. Check Output

Your consolidated specification will be at: `specs/openapi.yml`

---

## What Gets Merged?

| Source File | Service Name | URL Prefix | Schemas Count |
|------------|--------------|------------|---------------|
| `specs/auth.yml` | Auth | full public paths under `/api/v1/auth` | per file |
| `specs/user_profile.yml` | UserProfile | full public paths under `/api/v1/users` | per file |
| `specs/profile.yml` | Profile | full public paths under `/api/v1/profile` | per file |
| `specs/catalog.yml` | Catalog | full public paths under `/api/v1/catalog` | per file |
| `specs/menu.yml` | Menu | full public paths under `/api/v1/catalog/menu` | per file |
| `specs/order.yml` | Order | full public paths under `/api/v1/orders` | per file |
| `specs/gateway.yml` | Gateway | full public path `/api/v1/storefront/graphql` | per file |

Paths in each file are already full public paths, so the merge copies them unchanged. A path that
two files define is a merge error. Missing files are skipped with a warning.

**Total**: 88+ schemas, 67+ endpoints, 13+ parameters, 14+ tags

---

## Schema Naming Convention

Schemas are prefixed with their service name:

- `Category` → `Catalog_Category`
- `Order` → `Order_Order`  
- `UserDetails` → `UserProfile_UserDetails`
- `Address` → `UserProfile_Address`

---

## Path Transformation

Paths are not rewritten. Each per-service file already uses its full public path, for example
`/api/v1/catalog/item/list` in `specs/catalog.yml`, so the merged file uses the same paths.

---

## Common Commands

```bash
# View all available make targets
make help

# Install dependencies
make install

# Merge specs
make merge

# Build executable
make build

# Clean generated files
make clean

# Validate output (requires swagger-cli)
make validate
```

---

## Troubleshooting

**Error: specs directory not found**
```bash
# Make sure you're in the project root
cd /path/to/dhaam-api-spec
```

**Error: missing go.sum entry**
```bash
go mod download gopkg.in/yaml.v3
```

**Want to validate the output?**
```bash
# Install swagger-cli
npm install -g @apidevtools/swagger-cli

# Validate
swagger-cli validate specs/openapi.yml
```

---

## Using the Consolidated Spec

### With Swagger UI (Docker)
```bash
docker run -p 8080:8080 \
  -e SWAGGER_JSON=/app/specs/openapi.yml \
  -v $(pwd):/app \
  swaggerapi/swagger-ui
```

Then open: http://localhost:8080

### Generate Client Code
```bash
# Install openapi-generator
npm install -g @openapitools/openapi-generator-cli

# Generate TypeScript client
openapi-generator-cli generate \
  -i specs/openapi.yml \
  -g typescript-axios \
  -o ./generated/typescript-client

# Generate Go client
openapi-generator-cli generate \
  -i specs/openapi.yml \
  -g go \
  -o ./generated/go-client
```

### Import into Postman
1. Open Postman
2. Click "Import"
3. Select `specs/openapi.yml`
4. All endpoints will be available in a new collection

---

## Need Help?

📧 Contact: support@dhaam.ai
📖 Full Documentation: See `README_MERGE.md`
