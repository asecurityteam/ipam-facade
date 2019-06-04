package main

import (
	"context"
	"os"

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
	fetchHandler := &v1.FetchByIPAddressHandler{
		LogFn: domain.LoggerFromContext,
	}
	syncHandler := &v1.SyncIPAMDataHandler{
		LogFn: domain.LoggerFromContext,
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
