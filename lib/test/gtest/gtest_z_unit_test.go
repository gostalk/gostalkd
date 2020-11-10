package gtest_test

import (
	"github.com/sjatsh/beanstalk-go/lib/test/gtest"
	"testing"
)

func TestC(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.Assert(1, 1)
		t.AssertNE(1, 0)
		t.AssertEQ(float32(123.456), float32(123.456))
		t.AssertEQ(float64(123.456), float64(123.456))
	})
}

func TestCase(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.Assert(1, 1)
		t.AssertNE(1, 0)
		t.AssertEQ(float32(123.456), float32(123.456))
		t.AssertEQ(float64(123.456), float64(123.456))
	})
}
