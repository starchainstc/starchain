package config

import (
	"bytes"
	"log"
	"io/ioutil"
	"os"
	"encoding/json"
)

const (
	DefaultConfigFilename = "./config.json"
)

var Version string

type Configuration struct {
	Magic           int64              `json:"Magic"`  //magic code for blockchain data file
	Version         int                `json:"Version"`  //version of starchain
	SeedList        []string           `json:"SeedList"`  //list of seed
	BookKeepers     []string           `json:"BookKeepers"` // The default book keepers' publickey
	HttpRestPort    int                `json:"HttpRestPort"`
	RestCertPath    string             `json:"RestCertPath"` // certification file path
	RestKeyPath     string             `json:"RestKeyPath"` //certification key file path
	HttpInfoPort    uint16             `json:"HttpInfoPort"` // http api port
	HttpInfoStart   bool               `json:"HttpInfoStart"`
	HttpWsPort      int                `json:"HttpWsPort"`
	HttpJsonPort    int                `json:"HttpJsonPort"`
	OauthServerUrl  string             `json:"OauthServerUrl"`
	NoticeServerUrl string             `json:"NoticeServerUrl"`
	NodePort        int                `json:"NodePort"`
	NodeType        string             `json:"NodeType"`
	WebSocketPort   int                `json:"WebSocketPort"`
	PrintLevel      int                `json:"PrintLevel"`
	IsTLS           bool               `json:"IsTLS"`
	CertPath        string             `json:"CertPath"`
	KeyPath         string             `json:"KeyPath"`
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
