package main

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/vanus-labs/connect-runtime"
)

func main() {
	connectorEventHandlerFuncs := connect_runtime.ConnectorEventHandlerFuncs{
		AddFunc: func(connectorID, config string) error {
			fmt.Printf("===AddFunc=== %s\n", connectorID)
			return nil
		},
		UpdateFunc: func(connectorID, config string) error {
			fmt.Printf("===UpdateFunc=== %s\n", connectorID)
			return nil
		},
		DeleteFunc: func(connectorID string) error {
			fmt.Printf("===DeleteFunc=== %s\n", connectorID)
			return nil
		},
	}
	r, err := connect_runtime.New(connect_runtime.WithFilter("kind=source,type=chatgpt"), connect_runtime.WithEventHandler(connectorEventHandlerFuncs))
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	go r.Run(ctx)

	// waiting runtime running
	connectors, err := r.Lister().List(labels.Everything())
	if err != nil {
		fmt.Println("failed to list connectors")
		panic(err)
	}
	fmt.Printf("list connectors: %+v\n", connectors)
	key := "vanus/my-connector-name"
	connector, err := r.Lister().Get(key)
	if err != nil {
		fmt.Printf("failed to get connector %s\n", key)
		panic(err)
	}
	fmt.Printf("success to get connector %+v\n", connector)
	<-ctx.Done()
}
