package client

import (
	"time"
	"starchain/crypto"
	"starchain/common/log"
	"math/rand"
	"starchain/common/config"
	"fmt"
	"os"
	"starchain/account"
	"strconv"
	"bytes"
	"encoding/json"
	"github.com/urfave/cli"
	"starchain/common/passwd"
	"starchain/core/transaction"
	"starchain/core/signature"
	"starchain/core/contract"
)

var (
	Ip   string
	Port string
)

func init() {
	log.Init()
	crypto.SetAlg(config.Parameters.EncryptAlg)
	//seed transaction nonce
	rand.Seed(time.Now().UnixNano())
}

func OpenWallet(name string, passwd []byte) account.Client {
	if name == account.WalletFileName {
		fmt.Println("Using default wallet: ", account.WalletFileName)
	}
	wallet, err := account.Open(name, passwd)
	if err != nil {
		fmt.Println("Failed to open wallet: ", name)
		os.Exit(1)
	}
	return wallet
}


func NewIpFlag() cli.Flag {
	return cli.StringFlag{
		Name:        "ip",
		Usage:       "node's ip address",
		Value:       "localhost",
		Destination: &Ip,
	}
}

func NewPortFlag() cli.Flag {
	return cli.StringFlag{
		Name:        "port",
		Usage:       "node's RPC port",
		Value:       strconv.Itoa(config.Parameters.HttpJsonPort),
		Destination: &Port,
	}
}

func Address() string {
	address := "http://" + Ip + ":" + Port
	return address
}

func PrintError(c *cli.Context, err error, cmd string) {
	fmt.Println("Incorrect Usage:", err)
	fmt.Println("")
	cli.ShowCommandHelp(c, cmd)
}

func FormatOutput(o []byte) error {
	var out bytes.Buffer
	err := json.Indent(&out, o, "", "\t")
	if err != nil {
		return err
	}
	out.Write([]byte("\n"))
	_, err = out.WriteTo(os.Stdout)
	return err
}

// WalletPassword prompts user to input wallet password when password is not
// specified from command line
func WalletPassword(passwword string) []byte {
	if passwword == "" {
		tmppasswd, _ := passwd.GetPwd()
		return tmppasswd
	} else {
		return []byte(passwword)
	}
}

func SignTransaction(signer *account.Account, tx *transaction.Transaction) error {
	signature, err := signature.SignBySigner(tx, signer)
	if err != nil {
		fmt.Println("SignBySigner failed.")
		return err
	}
	transactionContract, err := contract.CreateSignatureContract(signer.PubKey())
	if err != nil {
		fmt.Println("CreateSignatureContract failed.")
		return err
	}
	transactionContractContext := contract.NewContractContext(tx)
	if err := transactionContractContext.AddContract(transactionContract, signer.PubKey(), signature); err != nil {
		fmt.Println("AddContract failed")
		return err
	}
	tx.SetPrograms(transactionContractContext.GetPrograms())
	return nil
}

