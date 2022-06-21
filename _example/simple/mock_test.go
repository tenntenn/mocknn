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
