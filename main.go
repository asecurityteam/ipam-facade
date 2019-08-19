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
	Postgres   *sqldb.PostgresConfig
	Device42 *ipamfetcher.Device42ClientConfig
}

func (*config) Name() string {
	return "ipamfacade"
}

type component struct {
	Producer *producer.Component
	Postgres *sqldb.PostgresComponent
	Device42 *ipamfetcher.Device42ClientComponent
}

func (c *component) Settings() *config {
	return &config{
		LambdaMode: false,
		Producer:   &producerConfig{c.Producer.Settings()},
		Postgres:   &sqldb.PostgresConfig{c.Postgres.Settings()},
		Device42: &ipamfetcher.Device42ClientConfig{c.Device42.Settings()},
	}
}

func newComponent() *component {
	return &component{
		Producer: producer.NewComponent(),
		Postgres: sqldb.NewPostgresComponent(),
		Device42: ipamfetcher.NewDevice42ClientComponent(),
	}
}

func (c *component) New(ctx context.Context, conf *config) (func(context.Context, settings.Source) error, error) {
	p, err := c.Producer.New(ctx, conf.Producer.Config)
	if err != nil {
		return nil, err
	}
	enqueueHandler := &v1.EnqueueHandler{
		RandomNumberGenerator: &randomnumbergenerator.UUIDGenerator{},
		Producer:              p,
		LogFn:                 domain.LoggerFromContext,
	}

	pgdb, err := c.Postgres.New(ctx, conf.PostgresConfig)
	if err != nil {
		return nil, err
	}

	dc, err := c.Device42.New(ctx, conf.Device42)
	if err != nil {
		return nil, err
	}

	deviceFetcher := ipamfetcher.NewDevice42DeviceFetcher(dc)
	subnetFetcher :=  ipamfetcher.NewDevice42SubnetFetcher(dc)
	customerFetcher := ipamfetcher.NewDevice42CustomerFetcher(dc)
	ipamDataFetcher := &ipamfetcher.Client{
		CustomerFetcher: customerFetcher,
		DeviceFetcher:   deviceFetcher,
		SubnetFetcher:   subnetFetcher,
	}

	assetFetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: pbdb}
	fetchHandler := &v1.FetchByIPAddressHandler{
		LogFn:                domain.LoggerFromContext,
		PhysicalAssetFetcher: assetFetcher,
	}
	assetStorer := &assetstorer.PostgresPhysicalAssetStorer{DB: pgdb}
	syncHandler := &v1.SyncIPAMDataHandler{
		IPAMDataFetcher:     ipamDataFetcher,
		LogFn:               domain.LoggerFromContext,
		PhysicalAssetStorer: assetStorer,
	}

	handlers := map[string]serverfull.Function{
		"fetchbyip": serverfull.NewFunction(fetchHandler.Handle),
		"sync":      serverfull.NewFunction(syncHandler.Handle),
		"enqueue":   serverfull.NewFunction(enqueueHandler.Handle),
	}

	fetcher := &serverfull.StaticFetcher{Functions: handlers}
	if conf.LambdaMode {
		return func(ctx context.Context, source settings.Source) error {
			return serverfull.StartLambda(ctx, source, fetcher, "filter")
		}, nil
	}
	return func(ctx context.Context, source settings.Source) error {
		return serverfull.StartHTTP(ctx, source, fetcher)
	}, nil
}

func main() {
	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}
	ctx := context.Background()
	runner := new(func(context.Context, settings.Source) error)
	cmp := newComponent()

	err = settings.NewComponent(ctx, source, cmp, runner)
	if err != nil {
		panic(err.Error())
	}
	if err := (*runner)(ctx, source); err != nil {
		panic(err.Error())
	}
}
