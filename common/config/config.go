package config

import (
	"bytes"
	"log"
	"io/ioutil"
	"os"
	"encoding/json"
)

const (
	DefaultConfigFilename = "./config/config.json"
)

var Version string

type Configuration struct {
	MagicCode       int64              `json:"MagicCode"`  //magic code for blockchain data file
	Version         int                `json:"Ver"`  //version of starchain
	SeedList        []string           `json:"VerifyList"`  //list of seed
	BookKeepers     []string           `json:"BookKeepers"` // The default book keepers' publickey
	HttpRestPort    int                `json:"RestPort"`
	HttpRestStart	bool			   `json:"RestStart"`
	RestCertPath    string             `json:"RestCertPath"` // certification file path
	RestKeyPath     string             `json:"RestKeyPath"` //certification key file path
	HttpInfoPort    uint16             `json:"HttpApiPort"` // http api port
	HttpInfoStart   bool               `json:"HttpApiStart"`
	HttpWsPort      int                `json:"WsPort"`
	HttpJsonPort    int                `json:"JsonPort"`
	OauthServerUrl  string             `json:"OauthUrl"`
	NoticeServerUrl string             `json:"NoticeUrl"`
	NodePort        int                `json:"NodePort"`
	NodeType        string             `json:"NodeType"`
	WebSocketPort   int                `json:"WebSocketPort"`
	PrintLevel      string             `json:"LogLevel"`
	IsTLS           bool               `json:"Tls"`
	AppKey        string             	`json:"AppKey"`
	SecretKey         string             `json:"SecretKey"`
	CertPath		string				`json:"CertPath"`
	KeyPath			string				`json:"KeyPath"`
	AllowIp			string 				`json:"AllowIp"`
	ChainPath		string				`json:"ChainPath"`
	CAPath          string             `json:"CAPath"`
	GenBlockTime    uint               `json:"GenBlockTime"`
	MultiCoreNum    uint               `json:"MultiCoreNum"`
	EncryptAlg      string             `json:"EncryptAlg"`
	MaxLogSize      int64              `json:"MaxLogSize"`
	MaxTxInBlock    int                `json:"MaxTransactionInBlock"`
	MaxHdrSyncReqs  int                `json:"MaxConcurrentSyncHeaderReqs"`
	TransactionFee  map[string]float64 `json:"TransactionFee"`
}

type ConfigFile struct {
	ConfigFile Configuration `json:"Configuration"`
}

var Parameters *Configuration
/**
 read config fiel init configurateion
 */
func init(){
	file,err := ioutil.ReadFile(DefaultConfigFilename)
	if err != nil {
		log.Fatalf("read config file error:%v\n",err)
		os.Exit(1)
	}
	file = bytes.TrimPrefix(file,[]byte("\xef\xbb\xbf"))
	config := ConfigFile{}
	err = json.Unmarshal(file,&config)
	if err != nil {
		log.Fatalf("unmarshal config file error %v",err)
		os.Exit(1)
	}
	Parameters = &(config.ConfigFile)
}
