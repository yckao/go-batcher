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
	cc := batcher.NewDefaultConcurrencyControl(100)
	batcher := batcher.New[string, *ExampleData](ctx, batchFn, batcher.WithConcurrencyControl[string, *ExampleData](
		cc,
	))

	thunk := batcher.Do(ctx, "World")
	// You can call shutdown to graceful shutdown batcher
	batcher.Shutdown()
	val, err := thunk.Await(ctx)
	fmt.Printf("value: %v, err: %v\n", val, err)
}
