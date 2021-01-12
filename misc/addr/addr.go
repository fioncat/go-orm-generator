package addr

import (
	"fmt"
	"strconv"
	"strings"
)

// Addr represents a resolved address.
// The format is: "{host}[:{port}]" where port can be default
type Addr struct {
	Host string
	Port int
}

// Parse converts the string of "{host}[:{port}]" into the
// data of the Addr structure. If port is not provided,
// defaultPort is used as the port. If the parsing error
// (for example, the port is not an integer) will return
// an error.
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
			return nil, fmt.Errorf("invalid port '%s'", tmp[1])
		}
		return addr, nil
	}
	return nil, fmt.Errorf("addr '%s' format error", s)
}
