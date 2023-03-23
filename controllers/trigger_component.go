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
	"github.com/vanus-labs/vanus-operator/internal/convert"
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

func (r *CoreReconciler) handleTrigger(ctx context.Context, logger logr.Logger, core *vanusv1alpha1.Core) (ctrl.Result, error) {
	trigger := r.generateTrigger(core)
	// Check if the statefulSet already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: trigger.Name, Namespace: trigger.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Trigger ConfigMap
			triggerConfigMap := r.generateConfigMapForTrigger(core)
			logger.Info("Creating a new Trigger ConfigMap.", "Namespace", triggerConfigMap.Namespace, "Name", triggerConfigMap.Name)
			err = r.Create(ctx, triggerConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Trigger ConfigMap", "Namespace", triggerConfigMap.Namespace, "Name", triggerConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Trigger ConfigMap")
			}
			// Create Trigger Deployment
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
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Update Trigger Deployment
	logger.Info("Updating Trigger Deployment.", "Namespace", trigger.Namespace, "Name", trigger.Name)
	err = r.Update(ctx, trigger)
	if err != nil {
		logger.Error(err, "Failed to update Trigger Deployment", "Namespace", trigger.Namespace, "Name", trigger.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update Trigger Deployment")
	return ctrl.Result{}, nil
}

// returns a Trigger Deployment object
func (r *CoreReconciler) generateTrigger(core *vanusv1alpha1.Core) *appsv1.Deployment {
	labels := genLabels(cons.DefaultTriggerComponentName)
	annotations := annotationsForTrigger()
	replicas, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentTriggerReplicasAnnotation])
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultTriggerComponentName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
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
						Name:            cons.DefaultTriggerContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.DefaultTriggerContainerImageName, core.Spec.Version),
						ImagePullPolicy: corev1.PullPolicy(core.Annotations[cons.CoreComponentImagePullPolicyAnnotation]),
						Resources:       getResourcesForTrigger(core),
						Env:             getEnvForTrigger(core),
						Ports:           getPortsForTrigger(core),
						VolumeMounts:    getVolumeMountsForTrigger(core),
					}},
					Volumes: getVolumesForTrigger(core),
				},
			},
		},
	}
	// Set Trigger instance as the owner and controller
	controllerutil.SetControllerReference(core, dep, r.Scheme)
	return dep
}

func getResourcesForTrigger(core *vanusv1alpha1.Core) corev1.ResourceRequirements {
	limits := make(map[corev1.ResourceName]resource.Quantity)
	if val, ok := core.Annotations[cons.CoreComponentTriggerResourceLimitsCpuAnnotation]; ok && val != "" {
		limits[corev1.ResourceCPU] = resource.MustParse(core.Annotations[cons.CoreComponentTriggerResourceLimitsCpuAnnotation])
	}
	if val, ok := core.Annotations[cons.CoreComponentTriggerResourceLimitsMemAnnotation]; ok && val != "" {
		limits[corev1.ResourceMemory] = resource.MustParse(core.Annotations[cons.CoreComponentTriggerResourceLimitsMemAnnotation])
	}
	defaultResources := corev1.ResourceRequirements{
		Limits: limits,
	}
	return defaultResources
}

func getEnvForTrigger(core *vanusv1alpha1.Core) []corev1.EnvVar {
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

func getPortsForTrigger(core *vanusv1alpha1.Core) []corev1.ContainerPort {
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameGrpc,
		ContainerPort: cons.DefaultTriggerContainerPortGrpc,
	}}
	return defaultPorts
}

func getVolumeMountsForTrigger(core *vanusv1alpha1.Core) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.DefaultConfigMountPath,
		Name:      cons.DefaultTriggerConfigMapName,
	}}
	return defaultVolumeMounts
}

func getVolumesForTrigger(core *vanusv1alpha1.Core) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.DefaultTriggerConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.DefaultTriggerConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func annotationsForTrigger() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.DefaultPortMetrics)}
}

func (r *CoreReconciler) generateConfigMapForTrigger(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString(fmt.Sprintf("port: %d\n", cons.DefaultTriggerContainerPortGrpc))
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString("controllers:\n")
	for i := int32(0); i < cons.DefaultControllerReplicas; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller.vanus.svc:%s\n", i, core.Annotations[cons.CoreComponentControllerSvcPortAnnotation]))
	}
	data["trigger.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.DefaultTriggerConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}
