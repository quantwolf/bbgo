package okex

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"
	"golang.org/x/time/rate"

	"github.com/c9s/bbgo/pkg/exchange/okex/okexapi"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

var (
	// clientOrderIdRegex combine of case-sensitive alphanumerics, all numbers, or all letters of up to 32 characters.
	clientOrderIdRegex = regexp.MustCompile("^[a-zA-Z0-9]{0,32}$")

	// Rate Limit: 20 requests per 2 seconds, Rate limit rule: IP + instrumentType.
	// Currently, calls are not made very frequently, so only IP is considered.
	queryMarketLimiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 1)
	// Rate Limit: 20 requests per 2 seconds, Rate limit rule: IP
	queryTickerLimiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 1)
	// Rate Limit: 20 requests per 2 seconds, Rate limit rule: IP
	queryTickersLimiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 1)
	// Rate Limit: 10 requests per 2 seconds, Rate limit rule: UserID
	queryAccountLimiter = rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
	// Rate Limit: 60 requests per 2 seconds, Rate limit rule (except Options): UserID + Instrument ID.
	// TODO: support UserID + Instrument ID
	placeOrderLimiter = rate.NewLimiter(rate.Every(33*time.Millisecond), 1)
	// Rate Limit: 60 requests per 2 seconds, Rate limit rule (except Options): UserID + Instrument ID
	// TODO: support UserID + Instrument ID
	batchCancelOrderLimiter = rate.NewLimiter(rate.Every(33*time.Millisecond), 1)
	// Rate Limit: 60 requests per 2 seconds, Rate limit rule: UserID
	queryOpenOrderLimiter = rate.NewLimiter(rate.Every(33*time.Millisecond), 1)
	// Rate Limit: 60 requests per 2 seconds, Rate limit rule: UserID
	queryAlgoOpenOrderLimiter = rate.NewLimiter(rate.Every(33*time.Millisecond), 1)
	// Rate Limit: 20 requests per 2 seconds, Rate limit rule: UserID
	queryClosedOrderRateLimiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 1)
	// Rate Limit: 10 requests per 2 seconds, Rate limit rule: UserID
	queryTradeLimiter = rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
	// Rate Limit: 40 requests per 2 seconds, Rate limit rule: IP
	queryKLineLimiter = rate.NewLimiter(rate.Every(50*time.Millisecond), 1)
)

const (
	ID = "okex"

	// PlatformToken is the platform currency of OKEx, pre-allocate static string here
	PlatformToken = "OKB"

	defaultQueryLimit = 100

	maxHistoricalDataQueryPeriod = 90 * 24 * time.Hour
	threeDaysHistoricalPeriod    = 3 * 24 * time.Hour
)

var log = logrus.WithFields(logrus.Fields{
	"exchange": ID,
})

var ErrSymbolRequired = errors.New("symbol is a required parameter")

type Exchange struct {
	types.MarginSettings

	key, secret, passphrase, brokerId string

	client      *okexapi.RestClient
	timeNowFunc func() time.Time
}

type Option func(exchange *Exchange)

func WithBrokerId(id string) Option {
	return func(exchange *Exchange) {
		exchange.brokerId = id
	}
}

func New(key, secret, passphrase string, opts ...Option) *Exchange {
	client := okexapi.NewClient()

	if len(key) > 0 && len(secret) > 0 {
		client.Auth(key, secret, passphrase)
	}

	ex := &Exchange{
		key:         key,
		secret:      secret,
		passphrase:  passphrase,
		client:      client,
		timeNowFunc: time.Now,
	}

	for _, o := range opts {
		o(ex)
	}

	return ex
}

func (e *Exchange) Name() types.ExchangeName {
	return types.ExchangeOKEx
}

func (e *Exchange) QueryMarkets(ctx context.Context) (types.MarketMap, error) {
	if err := queryMarketLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("markets rate limiter wait error: %w", err)
	}

	instruments, err := e.client.NewGetInstrumentsInfoRequest().Do(ctx)
	if err != nil {
		return nil, err
	}

	markets := types.MarketMap{}
	for _, instrument := range instruments {
		symbol := toGlobalSymbol(instrument.InstrumentID)
		market := types.Market{
			Exchange:    types.ExchangeOKEx,
			Symbol:      symbol,
			LocalSymbol: instrument.InstrumentID,

			QuoteCurrency: instrument.QuoteCurrency,
			BaseCurrency:  instrument.BaseCurrency,

			// convert tick size OKEx to precision
			PricePrecision:  instrument.TickSize.NumFractionalDigits(),
			VolumePrecision: instrument.LotSize.NumFractionalDigits(),

			// TickSize: OKEx's price tick, for BTC-USDT it's "0.1"
			TickSize: instrument.TickSize,

			// Quantity step size, for BTC-USDT, it's "0.00000001"
			StepSize: instrument.LotSize,

			// for BTC-USDT, it's "0.00001"
			MinQuantity: instrument.MinSize,

			// OKEx does not offer minimal notional, use 1 USD here.
			MinNotional: fixedpoint.One,
			MinAmount:   fixedpoint.One,
		}
		markets[symbol] = market
	}

	return markets, nil
}

