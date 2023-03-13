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
	"github.com/vanus-labs/vanus-operator/api/restapi/operations/cluster"
	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	"github.com/vanus-labs/vanus-operator/pkg/apiserver/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"
)

const (
	DefaultInitialVersion = "v0.7.0"
)

const (
	// ClusterStatusStatusHealthy captures enum value "Healthy"
	ClusterStatusStatusHealthy string = "Healthy"

	// ClusterStatusStatusUnhealthy captures enum value "Unhealthy"
	ClusterStatusStatusUnhealthy string = "Unhealthy"

	// ClusterStatusStatusUnknown captures enum value "Unknown"
	ClusterStatusStatusUnknown string = "Unknown"
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
		return utils.Response(400, err)
	}

	log.Infof("parse cluster params finish, config: %s\n", c.String())

	isVaild, err := a.checkParamsValid(params)
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

	defer func() {
		if failedToExit {
			if vanusDeployed {
				err = a.deleteCore(cons.DefaultVanusClusterName, c.namespace)
				if err != nil {
					log.Warningf("clear controller failed when failed to exit, err: %s\n", err.Error())
				}
			}
		}
	}()

	log.Infof("Creating a new Core cluster, Core.Namespace: %s, Core.Name: %s\n", c.namespace, cons.DefaultVanusClusterName)
	vanus := generateCore(c)
	resultCore, err := a.createCore(vanus, c.namespace)
	if err != nil {
		log.Errorf("Failed to create new Core cluster, Core.Namespace: %s, Core.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultVanusClusterName, err.Error())
		failedToExit = true
		return utils.Response(500, err)
	}
	vanusDeployed = true
	log.Infof("Successfully create Core cluster: %+v\n", resultCore)

	retcode := int32(200)
	msg := "success"
	return cluster.NewCreateClusterOK().WithPayload(&cluster.CreateClusterOKBody{
		Code:    &retcode,
		Message: &msg,
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

	err = a.deleteCore(cons.DefaultNamespace, cons.DefaultVanusClusterName)
	if err != nil {
		log.Errorf("delete vanus cluster failed, err: %s\n", err.Error())
		return utils.Response(500, err)
	}

	retcode := int32(200)
	msg := "success"
	return cluster.NewDeleteClusterOK().WithPayload(&cluster.DeleteClusterOKBody{
		Code:    &retcode,
		Message: &msg,
	})
}

func (a *Api) patchClusterHandler(params cluster.PatchClusterParams) middleware.Responder {
	vanus := &vanusv1alpha1.Core{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cons.DefaultNamespace,
			Name:      cons.DefaultVanusClusterName,
		},
		Spec: vanusv1alpha1.CoreSpec{
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
	resultCore, err := a.patchCore(vanus)
	if err != nil {
		log.Errorf("Failed to patch Core cluster, Core.Namespace: %s, Core.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultVanusClusterName, err.Error())
		return utils.Response(500, err)
	}
	log.Infof("Successfully patch Core cluster: %+v\n", resultCore)
	retcode := int32(200)
	msg := "success"
	return cluster.NewPatchClusterOK().WithPayload(&cluster.PatchClusterOKBody{
		Code:    &retcode,
		Message: &msg,
	})
}

func (a *Api) getClusterHandler(params cluster.GetClusterParams) middleware.Responder {
	vanus, err := a.getCore(cons.DefaultNamespace, cons.DefaultVanusClusterName, &metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to get Core cluster", "Core.Namespace: ", cons.DefaultNamespace, "Core.Name: ", cons.DefaultVanusClusterName)
		return utils.Response(500, err)
	}
	retcode := int32(200)
	msg := "success"

	status, err := a.getCoreStatus()
	if err != nil {
		log.Error(err, "Get Core status failed", "Core.Namespace: ", cons.DefaultNamespace, "Core.Name: ", cons.DefaultVanusClusterName)
		return utils.Response(500, err)
	}

	return cluster.NewGetClusterOK().WithPayload(&cluster.GetClusterOKBody{
		Code: &retcode,
		Data: &models.ClusterInfo{
			Version: vanus.Spec.Version,
			Status:  status,
		},
		Message: &msg,
	})
}

func (a *Api) checkParamsValid(params cluster.CreateClusterParams) (bool, error) {
	if strings.Compare(params.Create.Version, DefaultInitialVersion) < 0 {
		log.Errorf("Only supports %s and later version.\n", DefaultInitialVersion)
		return false, errors.New("unsupported version")
	}
	return true, nil
}

func (a *Api) checkClusterExist() (bool, error) {
	_, exist, err := a.existCore(cons.DefaultNamespace, cons.DefaultVanusClusterName, &metav1.GetOptions{})
	if err != nil {
		log.Errorf("Failed to get Core cluster, err: %s\n", err.Error())
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
		return nil, errors.New("cluster version is required parameters")
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
		c.triggerReplicas = *cluster.TriggerReplicas
	} else {
		c.triggerReplicas = 1
	}
	if cluster.TimerReplicas != nil {
		c.timerReplicas = *cluster.TimerReplicas
	} else {
		c.timerReplicas = 2
	}
	if cluster.GatewayReplicas != nil {
		c.gatewayReplicas = *cluster.GatewayReplicas
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

func generateCore(c *config) *vanusv1alpha1.Core {
	labels := genLabels(cons.DefaultVanusClusterName)
	requests := make(map[corev1.ResourceName]resource.Quantity)
	requests[corev1.ResourceStorage] = resource.MustParse(c.storeStorageSize)
	controller := &vanusv1alpha1.Core{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      cons.DefaultVanusClusterName,
		},
		Spec: vanusv1alpha1.CoreSpec{
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

func (a *Api) getCoreStatus() (string, error) {
	etcdSts := &appsv1.StatefulSet{}
	err := a.ctrl.Get(types.NamespacedName{Name: cons.DefaultEtcdName, Namespace: cons.DefaultNamespace}, etcdSts)
	if err != nil {
		return "", err
	}
	if etcdSts.Status.Replicas != etcdSts.Status.ReadyReplicas || etcdSts.Status.Replicas != etcdSts.Status.AvailableReplicas {
		log.Warningf("Etcd status unhealthy, replicas: %d, ready_replicas: %d, available_replicas: %d\n", etcdSts.Status.Replicas, etcdSts.Status.ReadyReplicas, etcdSts.Status.AvailableReplicas)
		return ClusterStatusStatusUnhealthy, nil
	}

	controllerSts := &appsv1.StatefulSet{}
	err = a.ctrl.Get(types.NamespacedName{Name: cons.DefaultControllerName, Namespace: cons.DefaultNamespace}, controllerSts)
	if err != nil {
		return "", err
	}
	if controllerSts.Status.Replicas != controllerSts.Status.ReadyReplicas || controllerSts.Status.Replicas != controllerSts.Status.AvailableReplicas {
		log.Warningf("Controller status unhealthy, replicas: %d, ready_replicas: %d, available_replicas: %d\n", controllerSts.Status.Replicas, controllerSts.Status.ReadyReplicas, controllerSts.Status.AvailableReplicas)
		return ClusterStatusStatusUnhealthy, nil
	}

	storeSts := &appsv1.StatefulSet{}
	err = a.ctrl.Get(types.NamespacedName{Name: cons.DefaultStoreName, Namespace: cons.DefaultNamespace}, storeSts)
	if err != nil {
		return "", err
	}
	if storeSts.Status.Replicas != storeSts.Status.ReadyReplicas || storeSts.Status.Replicas != storeSts.Status.AvailableReplicas {
		log.Warningf("Store status unhealthy, replicas: %d, ready_replicas: %d, available_replicas: %d\n", storeSts.Status.Replicas, storeSts.Status.ReadyReplicas, storeSts.Status.AvailableReplicas)
		return ClusterStatusStatusUnhealthy, nil
	}

	gatewayDep := &appsv1.Deployment{}
	err = a.ctrl.Get(types.NamespacedName{Name: cons.DefaultGatewayName, Namespace: cons.DefaultNamespace}, gatewayDep)
	if err != nil {
		return "", err
	}
	if gatewayDep.Status.Replicas != gatewayDep.Status.ReadyReplicas || gatewayDep.Status.Replicas != gatewayDep.Status.AvailableReplicas {
		log.Warningf("Gateway status unhealthy, replicas: %d, ready_replicas: %d, available_replicas: %d\n", gatewayDep.Status.Replicas, gatewayDep.Status.ReadyReplicas, gatewayDep.Status.AvailableReplicas)
		return ClusterStatusStatusUnhealthy, nil
	}

	triggerDep := &appsv1.Deployment{}
	err = a.ctrl.Get(types.NamespacedName{Name: cons.DefaultTriggerName, Namespace: cons.DefaultNamespace}, triggerDep)
	if err != nil {
		return "", err
	}
	if triggerDep.Status.Replicas != triggerDep.Status.ReadyReplicas || triggerDep.Status.Replicas != triggerDep.Status.AvailableReplicas {
		log.Warningf("Trigger status unhealthy, replicas: %d, ready_replicas: %d, available_replicas: %d\n", triggerDep.Status.Replicas, triggerDep.Status.ReadyReplicas, triggerDep.Status.AvailableReplicas)
		return ClusterStatusStatusUnhealthy, nil
	}

	timerDep := &appsv1.Deployment{}
	err = a.ctrl.Get(types.NamespacedName{Name: cons.DefaultTimerName, Namespace: cons.DefaultNamespace}, timerDep)
	if err != nil {
		return "", err
	}
	if timerDep.Status.Replicas != timerDep.Status.ReadyReplicas || timerDep.Status.Replicas != timerDep.Status.AvailableReplicas {
		log.Warningf("Timer status unhealthy, replicas: %d, ready_replicas: %d, available_replicas: %d\n", timerDep.Status.Replicas, timerDep.Status.ReadyReplicas, timerDep.Status.AvailableReplicas)
		return ClusterStatusStatusUnhealthy, nil
	}
	return ClusterStatusStatusHealthy, nil
}
