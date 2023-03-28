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

package ingress

import (
	"context"
	"fmt"
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
	"github.com/vanus-labs/vanus-operator/internal/convert"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/controller"
)

// IngressController is responsible for synchronizing Ingress objects stored
// in the system with actual running pods.
type IngressController struct {
	ctrl           *controller.Controller
	informer       networkinginformers.IngressInformer
	syncHandler    func(ctx context.Context, inKey string) error
	enqueueIngress func(ds *networkingv1.Ingress)
	inLister       networkinglisters.IngressLister
	inStoreSynced  cache.InformerSynced
	queue          workqueue.RateLimitingInterface
}

// NewIngressController creates a new IngressController
func NewIngressController(ctx context.Context, ctrl *controller.Controller) (*IngressController, error) {
	sharedInformers := informers.NewSharedInformerFactory(ctrl.K8SClientSet(), time.Minute)
	ingressInformer := sharedInformers.Networking().V1().Ingresses()
	inc := &IngressController{
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
func (inc *IngressController) Run(ctx context.Context, workers int) {
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

func (inc *IngressController) addIngress(obj interface{}) {
	ingress := obj.(*networkingv1.Ingress)
	if !isOperatorIngress(ingress.Namespace, ingress.Name) {
		return
	}
	log.Infof("Adding ingress: %+v\n", log.KObj(ingress))
	inc.enqueueIngress(ingress)
}

func (inc *IngressController) updateIngress(cur, old interface{}) {}

func (inc *IngressController) deleteIngress(obj interface{}) {
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
	if !isOperatorIngress(ingress.Namespace, ingress.Name) {
		return
	}
	log.Infof("Deleting ingress: %+v\n", ingress)

	// generate new ingress
	newIngress := rebuildIngress(ingress)
	_, err := inc.ctrl.K8SClientSet().NetworkingV1().Ingresses(ingress.Namespace).Create(context.Background(), newIngress, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("rebuild ingress failed, err: %s\n", err.Error())
		utilruntime.HandleError(fmt.Errorf("couldn't rebuild ingress object %+v, err: %s", ingress, err.Error()))
		return
	}
}

func (inc *IngressController) syncIngress(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	if !isOperatorIngress(namespace, name) {
		return nil
	}
	ins, err := inc.inLister.Ingresses(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Infof("Ingress has been deleted, key: %s\n", key)
			return nil
		}
		return fmt.Errorf("unable to retrieve ingress %s from store, err: %+v", key, err)
	}
	return inc.diffAndUpdate(ins)
}

func (inc *IngressController) enqueue(in *networkingv1.Ingress) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(in)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", in, err))
		return
	}
	inc.queue.Add(key)
}

func (inc *IngressController) runWorker(ctx context.Context) {
	for inc.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false when it's time to quit.
func (inc *IngressController) processNextWorkItem(ctx context.Context) bool {
	insKey, quit := inc.queue.Get()
	if quit {
		log.Error("quit process next work item")
		return false
	}
	defer inc.queue.Done(insKey)

	err := inc.syncHandler(ctx, insKey.(string))
	if err != nil {
		// check whether the ingress exists, if not, create a new ingress
		namespace, name, err := cache.SplitMetaNamespaceKey(insKey.(string))
		if err != nil {
			inc.queue.AddAfter(insKey, time.Minute)
			return true
		}
		_, err = inc.ctrl.K8SClientSet().NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				newIngress := newIngress()
				_, err := inc.ctrl.K8SClientSet().NetworkingV1().Ingresses(newIngress.Namespace).Create(context.Background(), newIngress, metav1.CreateOptions{})
				if err != nil {
					log.Errorf("create new ingress failed, err: %s\n", err.Error())
					inc.queue.AddAfter(insKey, time.Minute)
					return true
				}
				log.Infof("create new ingress success, key: %s\n", insKey)
				inc.queue.Add(insKey)
				return true
			}
		}
		inc.queue.AddAfter(insKey, time.Minute)
		return true
	}
	// rejoin delay queue
	inc.queue.AddAfter(insKey, time.Hour)
	return true
}

