package sql

import "github.com/fioncat/go-gendb/misc/col"

const (
	SELECT = "SELECT"
	FROM   = "FROM"
	UPDATE = "UPDATE"
	INSERT = "INSERT"
	DELETE = "DELETE"

	IFNULL = "IFNULL"
	COUNT  = "COUNT"

	COMMA    = ","
	DOT      = "."
	LPAREN   = "("
	RPAREN   = ")"
	LPREPARE = "${"
	LREPLACE = "#{"
	RBRACE   = "}"

	SPACE = " "
)

var Keywords = col.NewSetBySlice(
	SELECT, FROM, UPDATE, INSERT, DELETE,
	IFNULL, COUNT,
	COMMA, DOT, LPAREN, RPAREN, LPREPARE,
	LREPLACE, RBRACE,
)
