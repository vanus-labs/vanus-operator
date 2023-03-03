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

package controllers

import (
	"bytes"
	"context"
	stderr "errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
)

func (r *CoreReconciler) handleController(ctx context.Context, logger logr.Logger, core *vanusv1alpha1.Core) (ctrl.Result, error) {
	var (
		controller          *appsv1.StatefulSet
		controllerConfigMap *corev1.ConfigMap
	)
	if strings.Compare(core.Spec.Version, EtcdSeparateVersion) < 0 {
		controller = r.generateController(core)
		controllerConfigMap = r.generateConfigMapForController(core)
	} else {
		controller = r.generateNewController(core)
		controllerConfigMap = r.generateConfigMapForNewController(core)
	}
	// Create Controller StatefulSet
	// Check if the statefulSet already exists, if not create a new one
	sts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: controller.Name, Namespace: controller.Namespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Controller ConfigMap
			logger.Info("Creating a new Controller ConfigMap.", "Namespace", controllerConfigMap.Namespace, "Name", controllerConfigMap.Name)
			err = r.Create(ctx, controllerConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Controller ConfigMap", "Namespace", controllerConfigMap.Namespace, "Name", controllerConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Controller ConfigMap")
			}
			logger.Info("Creating a new Controller StatefulSet.", "Namespace", controller.Namespace, "Name", controller.Name)
			err = r.Create(ctx, controller)
			if err != nil {
				logger.Error(err, "Failed to create new Controller StatefulSet", "Namespace", controller.Namespace, "Name", controller.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Controller StatefulSet")
			}
			controllerSvc := r.generateSvcForController(core)
			// Create Controller Service
			// Check if the service already exists, if not create a new one
			svc := &corev1.Service{}
			err = r.Get(ctx, types.NamespacedName{Name: controllerSvc.Name, Namespace: controllerSvc.Namespace}, svc)
			if err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Creating a new Controller Service.", "Namespace", controllerSvc.Namespace, "Name", controllerSvc.Name)
					err = r.Create(ctx, controllerSvc)
					if err != nil {
						logger.Error(err, "Failed to create new Controller Service", "Namespace", controllerSvc.Namespace, "Name", controllerSvc.Name)
						return ctrl.Result{}, err
					} else {
						logger.Info("Successfully create Controller Service")
					}
				} else {
					logger.Error(err, "Failed to get Controller Service.")
					return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
				}
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Controller StatefulSet.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
	}

	logger.Info("Updating Controller ConfigMap.", "Namespace", controllerConfigMap.Namespace, "Name", controllerConfigMap.Name)
	err = r.Update(ctx, controllerConfigMap)
	if err != nil {
		logger.Error(err, "Failed to update Controller ConfigMap", "Namespace", controllerConfigMap.Namespace, "Name", controllerConfigMap.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update Controller ConfigMap")

	logger.Info("Updating Controller StatefulSet.", "Namespace", controller.Namespace, "Name", controller.Name)
	if strings.Compare(version(sts.Spec.Template.Spec.Containers[0].Image), EtcdSeparateVersion) < 0 && strings.Compare(core.Spec.Version, EtcdSeparateVersion) >= 0 {
		err = r.Delete(ctx, sts)
		if err != nil {
			logger.Error(err, "Failed to Delete Controller StatefulSet", "Namespace", sts.Namespace, "Name", sts.Name)
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, controller)
		if err != nil {
			logger.Error(err, "Failed to create Controller StatefulSet", "Namespace", controller.Namespace, "Name", controller.Name)
			return ctrl.Result{}, err
		}
	} else {
		err = r.Update(ctx, controller)
		if err != nil {
			logger.Error(err, "Failed to update Controller StatefulSet", "Namespace", controller.Namespace, "Name", controller.Name)
			return ctrl.Result{}, err
		}
	}
	logger.Info("Successfully update Controller StatefulSet")

	// Wait for Controller is ready
	start := time.Now()
	logger.Info("Wait for Controller is ready")
	t := time.NewTicker(defaultWaitForReadyTimeout)
	defer t.Stop()
	for {
		ready, err := r.waitControllerIsReady(ctx, core)
		if err != nil {
			logger.Error(err, "Wait for Controller is ready but got error")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
		if ready {
			break
		}
		select {
		case <-t.C:
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, stderr.New("controller isn't ready")
		default:
			time.Sleep(time.Second)
		}
	}
	logger.Info("Controller is ready", "WaitingTime", time.Since(start))

	return ctrl.Result{}, nil
}

// returns a Controller StatefulSet object
func (r *CoreReconciler) generateController(core *vanusv1alpha1.Core) *appsv1.StatefulSet {
	labels := genLabels(cons.DefaultControllerName)
	annotations := annotationsForController()
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultControllerName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &core.Spec.Replicas.Controller,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			ServiceName: cons.DefaultControllerName,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            cons.ControllerContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.ControllerImageName, core.Spec.Version),
						ImagePullPolicy: core.Spec.ImagePullPolicy,
						Resources:       core.Spec.Resources,
						Env:             getEnvForController(core),
						Ports:           getPortsForController(core),
						VolumeMounts:    getVolumeMountsForController(core),
						Command:         getCommandForController(core),
					}},
					Volumes: getVolumesForController(core),
				},
			},
			VolumeClaimTemplates: getVolumeClaimTemplatesForController(core),
		},
	}
	// Set Controller instance as the owner and controller
	controllerutil.SetControllerReference(core, sts, r.Scheme)

	return sts
}

