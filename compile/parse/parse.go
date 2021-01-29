package parse

import (
	"fmt"

	"github.com/fioncat/go-gendb/build"
	"github.com/fioncat/go-gendb/compile/mediate"
	"github.com/fioncat/go-gendb/compile/parse/psql"
	"github.com/fioncat/go-gendb/compile/scan/sgo"
	"github.com/fioncat/go-gendb/generate/coder"
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
	Do(r *sgo.Result) ([]mediate.Result, error)
}

var parsers = make(map[string]Parser)

func init() {
	parsers["sql"] = &psql.Parser{}
}

type StructResult struct {
	Structs []*coder.Struct
}

func (*StructResult) Type() string                   { return "struct" }
func (*StructResult) Key() string                    { return "Struct" }
func (sr *StructResult) GetStructs() []*coder.Struct { return sr.Structs }

// Do receives the result of scanning the go file, selects
// different parsers according to different Types, uses it
// to parse the scan results after detecting the parser,
// and returns the intermediate struct.
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
func Do(scanResult *sgo.Result) ([]mediate.Result, error) {
	parser := parsers[scanResult.Type]
	if parser == nil {
		return nil, fmt.Errorf(`unknown parser "%s"`, scanResult.Type)
	}

	results, err := parser.Do(scanResult)
	if err != nil {
		return nil, err
	}
	if build.DEBUG {
		return results, nil
	}

	structMap := make(map[string][]*coder.Struct, len(results))
	for _, result := range results {
		structs := result.GetStructs()
		for _, s := range structs {
			structMap[s.Name] = append(structMap[s.Name], s)
		}
	}

	sr := new(StructResult)
	for _, ss := range structMap {
		if len(ss) == 0 {
			continue
		}
		s := ss[0]
		if len(ss) == 1 {
			sr.Structs = append(sr.Structs, s)
			continue
		}

		ss = ss[1:]
		for _, os := range ss {
			s.Merge(os)
		}

		sr.Structs = append(sr.Structs, s)
	}

	if len(sr.Structs) > 0 {
		results = append(results, sr)
	}

	return results, nil
}
