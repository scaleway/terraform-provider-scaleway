package annotations

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

const (
	MaxAnnotationKeyNameLength   = 128
	MaxAnnotationValueNameLength = 256
)

var annotationNameRegexp = regexp.MustCompile(`^[_\pL][_\pL0-9]*$`)

func ValidateAnnotationKeyName(name string) error {
	return validateAnnotationName("annotation key", name, MaxAnnotationKeyNameLength)
}

func ValidateAnnotationValueName(name string) error {
	return validateAnnotationName("annotation value", name, MaxAnnotationValueNameLength)
}

func validateAnnotationName(kind, name string, maxLength int) error {
	if name == "" {
		return fmt.Errorf("%s name must not be empty", kind)
	}

	charCount := utf8.RuneCountInString(name)

	if charCount > maxLength {
		return fmt.Errorf("%s name must not exceed %d characters (got %d)", kind, maxLength, charCount)
	}

	if !annotationNameRegexp.MatchString(name) {
		return fmt.Errorf("%s name must start with a letter or an underscore, and contain only letters, digits and underscores", kind)
	}

	return nil
}