// returns a Controller StatefulSet object
func (r *CoreReconciler) generateNewController(core *vanusv1alpha1.Core) *appsv1.StatefulSet {
	labels := genLabels(cons.DefaultControllerName)
	annotations := annotationsForController()
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultControllerName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &core.Spec.Replicas.Controller,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			ServiceName: cons.DefaultControllerName,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            cons.ControllerContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.ControllerImageName, core.Spec.Version),
						ImagePullPolicy: core.Spec.ImagePullPolicy,
						Resources:       core.Spec.Resources,
						Env:             getEnvForController(core),
						Ports:           getPortsForNewController(core),
						VolumeMounts:    getVolumeMountsForNewController(core),
						Command:         getCommandForController(core),
					}},
					Volumes: getVolumesForController(core),
				},
			},
		},
	}
	// Set Controller instance as the owner and controller
	controllerutil.SetControllerReference(core, sts, r.Scheme)

	return sts
}

func (r *CoreReconciler) waitControllerIsReady(ctx context.Context, core *vanusv1alpha1.Core) (bool, error) {
	sts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: cons.DefaultControllerName, Namespace: cons.DefaultNamespace}, sts)
	if err != nil {
		return false, err
	}
	if sts.Status.Replicas == core.Spec.Replicas.Controller && sts.Status.ReadyReplicas == core.Spec.Replicas.Controller && sts.Status.AvailableReplicas == core.Spec.Replicas.Controller {
		return true, nil
	}
	return false, nil
}

func getEnvForController(core *vanusv1alpha1.Core) []corev1.EnvVar {
	defaultEnvs := []corev1.EnvVar{{
		Name:      cons.EnvPodIP,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
	}, {
		Name:      cons.EnvPodName,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
	}, {
		Name:  cons.EnvLogLevel,
		Value: "INFO",
	}}
	return defaultEnvs
}

func getPortsForController(core *vanusv1alpha1.Core) []corev1.ContainerPort {
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameGrpc,
		ContainerPort: cons.ControllerPortGrpc,
	}, {
		Name:          cons.ContainerPortNameEtcdClient,
		ContainerPort: cons.ControllerPortEtcdClient,
	}, {
		Name:          cons.ContainerPortNameEtcdPeer,
		ContainerPort: cons.ControllerPortEtcdPeer,
	}, {
		Name:          cons.ContainerPortNameMetrics,
		ContainerPort: cons.ControllerPortMetrics,
	}}
	return defaultPorts
}

func getPortsForNewController(core *vanusv1alpha1.Core) []corev1.ContainerPort {
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameGrpc,
		ContainerPort: cons.ControllerPortGrpc,
	}, {
		Name:          cons.ContainerPortNameMetrics,
		ContainerPort: cons.ControllerPortMetrics,
	}}
	return defaultPorts
}

func getVolumeMountsForController(core *vanusv1alpha1.Core) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.ConfigMountPath,
		Name:      cons.ControllerConfigMapName,
	}, {
		MountPath: cons.VolumeMountPath,
		Name:      cons.VolumeName,
	}}
	return defaultVolumeMounts
}