func (e *Exchange) QueryTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	if err := queryTickerLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("ticker rate limiter wait error: %w", err)
	}

	symbol = toLocalSymbol(symbol)
	marketTicker, err := e.client.NewGetTickerRequest().InstId(symbol).Do(ctx)
	if err != nil {
		return nil, err
	}

	if len(marketTicker) != 1 {
		return nil, fmt.Errorf("unexpected length of %s market ticker, got: %v", symbol, marketTicker)
	}

	return toGlobalTicker(marketTicker[0]), nil
}

func (e *Exchange) QueryTickers(ctx context.Context, symbols ...string) (map[string]types.Ticker, error) {
	if err := queryTickersLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("tickers rate limiter wait error: %w", err)
	}

	marketTickers, err := e.client.NewGetTickersRequest().Do(ctx)
	if err != nil {
		return nil, err
	}

	tickers := make(map[string]types.Ticker)
	for _, marketTicker := range marketTickers {
		symbol := toGlobalSymbol(marketTicker.InstrumentID)
		ticker := toGlobalTicker(marketTicker)
		tickers[symbol] = *ticker
	}

	if len(symbols) == 0 {
		return tickers, nil
	}

	selectedTickers := make(map[string]types.Ticker, len(symbols))
	for _, symbol := range symbols {
		if ticker, ok := tickers[symbol]; ok {
			selectedTickers[symbol] = ticker
		}
	}

	return selectedTickers, nil
}

func (e *Exchange) PlatformFeeCurrency() string {
	return PlatformToken
}

func (e *Exchange) QueryAccount(ctx context.Context) (*types.Account, error) {
	bals, err := e.QueryAccountBalances(ctx)
	if err != nil {
		return nil, err
	}

	account := types.NewAccount()
	account.UpdateBalances(bals)
	return account, nil
}

func (e *Exchange) QueryAccountBalances(ctx context.Context) (types.BalanceMap, error) {
	if err := queryAccountLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("account rate limiter wait error: %w", err)
	}

	accountBalances, err := e.client.NewGetAccountInfoRequest().Do(ctx)
	if err != nil {
		return nil, err
	}

	if len(accountBalances) != 1 {
		return nil, fmt.Errorf("unexpected length of balances: %v", accountBalances)
	}

	return toGlobalBalance(&accountBalances[0]), nil
}

func (e *Exchange) SubmitOrder(ctx context.Context, order types.SubmitOrder) (*types.Order, error) {
	if e.IsMargin {
		return e.submitMarginOrder(ctx, order)
	} else {
		return e.submitSpotOrder(ctx, order)
	}
}

func (e *Exchange) submitSpotOrder(ctx context.Context, order types.SubmitOrder) (*types.Order, error) {
	orderReq := e.client.NewPlaceOrderRequest()

	orderReq.TradeMode("cash")
	orderReq.InstrumentID(toLocalSymbol(order.Symbol))
	orderReq.Side(toLocalSideType(order.Side))
	orderReq.Size(order.Market.FormatQuantity(order.Quantity))

	// set price field for limit orders
	switch order.Type {
	case types.OrderTypeStopLimit, types.OrderTypeLimit, types.OrderTypeLimitMaker:
		orderReq.Price(order.Market.FormatPrice(order.Price))
	case types.OrderTypeMarket:
		// Because our order.Quantity unit is base coin, so we indicate the target currency to Base.
		if order.Side == types.SideTypeBuy {
			orderReq.Size(order.Market.FormatQuantity(order.Quantity))
			orderReq.TargetCurrency(okexapi.TargetCurrencyBase)
		} else {
			orderReq.Size(order.Market.FormatQuantity(order.Quantity))
			orderReq.TargetCurrency(okexapi.TargetCurrencyQuote)
		}
	}

	orderType, err := toLocalOrderType(order.Type)
	if err != nil {
		return nil, err
	}

	switch order.TimeInForce {
	case types.TimeInForceFOK:
		orderReq.OrderType(okexapi.OrderTypeFOK)
	case types.TimeInForceIOC:
		orderReq.OrderType(okexapi.OrderTypeIOC)
	default:
		orderReq.OrderType(orderType)
	}

	if err := placeOrderLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("place order rate limiter wait error: %w", err)
	}

	if len(order.ClientOrderID) > 0 {
		if ok := clientOrderIdRegex.MatchString(order.ClientOrderID); !ok {
			return nil, fmt.Errorf("client order id should be case-sensitive alphanumerics, all numbers, or all letters of up to 32 characters: %s", order.ClientOrderID)
		}
		orderReq.ClientOrderID(order.ClientOrderID)
	}

	if len(e.brokerId) != 0 {
		orderReq.Tag(e.brokerId)
	}

	orders, err := orderReq.Do(ctx)
	if err != nil {
		return nil, err
	}

	if len(orders) != 1 {
		return nil, fmt.Errorf("unexpected length of order response: %v", orders)
	}

	orderRes, err := e.QueryOrder(ctx, types.OrderQuery{
		Symbol:        order.Symbol,
		OrderID:       orders[0].OrderID,
		ClientOrderID: orders[0].ClientOrderID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query order by id: %s, clientOrderId: %s, err: %w", orders[0].OrderID, orders[0].ClientOrderID, err)
	}

	return orderRes, nil

	// TODO: move this to batch place orders interface
	/*
		batchReq := e.client.TradeService.NewBatchPlaceOrderRequest()
		batchReq.Add(reqs...)
		orderHeads, err := batchReq.Do(ctx)
		if err != nil {
			return nil, err
		}

		for idx, orderHead := range orderHeads {
			orderID, err := strconv.ParseInt(orderHead.OrderID, 10, 64)
			if err != nil {
				return createdOrder, err
			}

			submitOrder := order[idx]
			createdOrder = append(createdOrder, types.Order{
				SubmitOrder:      submitOrder,
				Exchange:         types.ExchangeOKEx,
				OrderID:          uint64(orderID),
				Status:           types.OrderStatusNew,
				ExecutedQuantity: fixedpoint.Zero,
				IsWorking:        true,
				CreationTime:     types.Time(time.Now()),
				UpdateTime:       types.Time(time.Now()),
				IsMargin:         false,
				IsIsolated:       false,
			})
		}
	*/
}

