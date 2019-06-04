package main

import (
	"context"
	"net/http"
	"os"

	"github.com/asecurityteam/ipam-facade/pkg/ipamfetcher"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	v1 "github.com/asecurityteam/ipam-facade/pkg/handlers/v1"
	"github.com/asecurityteam/serverfull"
	"github.com/asecurityteam/settings"
)

func main() {
	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}
	ipamClient, err := getIPAMClient(ctx, source)
	if err != nil {
		panic(err.Error())
	}
	fetchHandler := &v1.FetchByIPAddressHandler{
		LogFn: domain.LoggerFromContext,
	}
	syncHandler := &v1.SyncIPAMDataHandler{
		IPAMDataFetcher: ipamClient,
		LogFn:           domain.LoggerFromContext,
	}
	handlers := map[string]serverfull.Function{
		"fetchbyip": serverfull.NewFunction(fetchHandler.Handle),
		"sync":      serverfull.NewFunction(syncHandler.Handle),
	}

	fetcher := &serverfull.StaticFetcher{Functions: handlers}
	if err := serverfull.Start(ctx, source, fetcher); err != nil {
		panic(err.Error())
	}
}

func getIPAMClient(ctx context.Context, root settings.Source) (*ipamfetcher.Client, error) {
	ipsPrefixedEnv := &settings.PrefixSource{
		Source: root,
		Prefix: []string{"IPS"},
	}
	subnetsPrefixedEnv := &settings.PrefixSource{
		Source: root,
		Prefix: []string{"SUBNETS"},
	}
	customersPrefixedEnv := &settings.PrefixSource{
		Source: root,
		Prefix: []string{"CUSTOMERS"},
	}
	device42IPsClientComponent := &ipamfetcher.Device42ClientComponent{}
	device42IPsConfig := new(ipamfetcher.Device42ClientConfig)
	if err := settings.NewComponent(ctx, ipsPrefixedEnv, device42IPsClientComponent, device42IPsConfig); err != nil {
		return nil, err
	}
	devicesFetcher := &ipamfetcher.Device42DeviceFetcher{
		Iterator: &ipamfetcher.Device42PageIterator{
			Context: ctx,
			Limit:   device42IPsConfig.Limit,
			PageFetcher: &ipamfetcher.Device42PageFetcher{
				Client:   http.DefaultClient,
				Endpoint: device42IPsConfig.Endpoint,
			},
		},
	}

	device42SubnetsClientComponent := &ipamfetcher.Device42ClientComponent{}
	device42SubnetsConfig := new(ipamfetcher.Device42ClientConfig)
	if err := settings.NewComponent(ctx, subnetsPrefixedEnv, device42SubnetsClientComponent, device42SubnetsConfig); err != nil {
		return nil, err
	}
	subnetsFetcher := &ipamfetcher.Device42SubnetFetcher{
		Iterator: &ipamfetcher.Device42PageIterator{
			Context: ctx,
			Limit:   device42SubnetsConfig.Limit,
			PageFetcher: &ipamfetcher.Device42PageFetcher{
				Client:   http.DefaultClient,
				Endpoint: device42SubnetsConfig.Endpoint,
			},
		},
	}

	device42CustomersClientComponent := &ipamfetcher.Device42ClientComponent{}
	device42CustomersConfig := new(ipamfetcher.Device42ClientConfig)
	if err := settings.NewComponent(ctx, customersPrefixedEnv, device42CustomersClientComponent, device42CustomersConfig); err != nil {
		return nil, err
	}
	customerFetcher := &ipamfetcher.Device42CustomerFetcher{
		Client:   http.DefaultClient,
		Endpoint: device42CustomersConfig.Endpoint,
	}

	return &ipamfetcher.Client{
		CustomerFetcher: customerFetcher,
		DeviceFetcher:   devicesFetcher,
		SubnetFetcher:   subnetsFetcher,
	}, nil
}
