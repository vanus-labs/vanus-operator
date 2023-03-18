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
	stderr "errors"
	"fmt"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/vanus-labs/vanus-operator/api/models"
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/connector"
	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	"github.com/vanus-labs/vanus-operator/internal/convert"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/utils"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

const (
	SourceConnector = "source"
	SinkConnector   = "sink"
)

var (
	defaultConnectorImageTag string = "latest"
)

// All registered processing functions should appear under Registxxx in order
func RegistConnectorHandler(a *Api) {
	a.ConnectorCreateConnectorHandler = connector.CreateConnectorHandlerFunc(a.createConnectorHandler)
	a.ConnectorPatchConnectorsHandler = connector.PatchConnectorsHandlerFunc(a.patchConnectorsHandler)
	a.ConnectorPatchConnectorHandler = connector.PatchConnectorHandlerFunc(a.patchConnectorHandler)
	a.ConnectorDeleteConnectorHandler = connector.DeleteConnectorHandlerFunc(a.deleteConnectorHandler)
	a.ConnectorListConnectorHandler = connector.ListConnectorHandlerFunc(a.listConnectorHandler)
	a.ConnectorGetConnectorHandler = connector.GetConnectorHandlerFunc(a.getConnectorHandler)
}

func (a *Api) createConnectorHandler(params connector.CreateConnectorParams) middleware.Responder {
	// Check connector params
	if err := checkCreateConnectorParamsValid(params.Connector); err != nil {
		log.Errorf("create connector params invalid, err: %s\n", err.Error())
		return utils.Response(400, err)
	}

	log.Infof("show create connector params, name: %s, kind: %s, type: %s, version: %s, config: %+v, annotations: %+v\n",
		params.Connector.Name,
		params.Connector.Kind,
		params.Connector.Type,
		params.Connector.Version,
		params.Connector.Config,
		params.Connector.Annotations)

	// Check if the connector already exists, if exist, return error
	exist, err := a.checkConnectorExist(params.Connector.Name)
	if err != nil {
		log.Errorf("check connector exist failed, name: %s, err: %s\n", params.Connector.Name, err.Error())
		return utils.Response(500, err)
	}
	if exist {
		log.Warningf("Connector already exist, name: %s\n", params.Connector.Name)
		return utils.Response(500, stderr.New("connector already exist"))
	}

	log.Infof("Creating a new Connector %s/%s\n", cons.DefaultNamespace, params.Connector.Name)
	newConnector := generateConnector(params.Connector)
	resultConnector, err := a.createConnector(newConnector, cons.DefaultNamespace)
	if err != nil {
		log.Errorf("Failed to create new Connector %s/%s, err: %s\n", cons.DefaultNamespace, params.Connector.Name, err.Error())
		return utils.Response(500, err)
	}
	log.Infof("Successfully create Connector: %+v\n", resultConnector)

	return connector.NewCreateConnectorOK().WithPayload(&connector.CreateConnectorOKBody{
		Code:    convert.PtrInt32(200),
		Message: convert.PtrS("success"),
	})
}

func (a *Api) patchConnectorsHandler(params connector.PatchConnectorsParams) middleware.Responder {
	result := &models.ListOkErr{}
	for _, connector := range params.Connector {
		log.Infof("parse patch connector params finish, name: %s, version: %s, config: %+v, annotations: %+v\n",
			connector.Name,
			connector.Version,
			connector.Config,
			connector.Annotations)

		// Check if the connector exists, if not exist, return error
		oriConnector, err := a.getConnector(cons.DefaultNamespace, connector.Name, &metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				result.Failed = append(result.Failed, &models.ListOkErrFailedItems0{
					Name:   connector.Name,
					Reason: err.Error(),
				})
				continue
			}
			log.Errorf("get Connector %s/%s failed, err: %s\n", cons.DefaultNamespace, connector.Name, err.Error())
			return utils.Response(500, err)
		}

		log.Infof("Updating Connector %s/%s\n", cons.DefaultNamespace, connector.Name)
		newConnector := updateConnector(connector, oriConnector)
		_, err = a.patchConnector(newConnector)
		if err != nil {
			result.Failed = append(result.Failed, &models.ListOkErrFailedItems0{
				Name:   connector.Name,
				Reason: err.Error(),
			})
			continue
		}
		result.Successed = append(result.Successed, connector.Name)
	}

	log.Infof("Successfully patch Connectors, result: %+v\n", result)
	return connector.NewPatchConnectorsOK().WithPayload(&connector.PatchConnectorsOKBody{
		Code:    convert.PtrInt32(200),
		Data:    result,
		Message: convert.PtrS("success"),
	})
}

