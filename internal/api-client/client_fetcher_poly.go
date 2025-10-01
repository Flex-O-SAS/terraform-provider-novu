// Polymorphic fetcher for the api client
// TODO : à fusionner avec client_fetcher.go si on part sur cette solution
package api_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	noClientErr = "No client defined in the handler"
)

type handler[T any] struct {
	client *ApiClient
	path   string
	method string
}

func (h *handler[T]) call(ctx context.Context, body any) (*T, *ApiClientResponse, error) {
	if h.client == nil {
		return nil, nil, errors.New(noClientErr)
	}
	var bodyReader io.Reader
	switch b := body.(type) {
	case nil:
		// no body
	case io.Reader:
		bodyReader = b
	case []byte:
		bodyReader = bytes.NewReader(b)
	case json.RawMessage:
		bodyReader = bytes.NewReader(b)
	default:
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	apiRes, err := h.client.call(ctx, h.method, h.path, bodyReader)
	if err != nil {
		return nil, apiRes, err
	}

	var data T
	err = unmarshalBodyKey(apiRes.Body, &data, "data")

	if err != nil {
		return nil, apiRes, err
	}

	// TODO : faut-il vérifier que data n'est pas vide ?
	return &data, apiRes, nil
}

// methods cannot take type arguments, so we use a function with a type argument to return a generic handler
func fetcher[T any](client *ApiClient, method string, path string) *handler[T] {
	return &handler[T]{
		client: client,
		path:   path,
		method: method,
	}
}

// relicat, peut être fusionné avec call si pas de use case pour le moment
func (h *handler[T]) Do(ctx context.Context, body any) (*T, *ApiClientResponse, error) {
	return h.call(ctx, body)
}

func unmarshalBodyKey[T any](body []byte, p *T, dataKey string) error {
	if p == nil {
		return fmt.Errorf("pointer is nil")
	}
	if len(body) == 0 {
		return fmt.Errorf("body is empty")
	}

	// Use an envelope to check the data exists
	// Also allows any data key to be used
	var env map[string]json.RawMessage
	err := json.Unmarshal(body, &env)
	if err != nil {
		return err
	}
	if len(env) == 0 {
		return fmt.Errorf("body is empty")
	}
	raw, ok := env[dataKey]
	if !ok {
		return fmt.Errorf("%s is not present in the body", dataKey)
	}

	var dataEnv map[string]json.RawMessage
	err = json.Unmarshal(raw, &dataEnv)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, p)
	if err != nil {
		return err
	}
	return nil
}

/*
// unused, c'est ici pour comparer à unmarshalBodyKey durant la PR
// sera supprimé avant merge
// TODO: ne pas oublier de supprimer ce code
type bodyDataGeneric[T any] struct {
	Data T `json:"data"`
}

func unmarshalBodyData[T any](body []byte, p *T) error {
	if p == nil {
		return fmt.Errorf("pointer is nil")
	}
	if len(body) == 0 {
		return fmt.Errorf("body is empty")
	}
	var bodyData bodyDataGeneric[T]
	err := json.Unmarshal(body, &bodyData)
	if err != nil {
		return err
	}
	*p = bodyData.Data
	return nil
}
*/
