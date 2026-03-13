// Package main provides the openapi-gen CLI tool for generating OpenAPI specifications
// from Go source code.
//
// The tool parses RPC method constants from pkg/server/constants.go and updates
// api/openapi.yaml with the complete list of available JSON-RPC methods, organized
// by feature group.
//
// # Usage
//
//	openapi-gen -package pkg/server -spec api/openapi.yaml
//
// # Flags
//
//	-package: Package path to analyze (default: "pkg/server")
//	-spec: Path to OpenAPI spec file (default: "api/openapi.yaml")
//
// # Features
//
// The generator:
//   - Extracts RPC method names and their handler functions from constants
//   - Categorizes methods by feature group (Combat, Quest, Character, etc.)
//   - Preserves manual edits to request/response schemas
//   - Automatically syncs method lists with code changes
//
// # Integration
//
// Run via make target:
//
//	make openapi-gen
package main