func (a *Api) patchConnectorHandler(params connector.PatchConnectorParams) middleware.Responder {
	log.Infof("parse patch connector params finish, name: %s, version: %s, config: %+v, annotations: %+v\n",
		params.Connector.Name,
		params.Connector.Version,
		params.Connector.Config,
		params.Connector.Annotations)

	// Check if the connector exists, if not exist, return error
	oriConnector, err := a.getConnector(cons.DefaultNamespace, params.Connector.Name, &metav1.GetOptions{})
	if err != nil {
		log.Errorf("get Connector %s/%s failed, err: %s\n", cons.DefaultNamespace, params.Connector.Name, err.Error())
		return utils.Response(500, err)
	}

	log.Infof("Updating Connector %s/%s\n", cons.DefaultNamespace, params.Connector.Name)
	newConnector := updateConnector(params.Connector, oriConnector)
	ret, err := a.patchConnector(newConnector)
	if err != nil {
		log.Errorf("patch Connector %s/%s failed, err: %s\n", cons.DefaultNamespace, params.Connector.Name, err.Error())
		return utils.Response(500, err)
	}
	log.Infof("Successfully patch Connector %s\n", params.Name)

	return connector.NewPatchConnectorOK().WithPayload(&connector.PatchConnectorOKBody{
		Code: convert.PtrInt32(200),
		Data: &models.ConnectorInfo{
			Kind:        ret.Spec.Kind,
			Name:        ret.Spec.Name,
			Type:        ret.Spec.Type,
			Version:     getConnectorVersion(ret.Spec.Image),
			Annotations: ret.Annotations,
		},
		Message: convert.PtrS("success"),
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
		return utils.Response(400, stderr.New("connector not exist"))
	}

	c, err := a.getConnector(cons.DefaultNamespace, params.Name, &metav1.GetOptions{})
	if err != nil {
		log.Errorf("get connector failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}

	err = a.updateIngress(c)
	if err != nil {
		log.Errorf("update ingress failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}

	err = a.deleteConnector(cons.DefaultNamespace, params.Name)
	if err != nil {
		log.Errorf("delete connector failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}

	return connector.NewDeleteConnectorOK().WithPayload(&connector.DeleteConnectorOKBody{
		Code:    convert.PtrInt32(200),
		Message: convert.PtrS("success"),
	})
}

func (a *Api) listConnectorHandler(params connector.ListConnectorParams) middleware.Responder {
	connectors, err := a.listConnector(cons.DefaultNamespace, &metav1.ListOptions{})
	if err != nil {
		log.Errorf("Failed to list Connectors, err: %+v\n", err.Error())
		return utils.Response(500, err)
	}
	data := make([]*models.ConnectorInfo, 0)
	for _, c := range connectors.Items {
		status, reason, err := a.getConnectorStatus(c.Name)
		if err != nil {
			log.Errorf("Get Connector %s/%s status failed, err: %+v\n", cons.DefaultNamespace, c.Name, err.Error())
			return utils.Response(500, err)
		}
		data = append(data, &models.ConnectorInfo{
			Kind:        c.Spec.Kind,
			Name:        c.Spec.Name,
			Type:        c.Spec.Type,
			Version:     getConnectorVersion(c.Spec.Image),
			Annotations: c.Annotations,
			Status:      status,
			Reason:      reason,
		})
	}
	return connector.NewListConnectorOK().WithPayload(&connector.ListConnectorOKBody{
		Code:    convert.PtrInt32(200),
		Data:    data,
		Message: convert.PtrS("success"),
	})
}

func (a *Api) getConnectorHandler(params connector.GetConnectorParams) middleware.Responder {
	c, err := a.getConnector(cons.DefaultNamespace, params.Name, &metav1.GetOptions{})
	if err != nil {
		log.Errorf("Failed to get Connector %s/%s, err: %+v\n", cons.DefaultNamespace, params.Name, err.Error())
		return utils.Response(500, err)
	}
	status, reason, err := a.getConnectorStatus(params.Name)
	if err != nil {
		log.Errorf("Get Connector %s/%s status failed, err: %+v\n", cons.DefaultNamespace, params.Name, err.Error())
		return utils.Response(500, err)
	}

	return connector.NewGetConnectorOK().WithPayload(&connector.GetConnectorOKBody{
		Code: convert.PtrInt32(200),
		Data: &models.ConnectorInfo{
			Kind:        c.Spec.Kind,
			Name:        c.Spec.Name,
			Type:        c.Spec.Type,
			Version:     getConnectorVersion(c.Spec.Image),
			Annotations: c.Annotations,
			Status:      status,
			Reason:      reason,
		},
		Message: convert.PtrS("success"),
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

func checkCreateConnectorParamsValid(connector *models.ConnectorCreate) error {
	if connector.Kind != SourceConnector && connector.Kind != SinkConnector {
		return stderr.New("create connector kind params invalid")
	}
	return nil
}

func labelsForConnector(name string) map[string]string {
	return map[string]string{"app": name}
}

func generateConnector(c *models.ConnectorCreate) *vanusv1alpha1.Connector {
	labels := labelsForConnector(c.Name)
	config, _ := yaml.Marshal(c.Config)
	version := defaultConnectorImageTag
	if c.Version != "" {
		version = c.Version
	}
	connector := &vanusv1alpha1.Connector{
		ObjectMeta: metav1.ObjectMeta{
			Name:        c.Name,
			Namespace:   cons.DefaultNamespace,
			Labels:      labels,
			Annotations: c.Annotations,
		},
		Spec: vanusv1alpha1.ConnectorSpec{
			Name:            c.Name,
			Kind:            c.Kind,
			Type:            c.Type,
			Config:          string(config),
			Image:           fmt.Sprintf("public.ecr.aws/vanus/connector/%s-%s:%s", c.Kind, c.Type, version),
			ImagePullPolicy: corev1.PullAlways,
		},
	}
	return connector
}

func updateConnector(c *models.ConnectorPatch, ori *vanusv1alpha1.Connector) *vanusv1alpha1.Connector {
	newConnector := ori.DeepCopy()
	oriVersion := getConnectorVersion(ori.Spec.Image)
	if c.Version != "" && c.Version != oriVersion {
		newConnector.Spec.Image = fmt.Sprintf("public.ecr.aws/vanus/connector/%s-%s:%s", ori.Spec.Kind, ori.Spec.Type, c.Version)
	}
	config, _ := yaml.Marshal(c.Config)
	newConnector.Spec.Config = string(config)
	newConnector.Annotations = c.Annotations
	return newConnector
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
		log.Errorf("Get Connector %s/%s status failed, err: %+v\n", cons.DefaultNamespace, name, err.Error())
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
