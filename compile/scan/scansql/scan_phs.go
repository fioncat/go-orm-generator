package scansql

import (
	"strings"

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

// dynamic type enum
const (
	DynamicTypeConst = iota
	DynamicTypeIf
	DynamicTypeFor
)

// DynamicPart stores a part of the scanned dynamic SQL.
// It can be a constant sql segment, or an if/for conditional
// sql segment. The constant sql section is a statement
// that is not covered by the dynamic placeholder in the
// sql statement; the dynamic sql section is the statement
// section covered by the dynamic placeholder.
type DynamicPart struct {
	// Type represents the type of the sql statement segment.
	// The specific value is the "DynamicTypeXXX" constant
	// under the current package
	Type int

	// Cond represents dynamic condition content. Applies
	// only to if/for statement segments, storing original,
	// unresolved conditions or loop statements.
	Cond string

	// SQL represents the sql statement content of the statement
	// segment (replace and prepare placeholders can exist)
	SQL *Placeholders
}

// DoDynamic scans the dynamic sql statement, extracts all
// the dynamic placeholders in it, divides the sql statement
// into multiple statement segments and returns.
func DoDynamic(path string, lineNum int, sql string) ([]*DynamicPart, error) {
	var parts []*DynamicPart
	var bucket []rune
	flushConst := func() error {
		if len(bucket) > 0 {
			phs, err := DoPlaceholders(path, lineNum, string(bucket))
			if err != nil {
				return err
			}
			part := &DynamicPart{
				Type: DynamicTypeConst,
				SQL:  phs,
			}
			if part.SQL.SQL == "" {
				return nil
			}
			parts = append(parts, part)
		}
		bucket = nil
		return nil
	}

	hasCond := false

	iter := iter.New([]rune(sql))
	var ch rune
	var idx int

mainLoop:
	for {
		idx = iter.NextP(&ch)
		if idx < 0 {
			break
		}

		switch ch {
		case token.PLUS.Rune():
			idx = iter.NextP(&ch)
			if idx < 0 {
				bucket = append(bucket, token.PLUS.Rune())
				break mainLoop
			}
			if !token.LBRACE.EqRune(ch) {
				// not "+{", treat it as normal indent.
				bucket = append(bucket, token.PLUS.Rune())
				bucket = append(bucket, ch)
				continue mainLoop
			}

			// Save previous SQL(const)
			if err := flushConst(); err != nil {
				return nil, err
			}

			part, err := condition(path, lineNum, iter)
			if err != nil {
				return nil, err
			}
			parts = append(parts, part)

			hasCond = true

		default:
			bucket = append(bucket, ch)

		} // switch case

	} // main loop

	if err := flushConst(); err != nil {
		return nil, err
	}

	if !hasCond || len(parts) == 0 {
		return nil, errors.NewComp(path, lineNum,
			"can not find dynamic condition, please add condition"+
				" or change the sql method type to static(starts "+
				"with \"--!\")")
	}

	return parts, nil
}

// scan condition tag: "+{if/for xxx} xxx +{endif/endfor}"
func condition(path string, lineNum int, iter *iter.Iter) (*DynamicPart, error) {
	var ch rune
	var idx int
	// The next word must be 'if' or 'for
	var subBucket []rune
	var condType int
typeLoop:
	for {
		idx = iter.NextP(&ch)
		if idx < 0 {
			return nil, errors.NewComp(path, lineNum,
				"expect 'if' or 'for', but reach the end")
		}
		if token.SPACE.EqRune(ch) {
			continue typeLoop
		}
		subBucket = append(subBucket, ch)
		switch len(subBucket) {
		case 2:
			if subBucket[0] == 'f' && subBucket[1] == 'o' {
				continue typeLoop
			}

			s := string(subBucket)
			if !token.IF.EqString(s) {
				return nil, errors.NewComp(path, lineNum,
					"expect 'if' or 'fo', found: %s", s).
					WithCharNum(idx)
			}
			condType = DynamicTypeIf
			break typeLoop

		case 3:
			s := string(subBucket)
			if !token.FOR.EqString(s) {
				return nil, errors.NewComp(path, lineNum,
					"expect 'for', found: %s", s)
			}
			condType = DynamicTypeFor
			break typeLoop
		}
	}

	if condType == 0 {
		// Never trigger, prevent.
		return nil, errors.NewComp(path, lineNum,
			"unknown condType")
	}

	subBucket = nil
	// Cond
condLoop:
	for {
		idx = iter.NextP(&ch)
		if idx < 0 {
			return nil, errors.NewComp(path, lineNum,
				"expect <condition>, but reach the end")
		}
		if token.RBRACE.EqRune(ch) {
			// condition end
			break condLoop
		}
		subBucket = append(subBucket, ch)
	}
	if len(subBucket) == 0 {
		return nil, errors.NewComp(path, lineNum,
			"condition is empty").WithCharNum(idx)
	}

	condStr := string(subBucket)
	condStr = strings.TrimSpace(condStr)

	subBucket = nil
	// Body
bodyLoop:
	for {
		idx = iter.NextP(&ch)
		if idx < 0 {
			return nil, errors.NewComp(path, lineNum,
				"missing end tag for condition if/for")
		}
		if token.PLUS.EqRune(ch) {
			// +{ : end_tag
			var nextCh rune
			iter.Pick(&nextCh)
			if token.LBRACE.EqRune(nextCh) {
				iter.Next(nil)
				break bodyLoop
			}
		}
		subBucket = append(subBucket, ch)
	}
	if len(subBucket) == 0 {
		return nil, errors.NewComp(path, lineNum,
			"condition body is empty").WithCharNum(idx)
	}
	body := string(subBucket)
	phs, err := DoPlaceholders(path, lineNum, body)
	if err != nil {
		return nil, err
	}
	subBucket = nil
	// End Tag
endLoop:
	for {
		idx = iter.NextP(&ch)
		if idx < 0 {
			return nil, errors.NewComp(path, lineNum,
				"missing end tag")
		}
		subBucket = append(subBucket, ch)
		switch len(subBucket) {
		// endif}
		case 6:
			if condType == DynamicTypeFor {
				continue endLoop
			}
			s := string(subBucket)
			if s == "endif}" {
				break endLoop
			}
			return nil, errors.NewComp(path, lineNum,
				"expect end if tag, found: %s", s)

			// endfor}
		case 7:
			s := string(subBucket)
			if condType == DynamicTypeIf {
				return nil, errors.NewComp(path, lineNum,
					"for endtag mismatch, found: %s", s)
			}
			if s == "endfor}" {
				break endLoop
			}

			return nil, errors.NewComp(path, lineNum,
				"expect end for tag, found: %s", s)
		}
	}

	// all parse done, add the condition
	part := &DynamicPart{
		Type: condType,
		SQL:  phs,
		Cond: condStr,
	}
	return part, nil
}
