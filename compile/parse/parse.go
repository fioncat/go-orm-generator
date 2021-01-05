package parse

import (
	"fmt"

	"github.com/fioncat/go-gendb/compile/parse/parsesql"
	"github.com/fioncat/go-gendb/compile/scan/scango"
	"github.com/fioncat/go-gendb/generate"
)

// Parser is used to receive the scanned go file results,
// perform syntax analysis, and convert them into a struct
// that can be provided to the code generator for code
// generation.
type Parser interface {
	// Do receives the result of scanning the file and
	// returns the intermediate structure of the generated
	// code.
	// Different go file types have different specific
	// implementations.
	Do(r *scango.Result) ([]generate.Result, error)
}

var parsers = make(map[string]Parser)

func init() {
	parsers["sql"] = &parsesql.Parser{}
}

// Do receives the result of scanning the go file, selects
// different parsers according to different Types, uses it
// to parse the scan results after detecting the parser,
// and returns the code to generate an intermediate struct.
//
// It should be noted that only a few built-in parsers are
// supported. With the iteration of the version, the supported
// parsers will increase. For details, please refer to the
// online documentation.
//
// If the parser corresponding to the Type does not exist,
// an error will be returned. The parser will return an
// error if an IO error or syntax error occurs during the
// parsing process.
func Do(scanResult *scango.Result) ([]generate.Result, error) {
	parser := parsers[scanResult.Type]
	if parser == nil {
		return nil, fmt.Errorf(`unknown parser "%s"`, scanResult.Type)
	}

	return parser.Do(scanResult)
}
