package wallet

import (
	"fmt"
	"os"
	"starchain/account"
	"strconv"
	"starchain/common"
	"encoding/binary"
	"bytes"
	"syscall"
	"os/signal"
	"starchain/events/signalset"
	"github.com/urfave/cli"
	"starchain/client"
	"starchain/common/passwd"
)

func walletAction(c *cli.Context) error{
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")

	if name == ""{
		name = "wallet.dat"
	}

	pwd :=c.String("password")


	switch{
	case c.Bool("create"):
		if common.FileExisted(name){
			fmt.Printf("caution: %s already exists!\n",name)
			os.Exit(1)
		}else{
			wallet,err := account.Create(name,getConfirmPwd())
			if err != nil {
				fmt.Println("create wallet error")
				os.Exit(1)
			}
			showAccountInfo(wallet)
		}
	case c.String("list") != "":
		item := c.String("list")
		wallet, err := account.Open(name, client.WalletPassword(pwd))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		switch item {
		case "account":
			showAccountInfo(wallet)
		case "mainaccount":
			showDefaultAccountInfo(wallet)
		case "balance":
			showBalancesInfo(wallet)
		case "height":
			showHeightInfo(wallet)
		default:
			fmt.Println("missing parameter for [--list]")
		}
	case c.Bool("changepassword"):
		fmt.Printf("wallet:%s\n",name)
		wallet, err := account.Open(name, client.WalletPassword(pwd))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("# please input new password #")
		newPassword, _ := passwd.GetConfirePwd()
		if ok := wallet.ChangePassword([]byte(pwd), newPassword); !ok {
			fmt.Fprintln(os.Stderr, "failed to change your wallet password")
			os.Exit(1)
		}
		fmt.Println("password changed successfully!")

		return nil
	case c.Bool("reset"):
		wallet, err := account.Open(name, client.WalletPassword(pwd))
		if err != nil {
			fmt.Println("open wallet failt:", err)
			os.Exit(1)
		}
		if err := wallet.Rebuild(); err != nil {
			fmt.Fprintln(os.Stderr, "delete coins info from wallet file error")
			os.Exit(1)
		}
		fmt.Printf("%s was reset successfully\n", name)
		return nil

	case c.Int("addaccount") > 0:
		num := c.Int("addaccount")
		if num > 0{
			wallet, err := account.Open(name, client.WalletPassword(pwd))
			if err != nil {
				fmt.Println("open wallet failed :",err)
				os.Exit(1)
			}
			go processSignals(wallet)

			for i := 0; i < num; i++ {
				account, err := wallet.CreateAccount()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				if err := wallet.CreateContract(account); err != nil {
					wallet.DeleteAccount(account.ProgramHash)
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
			fmt.Printf("%d accounts created\n", num)
			return nil

		}else{
			fmt.Println("please input account number > 0")
		}
	}
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:		"wallet",
		Usage:		"user wallet operator",
		Description:	"you can control your wallet with in console",
		ArgsUsage:	"[args]",
		Flags:		[]cli.Flag{
			cli.BoolFlag{
				Name:"create,c",
				Usage:"create wallet",
			},
			cli.StringFlag{
				Name:"list,l",
				Usage:"list account info [account,mainaccount,balance,pubkey]",
			},
			cli.BoolFlag{
				Name: "changepassword",
				Usage:"change your wallet password",
			},
			cli.BoolFlag{
				Name:"reset,r",
				Usage:"rebuild wallet data from chaindata",
			},
			cli.IntFlag{
				Name:"addaccount",
				Usage:"add [value] account in wallet",
			},
			cli.StringFlag{
				Name:"password,p",
				Usage:"wallet passwrod",
			},
			cli.StringFlag{
				Name:"name",
				Usage:"the name of your wallet file name",
			},
		},
		Action:walletAction,
		OnUsageError:func(c *cli.Context,err error,isSubcommand bool)error {
			client.PrintError(c, err, "wallet")
			return cli.NewExitError("", 1)
		},
	}
}


func getConfirmPwd() []byte{
	tmp, err := passwd.GetConfirePwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return tmp
}

func showAccountInfo(wallet account.Client){
	accounts := wallet.GetAccounts()
	fmt.Println(" id        address \t\t\t\t public key ")
	fmt.Println("####       ####### \t\t\t\t ###########")
	for i,account := range accounts{
		address,_ := account.ProgramHash.ToAddress()
		pubkey,_ := account.PubKey().EncodePoint(true)
		fmt.Printf("%4s  %s %s\n", strconv.Itoa(i), address, common.BytesToHexString(pubkey))
	}
}


func showDefaultAccountInfo(wallet account.Client) {
	mainAccount, err := wallet.GetDefaultAccount()
	if nil == err {
		fmt.Println(" id   address\t\t\t\t public key")
		fmt.Println("####  #######\t\t\t\t ##########")

		address, _ := mainAccount.ProgramHash.ToAddress()
		publicKey, _ := mainAccount.PublicKey.EncodePoint(true)
		fmt.Printf("%4s  %s %s\n", strconv.Itoa(0), address, common.BytesToHexString(publicKey))
	} else {
		fmt.Println("GetDefaultAccount err! ", err.Error())
	}
}


func showPrivateKeysInfo(wallet account.Client) {
	accounts := wallet.GetAccounts()
	fmt.Println(" id   address\t\t\t\t public key\t\t\t\t\t\t\t    private key")
	fmt.Println("####  #######\t\t\t\t ##########\t\t\t\t\t\t\t    ###########")
	for i, account := range accounts {
		address, _ := account.ProgramHash.ToAddress()
		publicKey, _  := account.PublicKey.EncodePoint(true)
		privateKey := account.PrivateKey
		fmt.Printf("%4s  %s %s %s\n", strconv.Itoa(i), address, common.BytesToHexString(publicKey), common.BytesToHexString(privateKey))
	}
}


func showBalancesInfo(wallet account.Client) {
	coins := wallet.GetCoins()
	assets := make(map[common.Uint256]common.Fixed64)
	for _, out := range coins {
		if out.AddressType == account.SingleSign {
			if _, ok := assets[out.Output.AssetID]; !ok {
				assets[out.Output.AssetID] = out.Output.Value
			} else {
				assets[out.Output.AssetID] += out.Output.Value
			}
		}
	}
	if len(assets) == 0 {
		fmt.Println("no assets")
		return
	}
	fmt.Println(" id   asset id \t\t\t\t\t\t\t\t amount")
	fmt.Println("#####  ####### \t\t\t\t\t\t\t\t ########")
	i := 0
	for id, amount := range assets {
		fmt.Printf("%4s  %s  %v\n", strconv.Itoa(i), common.BytesToHexString(id.ToArrayReverse()), amount)
		i++
	}
}



func showHeightInfo(wallet *account.ClientImpl) {
	h, _ := wallet.LoadStoredData("Height")
	var height uint32
	binary.Read(bytes.NewBuffer(h), binary.LittleEndian, &height)
	fmt.Println("Height: ", height)
}



func processSignals(wallet *account.ClientImpl) {
	sigHandler := func(signal os.Signal, v interface{}) {
		switch signal {
		case syscall.SIGINT:
			fmt.Println("Caught SIGINT signal, existing...")
		case syscall.SIGTERM:
			fmt.Println("Caught SIGTERM signal, existing...")
		}
		// hold the mutex lock to prevent any wallet db changes
		wallet.FileStore.Lock()
		os.Exit(0)
	}
	signalSet := signalset.New()
	signalSet.Register(syscall.SIGINT, sigHandler)
	signalSet.Register(syscall.SIGTERM, sigHandler)
	sigChan := make(chan os.Signal, account.MaxSignalQueueLen)
	signal.Notify(sigChan)
	for {
		select {
		case sig := <-sigChan:
			signalSet.Handle(sig, nil)
		}
	}
}