func (e *Exchange) autoCancelAllAlgoOrders(ctx context.Context, symbol string) error {
	orders, err1 := e.QueryAlgoOpenOrders(ctx, symbol)
	ocoOrders, err2 := e.QueryOCOAlgoOpenOrders(ctx, symbol)

	log.WithField("orders", orders).
		WithField("ocoOrders", ocoOrders).
		WithField("error1", err1).
		WithField("error2", err2).
		Info("autoCancelAlgoOrders_QueryOpenOrders_result")

	if len(orders)+len(ocoOrders) > 0 {
		allOrders := make([]types.Order, 0)

		if len(orders) > 0 {
			allOrders = append(allOrders, orders...)
		}

		if len(ocoOrders) > 0 {
			allOrders = append(allOrders, ocoOrders...)
		}

		err := e.CancelAlgoOrders(ctx, allOrders...)
		if err != nil {
			log.WithField("Symbol", symbol).
				WithError(err).
				Error("autoCancelAlgoOrders_CancelOrders_fail")
		}

		log.WithField("Symbol", symbol).
			Info("autoCancelAlgoOrders_CancelOrders_ok")
	}

	return nil
}

func (e *Exchange) submitClosePositionOrder(ctx context.Context, order types.SubmitOrder) (*types.Order, error) {
	orderReq := e.client.NewClosePositionOrderRequest()

	orderReq.InstrumentID(toLocalSymbol(order.Symbol))

	orderReq.MarginMode("cross")
	orderReq.Ccy(order.Market.QuoteCurrency)

	if e.IsIsolatedMargin {
		orderReq.MarginMode("isolated")
	}

	orderReq.Tag(order.Tag)
	orderReq.AutoCxl(true)

	params, _ := orderReq.GetParameters()
	log.WithField("params", params).
		Info("order_req_start")

	orderHead, err := orderReq.Do(ctx)
	if err != nil {
		return nil, err
	}

	log.WithField("orderHead", orderHead).
		Info("order_req_end")

	return &types.Order{
		SubmitOrder:      order,
		Exchange:         types.ExchangeOKEx,
		OrderID:          uint64(0),
		Status:           types.OrderStatusNew,
		ExecutedQuantity: fixedpoint.Zero,
		IsWorking:        true,
		CreationTime:     types.Time(time.Now()),
		UpdateTime:       types.Time(time.Now()),
		IsMargin:         false,
		IsIsolated:       false,
	}, nil
}

