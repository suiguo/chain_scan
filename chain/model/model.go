package model

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/suiguo/chain_scan/chain/utils"
)

type DB interface {
	Save(key string, val int)
	Load(key string) int
}
type defaultDb struct {
	DB
}

func (db *defaultDb) Save(key string, val int) {
	ioutil.WriteFile("data", []byte(fmt.Sprintf("%s|%d", key, val)), fs.ModePerm)
}
func (db *defaultDb) Load(key string) int {
	d, err := ioutil.ReadFile("data")
	if err != nil {
		return 0
	}
	out := strings.Split(string(d), "|")
	if len(out) == 2 && out[0] == key {
		data, _ := strconv.ParseInt(out[1], 10, 32)
		return int(data)
	}
	return 0
}

func GetDb() DB {
	return &defaultDb{}
}

type ChainType string

const (
	ETH  ChainType = "eth"
	BSC  ChainType = "bsc"
	TRON ChainType = "tron"
)

type Cfg struct {
	Type         ChainType
	ApiUrl       string
	ApiKey       []string
	Start        int
	End          int
	ContractAddr string
	FromAddr     string
	ToAddr       string
	Interval     time.Duration
	Log          utils.Logger
}

type Options func(c *Cfg)

type Scan interface {
	Init(o ...Options) bool
	Run()
	Stop()
}
