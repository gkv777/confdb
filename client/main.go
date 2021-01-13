package main

import (
	"context"
	"fmt"

	cg "github.com/segmentio/consul-go"
)

func main() {
	// Queries Consul for a list of addresses where "my-service" is available,
	// the result will be sorted to get the addresses closest to the agent first.
	rslv := &cg.Resolver{
		Client:             &cg.Client{},
		ServiceTags:        []string{},
		NodeMeta:           map[string]string{},
		OnlyPassing:        false,
		AllowStale:         false,
		AllowCached:        false,
		DisableCoordinates: false,
		Cache:              &cg.ResolverCache{},
		Blacklist:          &cg.ResolverBlacklist{},
		Agent:              &cg.Agent{},
		Tomography:         &cg.Tomography{},
		Balancer:           nil,
		Sort: func([]cg.Endpoint) {
		},
	}

	addrs, err := rslv.LookupService(context.Background(), "hello")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, addr := range addrs {
		fmt.Printf("%#v \n", addr)
	}
}
