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

package handlers

import (
	"context"
	"encoding/json"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	ResourceCore      = "cores"
	ResourceConnector = "connectors"
)

func (a *Api) createCore(vanus *vanusv1alpha1.Core, namespace string) (*vanusv1alpha1.Core, error) {
	existCore, exist, err := a.existCore(namespace, vanus.Name, &metav1.GetOptions{})
	if err != nil {
		return existCore, err
	}
	if exist {
		return existCore, nil
	}
	result := &vanusv1alpha1.Core{}
	err = a.ctrl.ClientSet().
		Post().
		Namespace(namespace).
		Resource(ResourceCore).
		Body(vanus).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) patchCore(vanus *vanusv1alpha1.Core) (*vanusv1alpha1.Core, error) {
	result := &vanusv1alpha1.Core{}
	body, err := json.Marshal(vanus)
	if err != nil {
		return nil, err
	}
	err = a.ctrl.ClientSet().Patch(types.MergePatchType).
		Namespace(vanus.Namespace).
		Name(vanus.Name).
		Resource(ResourceCore).
		VersionedParams(&metav1.PatchOptions{}, scheme.ParameterCodec).
		Body(body).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) deleteCore(namespace string, name string) error {
	_, exist, err := a.existCore(namespace, name, &metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	result := &vanusv1alpha1.Core{}
	err = a.ctrl.ClientSet().
		Delete().
		Resource(ResourceCore).
		Namespace(namespace).
		Name(name).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return err
	}
	return nil
}

func (a *Api) listCore(namespace string, opts *metav1.ListOptions) (*vanusv1alpha1.CoreList, error) {
	result := &vanusv1alpha1.CoreList{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceCore).
		Namespace(namespace).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) getCore(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Core, error) {
	result := &vanusv1alpha1.Core{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceCore).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) existCore(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Core, bool, error) {
	result := &vanusv1alpha1.Core{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceCore).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, false, nil
		}
		return result, false, err
	}
	return result, true, nil
}

func (a *Api) createConnector(connector *vanusv1alpha1.Connector, namespace string) (*vanusv1alpha1.Connector, error) {
	result := &vanusv1alpha1.Connector{}
	err := a.ctrl.ClientSet().
		Post().
		Namespace(namespace).
		Resource(ResourceConnector).
		Body(connector).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) patchConnector(connector *vanusv1alpha1.Connector) (*vanusv1alpha1.Connector, error) {
	result := &vanusv1alpha1.Connector{}
	body, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}
	err = a.ctrl.ClientSet().Patch(types.MergePatchType).
		Namespace(connector.Namespace).
		Name(connector.Name).
		Resource(ResourceConnector).
		VersionedParams(&metav1.PatchOptions{}, scheme.ParameterCodec).
		Body(body).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) deleteConnector(namespace string, name string) error {
	_, exist, err := a.existConnector(namespace, name, &metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	result := &vanusv1alpha1.Connector{}
	err = a.ctrl.ClientSet().
		Delete().
		Resource(ResourceConnector).
		Namespace(namespace).
		Name(name).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return err
	}
	return nil
}

func (a *Api) listConnector(namespace string, opts *metav1.ListOptions) (*vanusv1alpha1.ConnectorList, error) {
	result := &vanusv1alpha1.ConnectorList{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceConnector).
		Namespace(namespace).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) getConnector(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Connector, error) {
	result := &vanusv1alpha1.Connector{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceConnector).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Api) existConnector(namespace string, name string, opts *metav1.GetOptions) (*vanusv1alpha1.Connector, bool, error) {
	result := &vanusv1alpha1.Connector{}
	err := a.ctrl.ClientSet().
		Get().
		Resource(ResourceConnector).
		Namespace(namespace).
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, false, nil
		}
		return result, false, err
	}
	return result, true, nil
}

func (a *Api) updateIngress(connector *vanusv1alpha1.Connector) error {
	ingress, err := a.ctrl.K8SClientSet().NetworkingV1().Ingresses(cons.DefaultNamespace).Get(context.TODO(), cons.DefaultOperatorName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	var newIngressRules []networkingv1.IngressRule = make([]networkingv1.IngressRule, 0)
	annotation, ok := connector.Annotations[cons.ConnectorNetworkHostDomainAnnotation]
	if !ok {
		return nil
	}
	for idx, rule := range ingress.Spec.Rules {
		if rule.Host == annotation {
			continue
		}
		newIngressRules = append(newIngressRules, ingress.Spec.Rules[idx])
	}
	ingress.Spec.Rules = newIngressRules
	_, err = a.ctrl.K8SClientSet().NetworkingV1().Ingresses(cons.DefaultNamespace).Update(context.TODO(), ingress, metav1.UpdateOptions{})
	return err
}
