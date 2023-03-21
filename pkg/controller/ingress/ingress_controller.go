/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ingress

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	log "k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	networkinginformers "k8s.io/client-go/informers/networking/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	coretypev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	networkinglisters "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/vanus-labs/vanus-operator/pkg/controller"
)

const (
	// BurstReplicas is a rate limiter for booting pods on a lot of pods.
	// The value of 250 is chosen b/c values that are too high can cause registry DoS issues.
	BurstReplicas = 250

	// StatusUpdateRetries limits the number of retries if sending a status update to API server fails.
	StatusUpdateRetries = 1

	// BackoffGCInterval is the time that has to pass before next iteration of backoff GC is run
	BackoffGCInterval = 1 * time.Minute
)

// Reasons for DaemonSet events
const (
	// SelectingAllReason is added to an event when a DaemonSet selects all Pods.
	SelectingAllReason = "SelectingAll"
	// FailedPlacementReason is added to an event when a DaemonSet can't schedule a Pod to a specified node.
	FailedPlacementReason = "FailedPlacement"
	// FailedDaemonPodReason is added to an event when the status of a Pod of a DaemonSet is 'Failed'.
	FailedDaemonPodReason = "FailedDaemonPod"
)

// controllerKind contains the schema.GroupVersionKind for this controller type.
var controllerKind = appsv1.SchemeGroupVersion.WithKind("Ingress")

// DaemonSetsController is responsible for synchronizing DaemonSet objects stored
// in the system with actual running pods.
type IngressController struct {
	kubeClient clientset.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	// An dsc is temporarily suspended after creating/deleting these many replicas.
	// It resumes normal action after observing the watch events for them.
	burstReplicas int

	// To allow injection of syncDaemonSet for testing.
	syncHandler func(ctx context.Context, inKey string) error
	// used for unit testing
	enqueueIngress func(ds *networkingv1.Ingress)
	// A TTLCache of pod creates/deletes each ds expects to see
	// dsLister can list/get daemonsets from the shared informer's store
	inLister networkinglisters.IngressLister
	// dsStoreSynced returns true if the daemonset store has been synced at least once.
	// Added as a member to the struct to allow injection for testing.
	inStoreSynced cache.InformerSynced

	// DaemonSet keys that need to be synced.
	queue workqueue.RateLimitingInterface
}

// NewDaemonSetsController creates a new DaemonSetsController
func NewIngressController(
	ctx context.Context,
	ingressInformer networkinginformers.IngressInformer,
	kubeClient clientset.Interface,
) (*IngressController, error) {
	eventBroadcaster := record.NewBroadcaster()
	inc := &IngressController{
		kubeClient:       kubeClient,
		eventBroadcaster: eventBroadcaster,
		eventRecorder:    eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "ingress-controller"}),
		burstReplicas:    BurstReplicas,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress"),
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

// Run begins watching and syncing daemon sets.
func (inc *IngressController) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()

	inc.eventBroadcaster.StartStructuredLogging(0)
	inc.eventBroadcaster.StartRecordingToSink(&coretypev1.EventSinkImpl{Interface: inc.kubeClient.CoreV1().Events("")})
	defer inc.eventBroadcaster.Shutdown()

	defer inc.queue.ShutDown()

	log.Info("Starting daemon sets controller")
	defer log.Info("Shutting down daemon sets controller")

	if !cache.WaitForNamedCacheSync("daemon sets", ctx.Done(), inc.inStoreSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, inc.runWorker, time.Second)
	}

	<-ctx.Done()
}

func (inc *IngressController) addIngress(obj interface{}) {
	ds := obj.(*networkingv1.Ingress)
	log.Info("Adding daemon set", "daemonset", log.KObj(ds))
	inc.enqueueIngress(ds)
}

func (inc *IngressController) updateIngress(cur, old interface{}) {
	// oldIngress := old.(*networkingv1.Ingress)
	// curIngress := cur.(*networkingv1.Ingress)

	// // TODO: make a KEP and fix informers to always call the delete event handler on re-create
	// if curIngress.UID != oldIngress.UID {
	// 	key, err := controller.KeyFunc(oldIngress)
	// 	if err != nil {
	// 		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", oldIngress, err))
	// 		return
	// 	}
	// 	inc.deleteIngress(logger, cache.DeletedFinalStateUnknown{
	// 		Key: key,
	// 		Obj: oldIngress,
	// 	})
	// }

	// logger.V(4).Info("Updating daemon set", "daemonset", klog.KObj(oldIngress))
	// inc.enqueueIngress(curIngress)
}

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
	log.Info("Deleting ingress", "ingress", log.KObj(ingress))

	key, err := controller.KeyFunc(ingress)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v, err: %s", ingress, err.Error()))
		return
	}

	// 监听到ingress delete事件时，直接将key扔到queue中，然后在queue中延迟创建新的ingress
	inc.queue.Add(key)
}

