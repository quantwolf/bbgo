package okexapi

import (
	"github.com/c9s/requestgen"
)

//go:generate -command GetRequest requestgen -method GET -responseType .APIResponse -responseDataField Data
//go:generate -command PostRequest requestgen -method POST -responseType .APIResponse -responseDataField Data
type AmendAlgoOrder struct {
	AlgoID      string `json:"algoId"`
	AlgoClOrdID string `json:"algoClOrdId"`
	ReqID       string `json:"reqId"`
	SCode       string `json:"sCode"`
	SMsg        string `json:"sMsg"`
}

//go:generate PostRequest -url "/api/v5/trade/amend-algos" -type AmendAlgoOrderRequest -responseDataType []AmendAlgoOrder
type AmendAlgoOrderRequest struct {
	client requestgen.AuthenticatedAPIClient

	// Instrument ID (required)
	instID string `param:"instId"`

	// Algo ID (conditional)
	// Either algoId or algoClOrdId is required. If both are passed, algoId will be used.
	algoID *string `param:"algoId"`

	// Client-supplied Algo ID (conditional)
	// Either algoId or algoClOrdId is required. If both are passed, algoId will be used.
	algoClOrdID *string `param:"algoClOrdId"`

	// Whether the order needs to be automatically canceled when the order amendment fails (optional)
	// Valid options: false or true, the default is false.
	cxlOnFail *bool `param:"cxlOnFail"`

	// Client Request ID as assigned by the client for order amendment (conditional)
	// A combination of case-sensitive alphanumerics, all numbers, or all letters of up to 32 characters.
	// The response will include the corresponding reqId to help you identify the request if you provide it in the request.
	reqID *string `param:"reqId"`

	// New quantity after amendment (conditional)
	newSz *string `param:"newSz"`

	// Take-profit trigger price (conditional)
	// Either the take-profit trigger price or order price is 0, it means that the take-profit is deleted
	newTpTriggerPx *string `param:"newTpTriggerPx"`

	// Take-profit order price (conditional)
	// If the price is -1, take-profit will be executed at the market price.
	newTpOrdPx *string `param:"newTpOrdPx"`

	// Stop-loss trigger price (conditional)
	// Either the stop-loss trigger price or order price is 0, it means that the stop-loss is deleted
	newSlTriggerPx *string `param:"newSlTriggerPx"`

	// Stop-loss order price (conditional)
	// If the price is -1, stop-loss will be executed at the market price.
	newSlOrdPx *string `param:"newSlOrdPx"`

	// Take-profit trigger price type (conditional)
	// last: last price
	// index: index price
	// mark: mark price
	newTpTriggerPxType *string `param:"newTpTriggerPxType"`

	// Stop-loss trigger price type (conditional)
	// last: last price
	// index: index price
	// mark: mark price
	newSlTriggerPxType *string `param:"newSlTriggerPxType"`
}

func (c *RestClient) NewAmendAlgoOrderRequest() *AmendAlgoOrderRequest {
	return &AmendAlgoOrderRequest{
		client: c,
	}
}
