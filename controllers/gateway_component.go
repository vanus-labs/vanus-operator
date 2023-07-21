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
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
)

func (r *CoreReconciler) handleGateway(ctx context.Context, logger logr.Logger, core *vanusv1alpha1.Core) (ctrl.Result, error) {
	gateway := r.generateGateway(core)
	// Check if the statefulSet already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Gateway ConfigMap
			gatewayConfigMap := r.generateConfigMapForGateway(core)
			logger.Info("Creating a new Gateway ConfigMap.", "Namespace", gatewayConfigMap.Namespace, "Name", gatewayConfigMap.Name)
			err = r.Create(ctx, gatewayConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Gateway ConfigMap", "Namespace", gatewayConfigMap.Namespace, "Name", gatewayConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Gateway ConfigMap")
			}
			// Create Gateway Deployment
			logger.Info("Creating a new Gateway Deployment.", "Namespace", gateway.Namespace, "Name", gateway.Name)
			err = r.Create(ctx, gateway)
			if err != nil {
				logger.Error(err, "Failed to create new Gateway Deployment", "Namespace", gateway.Namespace, "Name", gateway.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Gateway Deployment")
			}
			// Create Gateway Service
			gatewaySvc := r.generateSvcForGateway(core)
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
					return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
				}
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Gateway Deployment.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Update Gateway Deployment
	logger.Info("Updating Gateway Deployment.", "Namespace", gateway.Namespace, "Name", gateway.Name)
	err = r.Update(ctx, gateway)
	if err != nil {
		logger.Error(err, "Failed to update Gateway Deployment", "Namespace", gateway.Namespace, "Name", gateway.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update Gateway Deployment")

	return ctrl.Result{}, nil
}

// returns a Gateway Deployment object
func (r *CoreReconciler) generateGateway(core *vanusv1alpha1.Core) *appsv1.Deployment {
	labels := genLabels(cons.DefaultGatewayComponentName)
	annotations := annotationsForGateway()
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultGatewayComponentName,
			Namespace: core.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &cons.DefaultGatewayReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            cons.DefaultGatewayContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.DefaultGatewayContainerImageName, core.Spec.Version),
						ImagePullPolicy: corev1.PullPolicy(core.Annotations[cons.CoreComponentImagePullPolicyAnnotation]),
						Resources:       getResourcesForGateway(core),
						Env:             getEnvForGateway(core),
						Ports:           getPortsForGateway(core),
						VolumeMounts:    getVolumeMountsForGateway(core),
					}},
					Volumes: getVolumesForGateway(core),
				},
			},
		},
	}
	// Set trigger instance as the owner and controller
	controllerutil.SetControllerReference(core, dep, r.Scheme)
	return dep
}

func getResourcesForGateway(core *vanusv1alpha1.Core) corev1.ResourceRequirements {
	limits := make(map[corev1.ResourceName]resource.Quantity)
	if val, ok := core.Annotations[cons.CoreComponentGatewayResourceLimitsCpuAnnotation]; ok && val != "" {
		limits[corev1.ResourceCPU] = resource.MustParse(core.Annotations[cons.CoreComponentGatewayResourceLimitsCpuAnnotation])
	}
	if val, ok := core.Annotations[cons.CoreComponentGatewayResourceLimitsMemAnnotation]; ok && val != "" {
		limits[corev1.ResourceMemory] = resource.MustParse(core.Annotations[cons.CoreComponentGatewayResourceLimitsMemAnnotation])
	}
	defaultResources := corev1.ResourceRequirements{
		Limits: limits,
	}
	return defaultResources
}

func getEnvForGateway(core *vanusv1alpha1.Core) []corev1.EnvVar {
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

func getPortsForGateway(core *vanusv1alpha1.Core) []corev1.ContainerPort {
	portProxy, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentGatewayPortProxyAnnotation])
	// TODO(jiangkai): Currently, the port of CloudEvents configuration is not supported, default is portProxy+1.
	// portCloudEvents, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentGatewayPortCloudEventsAnnotation])
	portCloudEvents := portProxy + 1
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameProxy,
		ContainerPort: portProxy,
	}, {
		Name:          cons.ContainerPortNameCloudevents,
		ContainerPort: portCloudEvents,
	}, {
		Name:          cons.ContainerPortNameSinkProxy,
		ContainerPort: cons.DefaultGatewayContainerPortSinkProxy,
	}}
	return defaultPorts
}

func getVolumeMountsForGateway(core *vanusv1alpha1.Core) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.DefaultConfigMountPath,
		Name:      cons.DefaultGatewayConfigMapName,
	}}
	return defaultVolumeMounts
}

func getVolumesForGateway(core *vanusv1alpha1.Core) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.DefaultGatewayConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.DefaultGatewayConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func annotationsForGateway() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.DefaultPortMetrics)}
}

func (r *CoreReconciler) generateConfigMapForGateway(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	value := bytes.Buffer{}
	value.WriteString(fmt.Sprintf("port: %s\n", core.Annotations[cons.CoreComponentGatewayPortProxyAnnotation]))
	value.WriteString(fmt.Sprintf("sink_port: %d\n", cons.DefaultGatewayContainerPortSinkProxy))
	value.WriteString("controllers:\n")
	for i := int32(0); i < cons.DefaultControllerReplicas; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller:%s\n", i, core.Annotations[cons.CoreComponentControllerSvcPortAnnotation]))
	}
	value.WriteString("observability:\n")
	value.WriteString("  metrics:\n")
	value.WriteString("    enable: true\n")
	value.WriteString("  tracing:\n")
	value.WriteString("    enable: false\n")
	value.WriteString("    # OpenTelemetry Collector endpoint, https://opentelemetry.io/docs/collector/getting-started/\n")
	value.WriteString("    otel_collector: http://127.0.0.1:4318\n")
	data := make(map[string]string)
	data["gateway.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  core.Namespace,
			Name:       cons.DefaultGatewayConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}
	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}

func (r *CoreReconciler) generateSvcForGateway(core *vanusv1alpha1.Core) *corev1.Service {
	portProxy, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentGatewayPortProxyAnnotation])
	// TODO(jiangkai): Currently, the port of CloudEvents configuration is not supported, default is portProxy+1.
	// portCloudEvents, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentGatewayPortCloudEventsAnnotation])
	portCloudEvents := portProxy + 1
	nodePortProxy, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentGatewayNodePortProxyAnnotation])
	nodeportCloudEvents, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentGatewayNodePortCloudEventsAnnotation])
	labels := genLabels(cons.DefaultGatewayComponentName)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  core.Namespace,
			Name:       cons.DefaultGatewayComponentName,
			Labels:     labels,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "proxy",
					NodePort:   nodePortProxy,
					Port:       portProxy,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(int(portProxy)),
				},
				{
					Name:       "cloudevents",
					NodePort:   nodeportCloudEvents,
					Port:       portCloudEvents,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(int(portCloudEvents)),
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
	controllerutil.SetControllerReference(core, svc, r.Scheme)
	return svc
}
