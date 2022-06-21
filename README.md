# mocknn

[![pkg.go.dev][gopkg-badge]][gopkg]

mocknn is a mocking tool using `-overlay` option of `go test`.

This is experimental project. APIs may change, do not use your product development.
If you try to use this tool, please give me your feedback by filing an issue.

## How to install

```
$ go install github.com/tenntenn/mocknn@latest
```

## How to use

For example, there is a .go file which is test target.

```go
// simple.go: test target
package simple

import "fmt"

const msg string = "hello"

func New(n int) *T {
	return &T{n: n}
}

type T struct {
	n int
}

func (t *T) V() int {
	return t.n
}

func F(t *T) {
	fmt.Println(msg, t.V())
}
```

simple_test.go is a test file for simple.go.

```go
package simple_test

import (
	"testing"

	"github.com/tenntenn/mocknn/_example/simple"
)

func Test(t *testing.T) {
	v := simple.New(10)
	simple.F(v)
}
```

mock_test.go provides mockings for simple.go.

```go
package simple

//mocknn: msg
const msgMock = "mock"

//mocknn: T
type MockT struct {
	m int
}

//mocknn: New
func MockNew(m int) *MockT {
	return &MockT{m: m * 2}
}

func (t *MockT) V() int {
	return t.m
}
```

mocknn replaces implementation to mocking which has `//mocknn:` comment directive.
mocknn uses `-overlay` option to replace mocking files.

```
# use original implementation
$ go test
hello 10
PASS
ok  	github.com/tenntenn/mocknn/_example/simple	0.359s

# use mockings
$ go test -overlay=`mocknn`
mock 20
PASS
ok  	github.com/tenntenn/mocknn/_example/simple	0.273s

# same as go test -overlay=`mocknn`
$ mocknn test
mock 20
PASS
ok  	github.com/tenntenn/mocknn/_example/simple	0.273s
```

## Examples

See [_example](./_example) directory.

<!-- links -->
[gopkg]: https://pkg.go.dev/github.com/tenntenn/mocknn
[gopkg-badge]: https://pkg.go.dev/badge/github.com/tenntenn/testtime?status.svg
