package golang

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/token"
)

// IsSimpleType returns whether t is a simple type,
// for example: int32, int64, float64, string. Note that
// slice and map do not belong to it.
func IsSimpleType(t string) bool {
	switch {
	case strings.HasPrefix(t, "int"):
		return true
	case strings.HasPrefix(t, "uint"):
		return true

	case t == "string":
		return true

	case t == "bool":
		return true

	case strings.HasPrefix(t, "float"):
		return true

	}

	return false
}

// Go keywords that need to be concerned, not all keywords here.
const (
	_package   = token.Token("package")
	_import    = token.Token("import")
	_type      = token.Token("type")
	_interface = token.Token("interface")
	_struct    = token.Token("struct")
	_comment   = token.Token("//")
	_map       = token.Token("map")
)

var (
	// Uses to parse multi-line import:
	//   import (
	//     [{name}] {path}
	//     ...
	//   )
	_importsTokens = []token.Token{
		_import,
		token.LPAREN,
		token.RPAREN,
	}

	// Uses to parse struct definition:
	//   type {name} struct {
	_structTokens = []token.Token{
		_type,
		_struct,
		token.LBRACE,
	}

	// Uses to parse field definition:
	//   {name} {field} [// {comment}]
	_fieldTokens = []token.Token{
		_comment,
	}

	// Uses to parse interface definition:
	//   type {name} interface {
	_interfaceTokens = []token.Token{
		_type,
		_interface,
		token.LBRACE,
	}

	// Uses to parse method in interface:
	//   {name}({param}) [{ret}]
	_methodTokens = []token.Token{
		token.LPAREN,
		token.RPAREN,
		token.LBRACK,
		token.RBRACK,
		token.MUL,
		token.COMMA,
		token.PERIOD,
	}

	_typeTokens = []token.Token{
		token.LBRACK, token.RBRACK, token.MUL,
		token.PERIOD, _map,
	}
)

// _packageParser uses to parse package definition:
//   package {name}
type _packageParser struct{}

// Accept returns whether line is a package definition line.
func (*_packageParser) Accept(line string) bool {
	return _package.PrefixOf(line)
}

// Do parse package line and returns the package name(string).
func (*_packageParser) Do(line string) (interface{}, error) {
	line = _package.Trim(line)
	if line == "" {
		return nil, fmt.Errorf("missing package name")
	}
	return line, nil
}

// _singleImportParser uses to parse single-line import content.
//  import [{name}] {path}
type _singleImportParser struct{}

// Accept returns whether line is a import line.
func (*_singleImportParser) Accept(line string) bool {
	return _import.PrefixOf(line)
}

// Do parse the import line, returns *Import pointer.
func (*_singleImportParser) Do(line string) (interface{}, error) {
	line = _import.Trim(line)
	if line == "" {
		return nil, fmt.Errorf("missing import content")
	}
	s := token.NewScanner(line, nil)
	eles := s.Gets()
	imp := new(Import)
	switch len(eles) {
	case 1:
		// no name: "{path}"
		e := eles[0]
		if !e.String {
			return nil, e.NotMatch("STRING")
		}
		imp.Path = e.Get()

	case 2:
		// name and path: {name} "{path}"
		alias := eles[0]
		path := eles[1]
		if !alias.Indent {
			return nil, alias.NotMatch("INDENT")
		}
		if !path.String {
			return nil, path.NotMatch("STRING")
		}
		imp.Name = alias.Get()
		imp.Path = path.Get()

	default:
		// invalid
		return nil, fmt.Errorf("import statement bad format")
	}

	return imp, nil
}

// _multiPathsParser parse multi-lines import:
//   import (
//     [{name}] {path}
//     ...
//   )
type _multiPathsParser struct {
	imps []*Import
}

// If line is the beginning of a multi-line import, return
// the parser; otherwise, return nil. The condition is whether
// the line is "import ("
func acceptPaths(line string) base.ScanParser {
	tmp := strings.Fields(line)
	if len(tmp) <= 1 {
		return nil
	}
	if _import.Equal(tmp[0]) &&
		token.LPAREN.Equal(tmp[1]) {
		return new(_multiPathsParser)
	}
	return nil
}

