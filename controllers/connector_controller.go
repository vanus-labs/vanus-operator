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
	"context"
	"time"

	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
)

// ConnectorReconciler reconciles a Connector object
type ConnectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.vanus.ai,resources=connectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.vanus.ai,resources=connectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.vanus.ai,resources=connectors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Connector object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.Log.WithName("Connector")
	logger.Info("Reconciling Connector.")

	// TODO(user): your logic here
	// Fetch the Connector instance
	connector := &vanusv1alpha1.Connector{}
	err := r.Get(ctx, req.NamespacedName, connector)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Connector resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logger.Error(err, "Failed to get Connector.")
		return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
	}

	connectorDeployment := r.getDeploymentForConnector(connector)
	// Create Connector Deployment
	// Check if the Deployment already exists, if not create a new one
	sts := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: connectorDeployment.Name, Namespace: connectorDeployment.Namespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Connector ConfigMap
			connectorConfigMap, err := r.generateConfigMapForConnector(connector)
			if err != nil {
				logger.Error(err, "Failed to generate new Connector ConfigMap")
				return ctrl.Result{}, err
			}
			logger.Info("Creating a new Connector ConfigMap.", "ConfigMap.Namespace", connectorConfigMap.Namespace, "ConfigMap.Name", connectorConfigMap.Name)
			err = r.Create(ctx, connectorConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Connector ConfigMap", "ConfigMap.Namespace", connectorConfigMap.Namespace, "ConfigMap.Name", connectorConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Connector ConfigMap")
			}
			logger.Info("Creating a new Connector Deployment.", "Deployment.Namespace", connectorDeployment.Namespace, "Deployment.Name", connectorDeployment.Name)
			err = r.Create(ctx, connectorDeployment)
			if err != nil {
				logger.Error(err, "Failed to create new Connector Deployment", "Deployment.Namespace", connectorDeployment.Namespace, "Deployment.Name", connectorDeployment.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Connector Deployment")
			}
			connectorSvc := r.generateSvcForConnector(connector)
			// Create Connector Service
			// Check if the service already exists, if not create a new one
			svc := &corev1.Service{}
			err = r.Get(ctx, types.NamespacedName{Name: connectorSvc.Name, Namespace: connectorSvc.Namespace}, svc)
			if err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Creating a new Connector Service.", "Service.Namespace", connectorSvc.Namespace, "Service.Name", connectorSvc.Name)
					err = r.Create(ctx, connectorSvc)
					if err != nil {
						logger.Error(err, "Failed to create new Connector Service", "Service.Namespace", connectorSvc.Namespace, "Service.Name", connectorSvc.Name)
						return ctrl.Result{}, err
					} else {
						logger.Info("Successfully create Connector Service")
					}
				} else {
					logger.Error(err, "Failed to get Connector Service.")
					return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
				}
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Connector Deployment.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
	}

	// TODO(jiangkai): Update Connector Deployment
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1alpha1.Connector{}).
		Complete(r)
}

// returns a Connector Deployment object
func (r *ConnectorReconciler) getDeploymentForConnector(connector *vanusv1alpha1.Connector) *appsv1.Deployment {
	replicas := int32(1)
	labels := labelsForConnector(connector.Name)
	requests := make(map[corev1.ResourceName]resource.Quantity)
	requests[corev1.ResourceCPU] = resource.MustParse("100m")
	requests[corev1.ResourceMemory] = resource.MustParse("128Mi")
	limits := make(map[corev1.ResourceName]resource.Quantity)
	limits[corev1.ResourceCPU] = resource.MustParse("500m")
	limits[corev1.ResourceMemory] = resource.MustParse("512Mi")
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connector.Name,
			Namespace: connector.Namespace,
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
					Annotations: connector.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            cons.ConnectorContainerName,
						Image:           connector.Spec.Image,
						ImagePullPolicy: connector.Spec.ImagePullPolicy,
						Resources: corev1.ResourceRequirements{
							Requests: requests,
							Limits:   limits,
						},
						Env:          getEnvForConnector(connector),
						VolumeMounts: getVolumeMountsForConnector(connector),
					}},
					Volumes: getVolumesForConnector(connector),
				},
			},
		},
	}
	// Set Connector instance as the owner and connector
	controllerutil.SetControllerReference(connector, dep, r.Scheme)

	return dep
}

func getEnvForConnector(connector *vanusv1alpha1.Connector) []corev1.EnvVar {
	defaultEnvs := []corev1.EnvVar{{
		Name:  cons.EnvLogLevel,
		Value: "INFO",
	}}
	return defaultEnvs
}

func getVolumeMountsForConnector(connector *vanusv1alpha1.Connector) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		Name:      cons.VanceConfigMapName,
		MountPath: cons.VanceConfigMountPath,
	}}
	return defaultVolumeMounts
}

func getVolumesForConnector(connector *vanusv1alpha1.Connector) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.VanceConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: connector.Name,
				},
			}},
	}}
	return defaultVolumes
}

func labelsForConnector(name string) map[string]string {
	return map[string]string{"app": name}
}

func (r *ConnectorReconciler) generateSvcForConnector(connector *vanusv1alpha1.Connector) *corev1.Service {
	labels := labelsForConnector(connector.Name)
	connectorSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  connector.Namespace,
			Name:       connector.Name,
			Labels:     labels,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: cons.HeadlessService,
			Selector:  labels,
			Type:      corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: connector.Name,
					Port: 8080,
				},
			},
		},
	}

	controllerutil.SetControllerReference(connector, connectorSvc, r.Scheme)
	return connectorSvc
}

func (r *ConnectorReconciler) generateConfigMapForConnector(connector *vanusv1alpha1.Connector) (*corev1.ConfigMap, error) {
	labels := labelsForConnector(connector.Name)
	data := make(map[string]string)
	data["config.yml"] = connector.Spec.Config
	connectorConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       connector.Name,
			Labels:     labels,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(connector, connectorConfigMap, r.Scheme)
	return connectorConfigMap, nil
}
