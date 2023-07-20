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
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/vanus-labs/vanus-operator/api/models"
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/cluster"
	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	"github.com/vanus-labs/vanus-operator/internal/convert"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

const (
	DefaultInitialVersion = "v0.7.0"
)

func RegistClusterHandler(a *Api) {
	a.ClusterCreateClusterHandler = cluster.CreateClusterHandlerFunc(a.createClusterHandler)
	a.ClusterDeleteClusterHandler = cluster.DeleteClusterHandlerFunc(a.deleteClusterHandler)
	a.ClusterPatchClusterHandler = cluster.PatchClusterHandlerFunc(a.patchClusterHandler)
	a.ClusterGetClusterHandler = cluster.GetClusterHandlerFunc(a.getClusterHandler)
}

func (a *Api) createClusterHandler(params cluster.CreateClusterParams) middleware.Responder {
	log.Infof("show cluster create params, version: %s, annotations: %+v\n", params.Create.Version, params.Create.Annotations)
	isVaild, err := a.checkParamsValid(params.Create)
	if !isVaild {
		log.Errorf("cluster params invalid, err: %s\n", err.Error())
		return utils.Response(400, err)
	}

	// Check if the cluster namespace already exists, if exist, return error
	err = a.createNsIfNotExist(params.Create)
	if err != nil {
		log.Errorf("create namespace %s failed, err: %s\n", params.Create.Namespace, err.Error())
		return utils.Response(500, err)
	}

	// Check if the cluster already exists, if exist, return error
	exist, err := a.checkClusterExist(params.Create)
	if err != nil {
		log.Errorf("check cluster exist failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}
	if exist {
		log.Warningf("Cluster already exist in namespace %s\n", params.Create.Namespace)
		return utils.Response(400, stderr.New("cluster already exist"))
	}

	core := generateCore(params.Create)
	log.Infof("Creating a new cluster %s/%s\n", core.Namespace, core.Name)
	resultCore, err := a.createCore(core)
	if err != nil {
		log.Errorf("Failed to create new cluster %s/%s, err: %s\n", core.Namespace, core.Name, err.Error())
		return utils.Response(500, err)
	}
	log.Infof("Successfully create cluster: %+v\n", resultCore)

	return cluster.NewCreateClusterOK().WithPayload(&cluster.CreateClusterOKBody{
		Code:    convert.PtrInt32(200),
		Message: convert.PtrS("success"),
	})
}

func (a *Api) deleteClusterHandler(params cluster.DeleteClusterParams) middleware.Responder {
	_, err := a.getCore(params.Delete.Namespace, params.Delete.Name, &metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Errorf("Cluster %s/%s not found, err: %s\n", params.Delete.Namespace, params.Delete.Name, err.Error())
			return utils.Response(404, err)
		}
		log.Errorf("Failed to get cluster %s/%s, err: %s\n", params.Delete.Namespace, params.Delete.Name, err.Error())
		return utils.Response(500, err)
	}

	err = a.deleteCore(params.Delete.Namespace, params.Delete.Name)
	if err != nil {
		log.Errorf("delete cluster failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}

	return cluster.NewDeleteClusterOK().WithPayload(&cluster.DeleteClusterOKBody{
		Code:    convert.PtrInt32(200),
		Message: convert.PtrS("success"),
	})
}

func (a *Api) patchClusterHandler(params cluster.PatchClusterParams) middleware.Responder {
	_, err := a.getCore(params.Patch.Namespace, params.Patch.Name, &metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Errorf("Cluster %s/%s not found, err: %s\n", params.Patch.Namespace, params.Patch.Name, err.Error())
			return utils.Response(404, err)
		}
		log.Errorf("Failed to get cluster %s/%s, err: %s\n", params.Patch.Namespace, params.Patch.Name, err.Error())
		return utils.Response(500, err)
	}
	core := &vanusv1alpha1.Core{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   params.Patch.Namespace,
			Name:        params.Patch.Name,
			Annotations: params.Patch.Annotations,
		},
		Spec: vanusv1alpha1.CoreSpec{
			Version: params.Patch.Version,
		},
	}
	resultCore, err := a.patchCore(core)
	if err != nil {
		log.Errorf("Failed to patch cluster %s/%s, err: %s\n", params.Patch.Namespace, params.Patch.Name, err.Error())
		return utils.Response(500, err)
	}
	log.Infof("Successfully patch cluster: %+v\n", resultCore)
	return cluster.NewPatchClusterOK().WithPayload(&cluster.PatchClusterOKBody{
		Code:    convert.PtrInt32(200),
		Message: convert.PtrS("success"),
	})
}

func (a *Api) getClusterHandler(params cluster.GetClusterParams) middleware.Responder {
	core, err := a.getCore(params.Get.Namespace, params.Get.Name, &metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Errorf("Cluster %s/%s not found, err: %s\n", params.Get.Namespace, params.Get.Name, err.Error())
			return utils.Response(404, err)
		}
		log.Errorf("Failed to get cluster %s/%s, err: %s\n", params.Get.Namespace, params.Get.Name, err.Error())
		return utils.Response(500, err)
	}
	return cluster.NewGetClusterOK().WithPayload(&cluster.GetClusterOKBody{
		Code: convert.PtrInt32(200),
		Data: &models.ClusterInfo{
			Name:      core.Name,
			Namespace: core.Namespace,
			Version:   core.Spec.Version,
			Status:    "Running",
		},
		Message: convert.PtrS("success"),
	})
}

func (a *Api) checkParamsValid(cluster *models.ClusterCreate) (bool, error) {
	if cluster.Version == "" {
		return false, stderr.New("cluster version is required parameters")
	}
	if strings.Compare(cluster.Version, DefaultInitialVersion) < 0 {
		log.Errorf("Only supports %s and later version.\n", DefaultInitialVersion)
		return false, stderr.New("unsupported version")
	}
	return true, nil
}

func (a *Api) checkClusterExist(cluster *models.ClusterCreate) (bool, error) {
	clusterList, err := a.listCore(cluster.Namespace, &metav1.ListOptions{})
	if err != nil {
		log.Errorf("Failed to list cluster in namespace %s, err: %s\n", cluster.Namespace, err.Error())
		return false, err
	}
	return len(clusterList.Items) != 0, err
}

func (a *Api) createNsIfNotExist(cluster *models.ClusterCreate) error {
	exist, err := a.existNamespace(cluster.Namespace)
	if err != nil {
		log.Errorf("Failed to check namespace %s if exist, err: %s\n", cluster.Namespace, err.Error())
		return err
	}
	if exist {
		return nil
	}
	return a.createNamespace(cluster.Namespace)
}

func generateCore(cluster *models.ClusterCreate) *vanusv1alpha1.Core {
	controller := &vanusv1alpha1.Core{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cluster.Name,
			Namespace:   cluster.Namespace,
			Annotations: cluster.Annotations,
		},
		Spec: vanusv1alpha1.CoreSpec{
			Version: cluster.Version,
		},
	}
	return controller
}
