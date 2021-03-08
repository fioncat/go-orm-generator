package sql

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/token"
)

const (
	_select = token.Token("SELECT")
	_from   = token.Token("FROM")
	_inner  = token.Token("INNER")
	_left   = token.Token("LEFT")
	_right  = token.Token("RIGHT")
	_join   = token.Token("JOIN")
	_on     = token.Token("ON")
	_where  = token.Token("WHERE")
	_order  = token.Token("ORDER")
	_by     = token.Token("BY")
	_as     = token.Token("AS")
	_group  = token.Token("GROUP")
	_ifnull = token.Token("IFNULL")
	_limit  = token.Token("LIMIT")
	_count  = token.Token("COUNT")

	_update = token.Token("UPDATE")
	_delete = token.Token("DELETE")
	_insert = token.Token("INSERT")

	_preStart = token.Token("$")
	_repStart = token.Token("#")
	_varStart = token.Token("@")
	_dynStart = token.Token("%")
)

var sqlTokens = []token.Token{
	token.LPAREN,
	token.RPAREN,

	token.LBRACE,
	token.RBRACE,

	token.PERIOD,
	token.COMMA,
	token.COLON,
	token.PERCENT,

	_select, _from, _inner, _left, _right,
	_join, _on, _where, _order, _by, _as,
	_group, _ifnull, _limit, _count,

	_update, _delete, _insert,

	_preStart, _repStart, _varStart, _dynStart,
}

var phTokens = append(sqlTokens, token.SPACE)

var keywords = []string{
	_select.Get(),
	_from.Get(),
	_inner.Get(),
	_left.Get(),
	_right.Get(),
	_join.Get(),
	_on.Get(),
	_where.Get(),
	_order.Get(),
	_by.Get(),
	_as.Get(),
	_group.Get(),
	_ifnull.Get(),
	_limit.Get(),
	_count.Get(),
	_update.Get(),
	_delete.Get(),
	_insert.Get(),
}

var kwsLower = func() []string {
	kws := make([]string, len(keywords))
	for idx, kw := range keywords {
		kws[idx] = strings.ToLower(kw)
	}
	return kws
}()

type sqlVar struct {
	sqlEs []token.Element
	phEs  []token.Element
}

var globalVars = make(map[string]*sqlVar)

func parseNameFromTag(t string, tag *base.Tag) (string, error) {
	if tag.Name != t {
		return "", nil
	}
	if len(tag.Options) == 0 {
		return "", fmt.Errorf("missing name for %s", t)
	}

	nameOpt := tag.Options[0]
	if nameOpt.Key != "" && nameOpt.Key != "name" {
		return "", fmt.Errorf(`The 1st option `+
			`must be "name", found: "%s"`, nameOpt.Key)
	}
	if nameOpt.Value == "" {
		return "", fmt.Errorf("name is empty")
	}

	return nameOpt.Value, nil
}

type _varParser struct {
	name string
	sqls *token.Scanner
	phs  *token.Scanner
}

func acceptVar(tag *base.Tag) (base.ScanParser, error) {
	name, err := parseNameFromTag("var", tag)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, nil
	}

	p := new(_varParser)
	p.name = name
	p.sqls = token.EmptyScannerIC(sqlTokens)
	p.phs = token.EmptyScanner(phTokens)

	if _, ok := globalVars[name]; ok {
		return nil, fmt.Errorf(`var "%s" is duplicate`, name)
	}

	return p, nil
}

func (p *_varParser) Next(idx int, line string, _ []*base.Tag) (
	bool, error,
) {
	p.sqls.AddLine(idx, line)
	p.phs.AddLine(idx, line)
	return true, nil
}

func (p *_varParser) Get() interface{} {
	sqlEs := p.sqls.Gets()
	phEs := p.phs.Gets()
	if len(sqlEs) == 0 || len(phEs) == 0 {
		return p.phs.EarlyEndL("sql")
	}

	sv := &sqlVar{
		sqlEs: sqlEs,
		phEs:  phEs,
	}
	globalVars[p.name] = sv

	return nil
}

type _sqlParser struct {
	line int

	inter, name string

	sqls *token.Scanner
	phs  *token.Scanner

	dyn bool
}

