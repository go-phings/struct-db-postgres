package structdbpostgres

import (
	"reflect"
	"sort"

	stsql "github.com/go-phings/struct-sql-postgres"
)

// GetObjIDInterface returns an interface{} to ID field of an object
func (c *Controller) GetObjIDInterface(obj interface{}) interface{} {
	return reflect.ValueOf(obj).Elem().FieldByName("ID").Addr().Interface()
}

// GetObjIDValue returns value of ID field (int64) of an object
func (c *Controller) GetObjIDValue(obj interface{}) int64 {
	return reflect.ValueOf(obj).Elem().FieldByName("ID").Int()
}

// GetObjFieldInterfaces return list of interfaces to object's fields
// Argument includeID tells it to include or omit the ID field
func (c Controller) GetObjFieldInterfaces(obj interface{}, includeID bool) []interface{} {
	val := reflect.ValueOf(obj).Elem()

	var v []interface{}
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		if val.Type().Field(i).Name == "ID" && !includeID {
			continue
		}
		// struct-sql-postgres is used to generate SQL queries so here the same kinds must be supported
		if !stsql.IsFieldKindSupported(valueField.Kind()) {
			continue
		}

		v = append(v, valueField.Addr().Interface())
	}
	return v
}

// GetFiltersInterfaces returns list of interfaces from filters map (used in querying)
func (c Controller) GetFiltersInterfaces(mf map[string]interface{}) []interface{} {
	var xi []interface{}

	if len(mf) == 0 {
		return xi
	}

	sorted := []string{}
	for k := range mf {
		if k == "_raw" || k == "_rawConjuction" {
			continue
		}
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for _, v := range sorted {
		xi = append(xi, mf[v])
	}

	// Get pointers to values from raw query
	_, ok := mf["_raw"]
	if !ok {
		return xi
	}

	rt := reflect.TypeOf(mf["_raw"])
	if rt.Kind() != reflect.Slice && rt.Kind() != reflect.Array {
		return xi
	}
	if reflect.ValueOf(mf["_raw"]).Len() < 2 {
		return xi
	}

	for i := 1; i < reflect.ValueOf(mf["_raw"]).Len(); i++ {
		rt2 := reflect.TypeOf(mf["_raw"].([]interface{})[i])
		if rt2.Kind() == reflect.Slice || rt2.Kind() == reflect.Array {
			valInt8s, ok := mf["_raw"].([]interface{})[i].([]int8)
			if ok {
				for j := 0; j < len(valInt8s); j++ {
					xi = append(xi, valInt8s[j])
				}
				continue
			}
			valInt16s, ok := mf["_raw"].([]interface{})[i].([]int16)
			if ok {
				for j := 0; j < len(valInt16s); j++ {
					xi = append(xi, valInt16s[j])
				}
				continue
			}
			valInt32s, ok := mf["_raw"].([]interface{})[i].([]int32)
			if ok {
				for j := 0; j < len(valInt32s); j++ {
					xi = append(xi, valInt32s[j])
				}
				continue
			}
			valInt64s, ok := mf["_raw"].([]interface{})[i].([]int64)
			if ok {
				for j := 0; j < len(valInt64s); j++ {
					xi = append(xi, valInt64s[j])
				}
				continue
			}
			valInts, ok := mf["_raw"].([]interface{})[i].([]int)
			if ok {
				for j := 0; j < len(valInts); j++ {
					xi = append(xi, valInts[j])
				}
				continue
			}
			valUint8s, ok := mf["_raw"].([]interface{})[i].([]uint8)
			if ok {
				for j := 0; j < len(valUint8s); j++ {
					xi = append(xi, valUint8s[j])
				}
				continue
			}
			valUint16s, ok := mf["_raw"].([]interface{})[i].([]uint16)
			if ok {
				for j := 0; j < len(valUint16s); j++ {
					xi = append(xi, valUint16s[j])
				}
				continue
			}
			valUint32s, ok := mf["_raw"].([]interface{})[i].([]uint32)
			if ok {
				for j := 0; j < len(valUint32s); j++ {
					xi = append(xi, valUint32s[j])
				}
				continue
			}
			valUint64s, ok := mf["_raw"].([]interface{})[i].([]uint64)
			if ok {
				for j := 0; j < len(valUint64s); j++ {
					xi = append(xi, valUint64s[j])
				}
				continue
			}
			valUints, ok := mf["_raw"].([]interface{})[i].([]uint)
			if ok {
				for j := 0; j < len(valUints); j++ {
					xi = append(xi, valUints[j])
				}
				continue
			}
			valFloat32s, ok := mf["_raw"].([]interface{})[i].([]float32)
			if ok {
				for j := 0; j < len(valFloat32s); j++ {
					xi = append(xi, valFloat32s[j])
				}
				continue
			}
			valFloat64s, ok := mf["_raw"].([]interface{})[i].([]float64)
			if ok {
				for j := 0; j < len(valFloat64s); j++ {
					xi = append(xi, valFloat64s[j])
				}
				continue
			}
			valBools, ok := mf["_raw"].([]interface{})[i].([]bool)
			if ok {
				for j := 0; j < len(valBools); j++ {
					xi = append(xi, valBools[j])
				}
				continue
			}
			valStrings, ok := mf["_raw"].([]interface{})[i].([]string)
			if ok {
				for j := 0; j < len(valStrings); j++ {
					xi = append(xi, valStrings[j])
				}
			}
		} else {
			xi = append(xi, mf["_raw"].([]interface{})[i])
		}
	}

	return xi
}

// ResetFields zeroes object's field values
func (c Controller) ResetFields(obj interface{}) {
	val := reflect.ValueOf(obj).Elem()
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		k := f.Kind()

		if k == reflect.Ptr {
			f.Set(reflect.Zero(f.Type()))
		}
		if k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 {
			f.SetInt(0)
		}
		if k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 {
			f.SetInt(0)
		}
		if k == reflect.Float32 || k == reflect.Float64 {
			f.SetFloat(0.0)
		}
		if k == reflect.String {
			f.SetString("")
		}
		if k == reflect.Bool {
			f.SetBool(false)
		}
	}
}

// SetObjCreated sets object's CreatedAt and CreatedBy fields
func (c *Controller) SetObjCreated(obj interface{}, at int64, by int64) {
	reflect.ValueOf(obj).Elem().FieldByName("CreatedAt").SetInt(at)
	reflect.ValueOf(obj).Elem().FieldByName("CreatedBy").SetInt(by)
}

// SetObjLastModified sets object's LastModifiedAt and LastModifiedBy fields
func (c *Controller) SetObjLastModified(obj interface{}, at int64, by int64) {
	reflect.ValueOf(obj).Elem().FieldByName("LastModifiedAt").SetInt(at)
	reflect.ValueOf(obj).Elem().FieldByName("LastModifiedBy").SetInt(by)
}
