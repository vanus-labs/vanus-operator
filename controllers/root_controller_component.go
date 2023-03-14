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

func (r *CoreReconciler) handleRootController(ctx context.Context, logger logr.Logger, core *vanusv1alpha1.Core) (ctrl.Result, error) {
	rootController := r.generateRootController(core)
	// Check if the statefulSet already exists, if not create a new one
	sts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: cons.DefaultRootControllerName, Namespace: cons.DefaultNamespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create rootController ConfigMap
			rootControllerConfigMap := r.generateConfigMapForRootController(core)
			logger.Info("Creating a new rootController ConfigMap.", "Namespace", rootControllerConfigMap.Namespace, "Name", rootControllerConfigMap.Name)
			err = r.Create(ctx, rootControllerConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new rootController ConfigMap", "Namespace", rootControllerConfigMap.Namespace, "Name", rootControllerConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create rootController ConfigMap")
			}
			// Create rootController StatefulSet
			logger.Info("Creating a new rootController StatefulSet.", "Namespace", rootController.Namespace, "Name", rootController.Name)
			err = r.Create(ctx, rootController)
			if err != nil {
				logger.Error(err, "Failed to create new rootController StatefulSet", "Namespace", rootController.Namespace, "Name", rootController.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create rootController StatefulSet")
			}
			// Create rootController Service
			rootControllerSvc := r.generateSvcForRootController(core)
			// Check if the service already exists, if not create a new one
			svc := &corev1.Service{}
			err = r.Get(ctx, types.NamespacedName{Name: rootControllerSvc.Name, Namespace: rootControllerSvc.Namespace}, svc)
			if err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Creating a new rootController Service.", "Namespace", rootControllerSvc.Namespace, "Name", rootControllerSvc.Name)
					err = r.Create(ctx, rootControllerSvc)
					if err != nil {
						logger.Error(err, "Failed to create new rootController Service", "Namespace", rootControllerSvc.Namespace, "Name", rootControllerSvc.Name)
						return ctrl.Result{}, err
					} else {
						logger.Info("Successfully create rootController Service")
					}
				} else {
					logger.Error(err, "Failed to get rootController Service.")
					return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
				}
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get rootController StatefulSet.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Update rootController StatefulSet
	logger.Info("Updating rootController StatefulSet.", "Namespace", rootController.Namespace, "Name", rootController.Name)
	err = r.Update(ctx, rootController)
	if err != nil {
		logger.Error(err, "Failed to update rootController StatefulSet", "Namespace", rootController.Namespace, "Name", rootController.Name)
		return ctrl.Result{}, err
	}
	logger.Info("Successfully update rootController StatefulSet")

	// Wait for rootController is ready
	start := time.Now()
	logger.Info("Wait for rootController is ready")
	t := time.NewTicker(defaultWaitForReadyTimeout)
	defer t.Stop()
	for {
		ready, err := r.waitRootControllerIsReady(ctx, core)
		if err != nil {
			logger.Error(err, "Wait for rootController is ready but got error")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
		if ready {
			break
		}
		select {
		case <-t.C:
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, stderr.New("root-controller isn't ready")
		default:
			time.Sleep(time.Second)
		}
	}
	logger.Info("rootController is ready", "WaitingTime", time.Since(start))

	return ctrl.Result{}, nil
}

// returns a rootController StatefulSet object
func (r *CoreReconciler) generateRootController(core *vanusv1alpha1.Core) *appsv1.StatefulSet {
	labels := genLabels(cons.DefaultRootControllerName)
	annotations := annotationsForRootController()
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultRootControllerName,
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
			ServiceName: cons.DefaultRootControllerName,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            cons.RootControllerContainerName,
						Image:           fmt.Sprintf("%s:%s", cons.RootControllerImageName, core.Spec.Version),
						ImagePullPolicy: core.Spec.ImagePullPolicy,
						Resources:       core.Spec.Resources,
						Env:             getEnvForRootController(core),
						Ports:           getPortsForRootController(core),
						VolumeMounts:    getVolumeMountsForRootController(core),
						Command:         getCommandForRootController(core),
					}},
					Volumes: getVolumesForRootController(core),
				},
			},
		},
	}
	// Set rootController instance as the owner and root-controller
	controllerutil.SetControllerReference(core, sts, r.Scheme)

	return sts
}

