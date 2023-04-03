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
	"fmt"
	"time"

	cons "github.com/vanus-labs/vanus-operator/internal/constants"
	"github.com/vanus-labs/vanus-operator/internal/convert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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

//+kubebuilder:rbac:groups=vanus.ai,resources=connectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.ai,resources=connectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.ai,resources=connectors/finalizers,verbs=update

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
	// logger.Info("Reconciling Connector.")

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
		return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
	}

	// explicitly all supported annotations
	ExplicitConnectorAnnotations(connector)

	// Create Connector Deployment
	// Check if the Deployment already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: connector.Name, Namespace: connector.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Connector ConfigMap
			connectorConfigMap := r.generateConfigMapForConnector(connector)
			logger.Info("Creating a new Connector ConfigMap.", "ConfigMap.Namespace", connectorConfigMap.Namespace, "ConfigMap.Name", connectorConfigMap.Name)
			err = r.Create(ctx, connectorConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Connector ConfigMap", "ConfigMap.Namespace", connectorConfigMap.Namespace, "ConfigMap.Name", connectorConfigMap.Name)
				return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
			} else {
				logger.Info("Successfully create Connector ConfigMap")
			}

			// Create Connector Deployment
			connectorDeployment := r.generateDeploymentForConnector(connector)
			logger.Info("Creating a new Connector Deployment.", "Deployment.Namespace", connectorDeployment.Namespace, "Deployment.Name", connectorDeployment.Name)
			err = r.Create(ctx, connectorDeployment)
			if err != nil {
				logger.Error(err, "Failed to create new Connector Deployment", "Deployment.Namespace", connectorDeployment.Namespace, "Deployment.Name", connectorDeployment.Name)
				return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
			} else {
				logger.Info("Successfully create Connector Deployment")
			}

			// Create Connector Service
			// Check if the service already exists, if not create a new one
			connectorSvc := r.generateSvcForConnector(connector)
			svc := &corev1.Service{}
			err = r.Get(ctx, types.NamespacedName{Name: connectorSvc.Name, Namespace: connectorSvc.Namespace}, svc)
			if err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Creating a new Connector Service.", "Service.Namespace", connectorSvc.Namespace, "Service.Name", connectorSvc.Name)
					err = r.Create(ctx, connectorSvc)
					if err != nil {
						logger.Error(err, "Failed to create new Connector Service", "Service.Namespace", connectorSvc.Namespace, "Service.Name", connectorSvc.Name)
						return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
					} else {
						logger.Info("Successfully create Connector Service")
					}
				} else {
					logger.Error(err, "Failed to get Connector Service.")
					return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
				}
			}

			// requeue
			return ctrl.Result{Requeue: true}, err
		} else {
			logger.Error(err, "Failed to get Connector Deployment.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Check And Update Connector
	needUpdateConnector, err := r.isNeedUpdateConnector(ctx, connector)
	if err != nil {
		logger.Error(err, "Failed to check Connector", "Connector.Namespace", connector.Namespace, "Connector.Name", connector.Name)
		return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
	}
	if needUpdateConnector {
		err = r.Update(ctx, connector)
		if err != nil {
			logger.Error(err, "Failed to update Connector", "Connector.Namespace", connector.Namespace, "Connector.Name", connector.Name)
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
		logger.Info("Successfully Update Connector.", "Connector.Namespace", connector.Namespace, "Connector.Name", connector.Name)
		return ctrl.Result{Requeue: true}, err
	}

	// Check And Update Ingress
	needUpdateIngress, err := r.isNeedUpdateIngress(ctx, connector)
	if err != nil {
		logger.Error(err, "Failed to get Ingress", "Connector.Namespace", connector.Namespace, "Connector.Name", connector.Name)
		return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
	}
	if needUpdateIngress {
		ingress := &networkingv1.Ingress{}
		err = r.Get(ctx, types.NamespacedName{Name: connector.Annotations[cons.ConnectorIngressNameAnnotation], Namespace: cons.DefaultNamespace}, ingress)
		if err != nil {
			logger.Error(err, "Failed to get operator Ingress.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
		updateIngress := r.generateIngressForConnector(connector, ingress)
		err = r.Update(ctx, updateIngress)
		if err != nil {
			logger.Error(err, "Failed to update Ingress", "Ingress.Namespace", updateIngress.Namespace, "Ingress.Name", updateIngress.Name)
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
		logger.Info("Successfully update Ingress")
		return ctrl.Result{Requeue: true}, err
	}

	// Update Connector Deployment
	connectorDeployment := r.generateDeploymentForConnector(connector)
	err = r.Update(ctx, connectorDeployment)
	if err != nil {
		logger.Error(err, "Failed to update Connector Deployment", "Namespace", connectorDeployment.Namespace, "Name", connectorDeployment.Name)
		return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
	}
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1alpha1.Connector{}).
		Complete(r)
}

// returns a Connector Deployment object
func (r *ConnectorReconciler) generateDeploymentForConnector(connector *vanusv1alpha1.Connector) *appsv1.Deployment {
	replicas, _ := convert.StrToInt32(connector.Annotations[cons.ConnectorDeploymentReplicasAnnotation])
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
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            cons.DefaultConnectorContainerName,
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
	if val, ok := connector.Annotations[cons.ConnectorRestartAtAnnotation]; ok {
		dep.Spec.Template.Annotations = map[string]string{cons.ConnectorRestartAtAnnotation: val}
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
		Name:      cons.DefaultConnectorConfigMapName,
		MountPath: cons.DefaultConnectorConfigMountPath,
	}}
	return defaultVolumeMounts
}

func getVolumesForConnector(connector *vanusv1alpha1.Connector) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.DefaultConnectorConfigMapName,
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

func (r *ConnectorReconciler) generateConfigMapForConnector(connector *vanusv1alpha1.Connector) *corev1.ConfigMap {
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
	return connectorConfigMap
}

func (r *ConnectorReconciler) generateSvcForConnector(connector *vanusv1alpha1.Connector) *corev1.Service {
	labels := labelsForConnector(connector.Name)
	svcPort, _ := convert.StrToInt32(connector.Annotations[cons.ConnectorServicePortAnnotation])
	connectorSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   connector.Namespace,
			Name:        connector.Name,
			Labels:      labels,
			Annotations: connector.Annotations,
			Finalizers:  []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceType(connector.Annotations[cons.ConnectorServiceTypeAnnotation]),
			Ports: []corev1.ServicePort{
				{
					Name:       connector.Name,
					Port:       svcPort,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}

	controllerutil.SetControllerReference(connector, connectorSvc, r.Scheme)
	return connectorSvc
}

func (r *ConnectorReconciler) generateIngressForConnector(connector *vanusv1alpha1.Connector, ingress *networkingv1.Ingress) *networkingv1.Ingress {
	svcPort, _ := convert.StrToInt32(connector.Annotations[cons.ConnectorServicePortAnnotation])
	pathType := networkingv1.PathTypePrefix
	var httpIngressPath networkingv1.HTTPIngressPath = networkingv1.HTTPIngressPath{
		Path:     "/",
		PathType: &pathType,
		Backend: networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: connector.Name,
				Port: networkingv1.ServiceBackendPort{
					Number: svcPort,
				},
			},
		},
	}
	var ingressRuleValue = networkingv1.IngressRuleValue{
		HTTP: &networkingv1.HTTPIngressRuleValue{
			Paths: []networkingv1.HTTPIngressPath{httpIngressPath},
		},
	}
	var ingressRule = networkingv1.IngressRule{
		Host:             connector.Annotations[cons.ConnectorNetworkHostDomainAnnotation],
		IngressRuleValue: ingressRuleValue,
	}

	ingress.Spec.Rules = append(ingress.Spec.Rules, ingressRule)
	return ingress
}

func (r *ConnectorReconciler) isNeedUpdateConnector(ctx context.Context, connector *vanusv1alpha1.Connector) (bool, error) {
	need := false
	if _, ok := connector.Annotations[cons.ConnectorNetworkHostDomainAnnotation]; ok {
		if _, ok := connector.Annotations[cons.ConnectorIngressNameAnnotation]; !ok {
			connector.Annotations[cons.ConnectorIngressNameAnnotation] = cons.DefaultVanusOperatorName
			need = true
		}
		if _, ok := connector.Labels[cons.ConnectorNetworkHostDomainAnnotation]; !ok {
			connector.Labels[cons.ConnectorNetworkHostDomainAnnotation] = connector.Annotations[cons.ConnectorNetworkHostDomainAnnotation]
			need = true
		}
	}

	// Get Connector Configmap
	oriConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: connector.Name, Namespace: connector.Namespace}, oriConfigMap)
	if err != nil {
		return false, err
	}
	if oriConfigMap.Data["config.yml"] != connector.Spec.Config {
		// Update Connector Configmap
		connectorConfigMap := r.generateConfigMapForConnector(connector)
		err = r.Update(ctx, connectorConfigMap)
		if err != nil {
			return false, err
		}
		connector.Annotations[cons.ConnectorRestartAtAnnotation] = time.Now().Format("2006-01-02T15:04:05Z")
		need = true
	}
	return need, nil
}

