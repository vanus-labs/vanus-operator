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

func (r *CoreReconciler) handleTimer(ctx context.Context, logger logr.Logger, core *vanusv1alpha1.Core) (ctrl.Result, error) {
	timer := r.generateTimer(core)
	// Check if the statefulSet already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: timer.Name, Namespace: timer.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Timer ConfigMap
			timerConfigMap := r.generateConfigMapForTimer(core)
			logger.Info("Creating a new Timer ConfigMap.", "Namespace", timerConfigMap.Namespace, "Name", timerConfigMap.Name)
			err = r.Create(ctx, timerConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Timer ConfigMap", "Namespace", timerConfigMap.Namespace, "Name", timerConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Timer ConfigMap")
			}
			// Create Timer Deployment
			logger.Info("Creating a new Timer Deployment.", "Namespace", timer.Namespace, "Name", timer.Name)
			err = r.Create(ctx, timer)
			if err != nil {
				logger.Error(err, "Failed to create new Timer Deployment", "Namespace", timer.Namespace, "Name", timer.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Timer Deployment")
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Timer Deployment.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Update Timer Deployment
	logger.Info("Updating Timer Deployment.", "Namespace", timer.Namespace, "Name", timer.Name)
	err = r.Update(ctx, timer)
	if err != nil {
		logger.Error(err, "Failed to update Timer Deployment", "Namespace", timer.Namespace, "Name", timer.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update Timer Deployment")
	return ctrl.Result{}, nil
}

// returns a Timer Deployment object
func (r *CoreReconciler) generateTimer(core *vanusv1alpha1.Core) *appsv1.Deployment {
	labels := genLabels(cons.DefaultTimerComponentName)
	annotations := annotationsForTimer()
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultTimerComponentName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &cons.DefaultTimerReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cons.OperatorServiceAccountName,
					Containers: []corev1.Container{{
						Name:            cons.DefaultTimerContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.DefaultTimerContainerImageName, core.Spec.Version),
						ImagePullPolicy: core.Spec.ImagePullPolicy,
						Resources:       core.Spec.Resources,
						Env:             getEnvForTimer(core),
						VolumeMounts:    getVolumeMountsForTimer(core),
					}},
					Volumes: getVolumesForTimer(core),
				},
			},
		},
	}
	// Set trigger instance as the owner and controller
	controllerutil.SetControllerReference(core, dep, r.Scheme)

	return dep
}

func getEnvForTimer(core *vanusv1alpha1.Core) []corev1.EnvVar {
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

func getVolumeMountsForTimer(core *vanusv1alpha1.Core) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.DefaultConfigMountPath,
		Name:      cons.DefaultTimerConfigMapName,
	}}
	return defaultVolumeMounts
}

func getVolumesForTimer(core *vanusv1alpha1.Core) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.DefaultTimerConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.DefaultTimerConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func annotationsForTimer() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.DefaultPortMetrics)}
}

func (r *CoreReconciler) generateConfigMapForTimer(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("name: timer\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString("etcd:\n")
	for i := int32(0); i < cons.DefaultEtcdReplicas; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-etcd-%d.vanus-etcd:%s\n", i, core.Annotations[cons.CoreComponentEtcdPortClientAnnotation]))
	}
	value.WriteString("controllers:\n")
	for i := int32(0); i < cons.DefaultControllerReplicas; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller.vanus.svc:%s\n", i, core.Annotations[cons.CoreComponentControllerSvcPortAnnotation]))
	}
	value.WriteString("leader_election:\n")
	value.WriteString("  lease_duration: 15\n")
	value.WriteString("timingwheel:\n")
	value.WriteString(fmt.Sprintf("  tick: %s\n", core.Annotations[cons.CoreComponentTimerTimingWheelTickAnnotation]))
	value.WriteString(fmt.Sprintf("  wheel_size: %s\n", core.Annotations[cons.CoreComponentTimerTimingWheelSizeAnnotation]))
	value.WriteString(fmt.Sprintf("  layers: %s\n", core.Annotations[cons.CoreComponentTimerTimingWheelLayersAnnotation]))
	data["timer.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.DefaultTimerConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}
