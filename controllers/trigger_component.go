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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
)

func (r *VanusReconciler) handleTrigger(ctx context.Context, logger logr.Logger, vanus *vanusv1alpha1.Vanus) (ctrl.Result, error) {
	trigger := r.generateTrigger(vanus)
	// Create Trigger Deployment
	// Check if the statefulSet already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: trigger.Name, Namespace: trigger.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Trigger ConfigMap
			triggerConfigMap := r.generateConfigMapForTrigger(vanus)
			logger.Info("Creating a new Trigger ConfigMap.", "Namespace", triggerConfigMap.Namespace, "Name", triggerConfigMap.Name)
			err = r.Create(ctx, triggerConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Trigger ConfigMap", "Namespace", triggerConfigMap.Namespace, "Name", triggerConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Trigger ConfigMap")
			}
			logger.Info("Creating a new Trigger Deployment.", "Namespace", trigger.Namespace, "Name", trigger.Name)
			err = r.Create(ctx, trigger)
			if err != nil {
				logger.Error(err, "Failed to create new Trigger Deployment", "Namespace", trigger.Namespace, "Name", trigger.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Trigger Deployment")
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Trigger Deployment.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Update Trigger StatefulSet
	err = r.Update(ctx, trigger)
	if err != nil {
		logger.Error(err, "Failed to update Trigger Deployment", "Namespace", trigger.Namespace, "Name", trigger.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update Trigger Deployment")
	return ctrl.Result{}, nil
}

// returns a Trigger Deployment object
func (r *VanusReconciler) generateTrigger(vanus *vanusv1alpha1.Vanus) *appsv1.Deployment {
	labels := genLabels(cons.DefaultTriggerName)
	annotations := annotationsForTrigger()
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultTriggerName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &vanus.Spec.Replicas.Trigger,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cons.ServiceAccountName,
					Containers: []corev1.Container{{
						Name:            cons.TriggerContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.TriggerImageName, vanus.Spec.Version),
						ImagePullPolicy: vanus.Spec.ImagePullPolicy,
						Resources:       vanus.Spec.Resources,
						Env:             getEnvForTrigger(vanus),
						Ports:           getPortsForTrigger(vanus),
						VolumeMounts:    getVolumeMountsForTrigger(vanus),
					}},
					Volumes: getVolumesForTrigger(vanus),
				},
			},
		},
	}
	// Set Trigger instance as the owner and controller
	controllerutil.SetControllerReference(vanus, dep, r.Scheme)

	return dep
}

func getEnvForTrigger(vanus *vanusv1alpha1.Vanus) []corev1.EnvVar {
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

func getPortsForTrigger(vanus *vanusv1alpha1.Vanus) []corev1.ContainerPort {
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameGrpc,
		ContainerPort: cons.TriggerPortGrpc,
	}}
	return defaultPorts
}

func getVolumeMountsForTrigger(vanus *vanusv1alpha1.Vanus) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.ConfigMountPath,
		Name:      cons.TriggerConfigMapName,
	}}
	return defaultVolumeMounts
}

func getVolumesForTrigger(vanus *vanusv1alpha1.Vanus) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.TriggerConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.TriggerConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func annotationsForTrigger() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.ControllerPortMetrics)}
}

func (r *VanusReconciler) generateConfigMapForTrigger(vanus *vanusv1alpha1.Vanus) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("port: 2148\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString("controllers:\n")
	for i := int32(0); i < vanus.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller.vanus.svc:2048\n", i))
	}
	data["trigger.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.TriggerConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(vanus, cm, r.Scheme)
	return cm
}