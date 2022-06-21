package normal

//mocknn: T
type MockT struct {
	m  int
}

func (t *MockT) M() {
	println(t.m)
}

//mocknn: F
func MockF(t *T) {
	t.M()
}
