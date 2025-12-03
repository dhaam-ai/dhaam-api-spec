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

Your consolidated specification will be at: `consolidated-openapi.yml`

---

## What Gets Merged?

| Source File | Service Name | URL Prefix | Schemas Count |
|------------|--------------|------------|---------------|
| `specs/catalog.yml` | Catalog | `/catalog/v1` | ~40 |
| `specs/order.yml` | Order | `/order/v1` | ~15 |
| `specs/user_profile.yml` | UserProfile | `/profile/v1` | ~30 |

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

Paths are prefixed with service paths:

| Original | Consolidated |
|----------|--------------|
| `/category` | `/catalog/v1/category` |
| `/item` | `/catalog/v1/item` |
| `/quote` | `/order/v1/quote` |
| `/create` | `/order/v1/create` |
| `/health` | `/profile/v1/health` |
| `/tenant/register` | `/profile/v1/tenant/register` |

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
swagger-cli validate consolidated-openapi.yml
```

---

## Using the Consolidated Spec

### With Swagger UI (Docker)
```bash
docker run -p 8080:8080 \
  -e SWAGGER_JSON=/app/consolidated-openapi.yml \
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
  -i consolidated-openapi.yml \
  -g typescript-axios \
  -o ./generated/typescript-client

# Generate Go client
openapi-generator-cli generate \
  -i consolidated-openapi.yml \
  -g go \
  -o ./generated/go-client
```

### Import into Postman
1. Open Postman
2. Click "Import"
3. Select `consolidated-openapi.yml`
4. All endpoints will be available in a new collection

---

## Need Help?

📧 Contact: support@dhaam.ai
📖 Full Documentation: See `README_MERGE.md`
