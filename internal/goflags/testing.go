package goflags

import "flag"

func Testing(flags *flag.FlagSet) {
	// -v is defined in Build
	_ = flags.String("bench", "", "")
	_ = flags.Bool("benchmem", false, "")
	_ = flags.Duration("benchtime", 0, "")
	_ = flags.String("blockprofile", "", "")
	_ = flags.Int("blockprofilerate", 0, "")
	_ = flags.Uint("count", 0, "")
	_ = flags.String("coverprofile", "", "")
	_ = flags.String("cpu", "", "")
	_ = flags.String("cpuprofile", "", "")
	_ = flags.Bool("failfast", false, "")
	_ = flags.String("fuzz", "", "")
	_ = flags.String("fuzzcachedir", "", "")
	_ = flags.Duration("fuzzminimizetime", 0, "")
	_ = flags.Duration("fuzztime", 0, "")
	_ = flags.Bool("fuzzworker", false, "")
	_ = flags.String("list", "", "")
	_ = flags.String("memprofile", "", "")
	_ = flags.Int("memprofilerate", 0, "")
	_ = flags.String("mutexprofile", "", "")
	_ = flags.Int("mutexprofilefraction", 0, "")
	_ = flags.String("outputdir", "", "")
	_ = flags.Bool("paniconexit0", false, "")
	_ = flags.Int("parallel", 0, "")
	_ = flags.String("run", "", "")
	_ = flags.Bool("short", false, "")
	_ = flags.String("shuffle", "", "")
	_ = flags.String("testlogfile", "", "")
	_ = flags.Duration("timeout", 0, "")
	_ = flags.String("trace", "", "")
}
