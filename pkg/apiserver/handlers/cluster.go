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
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/vanus-labs/vanus-operator/api/models"
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/cluster"
	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	"github.com/vanus-labs/vanus-operator/internal/convert"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/utils"
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

	// Check if the cluster already exists, if exist, return error
	exist, err := a.checkClusterExist()
	if err != nil {
		log.Errorf("check cluster exist failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}
	if exist {
		log.Warning("Cluster already exist")
		return utils.Response(400, errors.New("cluster already exist"))
	}

	log.Infof("Creating a new cluster %s/%s\n", cons.DefaultNamespace, cons.DefaultVanusCoreName)
	core := generateCore(params.Create)
	resultCore, err := a.createCore(core, cons.DefaultNamespace)
	if err != nil {
		log.Errorf("Failed to create new cluster %s/%s, err: %s\n", cons.DefaultNamespace, cons.DefaultVanusCoreName, err.Error())
		return utils.Response(500, err)
	}
	log.Infof("Successfully create cluster: %+v\n", resultCore)

	return cluster.NewCreateClusterOK().WithPayload(&cluster.CreateClusterOKBody{
		Code:    convert.PtrInt32(200),
		Message: convert.PtrS("success"),
	})
}

func (a *Api) deleteClusterHandler(params cluster.DeleteClusterParams) middleware.Responder {
	// Check if the cluster already exists
	exist, err := a.checkClusterExist()
	if err != nil {
		log.Errorf("check cluster exist failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}
	if !exist {
		log.Warning("Cluster not exist")
		return utils.Response(400, errors.New("cluster not exist"))
	}

	err = a.deleteCore(cons.DefaultNamespace, cons.DefaultVanusCoreName)
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
	core := &vanusv1alpha1.Core{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   cons.DefaultNamespace,
			Name:        cons.DefaultVanusCoreName,
			Annotations: params.Patch.Annotations,
		},
		Spec: vanusv1alpha1.CoreSpec{
			Version: params.Patch.Version,
		},
	}
	resultCore, err := a.patchCore(core)
	if err != nil {
		log.Errorf("Failed to patch cluster %s/%s, err: %s\n", cons.DefaultNamespace, cons.DefaultVanusCoreName, err.Error())
		return utils.Response(500, err)
	}
	log.Infof("Successfully patch cluster: %+v\n", resultCore)
	return cluster.NewPatchClusterOK().WithPayload(&cluster.PatchClusterOKBody{
		Code:    convert.PtrInt32(200),
		Message: convert.PtrS("success"),
	})
}

func (a *Api) getClusterHandler(params cluster.GetClusterParams) middleware.Responder {
	vanus, err := a.getCore(cons.DefaultNamespace, cons.DefaultVanusCoreName, &metav1.GetOptions{})
	if err != nil {
		log.Errorf("Failed to get cluster %s/%s, err: %s\n", cons.DefaultNamespace, cons.DefaultVanusCoreName, err.Error())
		return utils.Response(500, err)
	}
	return cluster.NewGetClusterOK().WithPayload(&cluster.GetClusterOKBody{
		Code: convert.PtrInt32(200),
		Data: &models.ClusterInfo{
			Version: vanus.Spec.Version,
			Status:  "Running",
		},
		Message: convert.PtrS("success"),
	})
}

func (a *Api) checkParamsValid(cluster *models.ClusterCreate) (bool, error) {
	if cluster.Version == "" {
		return false, errors.New("cluster version is required parameters")
	}
	if strings.Compare(cluster.Version, DefaultInitialVersion) < 0 {
		log.Errorf("Only supports %s and later version.\n", DefaultInitialVersion)
		return false, errors.New("unsupported version")
	}
	return true, nil
}

func (a *Api) checkClusterExist() (bool, error) {
	_, exist, err := a.existCore(cons.DefaultNamespace, cons.DefaultVanusCoreName, &metav1.GetOptions{})
	if err != nil {
		log.Errorf("Failed to get cluster, err: %s\n", err.Error())
		return false, err
	}
	return exist, err
}

func generateCore(cluster *models.ClusterCreate) *vanusv1alpha1.Core {
	controller := &vanusv1alpha1.Core{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   cons.DefaultNamespace,
			Name:        cons.DefaultVanusCoreName,
			Annotations: cluster.Annotations,
		},
		Spec: vanusv1alpha1.CoreSpec{
			Version: cluster.Version,
		},
	}
	return controller
}
