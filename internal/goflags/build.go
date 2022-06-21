package goflags

import "flag"

func Build(flags *flag.FlagSet) {
	// -ovelay use in mocknn
	_ = flags.Bool("a", false, "")
	_ = flags.Bool("n", false, "")
	_ = flags.Int("p", 0, "")
	_ = flags.Bool("race", false, "")
	_ = flags.Bool("msan", false, "")
	_ = flags.Bool("asan", false, "")
	_ = flags.Bool("v", false, "")
	_ = flags.Bool("work", false, "")
	_ = flags.Bool("x", false, "")
	_ = flags.String("asmflags", "", "")
	_ = flags.String("buildmode", "", "")
	_ = flags.Bool("buildvcs", false, "")
	_ = flags.String("compiler", "gc", "")
	_ = flags.String("gccgoflags", "", "")
	_ = flags.String("gcflags", "", "")
	_ = flags.String("installsuffix", "", "")
	_ = flags.String("ldflags", "", "")
	_ = flags.Bool("linkshared", false, "")
	_ = flags.String("mode", "", "")
	_ = flags.Bool("modcacherw", false, "")
	_ = flags.String("modfile", "", "")
	_ = flags.String("pkgdir", "", "")
	_ = flags.String("tags", "", "")
	_ = flags.Bool("trimpath", false, "")
	_ = flags.String("toolexec", "", "")
}
