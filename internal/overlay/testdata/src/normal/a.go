package normal

type T struct {
	n int
}

func (t *T) M() {
	println(t.n)
}

func F(t *T) {
	t.M()
}
