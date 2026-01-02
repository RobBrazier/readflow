set dotenv-load := true

# Default task
@default:
    just --list

# Build the binary
build: generate
    CGO_ENABLED=0 go build -o readflow main.go

# Generate GraphQL clients
generate:
    go generate ./...

# Run tests
test:
    go test ./...

# Download GraphQL schemas
download-schemas:
    gquil introspection generate-sdl https://graphql.anilist.co > schemas/anilist/schema.gql
    gquil introspection generate-sdl https://api.hardcover.app/v1/graphql -H 'Authorization: Bearer {{env_var("HARDCOVER_TOKEN")}}' > schemas/hardcover/schema.gql

# Bump version
version:
    #!/usr/bin/env bash
    set -euo pipefail
    NEW_VERSION=$(git cliff --bumped-version)
    read -p "Bumping version to $NEW_VERSION, continue? "
    git tag -f "$NEW_VERSION"
