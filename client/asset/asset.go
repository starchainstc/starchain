package asset

import (
	"github.com/urfave/cli"
	"math/rand"
	."starchain/common"
	"fmt"
	"os"
	"bytes"
	"starchain/core/transaction"
	"starchain/client"
	"starchain/util"
	"starchain/net/rpchttp"
)

const (
	DEFNAMELEN = 4
)


//if name is nil the general a name in rand
func parseAssetName(c *cli.Context) string {
	name := c.String("name")
	if name == "" {
		rbuf := make([]byte, DEFNAMELEN)
		rand.Read(rbuf)
		name = "STC-" + BytesToHexString(rbuf)
	}

	return name
}

func parseAssetId(c *cli.Context) Uint256{
	id := c.String("asset")
	if id == "" {
		fmt.Println("missing parameter [--asset]")
		os.Exit(1)
	}
	var assetID Uint256
	assetBytes, err := HexStringToBytesReverse(id)
	if err != nil {
		fmt.Println("invalid asset ID")
		os.Exit(1)
	}
	if err := assetID.Deserialize(bytes.NewReader(assetBytes)); err != nil {
		fmt.Println("invalid asset hash")
		os.Exit(1)
	}
	return assetID
}

func parseAddress(c *cli.Context) string {
	if address := c.String("to"); address != "" {
		_, err := ToScriptHash(address)
		if err != nil {
			fmt.Println("invalid receiver address")
			os.Exit(1)
		}
		return address
	} else {
		fmt.Println("missing parameter [--to]")
		os.Exit(1)
	}
	return ""
}

func parseHeight(c *cli.Context) int64 {
	height := c.Int64("height")
	if height != -1 {
		return height
	} else {
		fmt.Println("invalid parameter [--height]")
		os.Exit(1)
	}

	return 0
}
func parseAction(c *cli.Context) error {
	if c.NumFlags() == 0{
		cli.ShowSubcommandHelp(c)
		return nil
	}
	value := c.String("value")
	if value == ""{
		fmt.Println("missing parameter [--value]")
		os.Exit(1)
	}
	walletName := c.String("wallet")
	if walletName == ""{
		walletName = "wallet.dat"
	}
	pwd := c.String("password")
	var txn *transaction.Transaction
	var buf bytes.Buffer
	var err error
	switch {
	case c.Bool("reg"):
		name := parseAssetName(c)
		wallet := client.OpenWallet(walletName,client.WalletPassword(pwd))
		txn,err = util.MakeRegTransaction(wallet,name,string(value))
		if err = txn.Serialize(&buf);err != nil{
			fmt.Println("transaction serialize err",err)
			return err
		}
		resp, err := rpchttp.Call(client.Address(), "sendrawtransaction", 0, []interface{}{BytesToHexString(buf.Bytes())})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		client.FormatOutput(resp)
		return nil
	case c.Bool("issue"):
		assetId := parseAssetId(c)
		addr:= parseAddress(c)
		wallet := client.OpenWallet(walletName,client.WalletPassword(pwd))
		txn, err = util.MakeIssueTransaction(wallet, assetId, addr, string(value))
		if err = txn.Serialize(&buf);err != nil{
			fmt.Println("transaction serialize err",err)
			return err
		}
		resp, err := rpchttp.Call(client.Address(), "sendrawtransaction", 0, []interface{}{BytesToHexString(buf.Bytes())})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		client.FormatOutput(resp)
	case c.Bool("transfer"):
		assetId := parseAssetId(c)
		addr:= parseAddress(c)
		client.OpenWallet(walletName,client.WalletPassword(pwd))
		resp,err := rpchttp.Call(client.Address(),"sendtoaddress",0,[]interface{}{assetId,addr,value})
		if err != nil {
			fmt.Println("transfer error:",err)
			os.Exit(1)
		}
		client.FormatOutput(resp)
		return nil
	default:
		cli.ShowSubcommandHelp(c)
		return nil
	}

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}

func NewCommand() *cli.Command{
	return &cli.Command{
		Name:		"asset",
		Usage:		"reg,issue,transacion asset",
		Description:	"control asset in console",
		ArgsUsage:	"[args]",
		Flags:		[]cli.Flag{
					cli.BoolFlag{
						Name:	"reg,r",
						Usage:	"regist a new asset",
					},
					cli.BoolFlag{
						Name: "issue,i",
						Usage:"issue a asset for some address",
					},
					cli.BoolFlag{
						Name:"transfer,t",
						Usage:"transfer some issue to some addresss",
					},
					cli.StringFlag{
						Name:"asset,a",
						Usage:"asset id",
					},
					cli.StringFlag{
						Name:"name",
						Usage:"the name of asset",
					},
					cli.StringFlag{
						Name:"to",
						Usage:"the address for receiver address",
					},
					cli.StringFlag{
						Name:"password,p",
						Usage:"the password of wallet",
					},
					cli.StringFlag{
						Name:"wallet,w",
						Usage:"the name of wallet",
					},
					cli.StringFlag{
						Name:"value",
						Usage:"the amount of transfer",
					},

		},
		Action:		parseAction,
		OnUsageError:	func(c *cli.Context,err error,isSubCommon bool)error {
			client.PrintError(c, err, "asset")
			return cli.NewExitError("", 1)
		},
	}
}



