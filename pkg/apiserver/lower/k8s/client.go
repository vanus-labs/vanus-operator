// Copyright 2017 EasyStack, Inc.
package k8s

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	clientgoscheme.AddToScheme(scheme) //nolint
}

// clientcache: with cache Get/List
type ClientCache struct {
	// reuse interface
	ctrlclient.Client

	// config which create informer client
	config *rest.Config

	// context
	ctx    context.Context
	stopfn func()

	// schema should add before start
	mu          sync.RWMutex
	cacheSchema map[schema.GroupVersionKind]ctrlcache.Cache
}

func NewClientCache(config *rest.Config, se *runtime.Scheme) *ClientCache {
	if se == nil {
		se = scheme
	}
	cli, err := ctrlclient.New(config, ctrlclient.Options{
		Scheme: se,
	})
	if err != nil {
		panic(err)
	}
	ctx, cancle := context.WithCancel(context.Background())
	var once sync.Once
	stopfn := func() {
		once.Do(cancle)
	}
	return &ClientCache{
		stopfn:      stopfn,
		ctx:         ctx,
		Client:      cli,
		config:      config,
		cacheSchema: make(map[schema.GroupVersionKind]ctrlcache.Cache),
	}
}

// must call before Start
func (c *ClientCache) AddObject(runobj ctrlclient.Object, ns string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if runobj == nil {
		return fmt.Errorf("object has no RuntimeObject")
	}
	gvk, err := apiutil.GVKForObject(runobj, scheme)
	if err != nil {
		return err
	}
	if strings.HasSuffix(gvk.Kind, "List") {
		return fmt.Errorf("could not add list RuntimeObject")
	}
	option := ctrlcache.Options{
		Scheme: scheme,
	}
	if ns != "" {
		option.Namespace = ns
	}

	reader, err := ctrlcache.New(c.config, option)
	if err != nil {
		return err
	}
	klog.Infof("add informer (%s) on namespace %s ", gvk.String(), ns)
	go func(gv schema.GroupVersionKind) {
		for {
			err := reader.Start(c.ctx)
			if err != nil {
				klog.Errorf("%s can not sync, wait next", gv.String())
			}
			time.Sleep(time.Second * 3)
		}
	}(gvk)
	c.cacheSchema[gvk] = reader
	return nil
}

func (c *ClientCache) Stop() {
	c.stopfn()
}

func (c *ClientCache) inCache(gvk *schema.GroupVersionKind) (ctrlcache.Cache, bool) {
	news := gvk.String()

	c.mu.RLock()
	defer c.mu.RUnlock()
	for k, v := range c.cacheSchema {
		if k.String() == news {
			return v, true
		}
	}

	return nil, false
}

func (c *ClientCache) Get(ctx context.Context, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return err
	}
	if reader, ok := c.inCache(&gvk); ok {
		if !reader.WaitForCacheSync(ctx) {
			klog.Warningf("%s could not sync, reader from server", gvk.String())
		} else {
			return reader.Get(ctx, key, obj)
		}
	}
	return c.Client.Get(ctx, key, obj)
}

func (c *ClientCache) List(ctx context.Context, obj ctrlclient.ObjectList, opts ...ctrlclient.ListOption) error {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return err
	}
	if strings.HasSuffix(gvk.Kind, "List") {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	if reader, ok := c.inCache(&gvk); ok {
		if !reader.WaitForCacheSync(ctx) {
			klog.Warningf("%s could not sync, reader from server", gvk.String())
		} else {
			return reader.List(ctx, obj, opts...)
		}
	}
	return c.Client.List(ctx, obj, opts...)
}

func buildLabels(la map[string]string) (labels.Selector, error) {
	if la == nil {
		return nil, nil
	}
	se := labels.NewSelector()
	for k, v := range la {
		req, err := labels.NewRequirement(k, selection.Equals, []string{v})
		if err != nil {
			return nil, err
		}
		se.Add(*req)
	}
	return se, nil
}
