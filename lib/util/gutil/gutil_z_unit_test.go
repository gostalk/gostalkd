package gutil_test

import (
	"github.com/sjatsh/beanstalk-go/lib/test/gtest"
	"github.com/sjatsh/beanstalk-go/lib/util/gutil"
	"testing"
)

func Test_Dump(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		gutil.Dump(map[string]interface{}{
			"test1": 100,
			"test2": map[string]interface{}{
				"test3": 22222,
				"test4": 44444444,
				"test5": map[string]interface{}{
					"test6": 333333333,
				},
			},
		})
	})
}

func Test_TryCatch(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		gutil.TryCatch(func() {
			panic("gutil TryCatch test")
		})
	})

	gtest.C(t, func(t *gtest.T) {
		gutil.TryCatch(func() {
			panic("gutil TryCatch test")

		}, func(err interface{}) {
			t.Assert(err, "gutil TryCatch test")
		})
	})
}

func Test_IsEmpty(t *testing.T) {

	gtest.C(t, func(t *gtest.T) {
		t.Assert(gutil.IsEmpty(1), false)
	})
}

func Test_Throw(t *testing.T) {

	gtest.C(t, func(t *gtest.T) {
		defer func() {
			t.Assert(recover(), "gutil Throw test")
		}()

		gutil.Throw("gutil Throw test")
	})
}

func Test_IsNil(t *testing.T) {

	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Name string
		}
		var user interface{}
		t.Assert(gutil.IsNil(user), true)
		user = new(User)
		t.Assert(gutil.IsNil(user), false)
	})
}
