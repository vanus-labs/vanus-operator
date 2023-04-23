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
	"fmt"
	"reflect"
	"strings"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	networkinginformers "k8s.io/client-go/informers/networking/v1"
	networkinglisters "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	log "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/scheme"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/controller"
)

// ConnectorController is responsible for synchronizing Ingress objects stored
// in the system with actual running pods.
type ConnectorController struct {
	ctrl           *controller.Controller
	informer       networkinginformers.IngressInformer
	syncHandler    func(ctx context.Context, inKey string) error
	enqueueIngress func(ds *networkingv1.Ingress)
	inLister       networkinglisters.IngressLister
	inStoreSynced  cache.InformerSynced
	queue          workqueue.RateLimitingInterface
}

// NewConnectorController creates a new ConnectorController
func NewConnectorController(ctx context.Context, sharedInformers informers.SharedInformerFactory, ctrl *controller.Controller) (*ConnectorController, error) {
	ctrl.K8SClientSet().
	ingressInformer := sharedInformers.Networking().V1().Ingresses()
	inc := &ConnectorController{
		ctrl:     ctrl,
		informer: ingressInformer,
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress"),
	}

	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			inc.addIngress(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			inc.updateIngress(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			inc.deleteIngress(obj)
		},
	})

	inc.inLister = ingressInformer.Lister()
	inc.inStoreSynced = ingressInformer.Informer().HasSynced

	inc.syncHandler = inc.syncIngress
	inc.enqueueIngress = inc.enqueue

	return inc, nil
}

// Run begins watching and syncing ingress.
func (inc *ConnectorController) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()
	defer inc.queue.ShutDown()

	log.Info("Starting Ingress controller")
	defer log.Info("Shutting down Ingress controller")

	go inc.informer.Informer().Run(ctx.Done())

	if !cache.WaitForNamedCacheSync("Ingress", ctx.Done(), inc.inStoreSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, inc.runWorker, time.Second)
	}

	<-ctx.Done()
}

func (inc *ConnectorController) addIngress(obj interface{}) {
	ingress := obj.(*networkingv1.Ingress)
	if !isBuiltInIngress(ingress) {
		return
	}
	log.Infof("Adding ingress: %+v\n", log.KObj(ingress))
	inc.enqueueIngress(ingress)
}

func (inc *ConnectorController) updateIngress(old, new interface{}) {
	oldIngress := old.(*networkingv1.Ingress)
	newIngress := new.(*networkingv1.Ingress)
	if !isBuiltInIngress(newIngress) {
		return
	}
	if reflect.DeepEqual(old, new) {
		// log.Infof("No changes on ingress. Skipping update, ingress: %+v\n", log.KObj(newIngress))
		return
	}
	log.Infof("Updating ingress, old: %+v, new: %+v\n", log.KObj(oldIngress), log.KObj(newIngress))
	inc.syncIngress(context.Background(), fmt.Sprintf("%s/%s", newIngress.Namespace, newIngress.Name))
}

func (inc *ConnectorController) deleteIngress(obj interface{}) {
	ingress, ok := obj.(*networkingv1.Ingress)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		ingress, ok = tombstone.Obj.(*networkingv1.Ingress)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a Ingress %#v", obj))
			return
		}
	}
	if !isBuiltInIngress(ingress) {
		return
	}
	log.Infof("Deleting ingress: %+v\n", log.KObj(ingress))
	// generate new ingress
	newIngress := rebuildIngress(ingress)
	_, err := inc.ctrl.K8SClientSet().NetworkingV1().Ingresses(ingress.Namespace).Create(context.Background(), newIngress, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("rebuild ingress failed, err: %s\n", err.Error())
		utilruntime.HandleError(fmt.Errorf("couldn't rebuild ingress object %+v, err: %s", log.KObj(ingress), err.Error()))
		return
	}
}

func (inc *ConnectorController) syncIngress(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	ingress, err := inc.inLister.Ingresses(namespace).Get(name)
	if err != nil {
		log.Errorf("failed to get ingress %s, err: %s\n", key, err.Error())
		return err
	}
	if !isBuiltInIngress(ingress) {
		return nil
	}
	log.Infof("Sync ingress: %+v\n", log.KObj(ingress))
	return inc.diffAndUpdate(ingress)
}

func (inc *ConnectorController) enqueue(in *networkingv1.Ingress) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(in)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %s: %v", in.Name, err))
		return
	}
	inc.queue.Add(key)
}

func (inc *ConnectorController) runWorker(ctx context.Context) {
	for inc.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false when it's time to quit.
func (inc *ConnectorController) processNextWorkItem(ctx context.Context) bool {
	insKey, quit := inc.queue.Get()
	if quit {
		log.Error("quit process next work item")
		return false
	}
	defer inc.queue.Done(insKey)

	err := inc.syncHandler(ctx, insKey.(string))
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Infof("Ingress not found. Ignoring since object must be deleted., key: %s\n", insKey)
			return true
		}
		log.Warningf("Sync handler failed, rejoin delay queue, key: %s\n", insKey)
		inc.queue.AddAfter(insKey, time.Minute)
		return true
	}
	// rejoin delay queue
	inc.queue.AddAfter(insKey, time.Hour)
	return true
}

func (inc *ConnectorController) diffAndUpdate(ingress *networkingv1.Ingress) error {
	needUpdate := false
	newRules := make([]networkingv1.IngressRule, 0)
	// newRules = append(newRules, defaultIngressRule())
	for _, rule := range ingress.Spec.Rules {
		if rule.Host == "" {
			needUpdate = true
			newRules = append(newRules, rule)
			continue
		}
		if strings.HasPrefix(rule.Host, cons.DefaultVanusOperatorHostPrefix) {
			continue
		}
		// get connector from rule host
		result := &vanusv1alpha1.ConnectorList{}
		controller.AddToScheme(scheme.Scheme)
		err := inc.ctrl.ClientSet().
			Get().
			Resource("connectors").
			Namespace(cons.DefaultNamespace).
			VersionedParams(&metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", cons.ConnectorNetworkHostDomainAnnotation, rule.Host)}, scheme.ParameterCodec).
			Do(context.TODO()).
			Into(result)
		if err != nil {
			log.Errorf("list connectors failed, err: %s\n", err.Error())
			return fmt.Errorf("unable to list connectors, err: %s", err)
		}
		connectorsNumber := len(result.Items)
		if connectorsNumber == 0 {
			// connector has been deleted, need update rule
			needUpdate = true
		} else if connectorsNumber > 1 {
			log.Warningf("there are %d connectors with the same network host domain\n", connectorsNumber)
		} else {
			newRules = append(newRules, rule)
		}
	}

	if needUpdate {
		ingress.Spec.Rules = newRules
		_, err := inc.ctrl.K8SClientSet().NetworkingV1().Ingresses(cons.DefaultNamespace).Update(context.TODO(), ingress, metav1.UpdateOptions{})
		if err != nil {
			log.Errorf("update ingress failed, err: %s\n", err.Error())
			return err
		}
	}
	return nil
}