func (e *Exchange) submitMarginOrder(ctx context.Context, order types.SubmitOrder) (*types.Order, error) {
	if order.ClosePosition {
		e.autoCancelAllAlgoOrders(ctx, order.Symbol)
		return e.submitClosePositionOrder(ctx, order)
	}

	orderReq := e.client.NewPlaceOrderRequest()

	orderType, err := toLocalOrderType(order.Type)
	if err != nil {
		return nil, err
	}

	// Margin mode cross isolated
	if e.IsIsolatedMargin {
		orderReq.TradeMode(okexapi.TradeModeIsolated)
	} else {
		orderReq.TradeMode(okexapi.TradeModeCross)
	}

	orderReq.MarginCurrency(order.Market.QuoteCurrency)
	orderReq.InstrumentID(toLocalSymbol(order.Symbol))
	orderReq.Side(toLocalSideType(order.Side))

	if order.Market.Symbol != "" {
		orderReq.Size(order.Market.FormatQuantity(order.Quantity))
	} else {
		// TODO report error
		orderReq.Size(order.Quantity.FormatString(8))
	}

	// set price field for limit orders
	switch order.Type {
	case types.OrderTypeStopLimit, types.OrderTypeLimit:
		if order.Market.Symbol != "" {
			orderReq.Price(order.Market.FormatPrice(order.Price))
		} else {
			// TODO report error
			orderReq.Price(order.Price.FormatString(8))
		}
	}

	switch order.TimeInForce {
	case "FOK":
		orderReq.OrderType(okexapi.OrderTypeFOK)
	case "IOC":
		orderReq.OrderType(okexapi.OrderTypeIOC)
	default:
		orderReq.OrderType(orderType)
	}

	// Set stop loss trigger price
	if order.StopPrice.Compare(fixedpoint.Zero) > 0 {
		orderReq.StopLossTriggerPxType("last")
		orderReq.StopLossTriggerPx(order.Market.FormatPrice(order.StopPrice))
		orderReq.StopLossOrdPx("-1")
	}

	// Set take profit trigger price
	if order.TakePrice.Compare(fixedpoint.Zero) > 0 {
		orderReq.TakeProfitTriggerPxType("last")
		orderReq.TakeProfitTriggerPx(order.Market.FormatPrice(order.TakePrice))
		orderReq.TakeProfitOrdPx("-1")
	}

	params, _ := orderReq.GetParameters()
	log.WithField("params", params).
		Info("submitMarginOrder_start")

	orders, err := orderReq.Do(ctx)
	if err != nil {
		return nil, err
	}

	log.WithField("orders", orders).
		Info("submitMarginOrder_result")

	if len(orders) != 1 {
		return nil, fmt.Errorf("unexpected length of order response: %v", orders)
	}

	orderRes, err := e.QueryOrder(ctx, types.OrderQuery{
		Symbol:        order.Symbol,
		OrderID:       orders[0].OrderID,
		ClientOrderID: orders[0].ClientOrderID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query order by id: %s, clientOrderId: %s, err: %w", orders[0].OrderID, orders[0].ClientOrderID, err)
	}

	return orderRes, nil
}

// QueryOpenOrders retrieves the pending orders. The data returned is ordered by createdTime, and we utilized the
// `After` parameter to acquire all orders.
func (e *Exchange) QueryOpenOrders(ctx context.Context, symbol string) (orders []types.Order, err error) {
	instrumentID := toLocalSymbol(symbol)

	nextCursor := int64(0)
	for {
		if err := queryOpenOrderLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("query open orders rate limiter wait error: %w", err)
		}

		req := e.client.NewGetOpenOrdersRequest().
			InstrumentID(instrumentID).
			After(strconv.FormatInt(nextCursor, 10))
		openOrders, err := req.Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query open orders: %w", err)
		}

		for _, o := range openOrders {
			o, err := orderDetailToGlobal(&o.OrderDetail)
			if err != nil {
				return nil, fmt.Errorf("failed to convert order, err: %v", err)
			}

			orders = append(orders, *o)
		}

		orderLen := len(openOrders)
		// a defensive programming to ensure the length of order response is expected.
		if orderLen > defaultQueryLimit {
			return nil, fmt.Errorf("unexpected open orders length %d", orderLen)
		}

		if orderLen < defaultQueryLimit {
			break
		}
		nextCursor = int64(openOrders[orderLen-1].OrderId)
	}

	return orders, err
}

func (e *Exchange) QueryRawAlgoOpenOrders(ctx context.Context, symbol string) (orders []okexapi.AlgoOrder, err error) {
	instrumentID := toLocalSymbol(symbol)

	nextCursor := ""
	for {
		if err := queryAlgoOpenOrderLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("query open orders rate limiter wait error: %w", err)
		}

		req := e.client.NewGetAlgoOrdersRequest()
		req.
			InstrumentID(instrumentID).
			After(nextCursor)

		params, _ := req.GetQueryParameters()
		log.WithField("symbol", symbol).
			WithField("params", params).
			Info("Exchange#QueryRawAlgoOpenOrders_start")

		openOrders, err := req.Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query open orders: %w", err)
		}

		orders = append(orders, openOrders...)

		log.WithField("symbol", symbol).
			WithField("openOrders", openOrders).
			WithField("orders", orders).
			Info("Exchange#QueryRawAlgoOpenOrders_result")

		orderLen := len(openOrders)
		// a defensive programming to ensure the length of order response is expected.
		if orderLen > defaultQueryLimit {
			return nil, fmt.Errorf("unexpected open orders length %d", orderLen)
		}

		if orderLen < defaultQueryLimit {
			break
		}

		log.WithField("symbol", symbol).
			WithField("openOrders", orders).
			Info("Exchange#QueryRawAlgoOpenOrders_retry")

		nextCursor = openOrders[orderLen-1].AlgoID
	}

	return orders, err
}

