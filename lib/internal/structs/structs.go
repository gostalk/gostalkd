// Package structs provides functions for struct conversion.
package structs

import "github.com/gqcn/structs"

// Field is alias of structs.Field.
type Field struct {
	*structs.Field
	// Retrieved tag name. There might be more than one tags in the field,
	// but only one can be retrieved according to calling function rules.
	Tag string
}
