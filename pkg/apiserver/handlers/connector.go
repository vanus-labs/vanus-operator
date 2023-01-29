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

package handlers

import (
	stderr "errors"
	"fmt"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/linkall-labs/vanus-operator/api/models"
	"github.com/linkall-labs/vanus-operator/api/restapi/operations/connector"
	vanusv1alpha1 "github.com/linkall-labs/vanus-operator/api/v1alpha1"
	cons "github.com/linkall-labs/vanus-operator/internal/constants"
	"github.com/linkall-labs/vanus-operator/pkg/apiserver/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

// All registered processing functions should appear under Registxxx in order
func RegistConnectorHandler(a *Api) {
	a.ConnectorCreateConnectorHandler = connector.CreateConnectorHandlerFunc(a.createConnectorHandler)
	a.ConnectorDeleteConnectorHandler = connector.DeleteConnectorHandlerFunc(a.deleteConnectorHandler)
	a.ConnectorListConnectorHandler = connector.ListConnectorHandlerFunc(a.listConnectorHandler)
	a.ConnectorGetConnectorHandler = connector.GetConnectorHandlerFunc(a.getConnectorHandler)
}

func (a *Api) createConnectorHandler(params connector.CreateConnectorParams) middleware.Responder {
	// Parse cluster params
	c, err := genConnectorConfig(params.Connector)
	if err != nil {
		log.Error(err, "parse connector params failed")
		return utils.Response(0, err)
	}

	log.Infof("parse cluster params finish, name: %s, kind: %s, namespace: %s\n", c.name, c.kind, c.namespace)

	// Check if the connector already exists, if exist, return error
	exist, err := a.checkConnectorExist(c.name)
	if err != nil {
		log.Errorf("check connector exist failed, name: %s, err: %s\n", c.name, err.Error())
		return utils.Response(0, err)
	}
	if exist {
		log.Warningf("Connector already exist, name: %s\n", c.name)
		return utils.Response(0, stderr.New("connector already exist"))
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
		return utils.Response(0, err)
	} else {
		log.Infof("Successfully create Connector: %+v\n", resultConnector)
	}

	return connector.NewCreateConnectorOK().WithPayload(nil)
}

func (a *Api) deleteConnectorHandler(params connector.DeleteConnectorParams) middleware.Responder {
	// Check if the connector already exists
	exist, err := a.checkConnectorExist(*params.Name)
	if err != nil {
		log.Errorf("check connector exist failed, err: %s\n", err.Error())
		return utils.Response(0, err)
	}
	if !exist {
		log.Warning("Connector not exist")
		return connector.NewDeleteConnectorOK().WithPayload(nil)
	}

	err = a.deleteConnector(cons.DefaultNamespace, *params.Name)
	if err != nil {
		log.Errorf("delete connector failed, err: %s\n", err.Error())
	}

	return connector.NewDeleteConnectorOK().WithPayload(nil)
}

func (a *Api) listConnectorHandler(params connector.ListConnectorParams) middleware.Responder {
	connectors, err := a.listConnector(cons.DefaultNamespace, &metav1.ListOptions{})
	if err != nil {
		log.Error(err, "Failed to list Connectors", "Connector.Namespace: ", cons.DefaultNamespace)
		return utils.Response(0, err)
	}
	retcode := int32(400)
	msg := "list connector success"
	data := make([]*models.ConnectorInfo, 0)
	for _, c := range connectors.Items {
		data = append(data, &models.ConnectorInfo{
			Kind:    c.Spec.Kind,
			Name:    c.Spec.Name,
			Version: strings.Split(c.Spec.Image, ":")[1],
			Status:  "Running",
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
		return utils.Response(0, err)
	}
	retcode := int32(400)
	msg := "get connector success"
	return connector.NewGetConnectorOK().WithPayload(&connector.GetConnectorOKBody{
		Code: &retcode,
		Data: &models.ConnectorInfo{
			Kind:    c.Spec.Kind,
			Name:    c.Spec.Name,
			Version: strings.Split(c.Spec.Image, ":")[1],
			Status:  "Running",
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
	name      string
	namespace string
	kind      string
	ctype     string
	version   string
	config    string
}

func genConnectorConfig(connector *models.ConnectorInfo) (*connectorConfig, error) {
	// check required parameters
	c := &connectorConfig{
		name:      connector.Name,
		namespace: cons.DefaultNamespace,
		kind:      connector.Kind,
		ctype:     connector.Type,
		version:   connector.Version,
		config:    connector.Config,
	}
	if connector.Version == "" {
		c.version = "latest"
	}
	return c, nil
}

func labelsForConnector(name string) map[string]string {
	return map[string]string{"app": name}
}

func generateConnector(c *connectorConfig) *vanusv1alpha1.Connector {
	labels := labelsForConnector(c.name)
	controller := &vanusv1alpha1.Connector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
			Labels:    labels,
		},
		Spec: vanusv1alpha1.ConnectorSpec{
			Name:            c.name,
			Kind:            c.kind,
			Config:          c.config,
			Image:           fmt.Sprintf("public.ecr.aws/vanus/connector/%s-%s:%s", c.kind, c.ctype, c.version),
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}
	return controller
}