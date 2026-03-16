// Package validation provides helper functions for common validation patterns
// to reduce code duplication across method-specific validators.
package validation

import (
	"fmt"
	"strings"
)

// extractParamMap extracts the parameter map from params and returns an error
// with the method name if the params are not a map.
func extractParamMap(params interface{}, methodName string) (map[string]interface{}, error) {
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s expects object parameters", methodName)
	}
	return paramMap, nil
}

// validateSessionAndExtract validates the session_id and returns the parameter map.
// Use this for methods that require a valid session_id.
func validateSessionAndExtract(params interface{}, methodName string) (map[string]interface{}, error) {
	paramMap, err := extractParamMap(params, methodName)
	if err != nil {
		return nil, err
	}
	if err := validateSessionIDFromMap(paramMap); err != nil {
		return nil, err
	}
	return paramMap, nil
}

// validateRequiredStringParam checks that a parameter exists and is a non-empty string.
// validator is optional - if nil, only type checking is performed.
func validateRequiredStringParam(paramMap map[string]interface{}, paramName, methodName string, validator func(string) error) error {
	value, exists := paramMap[paramName]
	if !exists {
		return fmt.Errorf("%s requires '%s' parameter", methodName, paramName)
	}

	valueStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s must be a string", paramName)
	}

	if validator != nil {
		return validator(valueStr)
	}
	return nil
}

// validateRequiredNumericParam checks that a parameter exists and is a number
// within the specified range (min, max inclusive).
func validateRequiredNumericParam(paramMap map[string]interface{}, paramName, methodName string, min, max float64) (float64, error) {
	value, exists := paramMap[paramName]
	if !exists {
		return 0, fmt.Errorf("%s requires '%s' parameter", methodName, paramName)
	}

	valueNum, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("%s must be a number", paramName)
	}

	if valueNum < min || valueNum > max {
		return 0, fmt.Errorf("%s must be between %.0f and %.0f", paramName, min, max)
	}
	return valueNum, nil
}

// validateOptionalStringParam checks an optional string parameter and validates it if present.
// Returns the string value and whether it was present. Errors if present but invalid.
func validateOptionalStringParam(paramMap map[string]interface{}, paramName string, validator func(string) error) (string, bool, error) {
	value, exists := paramMap[paramName]
	if !exists {
		return "", false, nil
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", true, fmt.Errorf("%s must be a string", paramName)
	}

	if validator != nil {
		if err := validator(valueStr); err != nil {
			return "", true, err
		}
	}
	return valueStr, true, nil
}

// validateEnumParam checks that a string parameter is one of the allowed values.
func validateEnumParam(value string, allowedValues []string, paramName string) error {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return fmt.Errorf("invalid %s: %s", paramName, value)
}

// validateNonEmpty is a validator function that checks a string is not empty.
func validateNonEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

// sessionRequiredValidatorFunc returns a validator that requires a session_id parameter.
// Use for methods that need only a valid session_id and nothing else.
func sessionRequiredValidatorFunc() func(interface{}) error {
	return func(params interface{}) error {
		return validateSessionID(params)
	}
}

// sessionAndExtractValidatorFunc returns a validator that validates session_id
// and extracts the parameter map, discarding it (since nothing else is checked).
// Use for methods that validate session + extract but have no additional param checks.
func sessionAndExtractValidatorFunc(methodName string) func(interface{}) error {
	return func(params interface{}) error {
		_, err := validateSessionAndExtract(params, methodName)
		return err
	}
}

// optionalSessionValidatorFunc returns a validator for methods where params are optional.
// If params are present and contain session_id, it validates the session_id.
// Otherwise, it accepts the request without error.
func optionalSessionValidatorFunc() func(interface{}) error {
	return func(params interface{}) error {
		if params == nil {
			return nil
		}
		paramMap, ok := params.(map[string]interface{})
		if !ok {
			return nil
		}
		if _, exists := paramMap["session_id"]; exists {
			return validateSessionIDFromMap(paramMap)
		}
		return nil
	}
}

// validateOptionalAppearanceFields validates optional cosmetic/biographical
// appearance fields in a createCharacter request. All fields are accepted
// but never required.
func validateOptionalAppearanceFields(paramMap map[string]interface{}) error {
	if skinTone, ok := paramMap["skin_tone"]; ok {
		v, vOk := skinTone.(float64)
		if !vOk || v < 1 || v > 10 {
			return fmt.Errorf("createCharacter: skin_tone must be 1-10")
		}
	}
	if bodyType, ok := paramMap["body_type"]; ok {
		v, vOk := bodyType.(float64)
		if !vOk || v < 0 || v > 5 {
			return fmt.Errorf("createCharacter: body_type must be 0-5")
		}
	}
	for _, field := range []string{
		"pronouns", "gender_expression",
		"hair_style", "hair_color",
		"romantic_orientation",
	} {
		if val, ok := paramMap[field]; ok {
			s, sOk := val.(string)
			if !sOk || len(s) > 100 {
				return fmt.Errorf(
					"createCharacter: %s must be a string ≤100 characters",
					field,
				)
			}
		}
	}
	return nil
}
