package structs

import (
	"reflect"

	"github.com/gqcn/structs"
)

// MapField retrieves struct field as map[name/tag]*Field from <pointer>, and returns the map.
//
// The parameter <pointer> should be type of struct/*struct.
//
// The parameter <priority> specifies the priority tag array for retrieving from high to low.
//
// The parameter <recursive> specifies whether retrieving the struct field recursively.
//
// Note that it only retrieves the exported attributes with first letter up-case from struct.
func MapField(pointer interface{}, priority []string, recursive bool) map[string]*Field {
	var (
		fields   []*structs.Field
		fieldMap = make(map[string]*Field)
	)
	if v, ok := pointer.(reflect.Value); ok {
		fields = structs.Fields(v.Interface())
	} else {
		fields = structs.Fields(pointer)
	}
	var (
		tag  = ""
		name = ""
	)
	for _, field := range fields {
		name = field.Name()
		// Only retrieve exported attributes.
		if name[0] < byte('A') || name[0] > byte('Z') {
			continue
		}
		fieldMap[name] = &Field{
			Field: field,
			Tag:   tag,
		}
		tag = ""
		for _, p := range priority {
			tag = field.Tag(p)
			if tag != "" {
				break
			}
		}
		if tag != "" {
			fieldMap[tag] = &Field{
				Field: field,
				Tag:   tag,
			}
		}
		if recursive {
			rv := reflect.ValueOf(field.Value())
			kind := rv.Kind()
			if kind == reflect.Ptr {
				rv = rv.Elem()
				kind = rv.Kind()
			}
			if kind == reflect.Struct {
				for k, v := range MapField(rv, priority, true) {
					if _, ok := fieldMap[k]; !ok {
						fieldMap[k] = v
					}
				}
			}
		}
	}
	return fieldMap
}
