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

	"github.com/go-openapi/runtime/middleware"
	"github.com/vanus-labs/vanus-operator/api/models"
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/cluster"
	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

var (
	WorkloadOldReplicAnno = "replic.ecns.io/workload"
	// revisionAnnoKey       = "deployment.kubernetes.io/revision"
	// rolloutReasonAnno     = "kubernetes.io/change-cause"
)

func RegistClusterHandler(a *Api) {
	a.ClusterCreateClusterHandler = cluster.CreateClusterHandlerFunc(a.createClusterHandler)
	a.ClusterDeleteClusterHandler = cluster.DeleteClusterHandlerFunc(a.deleteClusterHandler)
	a.ClusterPatchClusterHandler = cluster.PatchClusterHandlerFunc(a.patchClusterHandler)
	a.ClusterGetClusterHandler = cluster.GetClusterHandlerFunc(a.getClusterHandler)
}

func (a *Api) createClusterHandler(params cluster.CreateClusterParams) middleware.Responder {
	var (
		failedToExit  = false
		vanusDeployed = false
	)
	// Parse cluster params
	c, err := genClusterConfig(params.Create)
	if err != nil {
		log.Error(err, "parse cluster params failed")
		return utils.Response(0, err)
	}

	log.Infof("parse cluster params finish, config: %s\n", c.String())

	// Check if the cluster already exists, if exist, return error
	exist, err := a.checkClusterExist()
	if err != nil {
		log.Errorf("check cluster exist failed, err: %s\n", err.Error())
		return utils.Response(0, err)
	}
	if exist {
		log.Warning("Cluster already exist")
		return utils.Response(0, stderr.New("cluster already exist"))
	}

	defer func() {
		if failedToExit {
			if vanusDeployed {
				err = a.deleteVanus(cons.DefaultVanusClusterName, c.namespace)
				if err != nil {
					log.Warningf("clear controller failed when failed to exit, err: %s\n", err.Error())
				}
			}
		}
	}()

	log.Infof("Creating a new Vanus cluster, Vanus.Namespace: %s, Vanus.Name: %s\n", c.namespace, cons.DefaultVanusClusterName)
	vanus := generateVanus(c)
	resultVanus, err := a.createVanus(vanus, c.namespace)
	if err != nil {
		log.Errorf("Failed to create new Vanus cluster, Vanus.Namespace: %s, Vanus.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultVanusClusterName, err.Error())
		failedToExit = true
		return utils.Response(0, err)
	} else {
		vanusDeployed = true
		log.Infof("Successfully create Vanus cluster: %+v\n", resultVanus)
	}

	return cluster.NewCreateClusterOK().WithPayload(nil)
}

func (a *Api) deleteClusterHandler(params cluster.DeleteClusterParams) middleware.Responder {
	// Check if the cluster already exists
	exist, err := a.checkClusterExist()
	if err != nil {
		log.Errorf("check cluster exist failed, err: %s\n", err.Error())
		return utils.Response(0, err)
	}
	if !exist {
		log.Warning("Cluster not exist")
		return cluster.NewDeleteClusterOK().WithPayload(nil)
	}

	err = a.deleteVanus(cons.DefaultNamespace, cons.DefaultVanusClusterName)
	if err != nil {
		log.Errorf("delete vanus cluster failed, err: %s\n", err.Error())
		return utils.Response(0, err)
	}

	return cluster.NewDeleteClusterOK().WithPayload(nil)
}

func (a *Api) patchClusterHandler(params cluster.PatchClusterParams) middleware.Responder {
	vanus := &vanusv1alpha1.Vanus{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cons.DefaultNamespace,
			Name:      cons.DefaultVanusClusterName,
		},
		Spec: vanusv1alpha1.VanusSpec{
			Replicas: vanusv1alpha1.Replicas{
				Controller: params.Patch.ControllerReplicas,
				Store:      params.Patch.StoreReplicas,
				Trigger:    params.Patch.TriggerReplicas,
				Timer:      params.Patch.TimerReplicas,
				Gateway:    params.Patch.GatewayReplicas,
			},
			Version: params.Patch.Version,
		},
	}
	resultVanus, err := a.patchVanus(vanus)
	if err != nil {
		log.Errorf("Failed to patch Vanus cluster, Vanus.Namespace: %s, Vanus.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultVanusClusterName, err.Error())
		return utils.Response(0, err)
	}
	log.Infof("Successfully patch Vanus cluster: %+v\n", resultVanus)
	return cluster.NewPatchClusterOK().WithPayload(nil)
}

func (a *Api) getClusterHandler(params cluster.GetClusterParams) middleware.Responder {
	vanus, err := a.getVanus(cons.DefaultNamespace, cons.DefaultVanusClusterName, &metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to get Vanus cluster", "Vanus.Namespace: ", cons.DefaultNamespace, "Vanus.Name: ", cons.DefaultVanusClusterName)
		return utils.Response(0, err)
	}
	retcode := int32(400)
	msg := "get cluster success"
	return cluster.NewGetClusterOK().WithPayload(&cluster.GetClusterOKBody{
		Code: &retcode,
		Data: &models.ClusterInfo{
			Version: vanus.Spec.Version,
			Status:  "Running",
		},
		Message: &msg,
	})
}

func (a *Api) checkClusterExist() (bool, error) {
	// TODO(jiangkai): need to check other components
	_, exist, err := a.existVanus(cons.DefaultNamespace, cons.DefaultVanusClusterName, &metav1.GetOptions{})
	if err != nil {
		log.Errorf("Failed to get Vanus cluster, err: %s\n", err.Error())
		return false, err
	}
	return exist, err
}

type config struct {
	namespace           string
	version             string
	controllerReplicas  int32
	storeReplicas       int32
	triggerReplicas     int32
	timerReplicas       int32
	gatewayReplicas     int32
	metadataStorageSize string
	storeStorageSize    string
}

func (c *config) String() string {
	return fmt.Sprintf("namespace: %s, version: %s, controller_replicas: %d, store_replicas: %d, trigger_replicas: %d, timer_replicas: %d, gateway_replicas: %d, metadata_storage_size: %s, store_storage_size: %s\n",
		c.namespace,
		c.version,
		c.controllerReplicas,
		c.storeReplicas,
		c.triggerReplicas,
		c.timerReplicas,
		c.gatewayReplicas,
		c.metadataStorageSize,
		c.storeStorageSize)
}

func genClusterConfig(cluster *models.ClusterCreate) (*config, error) {
	// check required parameters
	if cluster.Version == "" {
		return nil, stderr.New("cluster version is required parameters")
	}

	c := &config{
		version:   cluster.Version,
		namespace: cons.DefaultNamespace,
	}
	if cluster.ControllerReplicas != nil {
		c.controllerReplicas = *cluster.ControllerReplicas
	} else {
		c.controllerReplicas = 3
	}
	if cluster.StoreReplicas != nil {
		c.storeReplicas = *cluster.StoreReplicas
	} else {
		c.storeReplicas = 3
	}
	if cluster.TriggerReplicas != nil {
		c.controllerReplicas = *cluster.TriggerReplicas
	} else {
		c.triggerReplicas = 1
	}
	if cluster.TimerReplicas != nil {
		c.controllerReplicas = *cluster.TimerReplicas
	} else {
		c.timerReplicas = 2
	}
	if cluster.GatewayReplicas != nil {
		c.controllerReplicas = *cluster.GatewayReplicas
	} else {
		c.gatewayReplicas = 1
	}
	if cluster.MetadataStorageSize != nil {
		c.metadataStorageSize = *cluster.MetadataStorageSize
	} else {
		c.metadataStorageSize = "10Gi"
	}
	if cluster.StoreStorageSize != nil {
		c.storeStorageSize = *cluster.StoreStorageSize
	} else {
		c.storeStorageSize = "10Gi"
	}
	return c, nil
}

func genLabels(name string) map[string]string {
	return map[string]string{"app": name}
}

func generateVanus(c *config) *vanusv1alpha1.Vanus {
	labels := genLabels(cons.DefaultVanusClusterName)
	requests := make(map[corev1.ResourceName]resource.Quantity)
	requests[corev1.ResourceStorage] = resource.MustParse(c.storeStorageSize)
	controller := &vanusv1alpha1.Vanus{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      cons.DefaultVanusClusterName,
		},
		Spec: vanusv1alpha1.VanusSpec{
			Replicas: vanusv1alpha1.Replicas{
				Controller: c.controllerReplicas,
				Store:      c.storeReplicas,
				Trigger:    c.triggerReplicas,
				Timer:      c.timerReplicas,
				Gateway:    c.gatewayReplicas,
			},
			Version:         c.version,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Resources:       corev1.ResourceRequirements{},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Name:   cons.VolumeName,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: requests,
					},
				},
			}},
		},
	}
	return controller
}
