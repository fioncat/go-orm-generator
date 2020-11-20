package addr

import (
	"strconv"
	"strings"

	"github.com/fioncat/go-gendb/misc/errors"
)

type Addr struct {
	Host string
	Port int
}

func Parse(s string, defaultPort int) (*Addr, error) {
	addr := new(Addr)
	tmp := strings.Split(s, ":")
	addr.Host = tmp[0]
	if len(tmp) == 1 {
		addr.Port = defaultPort
		return addr, nil
	}

	if len(tmp) == 2 {
		var err error
		addr.Port, err = strconv.Atoi(tmp[1])
		if err != nil || addr.Port <= 0 {
			return nil, errors.Fmt("invalid port '%s'", tmp[1])
		}
		return addr, nil
	}
	return nil, errors.Fmt("addr '%s' format error", s)
}
