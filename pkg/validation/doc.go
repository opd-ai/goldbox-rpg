// Package validation provides comprehensive input validation for JSON-RPC requests
// in the GoldBox RPG Engine.
//
// This package ensures all user inputs are sanitized and validated before processing
// to prevent security vulnerabilities, injection attacks, and denial-of-service
// conditions.
//
// # Creating a Validator
//
// Create an InputValidator with a maximum request size limit:
//
//	validator := validation.NewInputValidator(1024 * 1024) // 1MB limit
//
// # Validating Requests
//
// Validate incoming JSON-RPC requests before processing:
//
//	err := validator.ValidateRPCRequest(method, params, requestSize)
//	if err != nil {
//	    return fmt.Errorf("invalid request: %w", err)
//	}
//
// # Supported Methods
//
// The validator includes built-in validation for all standard JSON-RPC methods:
//
// Session/Player operations:
//   - ping, createPlayer, getPlayer, listPlayers
//
// Character management:
//   - createCharacter, getCharacter, updateCharacter, listCharacters
//
// Movement:
//   - move, getPosition
//
// Combat:
//   - attack, castSpell, getSpells
//
// World state:
//   - getWorld, getWorldState
//
// Equipment:
//   - equipItem, unequipItem, getInventory
//
// Other:
//   - useItem, leaveGame
//
// # Validation Rules
//
// Common validation patterns enforced:
//   - UUIDs: Must match 8-4-4-4-12 hexadecimal format
//   - Names: 1-50 characters, UTF-8, alphanumeric with limited punctuation
//   - Character classes: fighter, mage, cleric, thief, ranger, paladin
//   - Equipment slots: head, chest, main-hand, off-hand, etc.
//   - Spell IDs: Lowercase identifiers
//   - Coordinates: Range -10000 to 10000
//
// # Security Features
//
//   - Request size enforcement prevents DoS via large payloads
//   - Input sanitization prevents injection attacks
//   - Type validation prevents type confusion vulnerabilities
//   - Range validation prevents integer overflow attacks
//
// # Logging
//
// Structured logging with caller context is integrated throughout for
// debugging and security monitoring. All validation failures are logged
// with relevant context.
package validation
