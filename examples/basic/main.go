package main

import (
	"context"
	"fmt"

	"github.com/yckao/go-batcher"
)

type ExampleData struct {
	Message string
}

func batchFn(ctx context.Context, requests []string) []batcher.Response[*ExampleData] {
	response := make([]batcher.Response[*ExampleData], len(requests))
	for index, req := range requests {
		// do something
		response[index] = batcher.Response[*ExampleData]{
			Response: &ExampleData{
				Message: fmt.Sprintf("Hello %s", req),
			},
		}
	}
	return response
}

func main() {
	ctx := context.Background()
	loader := batcher.New[string, *ExampleData](ctx, batchFn, batcher.WithConcurrencyControl[string, *ExampleData](
		batcher.NewDefaultConcurrencyControl(100),
	))

	thunk := loader.Do(ctx, "World")
	val, err := thunk.Await(ctx)
	fmt.Printf("value: %v, err: %v\n", val, err)
}
