package v1

import (
	"context"
	"fmt"
)

// GreetingInput represents the given name of the requesting user.
type GreetingInput struct {
	Name string `json:"name"`
}

// GreetingOutput represents the greeting based on the requesting user's name.
type GreetingOutput struct {
	Greeting string `json:"greeting"`
}

// GreetingHandler processes incoming greeting requests.
type GreetingHandler struct {
}

// Handle constructs and return a greeting based on the given name.
func (h *GreetingHandler) Handle(ctx context.Context, in GreetingInput) (GreetingOutput, error) {
	greeting := fmt.Sprintf("Hello %s!", in.Name)
	return GreetingOutput{Greeting: greeting}, nil
}
