package store

type Conn struct {
	Key  string
	Addr string
	User string
	Pass string

	Database string
}

func SaveConn(conn *Conn) error {
	return Save("conn", conn.Key, conn, 0)
}

func GetConn(key string) (*Conn, error) {
	var conn Conn
	found, err := Get("conn", key, &conn)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &conn, nil
}

func DelConn(key string) error {
	return Remove("conn", key)
}
