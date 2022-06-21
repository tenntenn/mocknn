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
