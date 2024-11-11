package okexapi

import "github.com/c9s/requestgen"

type ClosePositionResponse struct {
	InstrumentID string `param:"instId"`
	PosSide      string `param:"posSide"`
	ClOrdId      string `param:"clOrdId"`
	Tag          string `param:"tag"`
}

//go:generate -command GetRequest requestgen -method GET -responseType .APIResponse -responseDataField Data
//go:generate -command PostRequest requestgen -method POST -responseType .APIResponse -responseDataField Data

//go:generate PostRequest -url "/api/v5/trade/close-position" -type ClosePositionRequest -responseDataType []ClosePositionResponse
type ClosePositionRequest struct {
	client requestgen.AuthenticatedAPIClient

	instrumentID string  `param:"instId"`
	posSide      *string `param:"posSide"`
	marginMode   string  `param:"mgnMode"`
	ccy          *string `param:"ccy"`
	autoCxl      bool    `param:"autoCxl"`
	clOrdId      *string `param:"clOrdId"`
	tag          *string `param:"tag"`
}

func (c *RestClient) NewClosePositionRequest() *ClosePositionRequest {
	return &ClosePositionRequest{
		client: c,
	}
}