func getVolumeMountsForNewController(core *vanusv1alpha1.Core) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.ConfigMountPath,
		Name:      cons.ControllerConfigMapName,
	}}
	return defaultVolumeMounts
}

func getCommandForController(core *vanusv1alpha1.Core) []string {
	defaultCommand := []string{"/bin/sh", "-c", "NODE_ID=${HOSTNAME##*-} /vanus/bin/controller"}
	return defaultCommand
}

func getVolumesForController(core *vanusv1alpha1.Core) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.ControllerConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.ControllerConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func getVolumeClaimTemplatesForController(core *vanusv1alpha1.Core) []corev1.PersistentVolumeClaim {
	labels := genLabels(cons.DefaultControllerName)
	requests := make(map[corev1.ResourceName]resource.Quantity)
	requests[corev1.ResourceStorage] = resource.MustParse(cons.VolumeStorage)
	defaultPersistentVolumeClaims := []corev1.PersistentVolumeClaim{{
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
	}}
	if len(core.Spec.VolumeClaimTemplates) != 0 {
		if core.Spec.VolumeClaimTemplates[0].Name != "" {
			defaultPersistentVolumeClaims[0].Name = core.Spec.VolumeClaimTemplates[0].Name
		}
		defaultPersistentVolumeClaims[0].Spec.Resources = core.Spec.VolumeClaimTemplates[0].Spec.Resources
	}
	return defaultPersistentVolumeClaims
}

func genLabels(name string) map[string]string {
	return map[string]string{"app": name}
}

func annotationsForController() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.ControllerPortMetrics)}
}

func (r *CoreReconciler) generateConfigMapForController(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("node_id: ${NODE_ID}\n")
	value.WriteString("name: ${POD_NAME}\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString("port: 2048\n")
	value.WriteString("etcd:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller:2379\n", i))
	}
	value.WriteString("data_dir: /data\n")
	value.WriteString(fmt.Sprintf("replicas: %d\n", core.Spec.Replicas.Controller))
	value.WriteString("metadata:\n")
	value.WriteString("  key_prefix: /vanus\n")
	value.WriteString("topology:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  vanus-controller-%d: vanus-controller-%d.vanus-controller.vanus.svc:2048\n", i, i))
	}
	value.WriteString("embed_etcd:\n")
	value.WriteString("  data_dir: etcd/data\n")
	value.WriteString("  listen_client_addr: 0.0.0.0:2379\n")
	value.WriteString("  listen_peer_addr: 0.0.0.0:2380\n")
	value.WriteString("  advertise_client_addr: ${POD_NAME}.vanus-controller:2379\n")
	value.WriteString("  advertise_peer_addr: ${POD_NAME}.vanus-controller:2380\n")
	value.WriteString("  clusters:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("    - vanus-controller-%d=http://vanus-controller-%d.vanus-controller:2380\n", i, i))
	}
	data["controller.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       "config-controller",
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}

func (r *CoreReconciler) generateConfigMapForNewController(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("node_id: ${NODE_ID}\n")
	value.WriteString("name: ${POD_NAME}\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString("port: 2048\n")
	value.WriteString("etcd:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-etcd-%d.vanus-etcd:2379\n", i))
	}
	value.WriteString("data_dir: /data\n")
	value.WriteString(fmt.Sprintf("replicas: %d\n", core.Spec.Replicas.Controller))
	value.WriteString("metadata:\n")
	value.WriteString("  key_prefix: /vanus\n")
	value.WriteString("leaderelection:\n")
	value.WriteString("  lease_duration: 15\n")
	value.WriteString("topology:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  vanus-controller-%d: vanus-controller-%d.vanus-controller.vanus.svc:2048\n", i, i))
	}
	data["controller.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       "config-controller",
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}

func (r *CoreReconciler) generateSvcForController(core *vanusv1alpha1.Core) *corev1.Service {
	labels := genLabels(cons.DefaultControllerName)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.DefaultControllerName,
			Labels:     labels,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: cons.HeadlessService,
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name:       cons.DefaultControllerName,
					Port:       cons.ControllerPortGrpc,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(cons.ControllerPortGrpc),
				},
			},
		},
	}

	controllerutil.SetControllerReference(core, svc, r.Scheme)
	return svc
}
