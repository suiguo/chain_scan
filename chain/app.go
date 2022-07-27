package chain

import (
	"github.com/suiguo/chain_scan/chain/model"
	"github.com/suiguo/chain_scan/chain/tron"
)

func NewScan(t model.ChainType) model.Scan {
	switch t {
	case model.TRON:
		return &tron.TronScan{}
	}
	return nil
}
