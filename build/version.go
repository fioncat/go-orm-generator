package build

import (
	"fmt"
	"runtime"
)

const VERSION = "0.1.0"

func ShowVersion() {
	fmt.Printf("go-gendb version %s on %s/%s\n", VERSION, runtime.GOOS, runtime.GOARCH)
}
