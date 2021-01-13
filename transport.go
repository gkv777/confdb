package confdb

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

type helloRequest struct {
	Name string `json:"name"`
}

type helloResponse struct {
	Message string `json:"message"`
}

type healthRequest struct {
}

type healthResponse struct {
	Status bool `json:"status"`
}

func DecodeHelloRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req helloRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func DecodeHealthRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return healthRequest{}, nil
}

func EncodeResponse(_ context.Context, w http.ResponseWriter, resp interface{}) error {
	return json.NewEncoder(w).Encode(resp)
}

func MakeHealthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		status := svc.HealthCheck()
		return healthResponse{Status: status}, nil
	}
}

func MakeHelloEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		r := req.(helloRequest)
		hello := svc.SayHello(r.Name)
		return helloResponse{Message: hello}, nil
	}
}

func EncodeJSONRequest(_ context.Context, req *http.Request, request interface{}) error {
	// simple JSON serialization to the request body.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
	return nil
}

func DecodeHelloResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response helloResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}
