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

func (r *CoreReconciler) handleController(ctx context.Context, logger logr.Logger, core *vanusv1alpha1.Core) (ctrl.Result, error) {
	controller := r.generateController(core)
	// Check if the statefulSet already exists, if not create a new one
	sts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: cons.DefaultControllerComponentName, Namespace: cons.DefaultNamespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Controller ConfigMap
			controllerConfigMap := r.generateConfigMapForController(core)
			logger.Info("Creating a new Controller ConfigMap.", "Namespace", controllerConfigMap.Namespace, "Name", controllerConfigMap.Name)
			err = r.Create(ctx, controllerConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Controller ConfigMap", "Namespace", controllerConfigMap.Namespace, "Name", controllerConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Controller ConfigMap")
			}
			// Create Controller StatefulSet
			logger.Info("Creating a new Controller StatefulSet.", "Namespace", controller.Namespace, "Name", controller.Name)
			err = r.Create(ctx, controller)
			if err != nil {
				logger.Error(err, "Failed to create new Controller StatefulSet", "Namespace", controller.Namespace, "Name", controller.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Controller StatefulSet")
			}
			// Create Controller Service
			controllerSvc := r.generateSvcForController(core)
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
					return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
				}
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Controller StatefulSet.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Update Controller StatefulSet
	logger.Info("Updating Controller StatefulSet.", "Namespace", controller.Namespace, "Name", controller.Name)
	err = r.Update(ctx, controller)
	if err != nil {
		logger.Error(err, "Failed to update Controller StatefulSet", "Namespace", controller.Namespace, "Name", controller.Name)
		return ctrl.Result{}, err
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
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
		if ready {
			break
		}
		select {
		case <-t.C:
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, stderr.New("controller isn't ready")
		default:
			time.Sleep(time.Second)
		}
	}
	logger.Info("Controller is ready", "WaitingTime", time.Since(start))

	return ctrl.Result{}, nil
}

// returns a Controller StatefulSet object
func (r *CoreReconciler) generateController(core *vanusv1alpha1.Core) *appsv1.StatefulSet {
	labels := genLabels(cons.DefaultControllerComponentName)
	annotations := annotationsForController()
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultControllerComponentName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &cons.DefaultControllerReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			ServiceName: cons.DefaultControllerComponentName,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            cons.DefaultControllerContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.DefaultControllerContainerImageName, core.Spec.Version),
						ImagePullPolicy: corev1.PullPolicy(core.Annotations[cons.CoreComponentImagePullPolicyAnnotation]),
						Resources:       getResourcesForController(core),
						Env:             getEnvForController(core),
						Ports:           getPortsForController(core),
						VolumeMounts:    getVolumeMountsForController(core),
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
	err := r.Get(ctx, types.NamespacedName{Name: cons.DefaultControllerComponentName, Namespace: cons.DefaultNamespace}, sts)
	if err != nil {
		return false, err
	}
	if sts.Status.Replicas == cons.DefaultControllerReplicas && sts.Status.ReadyReplicas == cons.DefaultControllerReplicas && sts.Status.AvailableReplicas == cons.DefaultControllerReplicas {
		return true, nil
	}
	return false, nil
}

func getResourcesForController(core *vanusv1alpha1.Core) corev1.ResourceRequirements {
	limits := make(map[corev1.ResourceName]resource.Quantity)
	if val, ok := core.Annotations[cons.CoreComponentControllerResourceLimitsCpuAnnotation]; ok && val != "" {
		limits[corev1.ResourceCPU] = resource.MustParse(core.Annotations[cons.CoreComponentControllerResourceLimitsCpuAnnotation])
	}
	if val, ok := core.Annotations[cons.CoreComponentControllerResourceLimitsMemAnnotation]; ok && val != "" {
		limits[corev1.ResourceMemory] = resource.MustParse(core.Annotations[cons.CoreComponentControllerResourceLimitsMemAnnotation])
	}
	defaultResources := corev1.ResourceRequirements{
		Limits: limits,
	}
	return defaultResources
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
	port, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentControllerSvcPortAnnotation])
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameGrpc,
		ContainerPort: port,
	}, {
		Name:          cons.ContainerPortNameMetrics,
		ContainerPort: cons.DefaultPortMetrics,
	}}
	return defaultPorts
}

