package util

import "flag"

func ParseFlags() (string, string, float64, int) {
	bootstrap := flag.String("b", "", "Bootstrap server")
	objectFile := flag.String("o", "", "Object file path")
	timeDelay := flag.Float64("d", 0.0, "Initial delay")
	testcase := flag.Int("t", 0, "Testcase object ID")

	// Parse command-line flags
	flag.Parse()

	return *bootstrap, *objectFile, *timeDelay, *testcase
}
