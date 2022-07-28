package main

import (
	"time"

	"github.com/suiguo/chain_scan/chain"
	"github.com/suiguo/chain_scan/chain/model"
	"github.com/suiguo/chain_scan/chain/tron"
	"github.com/suiguo/chain_scan/chain/utils"
)

func main() {
	// info := fmt.Sprintf("%x", 42801345)
	// fmt.Println(info)
	// return
	scan := chain.NewScan(model.TRON)
	if scan != nil {
		log, _ := utils.NewLog([]*utils.LoggerCfg{
			{
				Name:       "stdout",
				Maxsize:    500,
				Maxbackups: 5,
				Maxage:     20160,
				Compress:   false,
				Level:      0,
			}})
		scan.Init(tron.WithApiKey("bd1bdb00-34a1-41b0-9efb-84d422ce3cb2"),
			tron.WithUrl("https://api.trongrid.io"),
			tron.WithContractAddr("TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"),
			tron.WithToAddr("TSRFE6vCH5LBeNtSHa3wLKBEmNR3NHebbb"),
			tron.WithStartBlock(42801342),
			tron.WithInterval(time.Second),
			tron.WithLog(log))
		utils.Go(scan.Run)
		utils.GoWait()
		// scan.Stop()
	}

}