func (e *Exchange) QueryAlgoOpenOrders(ctx context.Context, symbol string) (orders []types.Order, err error) {
	instrumentID := toLocalSymbol(symbol)

	nextCursor := ""
	for {
		if err := queryAlgoOpenOrderLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("query open orders rate limiter wait error: %w", err)
		}

		req := e.client.NewGetAlgoOrdersRequest()
		req.
			InstrumentID(instrumentID).
			After(nextCursor)

		params, _ := req.GetQueryParameters()
		log.WithField("symbol", symbol).
			WithField("params", params).
			Info("QueryAlgoOpenOrders_start")

		openOrders, err := req.Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query open orders: %w", err)
		}

		log.WithField("symbol", symbol).
			WithField("openOrders", openOrders).
			Info("QueryAlgoOpenOrders_result")

		for _, o := range openOrders {
			amount, _ := fixedpoint.NewFromString(o.Sz)
			orders = append(orders, types.Order{
				SubmitOrder: types.SubmitOrder{
					Symbol:   symbol,
					Quantity: amount,
				},
				UUID: o.AlgoID,
			})
		}

		orderLen := len(openOrders)
		// a defensive programming to ensure the length of order response is expected.
		if orderLen > defaultQueryLimit {
			return nil, fmt.Errorf("unexpected open orders length %d", orderLen)
		}

		if orderLen < defaultQueryLimit {
			break
		}

		nextCursor = openOrders[orderLen-1].AlgoID
	}

	return orders, err
}

func (e *Exchange) QueryRawOCOAlgoOpenOrders(ctx context.Context, symbol string) (orders []okexapi.AlgoOrder, err error) {
	instrumentID := toLocalSymbol(symbol)

	nextCursor := ""
	for {
		if err := queryAlgoOpenOrderLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("query open orders rate limiter wait error: %w", err)
		}

		req := e.client.NewGetOCOAlgoOrdersRequest()
		req.
			InstrumentID(instrumentID).
			After(nextCursor)

		params, _ := req.GetQueryParameters()
		log.WithField("symbol", symbol).
			WithField("params", params).
			Info("Exchange#QueryRawOCOAlgoOpenOrders_start")

		openOrders, err := req.Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query open orders: %w", err)
		}

		orders = append(orders, openOrders...)

		log.WithField("symbol", symbol).
			WithField("openOrders", openOrders).
			WithField("orders", orders).
			Info("Exchange#QueryRawOCOAlgoOpenOrders_result")

		orderLen := len(openOrders)

		// a defensive programming to ensure the length of order response is expected.
		if orderLen > defaultQueryLimit {
			return nil, fmt.Errorf("unexpected open orders length %d", orderLen)
		}

		if orderLen < defaultQueryLimit {
			break
		}

		log.WithField("symbol", symbol).
			WithField("openOrders", orders).
			Info("Exchange#QueryRawOCOAlgoOpenOrders_retry")

		nextCursor = openOrders[orderLen-1].AlgoID
	}

	return orders, err
}

func (e *Exchange) QueryOCOAlgoOpenOrders(ctx context.Context, symbol string) (orders []types.Order, err error) {
	instrumentID := toLocalSymbol(symbol)

	nextCursor := ""
	for {
		if err := queryAlgoOpenOrderLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("query open orders rate limiter wait error: %w", err)
		}

		req := e.client.NewGetOCOAlgoOrdersRequest()
		req.
			InstrumentID(instrumentID).
			After(nextCursor)

		params, _ := req.GetQueryParameters()
		log.WithField("symbol", symbol).
			WithField("params", params).
			Info("QueryOCOAlgoOpenOrders_start")

		openOrders, err := req.Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query open orders: %w", err)
		}

		log.WithField("symbol", symbol).
			WithField("openOrders", openOrders).
			Info("QueryOCOAlgoOpenOrders_result")

		for _, o := range openOrders {
			orders = append(orders, types.Order{
				SubmitOrder: types.SubmitOrder{
					Symbol: symbol,
				},
				UUID: o.AlgoID,
			})
		}

		orderLen := len(openOrders)
		// a defensive programming to ensure the length of order response is expected.
		if orderLen > defaultQueryLimit {
			return nil, fmt.Errorf("unexpected open orders length %d", orderLen)
		}

		if orderLen < defaultQueryLimit {
			break
		}

		nextCursor = openOrders[orderLen-1].AlgoID
	}

	return orders, err
}

func (e *Exchange) CancelOrders(ctx context.Context, orders ...types.Order) error {
	if len(orders) == 0 {
		return nil
	}

	var reqs []*okexapi.CancelOrderRequest
	for _, order := range orders {
		if len(order.Symbol) == 0 {
			return ErrSymbolRequired
		}

		req := e.client.NewCancelOrderRequest()
		req.InstrumentID(toLocalSymbol(order.Symbol))
		req.OrderID(strconv.FormatUint(order.OrderID, 10))
		if len(order.ClientOrderID) > 0 {
			if ok := clientOrderIdRegex.MatchString(order.ClientOrderID); !ok {
				return fmt.Errorf("client order id should be case-sensitive alphanumerics, all numbers, or all letters of up to 32 characters: %s", order.ClientOrderID)
			}
			req.ClientOrderID(order.ClientOrderID)
		}
		reqs = append(reqs, req)
	}

	if err := batchCancelOrderLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("batch cancel order rate limiter wait error: %w", err)
	}
	batchReq := e.client.NewBatchCancelOrderRequest()
	batchReq.Add(reqs...)
	_, err := batchReq.Do(ctx)
	return err
}