func (inc *IngressController) diffAndUpdate(ins *networkingv1.Ingress) error {
	// 1. list connectors
	result := &vanusv1alpha1.ConnectorList{}
	controller.AddToScheme(scheme.Scheme)
	err := inc.ctrl.ClientSet().
		Get().
		Resource("connectors").
		Namespace(cons.DefaultNamespace).
		VersionedParams(&metav1.ListOptions{}, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		log.Errorf("list connectors failed, err: %s\n", err.Error())
		return fmt.Errorf("unable to list connectors, err: %s", err)
	}

	// 2. generate connector network host map, format: map[host]networkingv1.IngressBackend
	connectorMap := make(map[string]networkingv1.IngressBackend)
	for _, connector := range result.Items {
		if host, ok := connector.Annotations[cons.ConnectorNetworkHostDomainAnnotation]; ok {
			var svcPort int32
			if _, ok := connector.Annotations[cons.ConnectorServicePortAnnotation]; ok {
				svcPort, _ = convert.StrToInt32(connector.Annotations[cons.ConnectorServicePortAnnotation])
			} else {
				svcPort = int32(cons.DefaultConnectorServicePort)
			}
			backend := networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: connector.Name,
					Port: networkingv1.ServiceBackendPort{
						Number: svcPort,
					},
				},
			}
			connectorMap[host] = backend
		}
	}

	// 3. generate ingress host map, format: map[host]networkingv1.IngressBackend
	ingressMap := make(map[string]networkingv1.IngressBackend)
	for idx, rule := range ins.Spec.Rules {
		if rule.Host == cons.DefaultVanusOperatorHost {
			continue
		}
		ingressMap[rule.Host] = ins.Spec.Rules[idx].IngressRuleValue.HTTP.Paths[0].Backend
	}

	// 4. diff connectorMap and ingressMap
	needUpdate := false
	newRules := make([]networkingv1.IngressRule, 0)
	newRules = append(newRules, defaultIngressRule())
	for host, backend := range connectorMap {
		if _, ok := ingressMap[host]; !ok {
			needUpdate = true
		}
		prefix := networkingv1.PathTypePrefix
		paths := make([]networkingv1.HTTPIngressPath, 0)
		paths = append(paths, networkingv1.HTTPIngressPath{
			Path:     "/",
			PathType: &prefix,
			Backend:  backend,
		})
		newRules = append(newRules, networkingv1.IngressRule{
			Host: host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		})
	}

	for host := range ingressMap {
		if _, ok := connectorMap[host]; !ok {
			needUpdate = true
		}
	}

	if needUpdate {
		ins.Spec.Rules = newRules
		_, err = inc.ctrl.K8SClientSet().NetworkingV1().Ingresses(cons.DefaultNamespace).Update(context.TODO(), ins, metav1.UpdateOptions{})
		if err != nil {
			log.Errorf("update ingress failed, err: %s\n", err.Error())
			return err
		}
	}
	return nil
}

func rebuildIngress(ingress *networkingv1.Ingress) *networkingv1.Ingress {
	newIngress := new(networkingv1.Ingress)
	newIngress.Name = ingress.Name
	newIngress.Namespace = ingress.Namespace
	newIngress.Annotations = ingress.Annotations
	newIngress.Spec.Rules = ingress.Spec.Rules
	return newIngress
}

func newIngress() *networkingv1.Ingress {
	newIngress := new(networkingv1.Ingress)
	newIngress.Name = cons.DefaultVanusOperatorName
	newIngress.Namespace = cons.DefaultNamespace
	annotations := make(map[string]string)
	annotations[cons.DefaultIngressClassAnnotationKey] = cons.DefaultIngressClassAnnotationValue
	newIngress.Annotations = annotations
	prefix := networkingv1.PathTypePrefix
	paths := make([]networkingv1.HTTPIngressPath, 0)
	paths = append(paths, networkingv1.HTTPIngressPath{
		Path:     "/",
		PathType: &prefix,
		Backend: networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: cons.DefaultVanusOperatorName,
				Port: networkingv1.ServiceBackendPort{
					Number: cons.DefaultOperatorContainerPortApi,
				},
			},
		},
	})
	rules := make([]networkingv1.IngressRule, 0)
	rules = append(rules, networkingv1.IngressRule{
		Host: cons.DefaultVanusOperatorHost,
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: paths,
			},
		},
	})
	newIngress.Spec.Rules = rules
	return newIngress
}

func defaultIngressRule() networkingv1.IngressRule {
	prefix := networkingv1.PathTypePrefix
	paths := make([]networkingv1.HTTPIngressPath, 0)
	paths = append(paths, networkingv1.HTTPIngressPath{
		Path:     "/",
		PathType: &prefix,
		Backend: networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: cons.DefaultVanusOperatorName,
				Port: networkingv1.ServiceBackendPort{
					Number: cons.DefaultOperatorContainerPortApi,
				},
			},
		},
	})
	rule := networkingv1.IngressRule{
		Host: cons.DefaultVanusOperatorHost,
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: paths,
			},
		},
	}
	return rule
}

func isOperatorIngress(namespace, name string) bool {
	return (namespace == cons.DefaultNamespace && name == cons.DefaultVanusOperatorName)
}
