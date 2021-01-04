package build

import (
	"fmt"
	"runtime"
)

const VERSION = "0.1.0"

func ShowVersion() {
	fmt.Printf("go-gendb v%s on %s\n", VERSION, runtime.GOOS)
}
