package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gkv777/confdb"
	ht "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
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

	errChan := make(chan error)

	svc := confdb.NewService()
	r := mux.NewRouter()

	r.Handle("/hello", ht.NewServer(
		confdb.MakeHelloEndpoint(svc),
		confdb.DecodeHelloRequest,
		confdb.EncodeResponse,
	))

	r.Methods("GET").Path("/health").Handler(ht.NewServer(
		confdb.MakeHealthEndpoint(svc),
		confdb.DecodeHealthRequest,
		confdb.EncodeResponse,
	))

	registar, tls := confdb.ConsulRegister(consulAddr, consulPort, address, port)

	// HTTP transport
	go func() {
		log.Println("httpAddress", address+":"+port)
		tls.InsecureSkipVerify = true
		server := &http.Server{
			Addr:      address + ":" + port,
			Handler:   r,
			TLSConfig: tls,
		}
		fmt.Printf("\n\n%#v\n\n", tls)
		fmt.Printf("\n\n%#v\n\n", &tls)
		registar.Register()
		errChan <- server.ListenAndServeTLS("", "")
		//handler := r
		//errChan <- http.ListenAndServe(":"+port, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	registar.Deregister()
	log.Fatalln(error)
}
