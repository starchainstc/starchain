package main

import (
	"github.com/urfave/cli"
	"starchain/client"
	"starchain/client/asset"
	"os"
	"starchain/client/wallet"
	"starchain/client/bookeeper"
	"starchain/client/consensus"
)

func main(){
	app := cli.NewApp()
	app.Name = "stc-client"
	app.Version = "v0.0.1"
	app.HelpName = "stc-client"
	app.Usage = "command line tool for STC blockchain"
	app.UsageText = "stc-client [global options] command [command options] [args]"
	app.HideHelp = false
	app.HideVersion = false

	app.Flags = []cli.Flag{
		client.NewIpFlag(),
		client.NewPortFlag(),
	}
	app.Commands = []cli.Command{
		*asset.NewCommand(),
		*wallet.NewCommand(),
		*bookeeper.NewCommond(),
		*consensus.NewCommond(),
	}
	app.Run(os.Args)
}