func acceptSql(tag *base.Tag) (base.ScanParser, error) {
	name, err := parseNameFromTag("method", tag)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, nil
	}
	p := new(_sqlParser)
	p.line = tag.Line

	tmp := strings.Split(name, ".")
	switch len(tmp) {
	case 1:
		p.name = tmp[0]

	case 2:
		p.inter = tmp[0]
		p.name = tmp[1]

	default:
		return nil, fmt.Errorf(`name "%s" is `+
			`bad format`, name)
	}

	p.sqls = token.EmptyScannerIC(sqlTokens)
	p.phs = token.EmptyScanner(phTokens)
	p.phs.BreakLine = true

	opts := tag.Options[1:]
	for _, opt := range opts {
		if opt.Key == "dyn" {
			switch opt.Value {
			case "true":
				p.dyn = true
			}
		}
	}

	return p, nil
}

func (p *_sqlParser) Next(idx int, line string, _ []*base.Tag) (
	bool, error,
) {
	p.sqls.AddLine(idx, line)
	p.phs.AddLine(idx, line)
	return true, nil
}

func (p *_sqlParser) Get() interface{} {
	sqls, err := parseVars(p.sqls, true)
	if err != nil {
		return err
	}
	phs, err := parseVars(p.phs, false)
	if err != nil {
		return err
	}
	m, err := parseMethod(sqls, phs,
		p.name, p.inter, p.dyn)
	if err != nil {
		return err
	}
	m.line = p.line
	return m
}

func parseVars(s *token.Scanner, isSqls bool) (*token.Scanner, error) {
	var e token.Element
	var ok bool
	var bucket []token.Element
	var hasVar bool
	for {
		ok = s.Next(&e)
		if !ok {
			break
		}
		if e.Token != _varStart {
			bucket = append(bucket, e)
			continue
		}
		hasVar = true

		ok = s.Next(&e)
		if !ok {
			return nil, s.EarlyEndL("LBRACE")
		}
		if e.Token != token.LBRACE {
			return nil, e.NotMatchL("LBRACE")
		}

		var nameBucket []string
		for {
			ok = s.Next(&e)
			if !ok {
				return nil, s.EarlyEndL("RBRACE")
			}
			if e.Token == token.RBRACE {
				break
			}
			nameBucket = append(nameBucket, e.Get())
		}
		if len(nameBucket) == 0 {
			return nil, e.FmtErrL("name is empty")
		}
		name := strings.Join(nameBucket, "")
		sqlVar := globalVars[name]
		if sqlVar == nil {
			return nil, e.FmtErrL(`can not find var "%s"`, name)
		}
		var es []token.Element
		if isSqls {
			es = sqlVar.sqlEs
		} else {
			es = sqlVar.phEs
		}

		bucket = append(bucket, es...)
	}
	if !hasVar {
		s.Reset()
		return s, nil
	}
	return token.CopyScanner(s, bucket), nil
}

func parseMethod(sqls, phs *token.Scanner, name, inter string, dyn bool) (
	*Method, error,
) {
	var e token.Element
	ok := sqls.Next(&e)
	if !ok {
		return nil, sqls.EarlyEndL("KEYWORD")
	}
	m := new(Method)
	m.Name = name
	m.Inter = inter
	m.Dyn = dyn

	switch e.Token {
	case _select:

	case _insert, _update, _delete:
		m.Exec = true

	default:
		return nil, e.NotMatchL("SQL start")
	}

	var err error
	if !m.Exec {
		m.Fields, err = parseQuery(sqls, m.Name)
		if err != nil {
			return nil, err
		}
	}

	if dyn {
		m.Dps, err = parseDynamic(phs)
	} else {
		m.State, err = parsePh(phs)
	}
	if err != nil {
		return nil, err
	}

	return m, nil
}

func parseQuery(s *token.Scanner, name string) ([]*QueryField, error) {
	fs, err := parseSelect(s)
	if err != nil {
		return nil, err
	}

	tables, err := parseTables(s)
	if err != nil {
		return nil, err
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("[%s] can not "+
			"find table in query sql", name)
	}
	defaultTable := tables[0]

	nameMap := make(map[string]struct{}, len(tables))
	aliasMap := make(map[string]string, len(tables))
	for _, t := range tables {
		nameMap[t.name] = struct{}{}
		aliasMap[t.alias] = t.name
	}

	// ensure field's table name
	for _, f := range fs {
		if f.IsCount {
			continue
		}
		if f.Table == "" {
			// use the default table
			f.Table = defaultTable.name
			continue
		}
		_, ok := nameMap[f.Table]
		if ok {
			continue
		}

		name, ok := aliasMap[f.Table]
		if !ok {
			return nil, fmt.Errorf(`[%s] can not find `+
				`table "%s" for field "%s"`,
				name, f.Table, f.Name)
		}
		f.Table = name
	}

	return fs, nil
}

