[![Go Report Card](https://goreportcard.com/badge/github.com/yckao/go-batcher)](https://goreportcard.com/report/github.com/yckao/go-batcher)
[![codecov](https://codecov.io/gh/yckao/go-batcher/graph/badge.svg?token=LY64CZWZPN)](https://codecov.io/gh/yckao/go-batcher)
[![test](https://github.com/yckao/go-batcher/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/yckao/go-batcher/actions/workflows/test.yml)
[![CodeQL](https://github.com/yckao/go-batcher/actions/workflows/github-code-scanning/codeql/badge.svg?branch=main)](https://github.com/yckao/go-batcher/actions/workflows/github-code-scanning/codeql)
[![Godoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/yckao/go-batcher)

# go-batcher

This is a batcher inspired by [GraphQL's Dataloader](https://github.com/graphql/dataloader) and [yckao's dataloader implementation](https://github.com/yckao/go-dataloader). 
If dataloader is batch to load somedata, why not use same way to **do something?**
That's why this project start.

## Features

## Requirement

- go >= 1.18

## Getting Started

```go
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
	cc := batcher.NewLimitedConcurrencyControl(10)
	action := batcher.NewAction[string, *ExampleData](batchFn)
	batcher := batcher.New[string, *ExampleData](ctx, action, batcher.WithConcurrencyControl[string, *ExampleData](
		cc,
	))

	thunk := batcher.Do(ctx, "World")
	// You can call shutdown to graceful shutdown batcher
	batcher.Shutdown()
	val, err := thunk.Await(ctx)
	fmt.Printf("value: %v, err: %v\n", val, err)
}
```

