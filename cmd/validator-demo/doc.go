// Package main provides a demonstration application for the PCG content validation system.
//
// The validator-demo application showcases how to use the content validation
// system in the GoldBox RPG engine. It demonstrates:
//
//   - Validating character attributes and names
//   - Automatic fixing of invalid content using fallback handlers
//   - Quest validation and objective enforcement
//   - Accessing validation metrics for monitoring and reporting
//
// # Usage
//
// Run the demo directly:
//
//	go run ./cmd/validator-demo
//
// Or build and execute:
//
//	go build -o validator-demo ./cmd/validator-demo
//	./validator-demo
//
// # Validation Scenarios
//
// The demo includes several validation scenarios:
//
// 1. Valid Character - Demonstrates that properly configured characters pass validation
//
// 2. Invalid Character - Shows how the validator detects and automatically fixes:
//   - Out-of-range attribute values (clamped to 3-25)
//   - Empty character names (generates fallback name)
//
// 3. Invalid Quest - Demonstrates quest validation and automatic addition of
//    default objectives when missing
//
// 4. Validation Metrics - Shows how to access validation statistics including
//    success rates, average validation times, and critical failure counts
//
// # Integration Example
//
// The validation system can be integrated into game content pipelines:
//
//	validator := pcg.NewContentValidator(logger)
//	results, err := validator.ValidateContent(ctx, pcg.ContentTypeCharacters, character)
//	if err != nil {
//	    // Handle validation error
//	}
//	for _, result := range results {
//	    if !result.Passed {
//	        log.Printf("Validation failed: %s", result.Message)
//	    }
//	}
//
// For automatic fixing of validation issues:
//
//	fixedContent, results, err := validator.ValidateAndFix(ctx, contentType, content)
package main