func (r *CoreReconciler) waitRootControllerIsReady(ctx context.Context, core *vanusv1alpha1.Core) (bool, error) {
	sts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: cons.DefaultRootControllerName, Namespace: cons.DefaultNamespace}, sts)
	if err != nil {
		return false, err
	}
	if sts.Status.Replicas == cons.DefaultControllerReplicas && sts.Status.ReadyReplicas == cons.DefaultControllerReplicas && sts.Status.AvailableReplicas == cons.DefaultControllerReplicas {
		return true, nil
	}
	return false, nil
}

func getEnvForRootController(core *vanusv1alpha1.Core) []corev1.EnvVar {
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

func getPortsForRootController(core *vanusv1alpha1.Core) []corev1.ContainerPort {
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameGrpc,
		ContainerPort: cons.RootControllerPortGrpc,
	}, {
		Name:          cons.ContainerPortNameMetrics,
		ContainerPort: cons.RootControllerPortMetrics,
	}}
	return defaultPorts
}

func getVolumeMountsForRootController(core *vanusv1alpha1.Core) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.ConfigMountPath,
		Name:      cons.RootControllerConfigMapName,
	}}
	return defaultVolumeMounts
}

func getCommandForRootController(core *vanusv1alpha1.Core) []string {
	defaultCommand := []string{"/bin/sh", "-c", "NODE_ID=${HOSTNAME##*-} /vanus/bin/root-controller"}
	return defaultCommand
}

func getVolumesForRootController(core *vanusv1alpha1.Core) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.RootControllerConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.RootControllerConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func annotationsForRootController() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.RootControllerPortMetrics)}
}

func (r *CoreReconciler) generateConfigMapForRootController(core *vanusv1alpha1.Core) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("node_id: ${NODE_ID}\n")
	value.WriteString("name: ${POD_NAME}\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString(fmt.Sprintf("port: %d\n", cons.RootControllerPortGrpc))
	value.WriteString("observability:\n")
	value.WriteString("  metrics:\n")
	value.WriteString("    enable: true\n")
	value.WriteString("  tracing:\n")
	value.WriteString("    enable: false\n")
	value.WriteString("    # OpenTelemetry Collector endpoint, https://opentelemetry.io/docs/collector/getting-started/\n")
	value.WriteString("    otel_collector: http://127.0.0.1:4318\n")

	value.WriteString("cluster:\n")
	value.WriteString("  component_name: root-controller\n")
	value.WriteString("  lease_duration_in_sec: 15\n")
	value.WriteString("  etcd:\n")
	for i := int32(0); i < 3; i++ {
		value.WriteString(fmt.Sprintf("    - vanus-etcd-%d.vanus-etcd:%d\n", i, cons.EtcdPortClient))
	}
	value.WriteString("  topology:\n")
	for i := int32(0); i < cons.DefaultControllerReplicas; i++ {
		value.WriteString(fmt.Sprintf("    vanus-root-controller-%d: vanus-root-controller-%d.vanus-root-controller.vanus.svc:%d\n", i, i, cons.RootControllerPortGrpc))
	}
	data["root.yaml"] = value.String()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.RootControllerConfigMapName,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(core, cm, r.Scheme)
	return cm
}

func (r *CoreReconciler) generateSvcForRootController(core *vanusv1alpha1.Core) *corev1.Service {
	labels := genLabels(cons.DefaultRootControllerName)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.DefaultRootControllerName,
			Labels:     labels,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: cons.HeadlessService,
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name:       cons.DefaultRootControllerName,
					Port:       cons.RootControllerPortGrpc,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(cons.RootControllerPortGrpc),
				},
			},
		},
	}

	controllerutil.SetControllerReference(core, svc, r.Scheme)
	return svc
}
