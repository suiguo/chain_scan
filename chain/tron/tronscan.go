package tron

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/mr-tron/base58"
	"github.com/suiguo/chain_scan/chain/model"
	"github.com/suiguo/chain_scan/chain/utils"
	"github.com/suiguo/chain_scan/nettool"
	// "github.com/suiguo/esutils/client"
)

const logTag string = "Tron"

var tranHash string = utils.GenMethod("Transfer(address,address,uint256)")

//myaccount = 'TWgdDVZ1GCTmBeC1NWbJ6vWTYxtYrWrsxz'
var usdt_contract = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"

const (
	getNowBlock        = "now_block"
	getBlockByNums     = "get_block_by_nums"
	getTranByBlockNums = "get_tran_by_block_num"
)

var methodMap = map[string][]string{
	getNowBlock:        {"GET", "/walletsolidity/getnowblock"},
	getTranByBlockNums: {"POST", "/wallet/gettransactioninfobyblocknum", `{"num":%d,"visible":true}`},
	getBlockByNums:     {"POST", "/walletsolidity/gettransactioninfobyblocknum", `{"num":%d,"visible":true}`},
}

//json rpc url
func WithUrl(url string) model.Options {
	return func(c *model.Cfg) {
		c.ApiUrl = url
	}
}

//tron api key
func WithApiKey(api_key string) model.Options {
	return func(c *model.Cfg) {
		c.ApiKey = api_key
	}
}

//scan begin block
func WithStartBlock(block int) model.Options {
	return func(c *model.Cfg) {
		c.Start = block
	}
}

//scan interval
func WithInterval(interval time.Duration) model.Options {
	return func(c *model.Cfg) {
		c.Interval = interval
	}
}

func WithFromAddr(addr string) model.Options {
	return func(c *model.Cfg) {
		c.FromAddr = addr
	}
}
func WithToAddr(addr string) model.Options {
	return func(c *model.Cfg) {
		c.ToAddr = addr
	}
}

func WithContractAddr(addr string) model.Options {
	return func(c *model.Cfg) {
		c.ContractAddr = addr
	}
}

func WithLog(log utils.Logger) model.Options {
	return func(c *model.Cfg) {
		c.Log = log
	}
}

type TronScan struct {
	cfg  *model.Cfg
	from string
	to   string
	model.Scan
	log    utils.Logger
	signal chan bool
}

func (t *TronScan) checkCfg() bool {
	if t.cfg == nil {
		return false
	}
	if t.cfg.ApiUrl == "" || t.cfg.ApiKey == "" {
		return false
	}
	if t.cfg.Type == "" {
		t.cfg.Type = model.TRON
	}
	if t.cfg.Interval == 0 {
		t.cfg.Interval = time.Second * 5
	}
	if t.cfg.FromAddr != "" {
		tmp, err := base58.Decode(t.cfg.FromAddr)
		if err != nil {
			return false
		}
		t.from = hex.EncodeToString(tmp)[2:42]
	}
	if t.cfg.ToAddr != "" {
		tmp, err := base58.Decode(t.cfg.ToAddr)
		if err != nil {
			return false
		}
		fmt.Println(hex.EncodeToString(tmp))
		t.to = hex.EncodeToString(tmp)[2:42]
	}
	if t.cfg.ContractAddr == "" {
		t.cfg.ContractAddr = usdt_contract
	}
	return true
}

func (t *TronScan) Init(o ...model.Options) bool {
	cfg := &model.Cfg{}
	for _, opt := range o {
		opt(cfg)
	}
	t.cfg = cfg
	return t.checkCfg()
}

func (t *TronScan) getTranByBlock(block int) ([]*model.TranRecord, error) {
	method := methodMap[getTranByBlockNums]
	data := []byte(fmt.Sprintf(method[2], block))
	resp, err := nettool.GetMethod(t.cfg.ApiUrl, method[0], method[1], data, "TRON-PRO-API-KEY", t.cfg.ApiKey)
	if err != nil {
		return nil, err
	}
	trans := make([]*model.TranRecord, 0)
	err = json.Unmarshal(resp, &trans)
	if err != nil {
		return nil, err
	}
	// ioutil.WriteFile("result.json", resp, fs.ModePerm)
	if err != nil {
		return nil, err
	}
	return trans, nil
}

func (t *TronScan) getNowBlock() (int, error) {
	method := methodMap[getNowBlock]
	data := make([]byte, 0)
	resp, err := nettool.GetMethod(t.cfg.ApiUrl, method[0], method[1], data, "TRON-PRO-API-KEY", t.cfg.ApiKey)
	if err != nil || resp == nil {
		return 0, err
	}
	blockData := &model.TronBlock{}
	err = json.Unmarshal(resp, blockData)
	if err != nil {
		return 0, err
	}
	return blockData.BlockHeader.RawData.Number, nil
}
func (t *TronScan) rangeLog(tran *model.TranRecord) *big.Int {
	val := new(big.Int)
	if tran == nil {
		return nil
	}
	for _, tran_log := range tran.Log {
		if len(tran_log.Topics) != 3 {
			continue
		}
		if tran_log.Topics[0] != tranHash {
			continue
		}
		if tran_log.Topics[2] == "000000000000000000000000e337c883793111f218465b663f0a783a1b4b50fa" {
			fmt.Println("11")
		}
		if t.cfg.ToAddr != "" {
			if (tran_log.Topics[2])[24:] != t.to[2:] {
				continue
			}
		}
		if t.cfg.FromAddr != "" {
			if (tran_log.Topics[1])[24:] != t.from {
				continue
			}
		}
		tran_val, ok := big.NewInt(0).SetString(tran_log.Data, 16)
		if ok {
			val = val.Add(val, tran_val)
		} else {
			if t.log != nil {
				t.log.Error(logTag, "bigInt", tran_log)
			}
		}
	}
	return val
}
func (t *TronScan) work() {
	for {
		select {
		case <-time.After(t.cfg.Interval):
			current_block, err := t.getNowBlock()
			if err != nil {
				t.log.Error(logTag, "getNowBlock", err)
			}
			begin := t.cfg.Start
			if begin < 0 {
				begin = current_block
			}
			for start := begin; start < current_block+1; start++ {
				trans, err := t.getTranByBlock(start)
				if err != nil {
					if t.log != nil {
						t.log.Error(logTag, "getTranByBlock", err, "len", len(trans))
						break
					}
				}
				for _, tran := range trans {
					if tran.Receipt.Result != model.Success {
						continue
					}
					if t.cfg.ContractAddr != "" && tran.ContractAddress != t.cfg.ContractAddr {
						continue
					}
					tran_val := t.rangeLog(tran)
					if tran_val.String() != "0" {
						if t.cfg.Log != nil {
							t.cfg.Log.Info(logTag, "tnx", tran.ID, "val", tran_val)
						}
					}
				}
			}

		case <-t.signal:
			return
		}
	}
}
func (t *TronScan) Run() bool {
	// client.
	go t.work()
	return true
}
func (t *TronScan) Stop() {

}
