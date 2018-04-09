package bookeeper

import (
	"github.com/urfave/cli"
	"starchain/client"
	"encoding/hex"
	"fmt"
	"os"
	"starchain/account"
	"starchain/core/transaction"
	"starchain/crypto"
	"strconv"
	"math/rand"
	"bytes"
	"starchain/net/rpchttp"
)

func makeBookeeperTransaction(pubkey *crypto.PubKey,add bool,cert []byte,acc *account.Account) (string,error){
	tx,_ := transaction.NewBookKeeperTransaction(pubkey,add,cert,acc.PubKey())
	attr:= transaction.NewTxAttribute(transaction.Nonce,[]byte(strconv.FormatInt(rand.Int63(),10)))
	tx.Attributes = make([]*transaction.TxAttribute,0)
	tx.Attributes = append(tx.Attributes,&attr)
	if err := client.SignTransaction(acc,tx);err != nil{
		fmt.Println("sign transaction fail")
		os.Exit(1)
	}
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		fmt.Println("transaction serialize fail")
		os.Exit(1)
	}
	return hex.EncodeToString(buffer.Bytes()),nil


}

func bookeeperAction(c *cli.Context) error{
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	var pubkeyHex []byte
	var err error
	var add bool
	switch  {
	case c.String("add") != "":
		pubkeyHex,err = hex.DecodeString(c.String("add"))
		if err != nil{
			fmt.Println("decode pubkey error")
			os.Exit(1)
		}
		add = true
	case c.String("rm") != "":
		pubkeyHex,err = hex.DecodeString(c.String("rm"))
		if err != nil {
			fmt.Println("decode pubkey error")
			os.Exit(1)
		}
		add = false
	default:
		fmt.Println("missing vaild param: 'add' or 'rm' ")
		os.Exit(1)
	}
	cert := c.String("cert")
	if cert == ""{
		fmt.Println("missing param :'cert'")
		os.Exit(1)
	}
	walletName := c.String("name")
	if walletName == ""{
		walletName = account.WalletFileName
	}
	wallet := client.OpenWallet(walletName,client.WalletPassword(c.String("password")))
	if wallet != nil {
		fmt.Println("wallet not exists")
		os.Exit(1)
	}
	acc,_ := wallet.GetDefaultAccount()
	pubkey,_ := crypto.DecodePoint(pubkeyHex)
	tx,err := makeBookeeperTransaction(pubkey,add,[]byte(cert),acc)
	resp,err := rpchttp.Call(client.Address(),"sendrawtransaction",0,[]interface{}{tx})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	client.FormatOutput(resp)
	return nil

}

func NewCommond() *cli.Command{
	return &cli.Command{
		Name:"bookeeper",
		Usage:"add or remove bookeeper",
		Description:"add or remove bookeeper with console",
		ArgsUsage:"[args",
		Flags:[]cli.Flag{
					cli.StringFlag{
						Name:"add,a",
						Usage:"add a new bookeeyper",
					},
					cli.StringFlag{
						Name:"rm",
						Usage:"remove a exists bookeeper",
					},
					cli.StringFlag{
						Name:"cert,c",
						Usage:"authorized certificate",
					},
		},
		Action:bookeeperAction,
		OnUsageError:func(c *cli.Context,err error,issubcommond bool) error{
			client.PrintError(c,err,"bookeeper")
			return cli.NewExitError("",1)
		},
	}
}
