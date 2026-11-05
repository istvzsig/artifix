package artifact

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
)

type Coordinates struct {
	namespace string
	name      string
	version   string
}

func NewCoordinates(namespace, name, version string) (Coordinates, error) {
	if err := validateIdentifier(namespace); err != nil {
		return Coordinates{}, err
	}

	if err := validateIdentifier(name); err != nil {
		return Coordinates{}, err
	}

	if err := validateVersion(version); err != nil {
		return Coordinates{}, err
	}

	return Coordinates{
		namespace: namespace,
		name:      name,
		version:   version,
	}, nil
}

func (c Coordinates) IsValid() bool {
	return c.namespace != "" &&
		c.name != "" &&
		c.version != ""
}

func validateIdentifier(input string) error {
	if input == "" {
		return errors.New("identifier cannot be empty")
	}

	for _, char := range input {
		if char >= 'a' && char <= 'z' {
			continue
		}

		if char >= '0' && char <= '9' {
			continue
		}

		if isSeparator(char) {
			continue
		}

		return fmt.Errorf("invalid character %q", char)
	}

	first := input[0]
	if isSeparator(rune(first)) {
		return fmt.Errorf("invalid separator %q", first)
	}

	last := input[len(input)-1]
	if isSeparator(rune(last)) {
		return fmt.Errorf("invalid separator %q", last)
	}

	return nil
}

func isSeparator(char rune) bool {
	return char == '-' || char == '_' || char == '.'
}

func validateVersion(input string) error {
	if input == "" {
		return errors.New("version cannot be empty")
	}

	if _, err := semver.StrictNewVersion(input); err != nil {
		return fmt.Errorf("invalid semantic version %q: %w", input, err)
	}

	return nil
}
