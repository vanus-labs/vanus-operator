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

package runtime

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	log "k8s.io/klog/v2"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/pkg/apis/vanus/v1alpha1"
)

func (r *runtime) enqueueAddConnector(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	log.Infof("enqueue add connector %s", key)
	r.addConnectorQueue.Add(key)
}

func (r *runtime) enqueueUpdateConnector(old, new interface{}) {
	oldKey, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	newKey, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	log.Infof("old connector: %s\n", oldKey)
	log.Infof("new connector: %s\n", newKey)
	r.updateConnectorQueue.Add(newKey)
}

func (r *runtime) enqueueDeleteConnector(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	log.Infof("enqueue delete connector %s\n", key)
	r.deleteConnectorQueue.Add(obj)
}

func (r *runtime) runAddConnectorWorker() {
	for r.processNextAddConnectorWorkItem() {
	}
}

func (r *runtime) runUpdateConnectorWorker() {
	for r.processNextUpdateConnectorWorkItem() {
	}
}

func (r *runtime) runDeleteConnectorWorker() {
	for r.processNextDeleteConnectorWorkItem() {
	}
}

func (r *runtime) processNextAddConnectorWorkItem() bool {
	obj, shutdown := r.addConnectorQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer r.addConnectorQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			r.addConnectorQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := r.handleAddConnector(key); err != nil {
			r.addConnectorQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		r.addConnectorQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (r *runtime) processNextUpdateConnectorWorkItem() bool {
	obj, shutdown := r.updateConnectorQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer r.updateConnectorQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			r.updateConnectorQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := r.handleUpdateConnector(key); err != nil {
			r.updateConnectorQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		r.updateConnectorQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (r *runtime) processNextDeleteConnectorWorkItem() bool {
	obj, shutdown := r.deleteConnectorQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer r.deleteConnectorQueue.Done(obj)
		var connector *vanusv1alpha1.Connector
		var ok bool
		if connector, ok = obj.(*vanusv1alpha1.Connector); !ok {
			r.deleteConnectorQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected connector in workqueue but got %#v", obj))
			return nil
		}
		if err := r.handleDeleteConnector(connector); err != nil {
			r.deleteConnectorQueue.AddRateLimited(obj)
			return fmt.Errorf("error syncing '%s': %s, requeuing", connector.Name, err.Error())
		}
		r.deleteConnectorQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (r *runtime) handleAddConnector(key string) error {
	var err error
	cachedConnector, err := r.connectorsLister.Get(key)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	log.Infof("handle add connector %s", cachedConnector.Name)
	err = r.handler.OnAdd(cachedConnector.Name, cachedConnector.Spec.Config)
	if err != nil {
		log.Errorf("handle add connector %s failed: %+v", cachedConnector.Name, err)
		return err
	}
	return nil
}

func (r *runtime) handleUpdateConnector(key string) error {
	var err error
	cachedConnector, err := r.connectorsLister.Get(key)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	log.Infof("handle update connector %s", cachedConnector.Name)
	err = r.handler.OnUpdate(cachedConnector.Name, cachedConnector.Spec.Config)
	if err != nil {
		log.Errorf("handle update connector %s failed: %+v", cachedConnector.Name, err)
		return err
	}
	return nil
}

func (r *runtime) handleDeleteConnector(connector *vanusv1alpha1.Connector) error {
	log.Infof("handle delete connector %s", connector.Name)
	err := r.handler.OnDelete(connector.Name)
	if err != nil {
		log.Errorf("handle delete connector %s failed: %+v", connector.Name, err)
		return err
	}
	return nil
}
