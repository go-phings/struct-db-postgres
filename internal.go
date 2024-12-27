package structdbpostgres

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	stsql "github.com/go-phings/struct-sql-postgres"
)

// getSQLGenerator returns a special StructSQL instance which reflects the struct type to get SQL queries etc.
func (c *Controller) getSQLGenerator(obj interface{}, generators map[string]*stsql.StructSQL, forceName string) (*stsql.StructSQL, error) {
	n := c.getSQLGeneratorName(obj, false)
	if c.sqlGenerators[n] == nil {
		h := stsql.NewStructSQL(obj, stsql.StructSQLOptions{
			DatabaseTablePrefix:          c.dbTblPrefix,
			TagName:                      c.tagName,
			Joined:                       generators,
			ForceName:                    forceName,
			UseRootNameWhenJoinedPresent: true,
		})
		if h.Err() != nil {
			return nil, ErrController{
				Op:  "GetHelper",
				Err: fmt.Errorf("Error getting StructSQL: %w", h.Err()),
			}
		}
		c.sqlGenerators[n] = h
	}
	return c.sqlGenerators[n], nil
}

func (c *Controller) getSQLGeneratorName(obj interface{}, onlyRoot bool) string {
	v := reflect.ValueOf(obj)
	i := reflect.Indirect(v)
	s := i.Type()

	var n string

	if s.String() == "reflect.Value" {
		s = reflect.ValueOf(obj.(reflect.Value).Interface()).Type().Elem().Elem()
		n = s.Name()
		if strings.Contains(s.Name(), ".") {
			sArr := strings.Split(s.Name(), ".")
			n = sArr[1]
		}
	} else {
		n = s.Name()
	}

	if onlyRoot && strings.Contains(n, "_") {
		nArr := strings.SplitN(n, "_", 1)
		return nArr[0]
	}

	return n
}

func (c Controller) mapWithInterfacesToMapBool(m map[string]interface{}) map[string]bool {
	o := map[string]bool{}
	for k := range m {
		o[k] = true
	}
	return o
}

func (c Controller) runOnDelete(obj interface{}, tagName string, ids []int64, lastDepth int) error {
	v := reflect.ValueOf(obj)
	i := reflect.Indirect(v)
	s := i.Type()

	var isReflectValue bool
	if s.String() == "reflect.Value" {
		isReflectValue = true
		s = reflect.ValueOf(obj.(reflect.Value).Interface()).Type().Elem().Elem()
	}

	var structName string
	if !isReflectValue {
		structName = s.Name()
	} else {
		if strings.Contains(s.Name(), ".") {
			sArr := strings.Split(s.Name(), ".")
			structName = sArr[1]
		}
	}
	parentIDField := structName + "ID"

	tagRegexp := regexp.MustCompile(`[a-zA-Z0-9_]+\:[a-zA-Z0-9_-]+`)

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		k := f.Type.Kind()

		// Only field which are slices of pointers to struct instances
		if k != reflect.Slice || f.Type.Elem().Kind() != reflect.Ptr || f.Type.Elem().Elem().Kind() != reflect.Struct {
			continue
		}

		// Get 2db tag, loop through its value on determine action based on it
		tag := s.Field(i).Tag.Get(tagName)
		if tag == "" {
			continue
		}

		tags := strings.Split(tag, " ")
		tagsMap := map[string]string{}
		for _, t := range tags {
			if m := tagRegexp.MatchString(t); m {
				mArr := strings.Split(t, ":")
				tagsMap[mArr[0]] = mArr[1]
			}
		}

		// Perform delete
		if tagsMap["on_del"] == "del" {
			if tagsMap["del_field"] != "" {
				parentIDField = tagsMap["del_field"]
			}

			// Delete from children table where parent ID = id of deleted object
			errCtl := c.DeleteMultiple(reflect.New(f.Type.Elem()), DeleteMultipleOptions{
				Filters: map[string]interface{}{
					"_raw": []interface{}{
						fmt.Sprintf(".%s IN (?)", parentIDField),
						ids,
					},
				},
				CascadeDeleteDepth: lastDepth + 1,
			})
			if errCtl != nil {
				return ErrController{
					Op:  "CascadeDelete",
					Err: errors.New("Error from DeleteMultiple"),
				}
			}
		}

		// Perform update
		if tagsMap["on_del"] == "upd" {
			updField := tagsMap["del_upd_field"]
			updValue := tagsMap["del_upd_val"]
			if updField == "" {
				return ErrController{
					Op:  "CascadeDelete",
					Err: errors.New("missing update field in tags"),
				}
			}

			if tagsMap["del_field"] != "" {
				parentIDField = tagsMap["del_field"]
			}
			// Update children table where parent ID = id of deleted object
			errCtl := c.UpdateMultiple(reflect.New(f.Type.Elem()),
				map[string]interface{}{
					updField: updValue,
				},
				UpdateMultipleOptions{
					Filters: map[string]interface{}{
						"_raw": []interface{}{
							fmt.Sprintf(".%s IN (?)", parentIDField),
							ids,
						},
					},
					ConvertValuesFromString: true,
				},
			)
			if errCtl != nil {
				return ErrController{
					Op:  "CascadeDelete",
					Err: errors.New("Error from UpdateMultiple"),
				}
			}
		}
	}

	return nil
}
