package gconv

import (
	"reflect"

	"github.com/sjatsh/beanstalk-go/lib/internal/utils"
)

// SliceAny is alias of Interfaces.
func SliceAny(i interface{}) []interface{} {
	return Interfaces(i)
}

// Interfaces converts <i> to []interface{}.
func Interfaces(i interface{}) []interface{} {
	if i == nil {
		return nil
	}
	if r, ok := i.([]interface{}); ok {
		return r
	} else if r, ok := i.(apiInterfaces); ok {
		return r.Interfaces()
	} else {
		var array []interface{}
		switch value := i.(type) {
		case []string:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []int:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []int8:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []int16:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []int32:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []int64:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []uint:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []uint8:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []uint16:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []uint32:
			for _, v := range value {
				array = append(array, v)
			}
		case []uint64:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []bool:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []float32:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		case []float64:
			array = make([]interface{}, len(value))
			for k, v := range value {
				array[k] = v
			}
		default:
			// Finally we use reflection.
			var (
				rv   = reflect.ValueOf(i)
				kind = rv.Kind()
			)
			for kind == reflect.Ptr {
				rv = rv.Elem()
				kind = rv.Kind()
			}
			switch kind {
			case reflect.Slice, reflect.Array:
				array = make([]interface{}, rv.Len())
				for i := 0; i < rv.Len(); i++ {
					array[i] = rv.Index(i).Interface()
				}
			case reflect.Struct:
				rt := rv.Type()
				array = make([]interface{}, 0)
				for i := 0; i < rv.NumField(); i++ {
					// Only public attributes.
					if !utils.IsLetterUpper(rt.Field(i).Name[0]) {
						continue
					}
					array = append(array, rv.Field(i).Interface())
				}
			default:
				return []interface{}{i}
			}
		}
		return array
	}
}
