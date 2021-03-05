package base

import (
	"fmt"

	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/version"
)

type NextAction func(idx int, tag *Tag) (bool, error)

var ErrNotAccept = errors.New(`The "+gen" tag is not found before "package"`)

func Accept(lines []string, prefix string, next NextAction) (int, error) {
	hasTag := false
	verifyVersion := false

	var idx int
	for ; idx < len(lines); idx++ {
		line := lines[idx]
		tag, err := ParseTag(idx, prefix, line)
		if err != nil {
			return 0, err
		}
		if tag == nil {
			ok, err := next(idx, nil)
			if err != nil {
				return 0, err
			}
			if !ok {
				break
			}
			continue
		}
		hasTag = true
		err = checkMetaOptions(idx,
			tag.Options, &verifyVersion)
		if err != nil {
			return 0, err
		}
		ok, err := next(idx, tag)
		if err != nil {
			return 0, err
		}
		if !ok {
			break
		}
	}

	if !hasTag {
		return 0, ErrNotAccept
	}
	if !verifyVersion {
		return 0, fmt.Errorf(`missing "v" option in meta tag(s)`)
	}

	return idx, nil
}

func checkMetaOptions(idx int, opts []Option, vv *bool) error {
	for _, opt := range opts {
		if opt.Key != "v" {
			continue
		}
		if *vv {
			return errors.TraceFmt(idx+1,
				`version option "v" is duplcate`)
		}
		if opt.Value != version.Short {
			return errors.TraceFmt(idx+1,
				`version not match: "v"="%s", accept="%s"`,
				opt.Value, version.Short)
		}
		*vv = true
	}
	return nil
}
