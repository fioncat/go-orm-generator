package build

import "runtime"

// This variable represents the number of concurrent.
// Whenever encountering tasks that need to be executed
// concurrently, N_WORKERS workers will be created for
// execution. The default is equal to the number of CPU cores.
var N_WORKERS = runtime.NumCPU()
