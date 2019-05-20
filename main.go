package main

import (
	"context"
	"os"

	v1 "github.com/asecurityteam/ipam-facade/pkg/handlers/v1"
	"github.com/asecurityteam/serverfull"
	"github.com/asecurityteam/settings"
)

func main() {
	ctx := context.Background()
	greetingHandler := &v1.GreetingHandler{}
	handlers := map[string]serverfull.Function{
		"hello": serverfull.NewFunction(greetingHandler.Handle),
	}

	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}
	fetcher := &serverfull.StaticFetcher{Functions: handlers}
	if err := serverfull.Start(ctx, source, fetcher); err != nil {
		panic(err.Error())
	}
}
