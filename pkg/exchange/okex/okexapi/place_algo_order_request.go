package okexapi

import (
	"github.com/c9s/requestgen"
)

//go:generate -command GetRequest requestgen -method GET -responseType .APIResponse -responseDataField Data
//go:generate -command PostRequest requestgen -method POST -responseType .APIResponse -responseDataField Data

type PlaceAlgoOrderResponse struct {
	AlgoID      string `json:"algoId"`
	CClOrdID    string `json:"clOrdId"`
	AlgoClOrdID string `json:"algoClOrdId"`
	SCode       string `json:"sCode"`
	SMsg        string `json:"sMsg"`
}

//go:generate PostRequest -url "/api/v5/trade/order-algo" -type PlaceAlgoOrderRequest -responseDataType []PlaceAlgoOrderResponse
type PlaceAlgoOrderRequest struct {
	client requestgen.AuthenticatedAPIClient

	// Instrument ID, e.g. BTC-USDT (required)
	instID string `param:"instId"`

	// Trade mode (required)
	// Margin mode: cross, isolated
	// Non-Margin mode: cash
	// spot_isolated (only applicable to SPOT lead trading)
	tdMode TradeMode `param:"tdMode" validValues:"cross,isolated,cash"`

	// Margin currency (optional)
	// Only applicable to cross MARGIN orders in Single-currency margin.
	ccy *string `param:"ccy"`

	// Order side, buy or sell (required)
	side string `param:"side" validValues:"buy,sell"`

	// Position side (conditional)
	// Required in long/short mode and only be long or short
	posSide *string `param:"posSide"`

	// Order type (required)
	// conditional: One-way stop order
	// oco: One-cancels-the-other order
	// trigger: Trigger order
	// move_order_stop: Trailing order
	// twap: TWAP order
	ordType AlgoOrderType `param:"ordType"`

	// Quantity to buy or sell (conditional)
	// Either sz or closeFraction is required.
	sz *string `param:"sz"`

	// Order tag (optional)
	// A combination of case-sensitive alphanumerics, all numbers, or all letters of up to 16 characters.
	tag *string `param:"tag"`

	// Order quantity unit setting for sz (optional)
	// base_ccy: Base currency, quote_ccy: Quote currency
	// Only applicable to SPOT traded with Market buy conditional order
	// Default is quote_ccy for buy, base_ccy for sell
	tgtCcy *string `param:"tgtCcy"`

	// Client-supplied Algo ID (optional)
	// A combination of case-sensitive alphanumerics, all numbers, or all letters of up to 32 characters.
	algoClOrdId *string `param:"algoClOrdId"`

	// Fraction of position to be closed when the algo order is triggered. (conditional)
	// Currently the system supports fully closing the position only so the only accepted value is 1.
	// For the same position, only one TPSL pending order for fully closing the position is supported.
	// This is only applicable to FUTURES or SWAP instruments.
	// If posSide is net, reduceOnly must be true.
	// This is only applicable if ordType is conditional or oco.
	// This is only applicable if the stop loss and take profit order is executed as market order.
	// This is not supported in Portfolio Margin mode.
	// Either sz or closeFraction is required.
	closeFraction *string `param:"closeFraction"`

	// Take-profit trigger price (optional)
	// If you fill in this parameter, you should fill in the take-profit order price as well.
	tpTriggerPx *string `param:"tpTriggerPx"`

	// Take-profit trigger price type (optional)
	// last: last price
	// index: index price
	// mark: mark price
	// The default is last
	tpTriggerPxType *string `param:"tpTriggerPxType"`

	// Take-profit order price (optional)
	// If you fill in this parameter, you should fill in the take-profit trigger price as well.
	// If the price is -1, take-profit will be executed at the market price.
	tpOrdPx *string `param:"tpOrdPx"`

	// TP order kind (optional)
	// condition or limit
	// The default is condition
	tpOrdKind *string `param:"tpOrdKind"`

	// Stop-loss trigger price (optional)
	// If you fill in this parameter, you should fill in the stop-loss order price.
	slTriggerPx *string `param:"slTriggerPx"`

	// Stop-loss trigger price type (optional)
	// last: last price
	// index: index price
	// mark: mark price
	// The default is last
	slTriggerPxType *string `param:"slTriggerPxType"`

	// Stop-loss order price (optional)
	// If you fill in this parameter, you should fill in the stop-loss trigger price.
	// If the price is -1, stop-loss will be executed at the market price.
	slOrdPx *string `param:"slOrdPx"`

	// Whether the TP/SL order placed by the user is associated with the corresponding position of the instrument. (optional)
	// If it is associated, the TP/SL order will be canceled when the position is fully closed;
	// if it is not, the TP/SL order will not be affected when the position is fully closed.
	// Valid values:
	// true: Place a TP/SL order associated with the position
	// false: Place a TP/SL order that is not associated with the position
	// The default value is false. If true is passed in, users must pass reduceOnly = true as well,
	// indicating that when placing a TP/SL order associated with a position, it must be a reduceOnly order.
	// Only applicable to Single-currency margin and Multi-currency margin.
	cxlOnClosePos *bool `param:"cxlOnClosePos"`

	// Whether the order can only reduce the position size. (optional)
	// Valid options: true or false. The default value is false.
	reduceOnly *bool `param:"reduceOnly"`

	// Quick Margin type (optional)
	// Only applicable to Quick Margin Mode of isolated margin
	// manual, auto_borrow, auto_repay
	// The default value is manual (Deprecated)
	quickMgnType *string `param:"quickMgnType"`

	// Trigger Order Parameters
	//TriggerPx      *string        `json:"triggerPx"`
	//TriggerPxType  *string        `json:"triggerPxType"`
	//OrderPx        *string        `json:"orderPx"`
	//AttachAlgoOrds []AttachAlgoOr `json:"attachAlgoOrds"`

	// Trailing Stop Order Parameters
	//CallbackRatio *string `json:"callbackRatio"`

	// TWAP Order Parameters
	//SzLimit      *string `json:"szLimit"`
	//PxLimit      *string `json:"pxLimit"`
	//TimeInterval *string `json:"timeInterval"`
	//PxSpread     *string `json:"pxSpread"`
}

func (c *RestClient) NewPlaceOCOAlgoOrderRequest() *PlaceAlgoOrderRequest {
	return &PlaceAlgoOrderRequest{
		client:  c,
		ordType: AlgoOrderTypeOCO,
	}
}