func (p *_multiPathsParser) Next(_ int, line string, _ []*base.Tag) (bool, error) {
	if token.RPAREN.Equal(line) {
		return false, nil
	}
	s := token.NewScanner(line, _importsTokens)
	es := s.Gets()
	imp := new(Import)
	switch len(es) {
	case 1:
		if !es[0].String {
			return false, es[0].NotMatch("STRING")
		}
		imp.Path = es[0].Get()

	case 2:
		if !es[0].Indent {
			return false, es[0].NotMatch("INDENT")
		}
		if !es[1].String {
			return false, es[1].NotMatch("STRING")
		}
		imp.Name = es[0].Get()
		imp.Path = es[1].Get()

	default:
		return false, fmt.Errorf("import bad format")
	}

	p.imps = append(p.imps, imp)
	return true, nil
}

func (p *_multiPathsParser) Get() interface{} {
	return p.imps
}

type _structParser struct {
	Struct *Struct
}

func acceptStruct(idx int, line string, tags []*base.Tag, comms []string) (
	base.ScanParser, error,
) {
	if len(tags) == 0 {
		return nil, nil
	}
	s := token.NewScanner(line, _structTokens)
	es := s.Gets()
	if len(es) != 4 {
		return nil, nil
	}
	if es[0].Token != _type {
		return nil, nil
	}
	if !es[1].Indent {
		return nil, nil
	}
	if es[2].Token != _struct {
		return nil, nil
	}
	if es[3].Token != token.LBRACE {
		return nil, nil
	}

	_struct := new(Struct)
	_struct.Name = es[1].Get()
	_struct.Tags = tags
	_struct.Line = idx + 1
	if len(comms) > 0 {
		_struct.Comment = comms[0]
	}

	p := new(_structParser)
	p.Struct = _struct

	return p, nil
}

func (p *_structParser) Next(idx int, line string, tags []*base.Tag) (
	bool, error,
) {
	if token.RBRACE.Equal(line) {
		return false, nil
	}
	s := token.NewScanner(line, _fieldTokens)
	var e token.Element

	ok := s.Next(&e)
	if !ok {
		return false, s.EarlyEnd("INDENT")
	}
	f := new(Field)
	f.Name = e.Get()

	ok = s.Next(&e)
	if !ok {
		return false, s.EarlyEnd("INDENT")
	}
	f.Type = e.Get()

	ok = s.Cur(&e)
	if ok && e.String {
		s.Next(nil)
	}

	ok = s.Next(&e)
	if ok {
		if e.Token != _comment {
			return false, e.NotMatch("COMMENT")
		}
		var tmp []string
		for s.Next(&e) {
			tmp = append(tmp, e.Get())
		}
		f.Comment = strings.Join(tmp, " ")
	}
	f.Tags = tags
	f.Line = idx + 1

	p.Struct.Fields = append(p.Struct.Fields, f)

	return true, nil
}

func (p *_structParser) Get() interface{} {
	return p.Struct
}

type _interfaceParser struct {
	inter *Interface
}

func acceptInterface(idx int, line string, tags []*base.Tag, _ []string) (
	base.ScanParser, error,
) {
	if len(tags) == 0 {
		return nil, nil
	}
	s := token.NewScanner(line, _interfaceTokens)
	es := s.Gets()
	if len(es) != 4 {
		return nil, nil
	}

	if es[0].Token != _type {
		return nil, nil
	}
	if !es[1].Indent {
		return nil, nil
	}
	if es[2].Token != _interface {
		return nil, nil
	}
	if es[3].Token != token.LBRACE {
		return nil, nil
	}
	if len(tags) != 1 {
		return nil, fmt.Errorf("only allow one tag to mark interface")
	}
	p := new(_interfaceParser)
	p.inter = new(Interface)
	p.inter.Name = es[1].Get()
	p.inter.Tag = tags[0]
	p.inter.line = idx + 1

	return p, nil
}

