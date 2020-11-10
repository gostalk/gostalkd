package gutil

import (
	"github.com/sjatsh/beanstalk-go/lib/internal/empty"
)

// Throw throws out an exception, which can be caught be TryCatch or recover.
func Throw(exception interface{}) {
	panic(exception)
}

// TryCatch implements try...catch... logistics using internal panic...recover.
func TryCatch(try func(), catch ...func(exception interface{})) {
	defer func() {
		if e := recover(); e != nil && len(catch) > 0 {
			catch[0](e)
		}
	}()
	try()
}

// IsEmpty checks given <value> empty or not.
// It returns false if <value> is: integer(0), bool(false), slice/map(len=0), nil;
// or else returns true.
func IsEmpty(value interface{}) bool {
	return empty.IsEmpty(value)
}

// IsNil checks whether given <value> is nil.
// Note that it might use reflect feature which affects performance a little bit.
func IsNil(value interface{}) bool {
	return empty.IsNil(value)
}