func (inc *IngressController) syncIngress(ctx context.Context, key string) error {
	defer func() {
		log.Info("Finished syncing ingress", "key", key)
	}()

	log.Info("Start syncing ingress", "key", key)
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	ins, err := inc.inLister.Ingresses(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Ingress has been deleted", "key", key)
			return nil
		}
		return fmt.Errorf("unable to retrieve ingress %v from store: %v", key, err)
	}

	// everything := metav1.LabelSelector{}
	// if reflect.DeepEqual(in.Spec.Selector, &everything) {
	// 	dsc.eventRecorder.Eventf(ds, v1.EventTypeWarning, SelectingAllReason, "This daemon set is selecting all pods. A non-empty selector is required.")
	// 	return nil
	// }

	// Don't process a daemon set until all its creations and deletions have been processed.
	// For example if daemon set foo asked for 3 new daemon pods in the previous call to manage,
	// then we do not want to call manage on foo until the daemon pods have been created.
	// ingressKey, err := controller.KeyFunc(ins)
	// if err != nil {
	// 	return fmt.Errorf("couldn't get key for object %#v: %v", ins, err)
	// }

	// If the DaemonSet is being deleted (either by foreground deletion or
	// orphan deletion), we cannot be sure if the DaemonSet history objects
	// it owned still exist -- those history objects can either be deleted
	// or orphaned. Garbage collector doesn't guarantee that it will delete
	// DaemonSet pods before deleting DaemonSet history objects, because
	// DaemonSet history doesn't own DaemonSet pods. We cannot reliably
	// calculate the status of a DaemonSet being deleted. Therefore, return
	// here without updating status for the DaemonSet being deleted.
	if ins.DeletionTimestamp != nil {
		log.Info("Finish syncing ingress with DeletionTimestamp", "key", key)
		return nil
	}

	// Construct histories of the DaemonSet, and get the hash of current history
	// cur, old, err := inc.constructHistory(ctx, ins)
	// if err != nil {
	// 	return fmt.Errorf("failed to construct revisions of DaemonSet: %v", err)
	// }
	// hash := cur.Labels[appsv1.DefaultDaemonSetUniqueLabelKey]

	// if !dsc.expectations.SatisfiedExpectations(dsKey) {
	// 	// Only update status. Don't raise observedGeneration since controller didn't process object of that generation.
	// 	return dsc.updateDaemonSetStatus(ctx, ds, nodeList, hash, false)
	// }

	// err = inc.updateDaemonSet(ctx, ins, nodeList, hash, ingressKey, old)
	// statusErr := dsc.updateDaemonSetStatus(ctx, ds, nodeList, hash, true)
	// switch {
	// case err != nil && statusErr != nil:
	// 	// If there was an error, and we failed to update status,
	// 	// log it and return the original error.
	// 	logger.Error(statusErr, "Failed to update status", "daemonSet", klog.KObj(ds))
	// 	return err
	// case err != nil:
	// 	return err
	// case statusErr != nil:
	// 	return statusErr
	// }

	log.Info("Finish syncing ingress with end", "key", key)
	return nil
}

func (inc *IngressController) enqueue(in *networkingv1.Ingress) {
	key, err := controller.KeyFunc(in)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", in, err))
		return
	}

	// TODO: Handle overlapping controllers better. See comment in ReplicationManager.
	inc.queue.Add(key)
}

func (inc *IngressController) runWorker(ctx context.Context) {
	for inc.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false when it's time to quit.
func (inc *IngressController) processNextWorkItem(ctx context.Context) bool {
	log.Info("Start process next work item")
	dsKey, quit := inc.queue.Get()
	if quit {
		return false
	}
	defer inc.queue.Done(dsKey)
	log.Infof("processing next work item, key: %+v\n", dsKey)

	err := inc.syncHandler(ctx, dsKey.(string))
	if err == nil {
		inc.queue.Forget(dsKey)
		log.Infof("Success process next work item, key: %+v\n", dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	inc.queue.AddRateLimited(dsKey)
	log.Infof("Failed process next work item, key: %+v, err: %s\n", dsKey, err.Error())

	return true
}
