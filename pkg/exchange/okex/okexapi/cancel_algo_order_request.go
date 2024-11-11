package okexapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"

	"github.com/c9s/requestgen"
)

type CancelAlgoOrderResponse struct {
	AlgoID string `json:"algoId"`
	SCode  string `json:"sCode"`
	SMsg   string `json:"sMsg"`
}

type CancelAlgoOrder struct {
	InstrumentID string `json:"instId"`
	AlgoOrderID  string `json:"algoId"`
}

type CancelAlgoOrderRequest struct {
	client requestgen.AuthenticatedAPIClient

	Payload []*CancelAlgoOrder
}

func (c *RestClient) NewCancelAlgoOrderRequest() *CancelAlgoOrderRequest {
	return &CancelAlgoOrderRequest{
		client: c,
	}
}

func (c *CancelAlgoOrderRequest) SetPayload(CancelAlgoOrders []*CancelAlgoOrder) *CancelAlgoOrderRequest {
	c.Payload = CancelAlgoOrders
	return c
}

// GetQueryParameters builds and checks the query parameters and returns url.Values
func (c *CancelAlgoOrderRequest) GetQueryParameters() (url.Values, error) {
	var params = map[string]interface{}{}

	query := url.Values{}
	for _k, _v := range params {
		query.Add(_k, fmt.Sprintf("%v", _v))
	}

	return query, nil
}

// GetParameters builds and checks the parameters and return the result in a map object
func (c *CancelAlgoOrderRequest) GetParameters() (interface{}, error) {
	return c.Payload, nil
}

// GetParametersJSON converts the parameters from GetParameters into the JSON format
func (c *CancelAlgoOrderRequest) GetParametersJSON() ([]byte, error) {
	params, err := c.GetParameters()
	if err != nil {
		return nil, err
	}

	return json.Marshal(params)
}

// GetSlugParameters builds and checks the slug parameters and return the result in a map object
func (c *CancelAlgoOrderRequest) GetSlugParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}

	return params, nil
}

func (c *CancelAlgoOrderRequest) applySlugsToUrl(url string, slugs map[string]string) string {
	for _k, _v := range slugs {
		needleRE := regexp.MustCompile(":" + _k + "\\b")
		url = needleRE.ReplaceAllString(url, _v)
	}

	return url
}

func (c *CancelAlgoOrderRequest) iterateSlice(slice interface{}, _f func(it interface{})) {
	sliceValue := reflect.ValueOf(slice)
	for _i := 0; _i < sliceValue.Len(); _i++ {
		it := sliceValue.Index(_i).Interface()
		_f(it)
	}
}

func (c *CancelAlgoOrderRequest) isVarSlice(_v interface{}) bool {
	rt := reflect.TypeOf(_v)
	switch rt.Kind() {
	case reflect.Slice:
		return true
	}
	return false
}

func (c *CancelAlgoOrderRequest) GetSlugsMap() (map[string]string, error) {
	slugs := map[string]string{}
	params, err := c.GetSlugParameters()
	if err != nil {
		return slugs, nil
	}

	for _k, _v := range params {
		slugs[_k] = fmt.Sprintf("%v", _v)
	}

	return slugs, nil
}

// GetPath returns the request path of the API
func (c *CancelAlgoOrderRequest) GetPath() string {
	return "/api/v5/trade/cancel-algos"
}

// Do generates the request object and send the request object to the API endpoint
func (c *CancelAlgoOrderRequest) Do(ctx context.Context) ([]CancelAlgoOrderResponse, error) {

	params, err := c.GetParameters()
	if err != nil {
		return nil, err
	}
	query := url.Values{}

	var apiURL string

	apiURL = c.GetPath()

	req, err := c.client.NewAuthenticatedRequest(ctx, "POST", apiURL, query, params)
	if err != nil {
		return nil, err
	}

	response, err := c.client.SendRequest(req)
	if err != nil {
		return nil, err
	}

	var apiResponse APIResponse
	if err := response.DecodeJSON(&apiResponse); err != nil {
		return nil, err
	}

	type responseValidator interface {
		Validate() error
	}
	validator, ok := interface{}(apiResponse).(responseValidator)
	if ok {
		if err := validator.Validate(); err != nil {
			return nil, err
		}
	}
	var data []CancelAlgoOrderResponse
	if err := json.Unmarshal(apiResponse.Data, &data); err != nil {
		return nil, err
	}
	return data, nil
}
