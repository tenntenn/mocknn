package externalpkg

import (
	_ "io"
)

//mocknn: io.Writer
type MyWriter interface{
	Write(p []byte) (n int, err error)
}