func (e *Exchange) CancelAlgoOrders(ctx context.Context, orders ...types.Order) error {
	if len(orders) == 0 {
		return nil
	}

	var reqs []*okexapi.CancelAlgoOrder
	for _, order := range orders {
		if len(order.Symbol) == 0 {
			return ErrSymbolRequired
		}

		reqs = append(reqs, &okexapi.CancelAlgoOrder{
			InstrumentID: toLocalSymbol(order.Symbol),
			AlgoOrderID:  order.UUID,
		})
	}

	if err := batchCancelOrderLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("batch cancel order rate limiter wait error: %w", err)
	}
	batchReq := e.client.NewCancelAlgoOrderRequest()
	batchReq.SetPayload(reqs)

	params, _ := batchReq.GetParameters()
	log.WithField("params", params).
		Info("CancelAlgoOrders_start")

	resp, err := batchReq.Do(ctx)

	log.WithField("resp", resp).
		WithField("error", err).
		Info("CancelAlgoOrders_result")

	return err
}

func (e *Exchange) NewStream() types.Stream {
	return NewStream(e.client, e)
}

func (e *Exchange) QueryKLines(
	ctx context.Context, symbol string, interval types.Interval, options types.KLineQueryOptions,
) ([]types.KLine, error) {
	if err := queryKLineLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("query k line rate limiter wait error: %w", err)
	}

	intervalParam, err := toLocalInterval(interval)
	if err != nil {
		return nil, fmt.Errorf("failed to get interval: %w", err)
	}

	req := e.client.NewGetCandlesRequest().InstrumentID(toLocalSymbol(symbol))
	req.Bar(intervalParam)

	if options.StartTime != nil {
		req.After(*options.StartTime)
	}

	if options.EndTime != nil {
		req.Before(*options.EndTime)
	}

	candles, err := req.Do(ctx)
	if err != nil {
		return nil, err
	}

	var klines []types.KLine
	for _, candle := range candles {
		klines = append(klines, kLineToGlobal(candle, interval, symbol))
	}

	return klines, nil

}

func (e *Exchange) QueryOrder(ctx context.Context, q types.OrderQuery) (*types.Order, error) {
	if len(q.Symbol) == 0 {
		return nil, ErrSymbolRequired
	}
	if len(q.OrderID) == 0 && len(q.ClientOrderID) == 0 {
		return nil, errors.New("okex.QueryOrder: OrderId or ClientOrderId is required parameter")
	}
	req := e.client.NewGetOrderDetailsRequest()
	req.InstrumentID(toLocalSymbol(q.Symbol)).
		OrderID(q.OrderID).
		ClientOrderID(q.ClientOrderID)

	var order *okexapi.OrderDetails
	order, err := req.Do(ctx)

	if err != nil {
		return nil, err
	}

	return toGlobalOrder(order)
}

// QueryOrderTrades quires order trades can query trades in last 3 months.
func (e *Exchange) QueryOrderTrades(ctx context.Context, q types.OrderQuery) (trades []types.Trade, err error) {
	if len(q.ClientOrderID) != 0 {
		log.Warn("!!!OKEX EXCHANGE API NOTICE!!! Okex does not support searching for trades using OrderClientId.")
	}

	req := e.client.NewGetTransactionHistoryRequest()
	if len(q.Symbol) != 0 {
		req.InstrumentID(toLocalSymbol(q.Symbol))
	}

	if len(q.OrderID) != 0 {
		req.OrderID(q.OrderID)
	}

	if err := queryTradeLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("order trade rate limiter wait error: %w", err)
	}
	response, err := req.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query order trades, err: %w", err)
	}

	for _, trade := range response {
		trades = append(trades, tradeToGlobal(trade))
	}

	return trades, nil
}

