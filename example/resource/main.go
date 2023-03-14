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

package main

import (
	"context"
	"fmt"
	"log"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func generateVanus() *vanusv1alpha1.Core {
	requests := make(map[corev1.ResourceName]resource.Quantity)
	requests[corev1.ResourceStorage] = resource.MustParse("1Gi")
	vanus := &vanusv1alpha1.Core{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "vanus-controller-jk",
		},
		Spec: vanusv1alpha1.CoreSpec{
			ImagePullPolicy: corev1.PullIfNotPresent,
			Resources:       corev1.ResourceRequirements{},
		},
	}
	return vanus
}

func NewForConfig(c *rest.Config) (rest.Interface, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: "vanus.ai", Version: "v1alpha1"}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func Create(client rest.Interface, controller *vanusv1alpha1.Core) (*vanusv1alpha1.Core, error) {
	result := vanusv1alpha1.Core{}
	err := client.
		Post().
		Namespace("default").
		Resource("controllers").
		Body(controller).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func List(client rest.Interface, opts metav1.ListOptions) (*vanusv1alpha1.CoreList, error) {
	result := vanusv1alpha1.CoreList{}
	err := client.
		Get().
		Namespace("default").
		Resource("controllers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func Get(client rest.Interface, opts metav1.GetOptions) (*vanusv1alpha1.Core, error) {
	result := vanusv1alpha1.Core{}
	err := client.
		Get().
		Resource("controllers").
		Namespace("default").
		Name("vanus-controller-jk").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func GetNotExist(client rest.Interface, opts metav1.GetOptions) (*vanusv1alpha1.Core, error) {
	result := vanusv1alpha1.Core{}
	err := client.
		Get().
		Resource("controllers").
		Namespace("default").
		Name("vanus-controller-not-exist").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func Delete(client rest.Interface, opts metav1.DeleteOptions) (*vanusv1alpha1.Core, error) {
	result := vanusv1alpha1.Core{}
	err := client.
		Delete().
		Resource("controllers").
		Namespace("default").
		Name("vanus-controller-jk").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func DeleteNotExist(client rest.Interface, opts metav1.DeleteOptions) (*vanusv1alpha1.Core, error) {
	result := vanusv1alpha1.Core{}
	err := client.
		Delete().
		Resource("controllers").
		Namespace("default").
		Name("vanus-controller-not-exist").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

const GroupName = "vanus.ai"
const GroupVersion = "v1alpha1"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: GroupVersion}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&vanusv1alpha1.Core{},
		&vanusv1alpha1.CoreList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

func main() {
	var config *rest.Config
	var err error
	kubeconfig := "/home/ubuntu/.kube/config"
	log.Printf("using configuration from '%s'", kubeconfig)
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	AddToScheme(scheme.Scheme)

	clientSet, err := NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// 1. create
	fmt.Println("==================create=====================")
	newcontroller := generateVanus()
	result1, err := Create(clientSet, newcontroller)
	if err != nil {
		fmt.Printf("create cr failed, err: %s\n", err.Error())
		panic(err)
	}
	fmt.Printf("create cr success, result: %+v\n", result1)

	// 2. list
	fmt.Println("==================list=====================")
	result2, err := List(clientSet, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("list cr failed, err: %s\n", err.Error())
		panic(err)
	}
	fmt.Printf("list cr success, result: %+v\n", result2)

	// 3. get
	fmt.Println("==================get=====================")
	result3, err := Get(clientSet, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("get crd failed, err: %+v\n", err)
		return
	}
	fmt.Printf("get crd success, ret: %+v\n", result3)

	// 3. get
	fmt.Println("==================get=====================")
	result31, err := GetNotExist(clientSet, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Print("===Error is Not Found===")
			return
		}
		fmt.Printf("get crd failed, err: %+v\n", err)
		return
	}
	fmt.Printf("get crd success, ret: %+v\n", result31)

	// 4. delete
	fmt.Println("==================delete=====================")
	result4, err := Delete(clientSet, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("delete crd failed, err: %+v\n", err)
		return
	}
	fmt.Printf("delete crd success, ret: %+v\n", result4)

	// 4. delete
	fmt.Println("==================delete=====================")
	result41, err := DeleteNotExist(clientSet, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("delete crd failed, err: %+v\n", err)
		return
	}
	fmt.Printf("delete crd success, ret: %+v\n", result41)

}
