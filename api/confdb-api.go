package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gkv777/confdb"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	ht "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
)

func main() {
	consulAddr := os.Getenv("CONS_ADDR")
	if consulAddr == "" {
		consulAddr = "localhost"
	}
	consulPort := os.Getenv("CONS_PORT")
	if consulPort == "" {
		consulPort = "8500"
	}

	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "localhost"
	}

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println(`Service's port not set!`)
		os.Exit(1)
	}

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()

		consulConfig.Address = "http://" + consulAddr + ":" + consulPort
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	tags := []string{"hello", "playgound"}
	passingOnly := true
	duration := 500 * time.Millisecond
	var helloEndpoint endpoint.Endpoint

	ctx := context.Background()
	r := mux.NewRouter()

	factory := helloFactory(ctx, "GET", "/hello")
	instancer := consulsd.NewInstancer(client, logger, "hello", tags, passingOnly)
	endpointer := sd.NewEndpointer(instancer, factory, logger)
	//subscriber := consulsd.NewSubscriber(client, factory, logger, "hello", tags, passingOnly)
	//balancer := lb.NewRoundRobin(subscriber)
	balancer := lb.NewRoundRobin(endpointer)
	retry := lb.Retry(3, duration, balancer)
	helloEndpoint = retry

	r.Handle("/hello/rocket", ht.NewServer(helloEndpoint, confdb.DecodeHelloRequest, confdb.EncodeResponse))

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", address)
		errc <- http.ListenAndServe(address+":"+port, r)
	}()

	// Run!
	logger.Log("exit", <-errc)
}

func helloFactory(_ context.Context, method, path string) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}

		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		tgt.Path = path

		var (
			enc ht.EncodeRequestFunc
			dec ht.DecodeResponseFunc
		)
		enc, dec = confdb.EncodeJSONRequest, confdb.DecodeHelloResponse

		return ht.NewClient(method, tgt, enc, dec).Endpoint(), nil, nil
	}

}
