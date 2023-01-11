// Copyright 2017 EasyStack, Inc.
package controller

import (
	"context"
	"time"

	"github.com/linkall-labs/vanus-operator/pkg/apiserver/lower/cache"
	"github.com/linkall-labs/vanus-operator/pkg/apiserver/lower/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	config *rest.Config

	// k8s client and cache
	clientset *kubernetes.Clientset
	client    *k8s.ClientCache

	timeout     time.Duration
	singleCache *cache.Cache
	stopCh      chan struct{}

	ctx context.Context
}

func New(config *rest.Config) *Controller {
	return &Controller{
		config:      config,
		timeout:     time.Second * 10,
		clientset:   kubernetes.NewForConfigOrDie(config),
		client:      k8s.NewClientCache(config, nil),
		stopCh:      make(chan struct{}),
		singleCache: cache.NewCache(time.Second * 5),
		ctx:         context.Background(),
	}
}

func (c *Controller) Client() ctrlclient.Client {
	return c.client
}

func (c *Controller) List(obj ctrlclient.ObjectList, opt ctrlclient.ListOption) error {
	ctx, cancle := context.WithTimeout(c.ctx, c.timeout)
	defer cancle()
	return c.client.List(ctx, obj, opt)
}

func (c *Controller) Get(key ctrlclient.ObjectKey, obj ctrlclient.Object) error {
	ctx, cancle := context.WithTimeout(c.ctx, c.timeout)
	defer cancle()
	return c.client.Get(ctx, key, obj)
}

func (c *Controller) Create(obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
	ctx, cancle := context.WithTimeout(c.ctx, c.timeout)
	defer cancle()
	return c.client.Create(ctx, obj, opts...)
}

func (c *Controller) Delete(obj ctrlclient.Object, opts ...ctrlclient.DeleteOption) error {
	ctx, cancle := context.WithTimeout(c.ctx, c.timeout)
	defer cancle()
	return c.client.Delete(ctx, obj, opts...)
}

func (c *Controller) Patch(obj ctrlclient.Object, patch ctrlclient.Patch, opts ...ctrlclient.PatchOption) error {
	ctx, cancle := context.WithTimeout(c.ctx, c.timeout)
	defer cancle()
	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c *Controller) Update(obj ctrlclient.Object, opts ...ctrlclient.UpdateOption) error {
	ctx, cancle := context.WithTimeout(c.ctx, c.timeout)
	defer cancle()
	return c.client.Update(ctx, obj, opts...)
}

func (c *Controller) ClientSet() *kubernetes.Clientset {
	return c.clientset
}
