package scanner

import (
	"io/ioutil"
	"strings"

	"github.com/fioncat/go-gendb/misc/col"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/scanner/token"
)

type SQLResult struct {
	Path string `json:"path"`

	Sqls []SQL `json:"sqls"`

	nameSet col.Set
}

type SQL struct {
	Path     string        `json:"-"`
	Line     int           `json:"line"`
	Name     string        `json:"name"`
	Tokens   []token.Token `json:"-"`
	SQL      string        `json:"sql"`
	TokenStr string        `json:"tokens"`
}

var (
	errEmptyName      = errors.New("name is empty")
	errNameDuplcate   = errors.New("name is duplcate")
	errSQLUnknownType = errors.New("unknown sql type")
	errEmptySQL       = errors.New("sql statement is empty")
)

func SQLFile(path string, debug bool) (*SQLResult, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var r SQLResult
	r.nameSet = col.NewSet(0)
	scanner := newLines(string(data))
	for {
		line, num := scanner.next()
		if num == -1 {
			break
		}
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "-- !") {
			name := strings.TrimLeft(line, "-- !")
			name = strings.TrimSpace(name)
			if name == "" {
				return nil, errors.Line(errEmptyName, num)
			}
			if r.nameSet.Exists(name) {
				return nil, errors.Line(errNameDuplcate, num)
			}

			var sql SQL
			sql.Name = name
			r.nameSet.Add(name)

			// Scan sql statement...
			var sqlLines []string
			for {
				pickLine := scanner.pick()
				if strings.HasPrefix(pickLine, "-- !") {
					break
				}
				line, num := scanner.next()

				if num == -1 {
					break
				}
				if line == "" || strings.HasPrefix(line, "--") {
					continue
				}
				sqlLines = append(sqlLines, line)
			}
			sqlContent := strings.Join(sqlLines, " ")
			sql.Path = path
			sql.Line = num
			sql.SQL = sqlContent
			sql.Tokens, err = scanSQL(sqlContent)
			if err != nil {
				return nil, errors.Line(err, num)
			}
			if debug {
				sql.TokenStr = token.Join(sql.Tokens)
			}

			r.Sqls = append(r.Sqls, sql)
		}
	}
	return &r, nil
}

var SQLKeywords = col.NewSetBySlice(
	"SELECT", "FROM", "IFNULL", "COUNT",
)

func scanSQL(s string) ([]token.Token, error) {
	if len(s) <= 7 {
		return nil, errEmptySQL
	}
	master := s[:6]
	masterUp := strings.ToUpper(master)
	switch masterUp {
	case "UPDATE", "DELETE", "INSERT":
		return []token.Token{token.NewFlag(masterUp)}, nil
	case "SELECT":
	default:
		return nil, errSQLUnknownType
	}

	scanner := newChars(s)
	var tokens []token.Token
	var bucket []rune
	addBucket := func() {
		if len(bucket) == 0 {
			return
		}
		key := upBucket(bucket)
		if SQLKeywords.Exists(key) {
			tokens = append(tokens, token.NewFlag(key))
		} else {
			tokens = append(tokens,
				token.NewIndent(string(bucket)))
		}
		bucket = nil
	}
	addFlag := func(r rune) {
		addBucket()
		tokens = append(tokens, token.NewFlag(string(r)))
	}
	for {
		ch, ok := scanner.next()
		if !ok {
			addBucket()
			break
		}
		if ch == ' ' || ch == '\t' || ch == '\n' {
			addBucket()
			continue
		}
		switch ch {
		case '(':
			addFlag('(')

		case ')':
			addFlag(')')

		case '.':
			addFlag('.')

		case ',':
			addFlag(',')

		case '\'':
			fallthrough
		case '`':
			addBucket()
			var quotes []rune
			for {
				ch, ok := scanner.next()
				if !ok {
					return nil, errors.New("Quote bad format")
				}
				if ch == '\'' || ch == '`' {
					break
				}
				quotes = append(quotes, ch)
			}
			quoteToken := token.NewIndent(string(quotes))
			tokens = append(tokens, quoteToken)

		default:
			bucket = append(bucket, ch)
		}
	}
	addBucket()
	return tokens, nil
}

func upBucket(bucket []rune) string {
	s := string(bucket)
	return strings.ToUpper(s)
}

type Placeholders struct {
	SQL      string
	Replaces []string
	Prepares []string
}

func SQLPlaceholders(s string) (*Placeholders, error) {
	chs := make([]rune, 0, len(s))
	p := newChars(s)
	ph := new(Placeholders)
	for {
		ch, ok := p.next()
		if !ok {
			break
		}
		if ch != '$' && ch != '#' {
			chs = append(chs, ch)
			continue
		}
		nextCh, ok := p.next()
		if !ok || nextCh != '{' {
			return nil, errors.Fmt(`the char next '$'/'#'`+
				` must be '{', found: '%s'`, string(nextCh))
		}

		var nameChs []rune
		for {
			nextCh, ok = p.next()
			if !ok {
				return nil, errors.New(`can not find '}'` +
					` for placeholder`)
			}
			if nextCh == '}' {
				break
			}
			nameChs = append(nameChs, nextCh)
		}
		if len(nameChs) == 0 {
			return nil, errors.New("found empty placeholder")
		}
		switch ch {
		case '$':
			ph.Prepares = append(ph.Prepares, string(nameChs))
			chs = append(chs, '?')

		case '#':
			ph.Replaces = append(ph.Replaces, string(nameChs))
			chs = append(chs, '%')
			chs = append(chs, 'v')
		}
	}
	ph.SQL = string(chs)
	ph.SQL = strings.Join(strings.Fields(ph.SQL), " ")
	return ph, nil
}
