package cmd

import (
	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/database/tools/mock"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var mockCmd = &cmdt.Command{
	Name: "mock",
	Pv:   (*mock.Arg)(nil),

	Usage: help.MockUsage,
	Help:  help.Mock,

	Action: func(p interface{}) error {
		return mock.Do(p.(*mock.Arg))
	},
}
