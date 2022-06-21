package simple_test

import (
	"testing"

	"github.com/tenntenn/mocknn/_example/simple"
)

func Test(t *testing.T) {
	v := simple.New(10)
	simple.F(v)
}
