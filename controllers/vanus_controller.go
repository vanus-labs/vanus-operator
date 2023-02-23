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
	"strings"
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

// VanusReconciler reconciles a Vanus object
type VanusReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vanus.vanus.ai,resources=vanus,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.vanus.ai,resources=vanus/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.vanus.ai,resources=vanus/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Vanus object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *VanusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.Log.WithName("Vanus")
	logger.Info("Reconciling Vanus.")

	// Fetch the Vanus instance
	vanus := &vanusv1alpha1.Vanus{}
	err := r.Get(ctx, req.NamespacedName, vanus)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Vanus resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logger.Error(err, "Failed to get Vanus.")
		return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
	}

	result, err := r.handleEtcd(ctx, logger, vanus)
	if err != nil {
		return result, err
	}
	result, err = r.handleController(ctx, logger, vanus)
	if err != nil {
		return result, err
	}
	result, err = r.handleStore(ctx, logger, vanus)
	if err != nil {
		return result, err
	}
	result, err = r.handleTrigger(ctx, logger, vanus)
	if err != nil {
		return result, err
	}
	result, err = r.handleTimer(ctx, logger, vanus)
	if err != nil {
		return result, err
	}
	result, err = r.handleGateway(ctx, logger, vanus)
	if err != nil {
		return result, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VanusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1alpha1.Vanus{}).
		Complete(r)
}

func version(image string) string {
	return strings.Split(image, ":")[1]
}
