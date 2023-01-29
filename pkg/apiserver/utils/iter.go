// Copyright 2017 EasyStack, Inc.
package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	IterFactoryPool = sync.Pool{
		New: func() interface{} {
			return NewIterFactory()
		},
	}
)

func GetIter() *IterFactory {
	factory := IterFactoryPool.New().(*IterFactory)
	factory.Reset()
	return factory
}

func PutIter(i *IterFactory) {
	IterFactoryPool.Put(i)
}

type IterFactory struct {
	options []ctrlclient.ListOption

	object ctrlclient.ObjectList
	ctx    context.Context

	processFn func(runtime.Object) error
}

func NewIterFactory() *IterFactory {
	return &IterFactory{
		ctx: context.Background(),
	}
}

func (i *IterFactory) Reset() *IterFactory {
	i.options = i.options[:0]
	i.processFn = nil
	i.object = nil
	return i
}

func (i *IterFactory) Namespace(ns string) *IterFactory {
	i.options = append(i.options, ctrlclient.InNamespace(ns))
	return i
}

func (i *IterFactory) Labels(labels map[string]string) *IterFactory {
	if labels != nil {
		i.options = append(i.options, ctrlclient.MatchingLabels(labels))
	} else {

	}
	return i
}

func (i *IterFactory) Fiedls(fields map[string]string) *IterFactory {
	if fields != nil {
		i.options = append(i.options, ctrlclient.MatchingFields(fields))
	}
	return i
}

func (i *IterFactory) Object(object ctrlclient.ObjectList) *IterFactory {
	i.object = object
	return i
}

func (i *IterFactory) Fn(fn func(runtime.Object) error) *IterFactory {
	i.processFn = fn
	return i
}

func (i *IterFactory) preDoCheck() error {
	if i.object == nil {
		return fmt.Errorf("factory not set listObject")
	}
	if i.processFn == nil {
		return fmt.Errorf("factory not set process function")
	}
	return nil
}

func (i *IterFactory) Do(cli ctrlclient.Client, timeout time.Duration) error {
	var (
		err error
	)
	err = i.preDoCheck()
	if err != nil {
		return err
	}
	ctx, cancle := context.WithTimeout(i.ctx, timeout)
	defer cancle()

	err = cli.List(ctx, i.object, i.options...)
	if err != nil {
		return err
	}

	switch v := i.object.(type) {
	case *appsv1.DeploymentList:
		klog.V(3).Infof("deploy object length: %v", len(v.Items))
		for _, value := range v.Items {
			err = i.processFn(&value)
			if err != nil {
				return err
			}
		}
	case *appsv1.ReplicaSetList:
		for _, value := range v.Items {
			err = i.processFn(&value)
			if err != nil {
				return err
			}
		}
	case *corev1.EventList:
		for _, value := range v.Items {
			err = i.processFn(&value)
			if err != nil {
				return err
			}
		}
	case *corev1.NamespaceList:
		for _, value := range v.Items {
			err = i.processFn(&value)
			if err != nil {
				return err
			}
		}
	case *corev1.PodList:
		for _, value := range v.Items {
			err = i.processFn(&value)
			if err != nil {
				return err
			}
		}
	case *corev1.PersistentVolumeClaimList:
		for _, pvc := range v.Items {
			err = i.processFn(&pvc)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("please add more type in this check")
	}
	return nil
}
