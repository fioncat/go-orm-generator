package mediate

import "github.com/fioncat/go-gendb/generate/coder"

// Result represents the intermediate structure of code
// generation. The structure data should be directly used
// for code generation. Result is the final result of the
// compilation process, and its data structure should be
// highly consistent with the final code to be generated.
//
// Result is defined in the form of an interface here because
// the intermediate data of each generator is very different,
// and the generator should convert the interface data into
// a specific structure by itself. Each Result needs to
// implement this interface for the generator to select such
// necessary operations.
type Result interface {
	// Type returns the type of Result. This method is used
	// to select the generator. Each Result has a unique
	// generator to process it to generate code. If the
	// generator returned by Type does not support it,
	// passing the Result data to the generating function
	// will return an error.
	Type() string

	// Key returns the only one of Result, it will be used
	// in the file name. If there are multiple results with
	// the same Key return value, they will be overwritten
	// when the code is generated.
	Key() string

	// GetStructs returns the structure to be generated by
	// the Result. If there is no structure to be generated,
	// can return nil.
	GetStructs() []*coder.Struct
}