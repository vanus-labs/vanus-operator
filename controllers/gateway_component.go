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
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
)

func (r *VanusReconciler) handleGateway(ctx context.Context, logger logr.Logger, vanus *vanusv1alpha1.Vanus) (ctrl.Result, error) {
	gateway := r.generateGateway(vanus)
	// Create Gateway Deployment
	// Check if the statefulSet already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Gateway ConfigMap
			gatewayConfigMap := r.generateConfigMapForGateway(vanus)
			logger.Info("Creating a new Gateway ConfigMap.", "Namespace", gatewayConfigMap.Namespace, "Name", gatewayConfigMap.Name)
			err = r.Create(ctx, gatewayConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Gateway ConfigMap", "Namespace", gatewayConfigMap.Namespace, "Name", gatewayConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Gateway ConfigMap")
			}
			logger.Info("Creating a new Gateway Deployment.", "Namespace", gateway.Namespace, "Name", gateway.Name)
			err = r.Create(ctx, gateway)
			if err != nil {
				logger.Error(err, "Failed to create new Gateway Deployment", "Namespace", gateway.Namespace, "Name", gateway.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Gateway Deployment")
			}
			gatewaySvc := r.generateSvcForGateway(vanus)
			// Create Gateway Service
			// Check if the service already exists, if not create a new one
			svc := &corev1.Service{}
			err = r.Get(ctx, types.NamespacedName{Name: gatewaySvc.Name, Namespace: gatewaySvc.Namespace}, svc)
			if err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Creating a new Gateway Service.", "Service.Namespace", gatewaySvc.Namespace, "Service.Name", gatewaySvc.Name)
					err = r.Create(ctx, gatewaySvc)
					if err != nil {
						logger.Error(err, "Failed to create new Gateway Service", "Service.Namespace", gatewaySvc.Namespace, "Service.Name", gatewaySvc.Name)
						return ctrl.Result{}, err
					} else {
						logger.Info("Successfully create Gateway Service")
					}
				} else {
					logger.Error(err, "Failed to get Gateway Service.")
					return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
				}
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Gateway Deployment.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Update Gateway StatefulSet
	err = r.Update(ctx, gateway)
	if err != nil {
		logger.Error(err, "Failed to update Gateway Deployment", "Namespace", gateway.Namespace, "Name", gateway.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update Gateway Deployment")

	return ctrl.Result{}, nil
}

// returns a Gateway Deployment object
func (r *VanusReconciler) generateGateway(vanus *vanusv1alpha1.Vanus) *appsv1.Deployment {
	labels := genLabels(cons.DefaultGatewayName)
	annotations := annotationsForGateway()
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultGatewayName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &vanus.Spec.Replicas.Gateway,
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
						Name:            cons.GatewayContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.GatewayImageName, vanus.Spec.Version),
						ImagePullPolicy: vanus.Spec.ImagePullPolicy,
						Resources:       vanus.Spec.Resources,
						Env:             getEnvForGateway(vanus),
						Ports:           getPortsForGateway(vanus),
						VolumeMounts:    getVolumeMountsForGateway(vanus),
					}},
					Volumes: getVolumesForGateway(vanus),
				},
			},
		},
	}
	// Set trigger instance as the owner and controller
	controllerutil.SetControllerReference(vanus, dep, r.Scheme)

	return dep
}

func getEnvForGateway(vanus *vanusv1alpha1.Vanus) []corev1.EnvVar {
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

func getPortsForGateway(vanus *vanusv1alpha1.Vanus) []corev1.ContainerPort {
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameProxy,
		ContainerPort: cons.GatewayPortProxy,
	}, {
		Name:          cons.ContainerPortNameCloudevents,
		ContainerPort: cons.GatewayPortCloudevents,
	}, {
		Name:          cons.ContainerPortNameSinkProxy,
		ContainerPort: cons.GatewayPortSinkProxy,
	}}
	return defaultPorts
}

func getVolumeMountsForGateway(vanus *vanusv1alpha1.Vanus) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.ConfigMountPath,
		Name:      cons.GatewayConfigMapName,
	}}
	return defaultVolumeMounts
}

func getVolumesForGateway(vanus *vanusv1alpha1.Vanus) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.GatewayConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.GatewayConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func annotationsForGateway() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.ControllerPortMetrics)}
}

func (r *VanusReconciler) generateConfigMapForGateway(vanus *vanusv1alpha1.Vanus) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("port: 8080\n")
	value.WriteString("controllers:\n")
	for i := int32(0); i < vanus.Spec.Replicas.Controller; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller:2048\n", i))
	}
	data["gateway.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.GatewayConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(vanus, cm, r.Scheme)
	return cm
}

func (r *VanusReconciler) generateSvcForGateway(vanus *vanusv1alpha1.Vanus) *corev1.Service {
	labels := genLabels(cons.DefaultGatewayName)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.DefaultGatewayName,
			Labels:     labels,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "proxy",
					NodePort:   cons.GatewayNodePortProxy,
					Port:       cons.GatewayPortProxy,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(cons.GatewayPortProxy),
				},
				{
					Name:       "cloudevents",
					NodePort:   cons.GatewayNodePortCloudevents,
					Port:       cons.GatewayPortCloudevents,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(cons.GatewayPortCloudevents),
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	controllerutil.SetControllerReference(vanus, svc, r.Scheme)
	return svc
}
