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

package controller

import (
	"context"
	"time"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/lower/cache"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/lower/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	APIPath            = "apis"
	Group              = vanusv1alpha1.GroupVersion.Group
	Version            = vanusv1alpha1.GroupVersion.Version
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme        = SchemeBuilder.AddToScheme
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&vanusv1alpha1.Core{},
		&vanusv1alpha1.CoreList{},
		&vanusv1alpha1.Connector{},
		&vanusv1alpha1.ConnectorList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

type Controller struct {
	config *rest.Config

	// k8s client and cache
	clientset rest.Interface
	client    *k8s.ClientCache

	timeout     time.Duration
	singleCache *cache.Cache
	stopCh      chan struct{}

	ctx context.Context
}

func DefaultRestConfig(config *rest.Config) {
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: Group, Version: Version}
	config.APIPath = APIPath
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()
}

func RESTClientForConfigOrDie(config *rest.Config) *rest.RESTClient {
	clientset, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}
	return clientset
}

func New(config *rest.Config) *Controller {
	DefaultRestConfig(config)
	AddToScheme(scheme.Scheme)
	return &Controller{
		config:      config,
		timeout:     time.Second * 10,
		clientset:   RESTClientForConfigOrDie(config),
		client:      k8s.NewClientCache(config, nil),
		stopCh:      make(chan struct{}),
		singleCache: cache.NewCache(time.Second * 5),
		ctx:         context.Background(),
	}
}

func (c *Controller) Client() ctrlclient.Client {
	return c.client.Client
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

func (c *Controller) ClientSet() rest.Interface {
	return c.clientset
}
