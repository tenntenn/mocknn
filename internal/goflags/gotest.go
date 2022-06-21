package goflags

import "flag"

func GoTest(flags *flag.FlagSet) {
	_ = flags.Bool("args", false, "")
	_ = flags.Bool("c", false, "")
	_ = flags.String("exec", "", "")
	_ = flags.Bool("i", false, "")
	_ = flags.Bool("json", false, "")
	_ = flags.String("o", "", "")
}
