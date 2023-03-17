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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
)

const (
	EtcdSeparateVersion = "v0.7.0"
)

var (
	defaultWaitForReadyTimeout = 3 * time.Minute
)

// CoreReconciler reconciles a Core object
type CoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vanus.ai,resources=cores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.ai,resources=cores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.ai,resources=cores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Core object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.Log.WithName("Vanus")
	logger.Info("Reconciling Vanus.")

	// Fetch the Core instance
	core := &vanusv1alpha1.Core{}
	err := r.Get(ctx, req.NamespacedName, core)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Core resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logger.Error(err, "Failed to get Core.")
		return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
	}

	// explicitly all supported annotations
	ExplicitAnnotations(core)

	result, err := r.handleEtcd(ctx, logger, core)
	if err != nil {
		return result, err
	}
	result, err = r.handleRootController(ctx, logger, core)
	if err != nil {
		return result, err
	}
	result, err = r.handleController(ctx, logger, core)
	if err != nil {
		return result, err
	}
	result, err = r.handleStore(ctx, logger, core)
	if err != nil {
		return result, err
	}
	result, err = r.handleTrigger(ctx, logger, core)
	if err != nil {
		return result, err
	}
	result, err = r.handleTimer(ctx, logger, core)
	if err != nil {
		return result, err
	}
	result, err = r.handleGateway(ctx, logger, core)
	if err != nil {
		return result, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1alpha1.Core{}).
		Complete(r)
}

func ExplicitAnnotations(core *vanusv1alpha1.Core) {
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentEtcdPortClientAnnotation, fmt.Sprintf("%d", cons.DefaultEtcdPortClient))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentEtcdPortPeerAnnotation, fmt.Sprintf("%d", cons.DefaultEtcdPortPeer))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentEtcdStorageSizeAnnotation, cons.DefaultEtcdStorageSize)
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentControllerSvcPortAnnotation, fmt.Sprintf("%d", cons.DefaultControllerPortGrpc))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentControllerSegmentCapacityAnnotation, cons.DefaultControllerSegmentCapacity)
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentRootControllerSvcPortAnnotation, fmt.Sprintf("%d", cons.DefaultRootControllerPortGrpc))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentStoreReplicasAnnotation, fmt.Sprintf("%d", cons.DefaultStoreReplicas))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentStoreStorageSizeAnnotation, cons.DefaultStoreStorageSize)
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentGatewayPortProxyAnnotation, fmt.Sprintf("%d", cons.DefaultGatewayContainerPortProxy))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentGatewayPortCloudEventsAnnotation, fmt.Sprintf("%d", cons.DefaultGatewayContainerPortCloudevents))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentGatewayNodePortProxyAnnotation, fmt.Sprintf("%d", cons.DefaultGatewayServiceNodePortProxy))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentGatewayNodePortCloudEventsAnnotation, fmt.Sprintf("%d", cons.DefaultGatewayServiceNodePortCloudevents))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentTriggerReplicasAnnotation, fmt.Sprintf("%d", cons.DefaultTriggerReplicas))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentTimerTimingWheelTickAnnotation, fmt.Sprintf("%d", cons.DefaultTimerTimingWheelTick))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentTimerTimingWheelSizeAnnotation, fmt.Sprintf("%d", cons.DefaultTimerTimingWheelSize))
	ExplicitAnnotationWithDefaultValue(core, cons.CoreComponentTimerTimingWheelLayersAnnotation, fmt.Sprintf("%d", cons.DefaultTimerTimingWheelLayers))
}

func ExplicitAnnotationWithDefaultValue(core *vanusv1alpha1.Core, key, defaultValue string) {
	if val, ok := core.Annotations[key]; ok && val != "" {
		return
	} else {
		core.Annotations[key] = defaultValue
	}
}
