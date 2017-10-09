package azurerm

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/satori/uuid"
)

func validateRFC3339Date(v interface{}, k string) (ws []string, errors []error) {
	dateString := v.(string)

	if _, err := date.ParseTime(time.RFC3339, dateString); err != nil {
		errors = append(errors, fmt.Errorf("%q is an invalid RFC3339 date: %+v", k, err))
	}

	return
}

// validateIntInSlice returns a SchemaValidateFunc which tests if the provided value
// is of type int and matches the value of an element in the valid slice
func validateIntInSlice(valid []int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		for _, str := range valid {
			if v == str {
				return
			}
		}

		es = append(es, fmt.Errorf("expected %q to be one of %v, got %v", k, valid, v))
		return
	}
}

func validateUUID(v interface{}, k string) (ws []string, errors []error) {
	if _, err := uuid.FromString(v.(string)); err != nil {
		errors = append(errors, fmt.Errorf("%q is an invalid UUUID: %s", k, err))
	}
	return
}

func validateDBAccountName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	r, _ := regexp.Compile("^[a-z0-9\\-]+$")
	if !r.MatchString(value) {
		errors = append(errors, fmt.Errorf("Account Name can only contain lower-case characters, numbers and the `-` character."))
	}

	length := len(value)
	if length > 50 || 3 > length {
		errors = append(errors, fmt.Errorf("Account Name can only be between 3 and 50 seconds."))
	}

	return
}

func validateStringLength(maxLength int) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		value := v.(string)
		if len(value) > maxLength {
			errors = append(errors, fmt.Errorf(
				"The %q can be no longer than %d chars", k, maxLength))
		}
		return
	}
}
