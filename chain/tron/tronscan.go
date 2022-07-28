package tron

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/big"
	"sync"
	"time"

	"github.com/mr-tron/base58"
	"github.com/suiguo/chain_scan/chain/model"
	"github.com/suiguo/chain_scan/chain/utils"
	"github.com/suiguo/chain_scan/nettool"
	// "github.com/suiguo/esutils/client"
)

const logTag string = "Tron"

var wg sync.WaitGroup
var tranHash string = utils.GenMethod("Transfer(address,address,uint256)")

//myaccount = 'TWgdDVZ1GCTmBeC1NWbJ6vWTYxtYrWrsxz'
var usdt_contract = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"

const (
	getNowBlock        = "now_block"
	getBlockByNums     = "get_block_by_nums"
	getTranByBlockNums = "get_tran_by_block_num"
)

var methodMap = map[string][]string{
	getNowBlock:        {"GET", "/walletsolidity/getnowblock", `{"visible":true}`},
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
func WithApiKey(api_key ...string) model.Options {
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

//scan end block
func WithEndBlock(block int) model.Options {
	return func(c *model.Cfg) {
		c.End = block
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
	if t.cfg.ApiUrl == "" || len(t.cfg.ApiKey) == 0 {
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
		t.to = hex.EncodeToString(tmp)[2:42]
	}
	if t.cfg.ContractAddr == "" {
		t.cfg.ContractAddr = usdt_contract
	}
	if t.cfg.Start <= 0 {
		t.cfg.Start = -1
	}
	return true
}

func (t *TronScan) Init(o ...model.Options) bool {
	cfg := &model.Cfg{}
	for _, opt := range o {
		opt(cfg)
	}
	t.cfg = cfg
	t.log = cfg.Log
	if t.signal == nil {
		t.signal = make(chan bool)
	}
	return t.checkCfg()
}

func (t *TronScan) getTranByBlock(block int, api_key string) ([]*model.TranRecord, error) {
	method := methodMap[getTranByBlockNums]
	data := []byte(fmt.Sprintf(method[2], block))
	resp, err := nettool.GetMethod(t.cfg.ApiUrl, method[0], method[1], data, "TRON-PRO-API-KEY", api_key)
	if err != nil {
		return nil, err
	}
	trans := make([]*model.TranRecord, 0)
	err = json.Unmarshal(resp, &trans)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return trans, nil
}

func (t *TronScan) getBlockId(block_num int, api_key string) (string, error) {
	method := methodMap[getBlockByNums]
	data := []byte(fmt.Sprintf(method[2], block_num))
	resp, err := nettool.GetMethod(t.cfg.ApiUrl, method[0], method[1], data, "TRON-PRO-API-KEY", api_key)
	if err != nil {
		return "", err
	}
	blockData := &model.TronBlock{}
	err = json.Unmarshal(resp, blockData)
	if err != nil {
		return "", err
	}
	return blockData.BlockID, nil
}

func (t *TronScan) getNowBlock() (int, error) {
	method := methodMap[getNowBlock]
	// data :=
	resp, err := nettool.GetMethod(t.cfg.ApiUrl, method[0], method[1], []byte(method[2]), "TRON-PRO-API-KEY", t.cfg.ApiKey[0])
	if err != nil || resp == nil {
		return 0, err
	}
	ioutil.WriteFile("result.json", resp, fs.ModePerm)
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
		if t.cfg.ToAddr != "" {
			if (tran_log.Topics[2])[24:] != t.to {
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
func (t *TronScan) scanBlock(start_block int, end_block int, api_key string) error {
	begin := start_block
	end := end_block
	if begin > end {
		begin, end = end, begin
	}
	if t.log != nil {
		t.log.Info(logTag, "from", begin, "to", end)
	}
	for ; begin < end+1; begin++ {
		if t.log != nil {
			t.log.Info(logTag, "work idx", begin)
		}
		t.getNowBlock()
		trans, err := t.getTranByBlock(begin, api_key)
		if err != nil {
			if t.log != nil {
				t.log.Error(logTag, "getTranByBlock", err, "len", len(trans), "block", begin)
			}
			return err
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
		// model.GetDb().Save("done", begin)
	}
	if t.log != nil {
		t.log.Info(logTag, "work done", begin)
	}
	return nil
}
func (t *TronScan) Go(f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}
func (t *TronScan) Wait() {
	wg.Wait()
}
func (t *TronScan) work() {
	begin := t.cfg.Start
	for {
		select {
		case <-time.After(t.cfg.Interval):
			current_block, err := t.getNowBlock()
			if err != nil {
				t.log.Error(logTag, "getNowBlock", err)
				continue
			}
			end := t.cfg.End
			if end <= 0 || end > current_block {
				end = current_block
			}
			tmp := model.GetDb().Load("done")
			if tmp > begin {
				begin = tmp + 1
			}
			if begin < 0 {
				if t.log != nil {
					t.log.Info(logTag, "db done idx", tmp)
				}
				if tmp == 0 {
					begin = current_block
				} else {
					begin = tmp
				}
			}
			if begin > end {
				begin, end = end, begin
			}
			key_number := len(t.cfg.ApiKey)
			blocks := end - begin
			step := 0
			if blocks < key_number {
				key_number = blocks
			} else {
				step = blocks/key_number + 1
			}
			for i := 0; i < key_number; i++ {
				b := begin + i*step + 1
				e := begin + (i+1)*step
				if e > end {
					e = end
				}
				if b > end {
					b = end
				}
				// fmt.Println(b, e)
				func(param1 int, parm2 int, idx int) {
					t.Go(func() {
						err = t.scanBlock(b, e, t.cfg.ApiKey[idx])
						if err != nil {
							if t.log != nil {
								t.log.Error(logTag, "scan block", err)
							}
						}
					})
				}(b, e, i)
			}
			t.Wait()
			model.GetDb().Save("done", end)
		case <-t.signal:
			return
		}
	}
}
func (t *TronScan) Run() {
	t.work()
}
func (t *TronScan) Stop() {
	go func() {
		t.signal <- true
		fmt.Println("stop")
	}()
}
