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
	"errors"
	"fmt"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/vanus-labs/vanus-operator/api/models"
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/connector"
	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/utils"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// PodStatusStatusPending captures enum value "Pending"
	PodStatusStatusPending string = "Pending"

	// PodStatusStatusRunning captures enum value "Running"
	PodStatusStatusRunning string = "Running"

	// PodStatusStatusSucceeded captures enum value "Succeeded"
	PodStatusStatusSucceeded string = "Succeeded"

	// PodStatusStatusStarting captures enum value "Starting"
	PodStatusStatusStarting string = "Starting"

	// PodStatusStatusFailed captures enum value "Failed"
	PodStatusStatusFailed string = "Failed"

	// PodStatusStatusRemoving captures enum value "Removing"
	PodStatusStatusRemoving string = "Removing"

	// PodStatusStatusUnknown captures enum value "Unknown"
	PodStatusStatusUnknown string = "Unknown"
)

var (
	defaultConnectorImageTag string = "latest"
	defaultConnectorSvcPort  string = "80"
)

// All registered processing functions should appear under Registxxx in order
func RegistConnectorHandler(a *Api) {
	a.ConnectorCreateConnectorHandler = connector.CreateConnectorHandlerFunc(a.createConnectorHandler)
	a.ConnectorDeleteConnectorHandler = connector.DeleteConnectorHandlerFunc(a.deleteConnectorHandler)
	a.ConnectorListConnectorHandler = connector.ListConnectorHandlerFunc(a.listConnectorHandler)
	a.ConnectorGetConnectorHandler = connector.GetConnectorHandlerFunc(a.getConnectorHandler)
}

func (a *Api) createConnectorHandler(params connector.CreateConnectorParams) middleware.Responder {
	// Parse connector params
	c, err := genConnectorConfig(params.Connector)
	if err != nil {
		log.Error(err, "parse connector params failed")
		return utils.Response(400, err)
	}

	log.Infof("parse connector params finish, config: %s\n", c.String())

	// Check if the connector already exists, if exist, return error
	exist, err := a.checkConnectorExist(c.name)
	if err != nil {
		log.Errorf("check connector exist failed, name: %s, err: %s\n", c.name, err.Error())
		return utils.Response(500, err)
	}
	if exist {
		log.Warningf("Connector already exist, name: %s\n", c.name)
		return utils.Response(500, errors.New("connector already exist"))
	}

	defer func() {
		err = a.deleteConnector(c.name, c.namespace)
		if err != nil {
			log.Warningf("clear connector failed when failed to exit, err: %s\n", err.Error())
		}
	}()

	log.Infof("Creating a new Connector, Connector.Namespace: %s, Connector.Name: %s\n", c.namespace, c.name)
	newConnector := generateConnector(c)
	resultConnector, err := a.createConnector(newConnector, c.namespace)
	if err != nil {
		log.Errorf("Failed to create new Connector, Connector.Namespace: %s, Connector.Name: %s, err: %s\n", cons.DefaultNamespace, c.name, err.Error())
		return utils.Response(500, err)
	}
	log.Infof("Successfully create Connector: %+v\n", resultConnector)

	retcode := int32(200)
	msg := "create connector success"
	return connector.NewCreateConnectorOK().WithPayload(&connector.CreateConnectorOKBody{
		Code:    &retcode,
		Message: &msg,
	})
}

