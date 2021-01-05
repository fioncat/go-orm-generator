package scansql

import (
	"github.com/fioncat/go-gendb/compile/token"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/iter"
)

// Placeholders store all the placeholders in the sql
// statement, as well as the legal sql statement after
// the placeholder is processed. The original SQL
// statement may contain prepare and replace placeholders,
// which need to be extracted and replaced with "?"
// and "%v" respectively.
type Placeholders struct {
	SQL      string   `json:"sql"`
	Replaces []string `json:"replaces"`
	Prepares []string `json:"prepares"`
}

var phReplaceRunes = []rune{'%', 'v'}

// DoPlaceholders processes SQL statements containing
// placeholders and converts them into Placeholders
// struct to return.
func DoPlaceholders(path string, lineNum int, sql string) (*Placeholders, error) {
	var phs Placeholders
	iter := iter.New([]rune(sql))
	bucket := make([]rune, 0, len(sql))
	var r rune
	var idx int

phloop:
	for {
		idx = iter.NextP(&r)
		if idx < 0 {
			break
		}

		var isPre bool
		switch {
		case token.SQLPH_PRE.EqRune(r):
			isPre = true
			bucket = append(bucket, '?')

		case token.SQLPH_REP.EqRune(r):
			isPre = false
			bucket = append(bucket, phReplaceRunes...)

		default:
			if token.SPACE.EqRune(r) {
				// If it is a space, add only when
				// the previous character is not a space
				// and the entire bucket is not empty to
				// eliminate extra spaces.
				if len(bucket) == 0 {
					continue phloop
				}
				lastidx := len(bucket) - 1
				lastR := bucket[lastidx]
				if token.SPACE.EqRune(lastR) {
					continue phloop
				}
			}
			bucket = append(bucket, r)
			continue phloop
		}

		idx = iter.NextP(&r)
		if idx < 0 {
			return nil, errors.NewComp(path, lineNum,
				"the char must be '{', found: '%c'", r).
				WithCharNum(idx)
		}

		var nameBucket []rune
		for {
			idx = iter.NextP(&r)
			if idx < 0 {
				return nil, errors.NewComp(path, lineNum,
					"expect '}', but reach the end")
			}
			if token.RBRACE.EqRune(r) {
				break
			}
			nameBucket = append(nameBucket, r)
		}
		if len(nameBucket) == 0 {
			return nil, errors.NewComp(path, lineNum,
				"empty placeholder").WithCharNum(idx)
		}

		name := string(nameBucket)
		if isPre {
			phs.Prepares = append(phs.Prepares, name)
		} else {
			phs.Replaces = append(phs.Replaces, name)
		}
	}

	// If the last character is a space,
	// eliminate it (meaningless)
	lastidx := len(bucket) - 1
	if len(bucket) > 0 {
		last := bucket[lastidx]
		if token.SPACE.EqRune(last) {
			bucket = bucket[:lastidx]
		}
	}

	phs.SQL = string(bucket)
	return &phs, nil
}