func getVolumeMountsForController(core *vanusv1alpha1.Core) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.DefaultConfigMountPath,
		Name:      cons.DefaultControllerConfigMapName,
	}}
	return defaultVolumeMounts
}

func getCommandForController(core *vanusv1alpha1.Core) []string {
	defaultCommand := []string{"/bin/sh", "-c", "NODE_ID=${HOSTNAME##*-} /vanus/bin/controller"}
	return defaultCommand
}

func getVolumesForController(core *vanusv1alpha1.Core) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.DefaultControllerConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.DefaultControllerConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func genLabels(name string) map[string]string {
	return map[string]string{"app": name}
}

func annotationsForController() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.DefaultPortMetrics)}
}

func (r *CoreReconciler) generateConfigMapForController(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	value := bytes.Buffer{}
	value.WriteString("node_id: ${NODE_ID}\n")
	value.WriteString("name: ${POD_NAME}\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString(fmt.Sprintf("port: %s\n", core.Annotations[cons.CoreComponentControllerSvcPortAnnotation]))
	value.WriteString("replicas: 3\n")
	quantity := resource.MustParse(core.Annotations[cons.CoreComponentControllerSegmentCapacityAnnotation])
	value.WriteString(fmt.Sprintf("segment_capacity: %d\n", quantity.Value()))
	// TODO(jiangkai): automatic generation
	value.WriteString("secret_encryption_salt: encryption_salt\n")
	value.WriteString("root_controllers:\n")
	for i := int32(0); i < cons.DefaultControllerReplicas; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-root-controller-%d.vanus-root-controller:%s\n", i, core.Annotations[cons.CoreComponentRootControllerSvcPortAnnotation]))
	}
	value.WriteString("observability:\n")
	value.WriteString("  metrics:\n")
	value.WriteString("    enable: true\n")
	value.WriteString("  tracing:\n")
	value.WriteString("    enable: false\n")
	value.WriteString("    # OpenTelemetry Collector endpoint, https://opentelemetry.io/docs/collector/getting-started/\n")
	value.WriteString("    otel_collector: http://127.0.0.1:4318\n")

	value.WriteString("cluster:\n")
	value.WriteString("  component_name: controller\n")
	value.WriteString("  lease_duration_in_sec: 15\n")
	value.WriteString("  etcd:\n")
	for i := int32(0); i < cons.DefaultEtcdReplicas; i++ {
		value.WriteString(fmt.Sprintf("    - vanus-etcd-%d.vanus-etcd:%s\n", i, core.Annotations[cons.CoreComponentEtcdPortClientAnnotation]))
	}
	value.WriteString("  topology:\n")
	for i := int32(0); i < cons.DefaultControllerReplicas; i++ {
		value.WriteString(fmt.Sprintf("    vanus-controller-%d: vanus-controller-%d.vanus-controller.vanus.svc:%s\n", i, i, core.Annotations[cons.CoreComponentControllerSvcPortAnnotation]))
	}
	data := make(map[string]string)
	data["controller.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.DefaultControllerConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}

func (r *CoreReconciler) generateSvcForController(core *vanusv1alpha1.Core) *corev1.Service {
	port, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentControllerSvcPortAnnotation])
	labels := genLabels(cons.DefaultControllerComponentName)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.DefaultControllerComponentName,
			Labels:     labels,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: cons.HeadlessServiceClusterIP,
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name:       cons.DefaultControllerComponentName,
					Port:       port,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(int(port)),
				},
			},
		},
	}

	controllerutil.SetControllerReference(core, svc, r.Scheme)
	return svc
}
