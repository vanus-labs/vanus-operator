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

package controller

import (
	"context"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	log "k8s.io/klog/v2"

	"github.com/vanus-labs/vanus-operator/pkg/apiserver/controller"
)

var (
	defaultIngressControllerWorker int = 1
)

type Controller struct {
	ctrl              *controller.Controller
	sharedInformers   informers.SharedInformerFactory
	ingressController *IngressController
}

// NewController creates a new Controller manager
func NewController(ctx context.Context, ctrl *controller.Controller) (*Controller, error) {
	sharedInformers := informers.NewSharedInformerFactory(ctrl.K8SClientSet(), time.Minute)
	controller := &Controller{
		ctrl:            ctrl,
		sharedInformers: sharedInformers,
	}
	ingressCtrl, err := NewIngressController(ctx, sharedInformers, ctrl)
	if err != nil {
		return nil, err
	}
	controller.ingressController = ingressCtrl
	return controller, nil
}

// Run begins controller.
func (c *Controller) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()

	log.Info("Starting controller manager")
	defer log.Info("Shutting down controller manager")

	go c.ingressController.Run(ctx, defaultIngressControllerWorker)
}