func parseSelect(s *token.Scanner) ([]*QueryField, error) {
	var e token.Element
	var fs []*QueryField
	for {
		ok := s.Cur(&e)
		if !ok {
			return nil, s.EarlyEndL("FROM")
		}
		if e.Token == _from {
			break
		}

		field := new(QueryField)
		var err error
		if !e.Indent {
			switch e.Token {
			case _ifnull:
				s.Next(nil)
				err = parseIfnull(s, field)

			case _count:
				err = parseCount(s, field)
				s.Next(nil)

			default:
				return nil, e.NotMatchL("FUNCTION")
			}
		} else {
			err = parseField(s, field)
		}
		if err != nil {
			return nil, err
		}

		fs = append(fs, field)
	}
	return fs, nil
}

func parseIfnull(s *token.Scanner, f *QueryField) error {
	var e token.Element
	ok := s.Next(&e)
	if !ok {
		return s.EarlyEndL("LPAREN")
	}
	if e.Token != token.LPAREN {
		return e.NotMatchL("LPAREN")
	}

	var next token.Element
	for {
		ok = s.Next(&e)
		if !ok {
			return s.EarlyEndL("RPAREN")
		}
		if e.Token == token.RPAREN {
			break
		}

		if e.Indent && f.Name == "" {
			ok = s.Cur(&next)
			if !ok {
				return s.EarlyEndL("INDENT/PERIOD")
			}
			if next.Token == token.PERIOD {
				f.Table = e.Get()
				s.Next(nil)
				ok = s.Next(&e)
				if !ok {
					return s.EarlyEndL("INDENT")
				}
				f.Name = e.Get()
				continue
			}
			f.Name = e.Get()
		}
	}
	if f.Name == "" {
		return e.FmtErrL("IFNULL is empty")
	}

	err := parseAlias(s, f)
	if err != nil {
		return err
	}

	return nil
}

func parseCount(s *token.Scanner, f *QueryField) error {
	f.IsCount = true
	var e token.Element
	for {
		ok := s.Next(&e)
		if !ok {
			return s.EarlyEndL("RPAREN")
		}
		if e.Token == token.RPAREN {
			break
		}
	}
	return nil
}

func parseField(s *token.Scanner, f *QueryField) error {
	var e token.Element
	ok := s.Next(&e)
	if !ok {
		return s.EarlyEndL("INDENT")
	}
	if !e.Indent {
		return e.NotMatchL("INDENT")
	}

	var next token.Element
	s.Cur(&next)
	if next.Token == token.PERIOD {
		f.Table = e.Get()
		s.Next(nil)
		ok = s.Next(&e)
		if !ok {
			return s.EarlyEndL("INDENT")
		}
		if !e.Indent {
			return e.NotMatchL("INDENT")
		}
	}
	f.Name = e.Get()

	return parseAlias(s, f)
}

func parseAlias(s *token.Scanner, f *QueryField) error {
	var e token.Element
	var last token.Element
	for {
		ok := s.Cur(&e)
		if !ok {
			return s.EarlyEndL("FROM")
		}
		if e.Token == token.COMMA || e.Token == _from {
			s.Pervious(&last, 1)
			if (last.Indent || last.String) && last.Get() != f.Name {
				f.Alias = last.Get()
			}
			if e.Token == token.COMMA {
				s.Next(nil)
			}
			return nil
		}
		s.Next(nil)
	}
}

type table struct {
	name  string
	alias string
}

func parseTables(s *token.Scanner) ([]*table, error) {
	var e token.Element
	var ok bool
	var tables []*table
	for {
		ok = s.Next(&e)
		if !ok {
			return tables, nil
		}

		if e.Token == _from ||
			e.Token == _join {
			// FROM or JOIN
			table := new(table)

			ok = s.Next(&e)
			if !ok {
				return nil, s.EarlyEndL("INDENT/STRING")
			}

			if !e.Indent && !e.String {
				return nil, e.NotMatchL("INDENT/STRING")
			}

			table.name = e.Get()

			// check alias
			for {
				ok = s.Cur(&e)
				if !ok {
					break
				}
				if e.Token == _as {
					s.Next(nil)
					continue
				}
				if !e.Indent {
					break
				}
				table.alias = e.Get()
				s.Next(nil)
				break
			}

			tables = append(tables, table)
		}
	}
}

