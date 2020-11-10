package gutil

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"github.com/sjatsh/beanstalk-go/lib/internal/json"
	"github.com/sjatsh/beanstalk-go/lib/util/gconv"
)

// Dump prints variables <i...> to stdout with more manually readable.
func Dump(i ...interface{}) {
	s := Export(i...)
	if s != "" {
		fmt.Println(s)
	}
}

// Export returns variables <i...> as a string with more manually readable.
func Export(i ...interface{}) string {
	buffer := bytes.NewBuffer(nil)
	for _, v := range i {
		if b, ok := v.([]byte); ok {
			buffer.Write(b)
		} else {
			rv := reflect.ValueOf(v)
			kind := rv.Kind()
			if kind == reflect.Ptr {
				rv = rv.Elem()
				kind = rv.Kind()
			}
			switch kind {
			case reflect.Slice, reflect.Array:
				v = gconv.Interfaces(v)
			case reflect.Map, reflect.Struct:
				v = gconv.Map(v)
			}
			encoder := json.NewEncoder(buffer)
			encoder.SetEscapeHTML(false)
			encoder.SetIndent("", "\t")
			if err := encoder.Encode(v); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}
	}
	return buffer.String()
}