func (r *ConnectorReconciler) isNeedUpdateIngress(ctx context.Context, connector *vanusv1alpha1.Connector) (bool, error) {
	if _, ok := connector.Annotations[cons.ConnectorNetworkHostDomainAnnotation]; !ok {
		return false, nil
	}
	ingress := &networkingv1.Ingress{}
	err := r.Get(ctx, types.NamespacedName{Name: connector.Annotations[cons.ConnectorIngressNameAnnotation], Namespace: cons.DefaultNamespace}, ingress)
	if err != nil {
		return false, err
	}
	for _, rule := range ingress.Spec.Rules {
		if rule.Host == connector.Annotations[cons.ConnectorNetworkHostDomainAnnotation] {
			return false, nil
		}
	}
	return true, nil
}

func ExplicitConnectorAnnotations(connector *vanusv1alpha1.Connector) {
	if connector.Annotations == nil {
		connector.Annotations = make(map[string]string)
	}
	ExplicitConectorAnnotationWithDefaultValue(connector, cons.ConnectorDeploymentReplicasAnnotation, fmt.Sprintf("%d", cons.DefaultConnectorReplicas))
	ExplicitConectorAnnotationWithDefaultValue(connector, cons.ConnectorServiceTypeAnnotation, cons.DefaultConnectorServiceType)
	ExplicitConectorAnnotationWithDefaultValue(connector, cons.ConnectorServicePortAnnotation, fmt.Sprintf("%d", cons.DefaultConnectorServicePort))
}

func ExplicitConectorAnnotationWithDefaultValue(connector *vanusv1alpha1.Connector, key, defaultValue string) {
	if val, ok := connector.Annotations[key]; ok && val != "" {
		return
	} else {
		connector.Annotations[key] = defaultValue
	}
}
