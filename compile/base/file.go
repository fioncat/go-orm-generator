package base

import (
	"fmt"

	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/version"
)

// NextAction is an externally provided function for
// processing meta tags. It allows the outside to
// terminate the Accept process early by returning
// false or an error. If an error is returned, it will
// be returned by the Accept function a second time.
type NextAction func(idx int, tag *Tag) (bool, error)

// When the content is not accepted, the Accept function
// will return this error.
var ErrNotAccept = errors.New(`The "+gen" tag is not found before "package"`)

// Accept judges whether the given content is accepted.
// Two conditions need to be met:
//   - The pre-definition of "+gen" must be included in
//     the header of the file, which is called "meta tag".
//     There can be multiple meta tags.
//   - The "v" configuration must be included in the
//     pre-definition. And its value must be consistent with
//     the current version(version.Short).
// If the above conditions are not met, an ErrNotAccept
// error will be returned.
// Every time a meta tag is found, the passed next function
// will be called to allow the outside to customize the
// processing logic of the meta tag. If the content is accepted,
// the function returns the index of the last meta tag.
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

// Look for the version configuration in the meta tag. And
// verify that it complies.
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
