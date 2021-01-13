package build

import (
	"fmt"
	"runtime"
)

// version(build-in)
const VERSION = "0.1.2"

// ShowVersion show version in the terminal.
func ShowVersion() {
	fmt.Printf("go-gendb version %s on %s/%s\n", VERSION, runtime.GOOS, runtime.GOARCH)
}
