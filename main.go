package main

import (
	"context"
	"net/http"
	"os"

	producer "github.com/asecurityteam/component-producer"
	"github.com/asecurityteam/ipam-facade/pkg/assetfetcher"
	"github.com/asecurityteam/ipam-facade/pkg/assetstorer"
	"github.com/asecurityteam/ipam-facade/pkg/domain"
	v1 "github.com/asecurityteam/ipam-facade/pkg/handlers/v1"
	"github.com/asecurityteam/ipam-facade/pkg/ipamfetcher"
	"github.com/asecurityteam/ipam-facade/pkg/randomnumbergenerator"
	"github.com/asecurityteam/ipam-facade/pkg/sqldb"
	"github.com/asecurityteam/serverfull"
	"github.com/asecurityteam/settings"
)

type producerConfig struct {
	*producer.Config
}

func (*producerConfig) Name() string {
	return "producer"
}

type config struct {
	LambdaMode bool `description:"Use the Lambda SDK to start the system."`
	Producer   *producerConfig
}

func (*config) Name() string {
	return "ipamfacade"
}

type component struct {
	Producer *producer.Component
}

func (c *component) Settings() *config {
	return &config{
		LambdaMode: false,
		Producer:   &producerConfig{c.Producer.Settings()},
	}
}

func newComponent() *component {
	return &component{
		Producer: producer.NewComponent(),
	}
}

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

	postgresConfigComponent := &sqldb.PostgresConfigComponent{}
	postgresdb := new(sqldb.PostgresDB)
	if err = settings.NewComponent(ctx, source, postgresConfigComponent, postgresdb); err != nil {
		panic(err.Error())
	}
	assetFetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: postgresdb}
	fetchHandler := &v1.FetchByIPAddressHandler{
		LogFn:                domain.LoggerFromContext,
		PhysicalAssetFetcher: assetFetcher,
	}
	assetStorer := &assetstorer.PostgresPhysicalAssetStorer{DB: postgresdb}
	syncHandler := &v1.SyncIPAMDataHandler{
		IPAMDataFetcher:     ipamClient,
		LogFn:               domain.LoggerFromContext,
		PhysicalAssetStorer: assetStorer,
	}
	p, err := c.Producer.New(ctx, conf.Producer.Config)
	if err != nil {
		return nil, err
	}
	enqueueHandler := &v1.EnqueueHandler{
		RandomNumberGenerator: &randomnumbergenerator.UUIDGenerator{},
		Producer:              p,
		LogFn:                 domain.LoggerFromContext,
	}
	handlers := map[string]serverfull.Function{
		"fetchbyip": serverfull.NewFunction(fetchHandler.Handle),
		"sync":      serverfull.NewFunction(syncHandler.Handle),
		"enqueue":   serverfull.NewFunction(enqueueHandler.Handle),
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
		Limit: device42IPsConfig.Limit,
		PageFetcher: &ipamfetcher.Device42PageFetcher{
			Client:   http.DefaultClient,
			Endpoint: device42IPsConfig.Endpoint,
		},
	}

	device42SubnetsClientComponent := &ipamfetcher.Device42ClientComponent{}
	device42SubnetsConfig := new(ipamfetcher.Device42ClientConfig)
	if err := settings.NewComponent(ctx, subnetsPrefixedEnv, device42SubnetsClientComponent, device42SubnetsConfig); err != nil {
		return nil, err
	}
	subnetsFetcher := &ipamfetcher.Device42SubnetFetcher{
		Limit: device42SubnetsConfig.Limit,
		PageFetcher: &ipamfetcher.Device42PageFetcher{
			Client:   http.DefaultClient,
			Endpoint: device42SubnetsConfig.Endpoint,
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
