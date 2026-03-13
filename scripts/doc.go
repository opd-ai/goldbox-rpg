// Package main provides build and development scripts for the GoldBox RPG Engine.
//
// This directory contains Go-based utility scripts for project maintenance:
//
// # find_untested_files.go
//
// Identifies Go source files that lack corresponding test files:
//
//	go run scripts/find_untested_files.go ./pkg/...
//
// # verify_adventures.go
//
// Validates adventure YAML files and generates status reports:
//
//	go run scripts/verify_adventures.go
//
// # Shell Scripts
//
// The directory also contains shell scripts for various tasks:
//   - analyze_test_coverage.sh: Test coverage analysis and reporting
//   - generate-all.sh: Complete asset generation pipeline
//   - generate-priority1.sh: Priority asset generation for quick testing
//   - verify-assets.sh: Asset verification and validation
//   - post-process.sh: Asset optimization for production
//
// # Usage
//
// Most scripts are invoked via make targets:
//
//	make test-coverage      # Run analyze_test_coverage.sh
//	make find-untested      # Run find_untested_files.go
//	make assets             # Run generate-all.sh
//	make assets-verify      # Run verify-assets.sh
//
// Scripts in this package use //go:build ignore tags and are not part of the
// main build. They should be run directly with go run or via make targets.
package main
