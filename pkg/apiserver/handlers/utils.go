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

package handlers

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// PodStatusStatusPending captures enum value "Pending"
	PodStatusStatusPending string = "Pending"

	// PodStatusStatusRunning captures enum value "Running"
	PodStatusStatusRunning string = "Running"

	// PodStatusStatusSucceeded captures enum value "Succeeded"
	PodStatusStatusSucceeded string = "Succeeded"

	// PodStatusStatusStarting captures enum value "Starting"
	PodStatusStatusStarting string = "Starting"

	// PodStatusStatusFailed captures enum value "Failed"
	PodStatusStatusFailed string = "Failed"

	// PodStatusStatusRemoving captures enum value "Removing"
	PodStatusStatusRemoving string = "Removing"

	// PodStatusStatusUnknown captures enum value "Unknown"
	PodStatusStatusUnknown string = "Unknown"
)

func statusCheck(a *corev1.Pod) (string, string) {
	if a == nil {
		return PodStatusStatusUnknown, ""
	}
	if a.DeletionTimestamp != nil {
		return PodStatusStatusRemoving, ""
	}
	// Status: Pending/Succeeded/Failed/Unknown
	if a.Status.Phase != corev1.PodRunning {
		return string(a.Status.Phase), a.Status.Reason
	}
	// handle running
	var (
		containers = a.Status.ContainerStatuses
		rnum       int
	)

	for _, v := range containers {
		if v.Ready {
			rnum++
			continue
		}
		if v.State.Terminated != nil {
			if v.State.Terminated.ExitCode != 0 {
				return PodStatusStatusFailed, v.State.Terminated.Reason
			}
			if v.State.Waiting != nil {
				return PodStatusStatusStarting, v.State.Waiting.Reason
			}
		}
	}
	if rnum == len(containers) {
		return PodStatusStatusRunning, ""
	} else {
		return PodStatusStatusStarting, ""
	}
}