func (p *_interfaceParser) Next(idx int, line string, tags []*base.Tag) (bool, error) {
	if token.RBRACE.Equal(line) {
		return false, nil
	}

	s := token.NewScanner(line, _methodTokens)

	method := new(Method)
	method.line = idx + 1
	method.Def = line
	method.Tags = tags

	var e token.Element
	// Method Name
	ok := s.Next(&e)
	if !ok {
		return false, s.EarlyEnd("INDENT")
	}
	if !e.Indent {
		return false, e.NotMatch("INDENT")
	}
	method.Name = e.Get()

	// Param List
	ok = s.Next(&e)
	if !ok {
		return false, s.EarlyEnd("LPAREN")
	}
	if e.Token != token.LPAREN {
		return false, e.NotMatch("LPAREN")
	}

	// NOTE: We only handle imports in the param list
	for {
		ok = s.Next(&e)
		if !ok {
			return false, s.EarlyEnd("RPAREN")
		}
		if e.Token == token.RPAREN {
			break
		}
		if e.Token != token.PERIOD {
			continue
		}

		// handle import
		ok = s.Pervious(&e, 2)
		if !ok {
			// Never trigger
			return false, fmt.Errorf("[method] " +
				"No content before PERIOD")
		}

		if !e.Indent {
			return false, e.NotMatch("INDENT")
		}

		name := e.Get()
		method.Imports = append(method.Imports, name)
	}

	// Return List
	ok = s.Next(&e)
	if !ok {
		return false, s.EarlyEnd("LPAREN")
	}
	if e.Token != token.LPAREN {
		return false, e.NotMatch("LPAREN")
	}

	ok = s.Cur(&e)
	if !ok {
		return false, s.EarlyEnd("TOKEN")
	}
	if !e.Indent {
		switch e.Token {
		case token.LBRACK:
			s.Next(nil)
			ok = s.Next(&e)
			if !ok {
				return false, s.EarlyEnd("RBRACK")
			}
			if e.Token != token.RBRACK {
				return false, e.NotMatch("RBRACK")
			}
			method.RetSlice = true

			ok = s.Cur(&e)
			if !ok {
				return false, s.EarlyEnd("TOKEN")
			}
			if e.Token == token.MUL {
				method.RetPointer = true
				s.Next(nil)
			}

		case token.MUL:
			method.RetPointer = true
			s.Next(nil)
		}
	}

	ok = s.Next(&e)
	if !ok {
		return false, s.EarlyEnd("INDENT")
	}
	if !e.Indent {
		return false, e.NotMatch("INDENT")
	}

	var next token.Element
	s.Cur(&next)
	if next.Token == token.PERIOD {
		importName := e.Get()
		method.Imports = append(method.Imports,
			importName)
		method.RetType = importName + "."
		s.Next(nil)

		ok = s.Next(&e)
		if !ok {
			return false, s.EarlyEnd("INDENT")
		}
		if !e.Indent {
			return false, e.NotMatch("INDENT")
		}
	}
	method.RetType += e.Get()

	ok = s.Next(&e)
	if !ok {
		return false, s.EarlyEnd("COMMA")
	}
	if e.Token != token.COMMA {
		return false, e.NotMatch("COMMA")
	}

	ok = s.Next(&e)
	if !ok {
		return false, s.EarlyEnd("error")
	}
	if e.Get() != "error" {
		return false, e.FmtErr(`expect "error" for the `+
			`2nd returns, found: "%s"`, e.Get())
	}

	ok = s.Next(&e)
	if !ok {
		return false, s.EarlyEnd("RPAREN")
	}
	if e.Token != token.RPAREN {
		return false, e.NotMatch("RPAREN")
	}

	if IsSimpleType(method.RetType) {
		method.RetSimple = true
	}
	p.inter.Methods = append(p.inter.Methods, method)

	return true, nil
}

func (p *_interfaceParser) Get() interface{} {
	return p.inter
}
