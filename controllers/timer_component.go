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
	"strings"
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
	var (
		timerConfigMap *corev1.ConfigMap
	)
	if strings.Compare(core.Spec.Version, EtcdSeparateVersion) < 0 {
		timerConfigMap = r.generateConfigMapForTimer(core)
	} else {
		timerConfigMap = r.generateConfigMapForNewTimer(core)
	}
	// Create Timer Deployment
	// Check if the statefulSet already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: cons.DefaultTimerName, Namespace: cons.DefaultNamespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Timer ConfigMap
			logger.Info("Creating a new Timer ConfigMap.", "Namespace", timerConfigMap.Namespace, "Name", timerConfigMap.Name)
			err = r.Create(ctx, timerConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Timer ConfigMap", "Namespace", timerConfigMap.Namespace, "Name", timerConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Timer ConfigMap")
			}
			logger.Info("Creating a new Timer Deployment.", "Namespace", cons.DefaultNamespace, "Name", cons.DefaultTimerName)
			timer := r.generateTimer(core)
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
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
	}

	logger.Info("Updating Timer ConfigMap.", "Namespace", timerConfigMap.Namespace, "Name", timerConfigMap.Name)
	err = r.Update(ctx, timerConfigMap)
	if err != nil {
		logger.Error(err, "Failed to update Timer ConfigMap", "Namespace", timerConfigMap.Namespace, "Name", timerConfigMap.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update Timer ConfigMap")

	// Update Timer Deployment
	updateTimer(dep, core)
	err = r.Update(ctx, dep)
	if err != nil {
		logger.Error(err, "Failed to update Timer Deployment", "Namespace", dep.Namespace, "Name", dep.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update Timer Deployment")
	return ctrl.Result{}, nil
}

// returns a Timer Deployment object
func (r *CoreReconciler) generateTimer(core *vanusv1alpha1.Core) *appsv1.Deployment {
	labels := genLabels(cons.DefaultTimerName)
	annotations := annotationsForTimer()
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultTimerName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &core.Spec.Replicas.Timer,
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
						Name:            cons.TimerContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.TimerImageName, core.Spec.Version),
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

func updateTimer(dep *appsv1.Deployment, core *vanusv1alpha1.Core) {
	dep.Spec.Replicas = &core.Spec.Replicas.Timer
	dep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", cons.TimerImageName, core.Spec.Version)
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
		MountPath: cons.ConfigMountPath,
		Name:      cons.TimerConfigMapName,
	}}
	return defaultVolumeMounts
}

func getVolumesForTimer(core *vanusv1alpha1.Core) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.TimerConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.TimerConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func annotationsForTimer() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.ControllerPortMetrics)}
}

func (r *CoreReconciler) generateConfigMapForTimer(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("name: timer\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString("etcd:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller:2379\n", i))
	}
	value.WriteString("controllers:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller.vanus.svc:2048\n", i))
	}
	value.WriteString("metadata:\n")
	value.WriteString("  key_prefix: /vanus\n")
	value.WriteString("leaderelection:\n")
	value.WriteString("  lease_duration: 15\n")
	value.WriteString("timingwheel:\n")
	value.WriteString("  tick: 1\n")
	value.WriteString("  wheel_size: 32\n")
	value.WriteString("  layers: 4\n")
	data["timer.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.TimerConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}

func (r *CoreReconciler) generateConfigMapForNewTimer(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("name: timer\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString("etcd:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-etcd-%d.vanus-etcd:2379\n", i))
	}
	value.WriteString("controllers:\n")
	for i := int32(0); i < core.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller.vanus.svc:2048\n", i))
	}
	value.WriteString("metadata:\n")
	value.WriteString("  key_prefix: /vanus\n")
	value.WriteString("leader_election:\n")
	value.WriteString("  lease_duration: 15\n")
	value.WriteString("timingwheel:\n")
	value.WriteString("  tick: 1\n")
	value.WriteString("  wheel_size: 32\n")
	value.WriteString("  layers: 4\n")
	data["timer.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.TimerConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}
