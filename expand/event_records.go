package expand

import (
	"strings"

	"github.com/stafiprotocol/go-substrate-rpc-client/types"
	"github.com/yanyushr/stafi-substrate-go/v2/expand/acala"
	"github.com/yanyushr/stafi-substrate-go/v2/expand/base"
	"github.com/yanyushr/stafi-substrate-go/v2/expand/chainX"
	"github.com/yanyushr/stafi-substrate-go/v2/expand/crust"
	"github.com/yanyushr/stafi-substrate-go/v2/expand/darwinia"
	"github.com/yanyushr/stafi-substrate-go/v2/expand/ori"
	"github.com/yanyushr/stafi-substrate-go/v2/expand/polkadot"
	"github.com/yanyushr/stafi-substrate-go/v2/expand/stafi"
)

type IEventRecords interface {
	GetBalancesTransfer() []types.EventBalancesTransfer
	GetSystemExtrinsicSuccess() []types.EventSystemExtrinsicSuccess
	GetSystemExtrinsicFailed() []types.EventSystemExtrinsicFailed
}

/*
扩展： 解析event
*/
func DecodeEventRecords(meta *types.Metadata, rawData string, chainName string) (IEventRecords, error) {
	e := types.EventRecordsRaw(types.MustHexDecodeString(rawData))
	var ier IEventRecords
	switch strings.ToLower(chainName) {
	case "chainx":
		var events chainX.ChainXEventRecords
		err := e.DecodeEventRecords(meta, &events)
		if err != nil {
			return nil, err
		}
		ier = &events
	case "crab", "darwinia":
		var events darwinia.DarwiniaEventRecords
		err := e.DecodeEventRecords(meta, &events)
		if err != nil {
			return nil, err
		}
		ier = &events
	case "crust":
		var events crust.CRustEventRecords
		err := e.DecodeEventRecords(meta, &events)
		if err != nil {
			return nil, err
		}
		ier = &events
	case "mandala": // acala mandala 网络
		var events acala.AcalaEventRecords
		err := e.DecodeEventRecords(meta, &events)
		if err != nil {
			return nil, err
		}
		ier = &events
	case "node": //stafi
		var events stafi.StafiEventRecords
		err := e.DecodeEventRecords(meta, &events)
		if err != nil {
			return nil, err
		}
		ier = &events
	case "orion":
		var events ori.OrionEventRecords
		err := e.DecodeEventRecords(meta, &events)
		if err != nil {
			return nil, err
		}
		ier = &events
	case "polkadot":
		var events polkadot.PolkadotEventRecords
		err := e.DecodeEventRecords(meta, &events)
		if err != nil {
			return nil, err
		}
		ier = &events
	default:
		var events base.BaseEventRecords
		err := e.DecodeEventRecords(meta, &events)
		if err != nil {
			return nil, err
		}
		ier = &events
	}
	return ier, nil
}
