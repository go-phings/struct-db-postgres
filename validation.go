package structdbpostgres

import (
	validator "github.com/mikolajgs/struct-validator"
	"reflect"
	"regexp"
	"strings"
)

// Validate checks object's fields. It returns result of validation as a bool and list of fields with invalid value
func (c Controller) Validate(obj interface{}, filters map[string]interface{}) (bool, map[string]int, error) {
	if filters != nil {
		restrictFilters := make(map[string]bool)
		failedFiltersLike := make(map[string]int)
		for k, v := range filters {
			if !strings.Contains(k, ":") {
				restrictFilters[k] = true
				continue
			}
			kArr := strings.Split(k, ":")
			if kArr[1] != "%" {
				continue
			}
			r := reflect.TypeOf(v).Kind()
			if r != reflect.String {
				// TODO Use validator.FailType when added
				failedFiltersLike[k] = validator.FailRegexp
				continue
			}
			reLike := regexp.MustCompile(`^[a-zA-Z0-9\-+_\.: \$@!=|\[\]\{\}<>,\?/]{0,32}$`)
			if !reLike.MatchString(v.(string)) {
				failedFiltersLike[k] = validator.FailRegexp
				continue
			}
		}
		valid, failedFields := validator.Validate(obj, &validator.ValidationOptions{
			ValidateWhenSuffix:   true,
			OverwriteFieldValues: filters,
			RestrictFields:       restrictFilters,
			OverwriteTagName:     c.tagName,
		})
		for k, v := range failedFiltersLike {
			failedFields[k] = v
			valid = true
		}
		return valid, failedFields, nil
	}

	valid, failedFields := validator.Validate(obj, &validator.ValidationOptions{
		ValidateWhenSuffix: true,
		OverwriteTagName:   c.tagName,
	})
	return valid, failedFields, nil
}
