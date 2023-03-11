package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type ConversionResponse struct {
	// TODO make this a real date
	Date       string  `json:"date,omitempty"`
	Rate       float64 `json:"rate,omitempty"`
	From       string  `json:"from,omitempty"`
	FromAmount float64 `json:"fromAmount,omitempty"`
	To         string  `json:"to,omitempty"`
	ToAmount   float64 `json:"toAmount,omitempty"`
}

type Getter func(url string) (resp *http.Response, err error)

type ConvertService struct {
	getter Getter
}

// TODO use functional options API pattern here (finish github.com/fn-go/fnoptions)
func NewConvertService(getter Getter) (*ConvertService, error) {
	if getter == nil {
		return nil, errors.New("getter cannot be nil")
	}

	return &ConvertService{
		getter: getter,
	}, nil
}

func (s *ConvertService) Convert(from, to string, fromAmount float64) (ConversionResponse, error) {
	url := fmt.Sprintf(
		"https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies/%s/%s.json",
		from,
		to,
	)

	// TODO this could probably be cached
	// TODO we could examine what the error is and handle it gracefully
	// Possible Scenarios:
	// - throttled -> retry with exponential backoff
	// - bad request -> we did not validate the user input, so maybe we should do that
	// - bad gateway or other networking related errors -> retry with exponential backoff
	resp, err := s.getter(url)
	if err != nil {
		return ConversionResponse{}, errors.Wrap(err, "failed to make http request")
	}

	if resp.StatusCode != http.StatusOK {
		// TODO this isn't very helpful
		return ConversionResponse{}, errors.Errorf("upstream api get error: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	// TODO this could be an actual model
	// But that's currently out of scope due to the dynamic nature of the keys presented in the response
	// i.e. the rate given is keyed on the _to_ currency, e.g. eur: 123.4
	response := map[string]interface{}{}
	err = json.Unmarshal(body, &response)

	responseDate, ok := response["date"]
	if !ok {
		return ConversionResponse{}, errors.New("unknown response, no date found")
	}

	responseDateStr, ok := responseDate.(string)
	if !ok {
		return ConversionResponse{}, errors.New("unknown response, date is not a string")
	}

	responseConversionRate, ok := response[to]
	if !ok {
		return ConversionResponse{}, errors.Errorf("unknown response, key [%s] not found", to)
	}

	responseConversionRateFloat, ok := responseConversionRate.(float64)
	if !ok {
		return ConversionResponse{}, errors.Errorf("unknown response, value of [%s], [%s], is not a float", to, responseConversionRate)
	}

	return ConversionResponse{
		Date:       responseDateStr,
		Rate:       responseConversionRateFloat,
		From:       from,
		FromAmount: fromAmount,
		To:         to,
		// TODO this could result in overflow issues, could use some validation
		ToAmount: responseConversionRateFloat * fromAmount,
	}, nil
}