func (a *Api) deleteConnectorHandler(params connector.DeleteConnectorParams) middleware.Responder {
	// Check if the connector already exists
	exist, err := a.checkConnectorExist(params.Name)
	if err != nil {
		log.Errorf("check connector exist failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}
	if !exist {
		log.Warning("Connector not exist")
		return utils.Response(400, errors.New("connector not exist"))
	}

	err = a.deleteConnector(cons.DefaultNamespace, params.Name)
	if err != nil {
		log.Errorf("delete connector failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}

	retcode := int32(200)
	msg := "delete connector success"
	return connector.NewDeleteConnectorOK().WithPayload(&connector.DeleteConnectorOKBody{
		Code:    &retcode,
		Message: &msg,
	})
}

func (a *Api) listConnectorHandler(params connector.ListConnectorParams) middleware.Responder {
	connectors, err := a.listConnector(cons.DefaultNamespace, &metav1.ListOptions{})
	if err != nil {
		log.Error(err, "Failed to list Connectors", "Connector.Namespace: ", cons.DefaultNamespace)
		return utils.Response(500, err)
	}
	retcode := int32(200)
	msg := "list connectors success"
	data := make([]*models.ConnectorInfo, 0)
	for _, c := range connectors.Items {
		status, reason, err := a.getConnectorStatus(c.Name)
		if err != nil {
			log.Error(err, "Get Connector status failed", "Connector.Namespace: ", cons.DefaultNamespace, "Connector.Name: ", c.Name)
			return utils.Response(500, err)
		}
		data = append(data, &models.ConnectorInfo{
			Kind:        c.Spec.Kind,
			Name:        c.Spec.Name,
			Type:        c.Spec.Type,
			ServiceType: c.Spec.ServiceType,
			ServicePort: c.Spec.ServicePort,
			Version:     getConnectorVersion(c.Spec.Image),
			Status:      status,
			Reason:      reason,
		})
	}
	return connector.NewListConnectorOK().WithPayload(&connector.ListConnectorOKBody{
		Code:    &retcode,
		Data:    data,
		Message: &msg,
	})
}

func (a *Api) getConnectorHandler(params connector.GetConnectorParams) middleware.Responder {
	c, err := a.getConnector(cons.DefaultNamespace, params.Name, &metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to get Connector", "Connector.Namespace: ", cons.DefaultNamespace, "Connector.Name: ", params.Name)
		return utils.Response(500, err)
	}
	retcode := int32(200)
	msg := "get connector success"

	status, reason, err := a.getConnectorStatus(params.Name)
	if err != nil {
		log.Error(err, "Get Connector status failed", "Connector.Namespace: ", cons.DefaultNamespace, "Connector.Name: ", params.Name)
		return utils.Response(500, err)
	}

	return connector.NewGetConnectorOK().WithPayload(&connector.GetConnectorOKBody{
		Code: &retcode,
		Data: &models.ConnectorInfo{
			Kind:        c.Spec.Kind,
			Name:        c.Spec.Name,
			Type:        c.Spec.Type,
			ServiceType: c.Spec.ServiceType,
			ServicePort: c.Spec.ServicePort,
			Version:     getConnectorVersion(c.Spec.Image),
			Status:      status,
			Reason:      reason,
		},
		Message: &msg,
	})
}

func (a *Api) checkConnectorExist(name string) (bool, error) {
	// TODO(jiangkai): need to check other components
	_, exist, err := a.existConnector(cons.DefaultNamespace, name, &metav1.GetOptions{})
	if err != nil {
		log.Errorf("Failed to get Connector, name: %s, err: %s\n", name, err.Error())
		return false, err
	}
	if exist {
		return true, nil
	}
	return false, nil
}

type connectorConfig struct {
	name       string
	namespace  string
	kind       string
	ctype      string
	stype      string
	sport      string
	version    string
	config     map[string]interface{}
	annotaions map[string]string
}

func (c *connectorConfig) String() string {
	return fmt.Sprintf("name: %s, namespace: %s, kind: %s, type: %s, service_type: %s, service_port: %s, version: %s\n",
		c.name,
		c.namespace,
		c.kind,
		c.ctype,
		c.stype,
		c.sport,
		c.version)
}

func genConnectorConfig(connector *models.ConnectorCreate) (*connectorConfig, error) {
	// check required parameters
	c := &connectorConfig{
		name:      connector.Name,
		namespace: cons.DefaultNamespace,
		kind:      connector.Kind,
		ctype:     connector.Type,
		stype:     connector.ServiceType,
		sport:     connector.ServicePort,
		version:   connector.Version,
		config:    connector.Config,
	}
	if connector.Version == "" {
		c.version = defaultConnectorImageTag
	}
	if connector.ServicePort == "" {
		c.sport = defaultConnectorSvcPort
	}
	if len(connector.Annotations) != 0 {
		annotations := make(map[string]string, len(connector.Annotations))
		for key, value := range connector.Annotations {
			annotations[key] = value
		}
		c.annotaions = annotations
	}
	return c, nil
}

func labelsForConnector(name string) map[string]string {
	return map[string]string{"app": name}
}

func generateConnector(c *connectorConfig) *vanusv1alpha1.Connector {
	labels := labelsForConnector(c.name)
	config, _ := yaml.Marshal(c.config)
	connector := &vanusv1alpha1.Connector{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.name,
			Namespace:   c.namespace,
			Labels:      labels,
			Annotations: c.annotaions,
		},
		Spec: vanusv1alpha1.ConnectorSpec{
			Name:            c.name,
			Kind:            c.kind,
			Type:            c.ctype,
			ServiceType:     c.stype,
			ServicePort:     c.sport,
			Config:          string(config),
			Annotations:     c.annotaions,
			Image:           fmt.Sprintf("public.ecr.aws/vanus/connector/%s-%s:%s", c.kind, c.ctype, c.version),
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}
	return connector
}

func getConnectorVersion(s string) string {
	return strings.Split(s, ":")[1]
}

func (a *Api) getConnectorStatus(name string) (string, string, error) {
	pods := &corev1.PodList{}
	l := make(map[string]string)
	l["app"] = name
	opts := &ctrlclient.ListOptions{Namespace: cons.DefaultNamespace, LabelSelector: labels.SelectorFromSet(l)}
	err := a.ctrl.List(pods, opts)
	if err != nil {
		log.Error(err, "Get Connector status failed", "Connector.Namespace: ", cons.DefaultNamespace, "Connector.Name: ", name)
		return "", "", err
	}
	var status, reason string
	if len(pods.Items) != 0 {
		status, reason = statusCheck(&pods.Items[0])
	}
	return status, reason, nil
}

func statusCheck(a *corev1.Pod) (string, string) {
	if a == nil {
		return PodStatusStatusUnknown, ""
	}
	if a.DeletionTimestamp != nil {
		return PodStatusStatusRemoving, ""
	}
	// Status: Pending/Succeeded/Failed/Unknown
	if a.Status.Phase != corev1.PodRunning {
		return string(a.Status.Phase), a.Status.Reason
	}
	// handle running
	var (
		containers = a.Status.ContainerStatuses
		rnum       int
	)

	for _, v := range containers {
		if v.Ready {
			rnum++
			continue
		}
		if v.State.Terminated != nil {
			if v.State.Terminated.ExitCode != 0 {
				return PodStatusStatusFailed, v.State.Terminated.Reason
			}
			if v.State.Waiting != nil {
				return PodStatusStatusStarting, v.State.Waiting.Reason
			}
		}
	}
	if rnum == len(containers) {
		return PodStatusStatusRunning, ""
	} else {
		return PodStatusStatusStarting, ""
	}
}
