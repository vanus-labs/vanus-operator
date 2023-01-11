/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"time"

	cons "github.com/linkall-labs/vanus-operator/internal/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	vanusv1alpha1 "github.com/linkall-labs/vanus-operator/api/v1alpha1"
)

// GatewayReconciler reconciles a Gateway object
type GatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vanus.linkall.com,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=gateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=gateways/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Gateway object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.Log.WithName("Gateway")
	logger.Info("Reconciling Gateway.")

	// Fetch the Gateway instance
	gateway := &vanusv1alpha1.Gateway{}
	err := r.Get(ctx, req.NamespacedName, gateway)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Gateway resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logger.Error(err, "Failed to get Gateway.")
		return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
	}

	gatewayDeployment := r.getDeploymentForGateway(gateway)
	// Create Gateway Deployment
	// Check if the Deployment already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: gatewayDeployment.Name, Namespace: gatewayDeployment.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Gateway ConfigMap
			gatewayConfigMap := r.generateConfigMapForGateway(gateway)
			logger.Info("Creating a new Gateway ConfigMap.", "ConfigMap.Namespace", gatewayConfigMap.Namespace, "ConfigMap.Name", gatewayConfigMap.Name)
			err = r.Create(ctx, gatewayConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Gateway ConfigMap", "ConfigMap.Namespace", gatewayConfigMap.Namespace, "ConfigMap.Name", gatewayConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Gateway ConfigMap")
			}
			logger.Info("Creating a new Gateway Deployment.", "Deployment.Namespace", gatewayDeployment.Namespace, "Deployment.Name", gatewayDeployment.Name)
			err = r.Create(ctx, gatewayDeployment)
			if err != nil {
				logger.Error(err, "Failed to create new Gateway Deployment", "Deployment.Namespace", gatewayDeployment.Namespace, "Deployment.Name", gatewayDeployment.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Gateway Deployment")
			}
			gatewaySvc := r.generateSvcForGateway(gateway)
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

	// TODO(jiangkai): Update Gateway Deployment
	// TODO(jiangkai): Currently, only the updated mirror version is supported
	if gateway.Spec.Image != dep.Spec.Template.Spec.Containers[0].Image {
		logger.Info("Updating Gateway Deployment.", "Deployment.Namespace", gatewayDeployment.Namespace, "Deployment.Name", gatewayDeployment.Name)
		dep.Spec.Template.Spec.Containers[0].Image = gateway.Spec.Image
		err = r.Update(ctx, dep)
		if err != nil {
			logger.Error(err, "Failed to update Gateway Deployment", "Deployment.Namespace", gatewayDeployment.Namespace, "Deployment.Name", gatewayDeployment.Name)
			return ctrl.Result{}, err
		} else {
			logger.Info("Successfully update Gateway Deployment")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1alpha1.Gateway{}).
		Complete(r)
}

func (r *GatewayReconciler) generateConfigMapForGateway(gateway *vanusv1alpha1.Gateway) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("port: 8080\n")
	value.WriteString("controllers:\n")
	// TODO(jiangkai): The timer needs to know the number of replicas of the controller，current default 3 replicas. Suggestted to use the service domain name for forwarding.
	for i := int32(0); i < 3; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller:2048\n", i))
	}
	data["gateway.yaml"] = value.String()
	gatewayConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  gateway.Namespace,
			Name:       "config-gateway",
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(gateway, gatewayConfigMap, r.Scheme)
	return gatewayConfigMap
}

// returns a Gateway Deployment object
func (r *GatewayReconciler) getDeploymentForGateway(gateway *vanusv1alpha1.Gateway) *appsv1.Deployment {
	labels := labelsForGateway(gateway.Name)
	annotations := annotationsForGateway()
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gateway.Name,
			Namespace: gateway.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: gateway.Spec.Replicas,
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
						Name:            cons.GatewayContainerName,
						Image:           gateway.Spec.Image,
						ImagePullPolicy: gateway.Spec.ImagePullPolicy,
						Resources:       gateway.Spec.Resources,
						Env:             getEnvForGateway(gateway),
						Ports:           getPortsForGateway(gateway),
						VolumeMounts:    getVolumeMountsForGateway(gateway),
					}},
					Volumes: getVolumesForGateway(gateway),
				},
			},
		},
	}
	// Set Gateway instance as the owner and controller
	controllerutil.SetControllerReference(gateway, dep, r.Scheme)

	return dep
}

func (r *GatewayReconciler) generateSvcForGateway(gateway *vanusv1alpha1.Gateway) *corev1.Service {
	labels := labelsForController(gateway.Name)
	gatewaySvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  gateway.Namespace,
			Name:       gateway.Name,
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

	controllerutil.SetControllerReference(gateway, gatewaySvc, r.Scheme)
	return gatewaySvc
}

func getEnvForGateway(gateway *vanusv1alpha1.Gateway) []corev1.EnvVar {
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

func getPortsForGateway(gateway *vanusv1alpha1.Gateway) []corev1.ContainerPort {
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

func getVolumeMountsForGateway(gateway *vanusv1alpha1.Gateway) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.ConfigMountPath,
		Name:      cons.GatewayConfigMapName,
	}}
	return defaultVolumeMounts
}

func getVolumesForGateway(gateway *vanusv1alpha1.Gateway) []corev1.Volume {
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

func labelsForGateway(name string) map[string]string {
	return map[string]string{"app": name}
}

func annotationsForGateway() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.ControllerPortMetrics)}
}