func parsePh(s *token.Scanner) (*Statement, error) {
	state := new(Statement)
	var e token.Element
	var ok bool

	var sqls []string

	for {
		ok = s.Next(&e)
		if !ok {
			break
		}
		pre := e

		ph := new(placeholder)
		var phStr string
		switch e.Token {
		case _preStart:
			ph.pre = true
			phStr = "?"

		case _repStart:
			ph.pre = false
			phStr = "%v"

		case token.BREAK, token.SPACE:
			ok = s.Pervious(&pre, 2)
			if !ok {
				continue
			}
			if pre.Token == token.BREAK ||
				pre.Token == token.SPACE {
				continue
			}
			sqls = append(sqls, " ")
			continue

		default:
			str := e.Get()
			if e.String {
				quo := string(e.StringRune)
				str = quo + str + quo
			}
			sqls = append(sqls, str)
			continue
		}

		ok = s.Next(&e)
		if !ok {
			sqls = append(sqls, pre.Get())
			break
		}
		if e.Token != token.LBRACE {
			sqls = append(sqls, pre.Get(), e.Get())
			continue
		}

		var nameBucket []string
		for {
			ok = s.Next(&e)
			if !ok {
				return nil, s.EarlyEndL("RBRACE")
			}
			if e.Token == token.RBRACE {
				break
			}
			name := e.Get()
			if e.String {
				quo := string(e.StringRune)
				name = quo + name + quo
			}
			nameBucket = append(nameBucket, name)
		}
		if len(nameBucket) == 0 {
			return nil, e.FmtErrL("name is empty")
		}
		ph.name = strings.Join(nameBucket, "")

		sqls = append(sqls, phStr)
		state.phs = append(state.phs, ph)
	}
	state.Sql = strings.Join(sqls, "")
	state.Sql = strings.TrimSpace(state.Sql)

	return state, nil
}

func parseDynamic(s *token.Scanner) ([]*DynamicPart, error) {
	var dps []*DynamicPart
	var bucket []token.Element

	flushConst := func() error {
		if len(bucket) == 0 {
			return nil
		}
		os := token.CopyScanner(s, bucket)
		state, err := parsePh(os)
		if err != nil {
			return err
		}
		if len(state.Sql) == 0 {
			return nil
		}
		dp := &DynamicPart{
			Type:  DynamicTypeConst,
			State: state,
		}
		dps = append(dps, dp)
		bucket = nil
		return nil
	}

	hasCond := false

	var e token.Element
	for {
		ok := s.Next(&e)
		if !ok {
			break
		}
		pre := e

		if e.Token != token.PERCENT {
			// do not care not '%'
			bucket = append(bucket, e)
			continue
		}

		ok = s.Next(&e)
		if !ok {
			bucket = append(bucket, pre)
			break
		}
		if e.Token != token.LBRACE {
			bucket = append(bucket, pre, e)
			continue
		}

		if err := flushConst(); err != nil {
			return nil, err
		}

		dp, err := parseCond(s)
		if err != nil {
			return nil, err
		}
		dps = append(dps, dp)
		hasCond = true
	}
	if err := flushConst(); err != nil {
		return nil, err
	}

	if !hasCond || len(dps) == 0 {
		return nil, e.FmtErrL("can not find dynamic " +
			"condition, please add dynamic tag or " +
			"change sql type to static(remove dyn=true)")
	}

	return dps, nil
}

