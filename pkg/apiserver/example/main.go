// Copyright 2022 Linkall Inc.
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
