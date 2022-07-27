package main

import (
	"time"

	"github.com/suiguo/chain_scan/chain"
	"github.com/suiguo/chain_scan/chain/model"
	"github.com/suiguo/chain_scan/chain/tron"
	"github.com/suiguo/chain_scan/chain/utils"
)

func main() {

	scan := chain.NewScan(model.TRON)
	if scan != nil {
		/*
					 "maxsize": 500,
			        "maxbackups": 5,
			        "maxage": 20160,
			        "compress": false,
			        "level": 0
		*/
		// tmp, err := base58.Decode("TWgdDVZ1GCTmBeC1NWbJ6vWTYxtYrWrsxz")
		// if err != nil {
		// 	return
		// }
		// fmt.Println(hex.EncodeToString(tmp))
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
			tron.WithToAddr("TWgdDVZ1GCTmBeC1NWbJ6vWTYxtYrWrsxz"),
			tron.WithStartBlock(42723579),
			tron.WithInterval(time.Second),
			tron.WithLog(log))
		scan.Run()
		scan.Stop()
	}
	for {
		time.Sleep(time.Second * 10)
	}
}