func parseCond(s *token.Scanner) (*DynamicPart, error) {
	skipSpace := func() {
		var tmp token.Element
		for {
			ok := s.Cur(&tmp)
			if !ok {
				return
			}
			if tmp.Token == token.SPACE ||
				tmp.Token == token.BREAK {
				s.Next(nil)
				continue
			}
			return
		}
	}
	var e token.Element
	var subBucket []token.Element
	dp := new(DynamicPart)
	skipSpace()
	ok := s.Next(&e)
	if !ok {
		return nil, s.EarlyEndL("IF/FOR")
	}
	if !e.Indent {
		return nil, e.NotMatchL("IF/FOR")
	}

	switch e.Get() {
	case "if":
		dp.Type = DynamicTypeIf

	case "for":
		dp.Type = DynamicTypeFor

	default:
		return nil, e.NotMatchL("IF/FOR")
	}
	if dp.Type == 0 {
		// Never trigger, prevent
		return nil, e.FmtErrL("[parseCond] CondType not found")
	}

	skipSpace()
	subBucket = nil
	for {
		ok := s.Next(&e)
		if !ok {
			return nil, s.EarlyEndL("RBRACE")
		}
		if e.Token == token.SPACE ||
			e.Token == token.BREAK {
			if dp.Type == DynamicTypeIf {
				subBucket = append(subBucket, e)
			}
			continue
		}
		if e.Token == token.RBRACE {
			break
		}
		subBucket = append(subBucket, e)
	}
	if len(subBucket) == 0 {
		return nil, e.FmtErrL("condition is empty")
	}

	os := token.CopyScanner(s, subBucket)
	err := parseCondState(os, dp)
	if err != nil {
		return nil, err
	}

	subBucket = nil
	for {
		ok := s.Next(&e)
		if !ok {
			return nil, s.EarlyEndL("EndTag")
		}
		if e.Token == token.PERCENT {
			var next token.Element
			ok = s.Cur(&next)
			if !ok {
				return nil, s.EarlyEndL("EndTag")
			}
			if next.Token == token.LBRACE {
				s.Next(nil)
				break
			}
		}
		subBucket = append(subBucket, e)
	}
	if len(subBucket) == 0 {
		return nil, e.FmtErrL("condition body is empty")
	}
	os = token.CopyScanner(s, subBucket)
	state, err := parsePh(os)
	if err != nil {
		return nil, err
	}
	dp.State = state

	subBucket = nil
	for {
		ok := s.Next(&e)
		if !ok {
			return nil, s.EarlyEndL("EndBody")
		}
		if e.Token == token.RBRACE {
			break
		}
		subBucket = append(subBucket, e)
	}
	if len(subBucket) != 1 {
		return nil, e.NotMatchL("EndBody")
	}
	e = subBucket[0]
	var expectEnd string
	switch dp.Type {
	case DynamicTypeIf:
		expectEnd = "endif"

	case DynamicTypeFor:
		expectEnd = "endfor"
	}

	if e.Get() != expectEnd {
		return nil, e.NotMatchL(expectEnd)
	}

	return dp, nil
}

func parseCondState(s *token.Scanner, dp *DynamicPart) error {
	switch dp.Type {
	case DynamicTypeIf:
		parseCondIf(s, dp)

	case DynamicTypeFor:
		return parseCondFor(s, dp)
	}
	return nil
}

func parseCondIf(s *token.Scanner, dp *DynamicPart) {
	es := s.Gets()
	bucket := make([]string, 0, len(es))
	for _, e := range es {
		if e.String {
			quo := string(e.StringRune)
			bucket = append(bucket, quo+e.Get()+quo)
		} else {
			bucket = append(bucket, e.Get())
		}
	}
	dp.IfCond = strings.Join(bucket, "")
}

func parseCondFor(s *token.Scanner, dp *DynamicPart) error {
	// %{for u in users join 'sss'} %{endfor}
	var e token.Element
	ok := s.Next(&e)
	if !ok {
		return s.EarlyEndL("INDENT")
	}
	if !e.Indent {
		return e.NotMatchL("INDENT")
	}

	dp.ForEle = e.Get()

	ok = s.Next(&e)
	if !ok {
		return s.EarlyEndL("in")
	}
	if e.Get() != "in" {
		return e.NotMatchL("in")
	}

	ok = s.Next(&e)
	if !ok {
		return s.EarlyEndL("INDENT")
	}
	if !e.Indent {
		return e.NotMatchL("INDENT")
	}
	dp.ForSlice = e.Get()

	ok = s.Next(&e)
	if !ok {
		return nil
	}
	if e.Get() != "join" {
		return e.NotMatchL("join")
	}

	ok = s.Next(&e)
	if !ok {
		return s.EarlyEndL("STRING")
	}
	if !e.String {
		return e.NotMatchL("STRING")
	}
	dp.ForJoin = e.Get()

	return nil
}
