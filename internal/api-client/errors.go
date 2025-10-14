package api_client

import "errors"

var (
	ErrClientResponseNil              = errors.New("client response is nil")
	ErrClientResponseStatusCodeNot200 = errors.New("client response status code is not 200")
	ErrClientExpected200              = errors.New("expected 200 status code")
	ErrDataNil                        = errors.New("data is nil or empty")
	ErrDataEmpty                      = errors.New("data is empty")
	ErrUnmarshalData                  = errors.New("error unmarshalling data")
	ErrFetcherNil                     = errors.New("could not create fetcher")
)
