package confdb

import (
	"crypto/tls"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/connect"

	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
)

// ConsulRegister ...
func ConsulRegister(consulAddress string,
	consulPort string,
	srvAddress string,
	srvPort string) (registar sd.Registrar, tlsConfig *tls.Config) {
	// Logging domain
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	// Service discovery domain.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = consulAddress + ":" + consulPort
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		srv, _ := connect.NewService("hello", consulClient)
		tlsConfig = srv.ServerTLSConfig()
		client = consulsd.NewClient(consulClient)
	}

	check := api.AgentServiceCheck{
		HTTP:     "https://" + srvAddress + ":" + srvPort + "/health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	port, _ := strconv.Atoi(srvPort)
	num := rand.Intn(100)
	asr := api.AgentServiceRegistration{
		ID:      "hello" + string(num),
		Name:    "hello",
		Address: srvAddress,
		Port:    port,
		Tags:    []string{"hello", "playgound"},
		Check:   &check,
	}
	registar = consulsd.NewRegistrar(client, &asr, logger)
	return
}