/*
QueryClosedOrders can query closed orders in last 3 months, there are no time interval limitations, as long as until >= since.
Please Use lastOrderID as cursor, only return orders later than that order, that order is not included.
If you want to query all orders within a large time range (e.g. total orders > 100), we recommend using batch.ClosedOrderBatchQuery.

** since and until are inclusive, you can include the lastTradeId as well. **
*/
func (e *Exchange) QueryClosedOrders(
	ctx context.Context, symbol string, since, until time.Time, lastOrderID uint64,
) (orders []types.Order, err error) {
	if symbol == "" {
		return nil, ErrSymbolRequired
	}

	newSince := since
	now := time.Now()

	if time.Since(newSince) > maxHistoricalDataQueryPeriod {
		newSince = now.Add(-maxHistoricalDataQueryPeriod)
		log.Warnf("!!!OKX EXCHANGE API NOTICE!!! The closed order API cannot query data beyond 90 days from the current date, update %s -> %s", since, newSince)
	}
	if until.Before(newSince) {
		log.Warnf("!!!OKX EXCHANGE API NOTICE!!! The 'until' comes before 'since', update until to now(%s -> %s).", until, now)
		until = now
	}
	if until.Sub(newSince) > maxHistoricalDataQueryPeriod {
		return nil, fmt.Errorf("the start time %s and end time %s cannot exceed 90 days", newSince, until)
	}

	if err := queryClosedOrderRateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("query closed order rate limiter wait error: %w", err)
	}

	res, err := e.client.NewGetOrderHistoryRequest().
		InstrumentID(toLocalSymbol(symbol)).
		StartTime(since).
		EndTime(until).
		Limit(defaultQueryLimit).
		Before(strconv.FormatUint(lastOrderID, 10)).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to call get order histories error: %w", err)
	}

	for _, order := range res {
		o, err2 := orderDetailToGlobal(&order)
		if err2 != nil {
			err = multierr.Append(err, err2)
			continue
		}

		orders = append(orders, *o)
	}
	if err != nil {
		return nil, err
	}

	return types.SortOrdersAscending(orders), nil
}

/*
QueryTrades can query trades in last 3 months, there are no time interval limitations, as long as end_time >= start_time.
okx does not provide an API to query by trade ID, so we use the bill ID to do it. The trades result is ordered by timestamp.

REMARK: If your start time is 90 days earlier, we will update it to now - 90 days.
** StartTime and EndTime are inclusive. **
** StartTime and EndTime cannot exceed 90 days.  **
** StartTime, EndTime, FromTradeId can be used together. **

If you want to query all trades within a large time range (e.g. total orders > 100), we recommend using batch.TradeBatchQuery.
We don't support the last trade id as a filter because okx supports bill ID only.
*/
func (e *Exchange) QueryTrades(ctx context.Context, symbol string, options *types.TradeQueryOptions) (trades []types.Trade, err error) {
	if symbol == "" {
		return nil, ErrSymbolRequired
	}

	limit := options.Limit
	if limit > defaultQueryLimit || limit <= 0 {
		log.Infof("limit is exceeded default limit %d or zero, got: %d, use default limit", defaultQueryLimit, limit)
		limit = defaultQueryLimit
	}

	timeNow := e.timeNowFunc()
	newStartTime := timeNow.Add(-threeDaysHistoricalPeriod)
	if options.StartTime != nil {
		newStartTime = *options.StartTime
		if timeNow.Sub(newStartTime) > maxHistoricalDataQueryPeriod {
			newStartTime = timeNow.Add(-maxHistoricalDataQueryPeriod)
			log.Warnf("!!!OKX EXCHANGE API NOTICE!!! The trade API cannot query data beyond 90 days from the current date, update %s -> %s", *options.StartTime, newStartTime)
		}
	}

	endTime := timeNow
	if options.EndTime != nil {
		if options.EndTime.Before(newStartTime) {
			return nil, fmt.Errorf("end time %s before start %s", *options.EndTime, newStartTime)
		}
		if options.EndTime.Sub(newStartTime) > maxHistoricalDataQueryPeriod {
			return nil, fmt.Errorf("start time %s and end time %s cannot greater than 90 days", newStartTime, options.EndTime)
		}
		endTime = *options.EndTime
	}

	if options.LastTradeID != 0 {
		// we don't support the last trade id as a filter because okx supports bill ID only.
		// we don't have any more fields (types.Trade) to store it.
		log.Infof("Last trade id not supported on QueryTrades")
	}

	if timeNow.Sub(newStartTime) <= threeDaysHistoricalPeriod {
		c := e.client.NewGetThreeDaysTransactionHistoryRequest().
			InstrumentID(toLocalSymbol(symbol)).
			StartTime(newStartTime).
			EndTime(endTime).
			Limit(uint64(limit))
		return getTrades(ctx, limit, func(ctx context.Context, billId string) ([]okexapi.Trade, error) {
			c.Before(billId)
			return c.Do(ctx)
		})
	}

	c := e.client.NewGetTransactionHistoryRequest().
		InstrumentID(toLocalSymbol(symbol)).
		StartTime(newStartTime).
		EndTime(endTime).
		Limit(uint64(limit))
	return getTrades(ctx, limit, func(ctx context.Context, billId string) ([]okexapi.Trade, error) {
		c.Before(billId)
		return c.Do(ctx)
	})
}

func getTrades(ctx context.Context, limit int64, doFunc func(ctx context.Context, billId string) ([]okexapi.Trade, error)) (trades []types.Trade, err error) {
	billId := "0"
	for {
		response, err := doFunc(ctx, billId)
		if err != nil {
			return nil, fmt.Errorf("failed to query trades, err: %w", err)
		}

		for _, trade := range response {
			trades = append(trades, tradeToGlobal(trade))
		}

		tradeLen := int64(len(response))
		// a defensive programming to ensure the length of order response is expected.
		if tradeLen > limit {
			return nil, fmt.Errorf("unexpected trade length %d", tradeLen)
		}

		if tradeLen < limit {
			break
		}
		// use Before filter to get all data.
		billId = response[tradeLen-1].BillId.String()
	}
	return trades, nil
}

