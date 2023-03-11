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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
)

func (r *CoreReconciler) handleStore(ctx context.Context, logger logr.Logger, core *vanusv1alpha1.Core) (ctrl.Result, error) {
	store := r.generateStore(core)
	// Check if the statefulSet already exists, if not create a new one
	sts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: store.Name, Namespace: store.Namespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Store ConfigMap
			storeConfigMap := r.generateConfigMapForStore(core)
			logger.Info("Creating a new Store ConfigMap.", "Namespace", storeConfigMap.Namespace, "Name", storeConfigMap.Name)
			err = r.Create(ctx, storeConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Store ConfigMap", "Namespace", storeConfigMap.Namespace, "Name", storeConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Store ConfigMap")
			}
			// Create Store StatefulSet
			logger.Info("Creating a new Store StatefulSet.", "Namespace", store.Namespace, "Name", store.Name)
			err = r.Create(ctx, store)
			if err != nil {
				logger.Error(err, "Failed to create new Store StatefulSet", "Namespace", store.Namespace, "Name", store.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Store StatefulSet")
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Store StatefulSet.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Update Store StatefulSet
	logger.Info("Updating Store StatefulSet.", "Namespace", store.Namespace, "Name", store.Name)
	err = r.Update(ctx, store)
	if err != nil {
		logger.Error(err, "Failed to update Store StatefulSet", "Namespace", store.Namespace, "Name", store.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update Store StatefulSet")

	return ctrl.Result{}, nil
}

// returns a Store StatefulSet object
func (r *CoreReconciler) generateStore(core *vanusv1alpha1.Core) *appsv1.StatefulSet {
	labels := genLabels(cons.DefaultStoreName)
	annotations := annotationsForStore()
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultStoreName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &core.Spec.Replicas.Store,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			ServiceName: cons.DefaultStoreName,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            cons.StoreContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.StoreImageName, core.Spec.Version),
						ImagePullPolicy: core.Spec.ImagePullPolicy,
						Resources:       core.Spec.Resources,
						Env:             getEnvForStore(core),
						Ports:           getPortsForStore(core),
						VolumeMounts:    getVolumeMountsForStore(core),
						Command:         getCommandForStore(core),
					}},
					Volumes: getVolumesForStore(core),
				},
			},
			VolumeClaimTemplates: getVolumeClaimTemplatesForStore(core),
		},
	}
	// Set Store instance as the owner and controller
	controllerutil.SetControllerReference(core, sts, r.Scheme)

	return sts
}

func getEnvForStore(core *vanusv1alpha1.Core) []corev1.EnvVar {
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

func getPortsForStore(core *vanusv1alpha1.Core) []corev1.ContainerPort {
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameGrpc,
		ContainerPort: cons.StorePortGrpc,
	}}
	return defaultPorts
}

func getVolumeMountsForStore(core *vanusv1alpha1.Core) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.ConfigMountPath,
		Name:      cons.StoreConfigMapName,
	}, {
		MountPath: cons.VolumeMountPath,
		Name:      cons.VolumeName,
	}}
	return defaultVolumeMounts
}

func getCommandForStore(core *vanusv1alpha1.Core) []string {
	defaultCommand := []string{"/bin/sh", "-c", "VOLUME_ID=${HOSTNAME##*-} /vanus/bin/store"}
	return defaultCommand
}

func getVolumesForStore(core *vanusv1alpha1.Core) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.StoreConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.StoreConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func getVolumeClaimTemplatesForStore(core *vanusv1alpha1.Core) []corev1.PersistentVolumeClaim {
	labels := genLabels(cons.DefaultStoreName)
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
		if core.Spec.VolumeClaimTemplates[0].Spec.StorageClassName != nil {
			defaultPersistentVolumeClaims[0].Spec.StorageClassName = core.Spec.VolumeClaimTemplates[0].Spec.StorageClassName
		}
	}
	return defaultPersistentVolumeClaims
}

func annotationsForStore() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.ControllerPortMetrics)}
}

func (r *CoreReconciler) generateConfigMapForStore(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("port: 11811\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString("controllers:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller:2048\n", i))
	}
	value.WriteString("volume:\n")
	value.WriteString("  id: ${VOLUME_ID}\n")
	value.WriteString("  dir: /data\n")
	value.WriteString("  capacity: 1073741824\n")
	value.WriteString("meta_store:\n")
	value.WriteString("  wal:\n")
	value.WriteString("    io:\n")
	value.WriteString("      engine: psync\n")
	value.WriteString("offset_store:\n")
	value.WriteString("  wal:\n")
	value.WriteString("    io:\n")
	value.WriteString("      engine: psync\n")
	value.WriteString("raft:\n")
	value.WriteString("  wal:\n")
	value.WriteString("    io:\n")
	value.WriteString("      engine: psync\n")
	data["store.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.StoreConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}
