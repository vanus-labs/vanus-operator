package main

import (
	"context"
	"fmt"

	"github.com/vanus-labs/vanus-operator/pkg/runtime"
	"k8s.io/apimachinery/pkg/labels"
)

func main() {
	connectorEventHandlerFuncs := runtime.ConnectorEventHandlerFuncs{
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
	r, err := runtime.New(runtime.WithFilter("kind=source,type=chatgpt"), runtime.WithEventHandler(connectorEventHandlerFuncs))
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
