package goflags

import "flag"

func All(flags *flag.FlagSet) {
	Testing(flags)
	Build(flags)
	GoTest(flags)
}
