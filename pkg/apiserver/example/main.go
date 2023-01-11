// Copyright 2017 EasyStack, Inc.
package main

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func main() {
	ll := &appsv1.DeploymentList{}
	l := &appsv1.Deployment{}
	gvk, err := apiutil.GVKForObject(ll, scheme.Scheme)
	if err != nil {
		fmt.Println(err)
		return
	}
	klog.Infof("list gvk: %v", gvk)

	gvk, err = apiutil.GVKForObject(l, scheme.Scheme)
	if err != nil {
		fmt.Println(err)
		return
	}
	klog.Infof("list gvk: %v", gvk)

}
