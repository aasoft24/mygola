// pkg/validation/validator.go
package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	Data   map[string]interface{}
	Errors map[string][]string
}

func NewValidator(data map[string]interface{}) *Validator {
	return &Validator{
		Data:   data,
		Errors: make(map[string][]string),
	}
}

func (v *Validator) Validate(rules map[string]string) bool {
	for field, ruleStr := range rules {
		rules := strings.Split(ruleStr, "|")
		value, exists := v.Data[field]

		for _, rule := range rules {
			parts := strings.SplitN(rule, ":", 2)
			ruleName := parts[0]
			var ruleValue string
			if len(parts) > 1 {
				ruleValue = parts[1]
			}

			// Skip validation if field doesn't exist and isn't required
			if !exists && ruleName != "required" {
				continue
			}

			switch ruleName {
			case "required":
				if !v.validateRequired(field, value) {
					v.addError(field, "The %s field is required", field)
				}
			case "email":
				if !v.validateEmail(field, value) {
					v.addError(field, "The %s must be a valid email address", field)
				}
			case "min":
				if !v.validateMin(field, value, ruleValue) {
					v.addError(field, "The %s must be at least %s characters", field, ruleValue)
				}
			case "max":
				if !v.validateMax(field, value, ruleValue) {
					v.addError(field, "The %s may not be greater than %s characters", field, ruleValue)
				}
			case "numeric":
				if !v.validateNumeric(field, value) {
					v.addError(field, "The %s must be a number", field)
				}
				// Add more validation rules as needed
			}
		}
	}

	return len(v.Errors) == 0
}

func (v *Validator) validateRequired(field string, value interface{}) bool {
	if value == nil {
		return false
	}

	if str, ok := value.(string); ok {
		return strings.TrimSpace(str) != ""
	}

	return true
}

func (v *Validator) validateEmail(field string, value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(str)
}

func (v *Validator) validateMin(field string, value interface{}, minStr string) bool {
	min := parseInt(minStr)

	switch val := value.(type) {
	case string:
		return utf8.RuneCountInString(val) >= min
	case int:
		return val >= min
	case float64:
		return val >= float64(min)
	default:
		return false
	}
}

func (v *Validator) validateMax(field string, value interface{}, maxStr string) bool {
	max := parseInt(maxStr)

	switch val := value.(type) {
	case string:
		return utf8.RuneCountInString(val) <= max
	case int:
		return val <= max
	case float64:
		return val <= float64(max)
	default:
		return false
	}
}

func (v *Validator) validateNumeric(field string, value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case string:
		// Check if it's a numeric string
		str := value.(string)
		matched, _ := regexp.MatchString(`^\-?\d+(\.\d+)?$`, str)
		return matched
	default:
		return false
	}
}

func (v *Validator) addError(field, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	v.Errors[field] = append(v.Errors[field], message)
}

func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v *Validator) GetErrors() map[string][]string {
	return v.Errors
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