func (e *Exchange) SupportedInterval() map[types.Interval]int {
	return SupportedIntervals
}

func (e *Exchange) IsSupportedInterval(interval types.Interval) bool {
	_, ok := SupportedIntervals[interval]
	return ok
}

func (e *Exchange) PlaceTakeProfitAndStopLossOrder(ctx context.Context, position *types.Position) error {
	orderReq := e.client.NewPlaceOCOAlgoOrderRequest()

	orderReq.InstID(toLocalSymbol(position.Symbol))

	// Margin mode cross isolated
	if e.IsIsolatedMargin {
		orderReq.TdMode(okexapi.TradeModeIsolated)
	} else {
		orderReq.TdMode(okexapi.TradeModeCross)
	}

	orderReq.Ccy(position.Market.QuoteCurrency)

	if position.IsLong() {
		orderReq.Side("sell")
		orderReq.Sz(position.Market.FormatQuantity(position.Quote.Abs()))
	} else if position.IsShort() {
		orderReq.Side("buy")
		orderReq.Sz(position.Market.FormatQuantity(position.Base.Abs()))
	}

	if position.TpTriggerPx != nil {
		orderReq.TpTriggerPxType("last")
		orderReq.TpTriggerPx(position.Market.FormatPrice(*position.TpTriggerPx))
		orderReq.TpOrdPx("-1")
	}

	if position.SlTriggerPx != nil {
		orderReq.SlTriggerPxType("last")
		orderReq.SlTriggerPx(position.Market.FormatPrice(*position.SlTriggerPx))
		orderReq.SlOrdPx("-1")
	}

	params, _ := orderReq.GetParameters()
	log.WithField("params", params).
		Info("PlaceTakeProfitStopLossOrder_start")

	orders, err := orderReq.Do(ctx)
	if err != nil {
		log.WithField("params", params).
			WithError(err).
			Error("PlaceTakeProfitStopLossOrder_fail")

		return err
	}

	log.WithField("orders", orders).
		Info("PlaceTakeProfitStopLossOrder_ok")

	return nil
}

func (e *Exchange) UpdatePosition(ctx context.Context, position *types.Position) error {
	e.autoCancelAllAlgoOrders(ctx, position.Symbol)
	return e.PlaceTakeProfitAndStopLossOrder(ctx, position)
}

func (e *Exchange) QueryPositionInfo(ctx context.Context, symbol string) (*types.PositionInfo, error) {
	position := &types.PositionInfo{
		Symbol: symbol,
	}

	orders, err := e.QueryRawAlgoOpenOrders(context.Background(), symbol)
	if err != nil {
		log.WithError(err).Error("handlePositionDetailsEvent_QueryRawAlgoOpenOrders_error")
		return nil, errors.Wrap(err, "QueryPositionInfo_QueryRawAlgoOpenOrders_error")
	}

	ocoOrders, err := e.QueryRawOCOAlgoOpenOrders(context.Background(), symbol)
	if err != nil {
		log.WithError(err).Error("handlePositionDetailsEvent_QueryRawOCOAlgoOpenOrders_error")
		return nil, errors.Wrap(err, "QueryPositionInfo_QueryRawOCOAlgoOpenOrders_error")
	}

	orders = append(orders, ocoOrders...)

	if len(orders) > 0 {
		algoOrder := orders[0]

		slTriggerPx, err := fixedpoint.NewFromString(algoOrder.SlTriggerPx)
		if err != nil {
			log.WithError(err).Error("handlePositionDetailsEvent_parse_SlTriggerPx_error")
			return nil, errors.Wrap(err, "handlePositionDetailsEvent_parse_SlTriggerPx_error")
		} else {
			position.SlTriggerPxType = algoOrder.SlTriggerPxType
			position.SlOrdPx = algoOrder.SlOrdPx
			position.SlTriggerPx = &slTriggerPx
		}

		tpTriggerPx, err := fixedpoint.NewFromString(algoOrder.TpTriggerPx)
		if err != nil {
			log.WithError(err).Error("handlePositionDetailsEvent_parse_TpTriggerPx_error")
			return nil, errors.Wrap(err, "handlePositionDetailsEvent_parse_TpTriggerPx_error")
		} else {
			position.TpTriggerPxType = algoOrder.TpTriggerPxType
			position.TpOrdPx = algoOrder.TpOrdPx
			position.TpTriggerPx = &tpTriggerPx
		}
	}

	log.WithField("symbol", symbol).
		WithField("position", position).
		WithField("orders", orders).
		Info("Exchange#QueryPositionInfo_result")

	return position, nil
}
