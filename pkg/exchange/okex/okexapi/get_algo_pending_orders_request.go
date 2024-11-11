package okexapi

import (
	"github.com/c9s/requestgen"
)

//go:generate -command GetRequest requestgen -method GET -responseType .APIResponse -responseDataField Data
//go:generate -command PostRequest requestgen -method POST -responseType .APIResponse -responseDataField Data

type AlgoOrder struct {
	InstType             string          `json:"instType"`
	InstID               string          `json:"instId"`
	Ccy                  string          `json:"ccy"`
	OrdID                string          `json:"ordId"`
	OrdIDList            []string        `json:"ordIdList"`
	AlgoID               string          `json:"algoId"`
	ClOrdID              string          `json:"clOrdId"`
	Sz                   string          `json:"sz"`
	CloseFraction        string          `json:"closeFraction"`
	OrdType              string          `json:"ordType"`
	Side                 string          `json:"side"`
	PosSide              string          `json:"posSide"`
	TdMode               string          `json:"tdMode"`
	TgtCcy               string          `json:"tgtCcy"`
	State                string          `json:"state"`
	Lever                string          `json:"lever"`
	TpTriggerPx          string          `json:"tpTriggerPx"`
	TpTriggerPxType      string          `json:"tpTriggerPxType"`
	TpOrdPx              string          `json:"tpOrdPx"`
	SlTriggerPx          string          `json:"slTriggerPx"`
	SlTriggerPxType      string          `json:"slTriggerPxType"`
	SlOrdPx              string          `json:"slOrdPx"`
	TriggerPx            string          `json:"triggerPx"`
	TriggerPxType        string          `json:"triggerPxType"`
	OrdPx                string          `json:"ordPx"`
	ActualSz             string          `json:"actualSz"`
	Tag                  string          `json:"tag"`
	ActualPx             string          `json:"actualPx"`
	ActualSide           string          `json:"actualSide"`
	TriggerTime          string          `json:"triggerTime"`
	PxVar                string          `json:"pxVar"`
	PxSpread             string          `json:"pxSpread"`
	SzLimit              string          `json:"szLimit"`
	PxLimit              string          `json:"pxLimit"`
	TimeInterval         string          `json:"timeInterval"`
	CallbackRatio        string          `json:"callbackRatio"`
	CallbackSpread       string          `json:"callbackSpread"`
	ActivePx             string          `json:"activePx"`
	MoveTriggerPx        string          `json:"moveTriggerPx"`
	ReduceOnly           string          `json:"reduceOnly"`
	QuickMgnType         string          `json:"quickMgnType"`
	Last                 string          `json:"last"`
	FailCode             string          `json:"failCode"`
	AlgoClOrdID          string          `json:"algoClOrdId"`
	AmendPxOnTriggerType string          `json:"amendPxOnTriggerType"`
	AttachAlgoOrds       []AttachAlgoOrd `json:"attachAlgoOrds"`
	LinkedOrd            LinkedOrd       `json:"linkedOrd"`
	CTime                string          `json:"cTime"`
	UTime                string          `json:"uTime"`
}

type AttachAlgoOrd struct {
	AttachAlgoClOrdID string `json:"attachAlgoClOrdId"`
	TpTriggerPx       string `json:"tpTriggerPx"`
	TpTriggerPxType   string `json:"tpTriggerPxType"`
	TpOrdPx           string `json:"tpOrdPx"`
	SlTriggerPx       string `json:"slTriggerPx"`
	SlTriggerPxType   string `json:"slTriggerPxType"`
	SlOrdPx           string `json:"slOrdPx"`
}

type LinkedOrd struct {
	OrdID string `json:"ordId"`
}

type AlgoOrderType string

const (
	AlgoOrderTypeConditional   AlgoOrderType = "conditional"
	AlgoOrderTypeOCO           AlgoOrderType = "oco"
	AlgoOrderTypeTrigger       AlgoOrderType = "trigger"
	AlgoOrderTypeMoveOrderStop AlgoOrderType = "move_order_stop"
	AlgoOrderTypeIceberg       AlgoOrderType = "iceberg"
	AlgoOrderTypeTWAP          AlgoOrderType = "twap"
)

//go:generate GetRequest -url "/api/v5/trade/orders-algo-pending" -type GetAlgoPendingOrdersRequest -responseDataType []AlgoOrder
type GetAlgoPendingOrdersRequest struct {
	client requestgen.AuthenticatedAPIClient

	// Order type (required)
	// conditional: One-way stop order
	// oco: One-cancels-the-other order
	// trigger: Trigger order
	// move_order_stop: Trailing order
	// iceberg: Iceberg order
	// twap: TWAP order
	ordType AlgoOrderType `param:"ordType,query"`

	// Instrument type (optional)
	// SPOT, SWAP, FUTURES, MARGIN
	instrumentType *InstrumentType `param:"instType,query"`

	// Instrument ID, e.g. BTC-USDT (optional)
	instrumentID *string `param:"instId,query"`

	// Algo ID (optional)
	algoID *string `param:"algoId,query"`

	// Client-supplied Algo ID (optional)
	// A combination of case-sensitive alphanumerics, all numbers, or all letters of up to 32 characters.
	algoClOrdID *string `param:"algoClOrdId,query"`

	// Pagination of data to return records earlier than the requested algoId (optional)
	after *string `param:"after,query,timestamp"`

	// Pagination of data to return records newer than the requested algoId (optional)
	before *string `param:"before,query,timestamp"`

	// Number of results per request. The maximum is 100. The default is 100 (optional)
	limit *string `param:"limit,query"`
}

func (c *RestClient) NewGetAlgoOrdersRequest() *GetAlgoPendingOrdersRequest {
	return &GetAlgoPendingOrdersRequest{
		client:  c,
		ordType: AlgoOrderTypeConditional,
	}
}

func (c *RestClient) NewGetOCOAlgoOrdersRequest() *GetAlgoPendingOrdersRequest {
	return &GetAlgoPendingOrdersRequest{
		client:  c,
		ordType: AlgoOrderTypeOCO,
	}
}
