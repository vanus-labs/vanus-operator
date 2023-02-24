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
		failedToExit       = false
		controllerDeployed = false
		storeDeployed      = false
		triggerDeployed    = false
		timerDeployed      = false
	)
	// Parse cluster params
	c, err := genClusterConfig(params.Create)
	if err != nil {
		log.Error(err, "parse cluster params failed")
		return utils.Response(0, err)
	}

	log.Infof("parse cluster params finish, version: %s, namespace: %s, controller_replicas: %d, controller_storage_size: %s, store_replicas: %d, store_storage_size: %s\n", c.version, c.namespace, c.controllerReplicas, c.controllerStorageSize, c.storeReplicas, c.storeStorageSize)

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
			if controllerDeployed {
				err = a.deleteController(cons.DefaultControllerName, c.namespace)
				if err != nil {
					log.Warningf("clear controller failed when failed to exit, err: %s\n", err.Error())
				}
			}
			if storeDeployed {
				err = a.deleteStore(c.namespace, cons.DefaultStoreName)
				if err != nil {
					log.Warningf("clear store failed when failed to exit, err: %s\n", err.Error())
				}
			}
			if triggerDeployed {
				err = a.deleteTrigger(c.namespace, cons.DefaultTriggerName)
				if err != nil {
					log.Warningf("clear trigger failed when failed to exit, err: %s\n", err.Error())
				}
			}
			if timerDeployed {
				err = a.deleteTimer(c.namespace, cons.DefaultTimerName)
				if err != nil {
					log.Warningf("clear timer failed when failed to exit, err: %s\n", err.Error())
				}
			}
		}
	}()

	log.Infof("Creating a new Controller, Controller.Namespace: %s, Controller.Name: %s\n", c.namespace, cons.DefaultControllerName)
	controller := generateController(c)
	resultController, err := a.createController(controller, c.namespace)
	if err != nil {
		log.Errorf("Failed to create new Controller, Controller.Namespace: %s, Controller.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultControllerName, err.Error())
		failedToExit = true
		return utils.Response(0, err)
	} else {
		controllerDeployed = true
		log.Infof("Successfully create Controller: %+v\n", resultController)
	}

	log.Infof("Creating a new Store, Store.Namespace: %s, Store.Name: %s\n", c.namespace, cons.DefaultStoreName)
	store := generateStore(c)
	resultStore, err := a.createStore(store, c.namespace)
	if err != nil {
		log.Errorf("Failed to create new Store, Store.Namespace: %s, Store.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultStoreName, err.Error())
		failedToExit = true
		return utils.Response(0, err)
	} else {
		storeDeployed = true
		log.Infof("Successfully create Store: %+v\n", resultStore)
	}

	log.Infof("Creating a new Trigger, Trigger.Namespace: %s, Trigger.Name: %s\n", c.namespace, cons.DefaultTriggerName)
	trigger := generateTrigger(c)
	resultTrigger, err := a.createTrigger(trigger, c.namespace)
	if err != nil {
		log.Errorf("Failed to create new Trigger, Trigger.Namespace: %s, Trigger.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultTriggerName, err.Error())
		failedToExit = true
		return utils.Response(0, err)
	} else {
		triggerDeployed = true
		log.Infof("Successfully create Trigger: %+v\n", resultTrigger)
	}

	log.Infof("Creating a new Timer, Timer.Namespace: %s, Timer.Name: %s\n", c.namespace, cons.DefaultTimerName)
	timer := generateTimer(c)
	resultTimer, err := a.createTimer(timer, c.namespace)
	if err != nil {
		log.Errorf("Failed to create new Timer, Timer.Namespace: %s, Timer.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultTimerName, err.Error())
		failedToExit = true
		return utils.Response(0, err)
	} else {
		timerDeployed = true
		log.Infof("Successfully create Timer: %+v\n", resultTimer)
	}

	log.Infof("Creating a new Gateway, Gateway.Namespace: %s, Gateway.Name: %s\n", c.namespace, cons.DefaultGatewayName)
	gateway := generateGateway(c)
	resultGateway, err := a.createGateway(gateway, c.namespace)
	if err != nil {
		log.Errorf("Failed to create new Gateway, Gateway.Namespace: %s, Gateway.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultGatewayName, err.Error())
		failedToExit = true
		return utils.Response(0, err)
	} else {
		log.Info("Successfully create Gateway: %+v\n", resultGateway)
	}

	return cluster.NewCreateClusterOK().WithPayload(nil)
}

func (a *Api) deleteClusterHandler(params cluster.DeleteClusterParams) middleware.Responder {
	var (
		result = 0
	)

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

	err = a.deleteController(cons.DefaultNamespace, cons.DefaultControllerName)
	if err != nil {
		result++
		log.Errorf("delete controller failed, err: %s\n", err.Error())
	}
	err = a.deleteStore(cons.DefaultNamespace, cons.DefaultStoreName)
	if err != nil {
		result++
		log.Errorf("delete store failed, err: %s\n", err.Error())
	}
	err = a.deleteTrigger(cons.DefaultNamespace, cons.DefaultTriggerName)
	if err != nil {
		result++
		log.Errorf("delete trigger failed, err: %s\n", err.Error())
	}
	err = a.deleteTimer(cons.DefaultNamespace, cons.DefaultTimerName)
	if err != nil {
		result++
		log.Errorf("delete timer failed, err: %s\n", err.Error())
	}
	err = a.deleteGateway(cons.DefaultNamespace, cons.DefaultGatewayName)
	if err != nil {
		result++
		log.Errorf("delete gateway failed, err: %s\n", err.Error())
	}

	if result != 0 {
		log.Errorf("delete cluster failed, have %d components failed to delete.\n", result)
		return utils.Response(0, err)
	}

	return cluster.NewDeleteClusterOK().WithPayload(nil)
}

func (a *Api) patchClusterHandler(params cluster.PatchClusterParams) middleware.Responder {
	if params.Patch.Version != "" {
		controller := &vanusv1alpha1.Controller{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cons.DefaultNamespace,
				Name:      cons.DefaultControllerName,
			},
			Spec: vanusv1alpha1.ControllerSpec{
				Image: fmt.Sprintf("public.ecr.aws/vanus/controller:%s", params.Patch.Version),
			},
		}
		resultController, err := a.patchController(controller)
		if err != nil {
			log.Errorf("Failed to patch Controller, Controller.Namespace: %s, Controller.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultControllerName, err.Error())
			return utils.Response(0, err)
		} else {
			log.Infof("Successfully patch Controller: %+v\n", resultController)
		}
		store := &vanusv1alpha1.Store{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cons.DefaultNamespace,
				Name:      cons.DefaultStoreName,
			},
			Spec: vanusv1alpha1.StoreSpec{
				Image: fmt.Sprintf("public.ecr.aws/vanus/store:%s", params.Patch.Version),
			},
		}
		if params.Patch.StoreReplicas != 0 {
			store.Spec.Replicas = &params.Patch.StoreReplicas
		}
		resultStore, err := a.patchStore(store)
		if err != nil {
			log.Errorf("Failed to patch Store, Store.Namespace: %s, Store.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultStoreName, err.Error())
			return utils.Response(0, err)
		} else {
			log.Infof("Successfully patch Store: %+v\n", resultStore)
		}
		trigger := &vanusv1alpha1.Trigger{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cons.DefaultNamespace,
				Name:      cons.DefaultTriggerName,
			},
			Spec: vanusv1alpha1.TriggerSpec{
				Image: fmt.Sprintf("public.ecr.aws/vanus/trigger:%s", params.Patch.Version),
			},
		}
		resultTrigger, err := a.patchTrigger(trigger)
		if err != nil {
			log.Errorf("Failed to patch Trigger, Trigger.Namespace: %s, Trigger.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultTriggerName, err.Error())
			return utils.Response(0, err)
		} else {
			log.Infof("Successfully patch Trigger: %+v\n", resultTrigger)
		}
		timer := &vanusv1alpha1.Timer{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cons.DefaultNamespace,
				Name:      cons.DefaultTimerName,
			},
			Spec: vanusv1alpha1.TimerSpec{
				Image: fmt.Sprintf("public.ecr.aws/vanus/timer:%s", params.Patch.Version),
			},
		}
		resultTimer, err := a.patchTimer(timer)
		if err != nil {
			log.Errorf("Failed to patch Timer, Timer.Namespace: %s, Timer.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultTimerName, err.Error())
			return utils.Response(0, err)
		} else {
			log.Infof("Successfully patch Timer: %+v\n", resultTimer)
		}
		gateway := &vanusv1alpha1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cons.DefaultNamespace,
				Name:      cons.DefaultGatewayName,
			},
			Spec: vanusv1alpha1.GatewaySpec{
				Image: fmt.Sprintf("public.ecr.aws/vanus/timer:%s", params.Patch.Version),
			},
		}
		resultGateway, err := a.patchGateway(gateway)
		if err != nil {
			log.Errorf("Failed to patch Gateway, Gateway.Namespace: %s, Gateway.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultGatewayName, err.Error())
			return utils.Response(0, err)
		} else {
			log.Infof("Successfully patch Gateway: %+v\n", resultGateway)
		}
	} else if params.Patch.ControllerReplicas != 0 {
		controller := &vanusv1alpha1.Controller{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cons.DefaultNamespace,
				Name:      cons.DefaultControllerName,
			},
			Spec: vanusv1alpha1.ControllerSpec{
				Replicas: &params.Patch.ControllerReplicas,
			},
		}
		resultController, err := a.patchController(controller)
		if err != nil {
			log.Errorf("Failed to patch Controller, Controller.Namespace: %s, Controller.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultControllerName, err.Error())
			return utils.Response(0, err)
		} else {
			log.Infof("Successfully patch Controller: %+v\n", resultController)
		}
	} else if params.Patch.StoreReplicas != 0 {
		store := &vanusv1alpha1.Store{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cons.DefaultNamespace,
				Name:      cons.DefaultStoreName,
			},
			Spec: vanusv1alpha1.StoreSpec{
				Replicas: &params.Patch.StoreReplicas,
			},
		}
		resultStore, err := a.patchStore(store)
		if err != nil {
			log.Errorf("Failed to patch Store, Store.Namespace: %s, Store.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultStoreName, err.Error())
			return utils.Response(0, err)
		} else {
			log.Infof("Successfully patch Store: %+v\n", resultStore)
		}
	} else if params.Patch.TriggerReplicas != 0 {
		trigger := &vanusv1alpha1.Trigger{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cons.DefaultNamespace,
				Name:      cons.DefaultTriggerName,
			},
			Spec: vanusv1alpha1.TriggerSpec{
				Replicas: &params.Patch.TriggerReplicas,
			},
		}
		resultTrigger, err := a.patchTrigger(trigger)
		if err != nil {
			log.Errorf("Failed to patch Trigger, Trigger.Namespace: %s, Trigger.Name: %s, err: %s\n", cons.DefaultNamespace, cons.DefaultTriggerName, err.Error())
			return utils.Response(0, err)
		} else {
			log.Infof("Successfully patch Trigger: %+v\n", resultTrigger)
		}
	}
	return cluster.NewPatchClusterOK().WithPayload(nil)
}

func (a *Api) getClusterHandler(params cluster.GetClusterParams) middleware.Responder {
	controller, err := a.getController(cons.DefaultNamespace, cons.DefaultControllerName, &metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to get Controller", "Controller.Namespace: ", cons.DefaultNamespace, "Controller.Name: ", cons.DefaultControllerName)
		return utils.Response(0, err)
	}
	retcode := int32(400)
	msg := "get cluster success"
	return cluster.NewGetClusterOK().WithPayload(&cluster.GetClusterOKBody{
		Code: &retcode,
		Data: &models.ClusterInfo{
			Version: strings.Split(controller.Spec.Image, ":")[1],
			Status:  "Running",
		},
		Message: &msg,
	})
}

func (a *Api) checkClusterExist() (bool, error) {
	// TODO(jiangkai): need to check other components
	_, exist, err := a.existController(cons.DefaultNamespace, cons.DefaultControllerName, &metav1.GetOptions{})
	if err != nil {
		log.Errorf("Failed to get Controller, err: %s\n", err.Error())
		return false, err
	}
	if exist {
		return true, nil
	}
	return false, nil
}

type config struct {
	namespace             string
	version               string
	controllerReplicas    int32
	controllerStorageSize string
	storeReplicas         int32
	storeStorageSize      string
}

func genClusterConfig(cluster *models.ClusterCreate) (*config, error) {
	// check required parameters
	if cluster.Version == "" {
		return nil, stderr.New("cluster version is required parameters")
	}
	c := &config{
		version:               cluster.Version,
		namespace:             cons.DefaultNamespace,
		controllerReplicas:    cons.DefaultControllerReplicas,
		controllerStorageSize: cons.VolumeStorage,
		storeReplicas:         cons.DefaultStoreReplicas,
		storeStorageSize:      cons.VolumeStorage,
	}
	if cluster.ControllerReplicas != 0 && cluster.ControllerReplicas != cons.DefaultControllerReplicas {
		c.controllerReplicas = cluster.ControllerReplicas
	}
	if cluster.ControllerStorageSize != "" {
		c.controllerStorageSize = cluster.ControllerStorageSize
	}
	if cluster.StoreReplicas != 0 && cluster.StoreReplicas != cons.DefaultStoreReplicas {
		c.storeReplicas = cluster.StoreReplicas
	}
	if cluster.ControllerStorageSize != "" {
		c.storeStorageSize = cluster.StoreStorageSize
	}
	return c, nil
}

func labelsForController(name string) map[string]string {
	return map[string]string{"app": name}
}

func generateController(c *config) *vanusv1alpha1.Controller {
	labels := labelsForController(cons.DefaultControllerName)
	requests := make(map[corev1.ResourceName]resource.Quantity)
	requests[corev1.ResourceStorage] = resource.MustParse(c.controllerStorageSize)
	controller := &vanusv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      cons.DefaultControllerName,
		},
		Spec: vanusv1alpha1.ControllerSpec{
			Replicas:        &c.controllerReplicas,
			Image:           fmt.Sprintf("public.ecr.aws/vanus/controller:%s", c.version),
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

func generateStore(c *config) *vanusv1alpha1.Store {
	labels := labelsForController(cons.DefaultStoreName)
	requests := make(map[corev1.ResourceName]resource.Quantity)
	requests[corev1.ResourceStorage] = resource.MustParse(c.storeStorageSize)
	store := &vanusv1alpha1.Store{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      cons.DefaultStoreName,
		},
		Spec: vanusv1alpha1.StoreSpec{
			Replicas:        &c.storeReplicas,
			Image:           fmt.Sprintf("public.ecr.aws/vanus/store:%s", c.version),
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
	return store
}

func generateTrigger(c *config) *vanusv1alpha1.Trigger {
	trigger := &vanusv1alpha1.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      cons.DefaultTriggerName,
		},
		Spec: vanusv1alpha1.TriggerSpec{
			Image:           fmt.Sprintf("public.ecr.aws/vanus/trigger:%s", c.version),
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}
	return trigger
}

func generateTimer(c *config) *vanusv1alpha1.Timer {
	timer := &vanusv1alpha1.Timer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      cons.DefaultTimerName,
		},
		Spec: vanusv1alpha1.TimerSpec{
			Image:           fmt.Sprintf("public.ecr.aws/vanus/timer:%s", c.version),
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}
	return timer
}

func generateGateway(c *config) *vanusv1alpha1.Gateway {
	gateway := &vanusv1alpha1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.namespace,
			Name:      cons.DefaultGatewayName,
		},
		Spec: vanusv1alpha1.GatewaySpec{
			Image:           fmt.Sprintf("public.ecr.aws/vanus/gateway:%s", c.version),
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}
	return gateway
}
