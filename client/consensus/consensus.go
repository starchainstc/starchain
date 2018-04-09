package consensus

import (
	"github.com/urfave/cli"
	"fmt"
	"os"
	"starchain/net/rpchttp"
	"starchain/client"
)

func consAction(c *cli.Context) error{
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	switch {
	case c.Bool("start"):
		resp, err := rpchttp.Call(client.Address(), "startconsensus", 0, []interface{}{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		client.FormatOutput(resp)
	case c.Bool("stop"):
		resp, err := rpchttp.Call(client.Address(), "stopconsensus", 0, []interface{}{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		client.FormatOutput(resp)
	}
	return nil
}

func NewCommond() *cli.Command{
	return &cli.Command{
		Name:"consensus",
		Usage:"start or stop consensus with args --start, --stop",
		Description:"control consensus of node with console",
		Flags:[]cli.Flag{
			cli.BoolFlag{
				Name:"start",
				Usage:"start consensus",
			},
			cli.BoolFlag{
				Name:"stop",
				Usage:"stop consensus",
			},
		},
		Action:consAction,
		OnUsageError: func(context *cli.Context, err error, isSubcommand bool) error {
			client.PrintError(context,err,"consensus")
			return cli.NewExitError("",1)
		},
	}
}
