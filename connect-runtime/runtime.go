// Copyright 2023 Linkall Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connect_runtime

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	log "k8s.io/klog/v2"

	vanusinformer "github.com/vanus-labs/connect-runtime/client/informers/externalversions"
	vanuslister "github.com/vanus-labs/connect-runtime/client/listers/vanus/v1alpha1"
)

type Runtime interface {
	Run(ctx context.Context)
	Lister() vanuslister.ConnectorLister
}

type runtime struct {
	connectorsLister     vanuslister.ConnectorLister
	connectorSynced      cache.InformerSynced
	addConnectorQueue    workqueue.RateLimitingInterface
	updateConnectorQueue workqueue.RateLimitingInterface
	deleteConnectorQueue workqueue.RateLimitingInterface
	vanusInformerFactory vanusinformer.SharedInformerFactory

	handler ConnectorEventHandler
}

// NewController creates a new Controller manager
func New(opts ...ConnectorOption) (Runtime, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, err
	}

	defaultOpts := defaultConnectorOptions()
	for _, apply := range opts {
		apply(&defaultOpts)
	}

	vanusInformerFactory := vanusinformer.NewSharedInformerFactoryWithOptions(config.VanusFactoryClient, 0,
		vanusinformer.WithTweakListOptions(func(listOption *metav1.ListOptions) {
			listOption.AllowWatchBookmarks = true
			listOption.LabelSelector = defaultOpts.labelSelector
		}))

	connectorInformer := vanusInformerFactory.Vanus().V1alpha1().Connectors()
	r := &runtime{
		connectorsLister:     connectorInformer.Lister(),
		connectorSynced:      connectorInformer.Informer().HasSynced,
		addConnectorQueue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "AddConnector"),
		updateConnectorQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "UpdateConnector"),
		deleteConnectorQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "DeleteConnector"),
		vanusInformerFactory: vanusInformerFactory,
		handler:              defaultOpts.handler,
	}

	if _, err = connectorInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.enqueueAddConnector,
		UpdateFunc: r.enqueueUpdateConnector,
		DeleteFunc: r.enqueueDeleteConnector,
	}); err != nil {
		log.Errorf("failed to add connector event handler: %+v\n", err)
		return nil, err
	}

	return r, nil
}

// Run begins runtime.
func (r *runtime) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()
	defer r.shutdown()

	log.Info("Starting controller manager")
	defer log.Info("Shutting down controller manager")

	// Wait for the caches to be synced before starting workers
	r.vanusInformerFactory.Start(ctx.Done())

	log.Info("Waiting for informer caches to sync")
	cacheSyncs := []cache.InformerSynced{
		r.connectorSynced,
	}

	if ok := cache.WaitForCacheSync(ctx.Done(), cacheSyncs...); !ok {
		log.Fatal("failed to wait for caches to sync")
	}

	// start workers to do all the connectors operations
	r.startWorkers(ctx)
	<-ctx.Done()
	log.Info("Shutting down workers")
}

func (r *runtime) Lister() vanuslister.ConnectorLister {
	return r.connectorsLister
}

func (r *runtime) startWorkers(ctx context.Context) {
	log.Info("Starting workers")

	go wait.Until(r.runAddConnectorWorker, time.Second, ctx.Done())
	go wait.Until(r.runUpdateConnectorWorker, time.Second, ctx.Done())
	go wait.Until(r.runDeleteConnectorWorker, time.Second, ctx.Done())
}

func (r *runtime) shutdown() {
	r.addConnectorQueue.ShutDown()
	r.updateConnectorQueue.ShutDown()
	r.deleteConnectorQueue.ShutDown()
}
