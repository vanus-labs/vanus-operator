// Copyright 2017 EasyStack, Inc.
package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// (TODO) deprecated
type Cache struct {
	kubernetes.Interface

	stopCh       chan struct{}
	podsLister   corev1.PodLister
	pvcLister    corev1.PersistentVolumeClaimLister
	pvLister     corev1.PersistentVolumeLister
	deployLister appsv1.DeploymentLister
}

func New(clientSet kubernetes.Interface) *Cache {
	sharedInformer := informers.NewSharedInformerFactoryWithOptions(clientSet, 0,
		informers.WithTweakListOptions(func(lo *metav1.ListOptions) {
			lo.AllowWatchBookmarks = true
		}))

	podInformer := sharedInformer.Core().V1().Pods()
	pvcInfomer := sharedInformer.Core().V1().PersistentVolumeClaims()
	pvInfomer := sharedInformer.Core().V1().PersistentVolumes()
	deployInformer := sharedInformer.Apps().V1().Deployments()

	c := &Cache{
		stopCh:       make(chan struct{}),
		Interface:    clientSet,
		podsLister:   podInformer.Lister(),
		pvcLister:    pvcInfomer.Lister(),
		pvLister:     pvInfomer.Lister(),
		deployLister: deployInformer.Lister(),
	}

	sharedInformer.Start(c.stopCh)

	cacheSyncs := []cache.InformerSynced{
		podInformer.Informer().HasSynced,
		pvcInfomer.Informer().HasSynced,
		pvInfomer.Informer().HasSynced,
		deployInformer.Informer().HasSynced,
	}
	if ok := cache.WaitForCacheSync(c.stopCh, cacheSyncs...); !ok {
		klog.Fatal("informer sync failed")
	}

	return c
}

func (c *Cache) Stop() {
	close(c.stopCh)
}

func (c *Cache) PodCache() corev1.PodLister {
	return c.podsLister
}

func (c *Cache) PVCCache() corev1.PersistentVolumeClaimLister {
	return c.pvcLister
}

func (c *Cache) PVCache() corev1.PersistentVolumeLister {
	return c.pvLister
}

func (c *Cache) DeploymentCache() appsv1.DeploymentLister {
	return c.deployLister
}
