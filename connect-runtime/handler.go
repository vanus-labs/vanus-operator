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

package connect_runtime

type FilterConnector struct {
	Kind string
	Type string
}

type ConnectorEventHandler interface {
	OnAdd(connectorID, config string) error
	OnUpdate(connectorID, config string) error
	OnDelete(connectorID string) error
}

type ConnectorEventHandlerFuncs struct {
	AddFunc    func(connectorID, config string) error
	UpdateFunc func(connectorID, config string) error
	DeleteFunc func(connectorID string) error
}

// OnAdd calls AddFunc if it's not nil.
func (r ConnectorEventHandlerFuncs) OnAdd(connectorID, config string) error {
	if r.AddFunc != nil {
		return r.AddFunc(connectorID, config)
	}
	return nil
}

// OnUpdate calls UpdateFunc if it's not nil.
func (r ConnectorEventHandlerFuncs) OnUpdate(connectorID, config string) error {
	if r.UpdateFunc != nil {
		return r.UpdateFunc(connectorID, config)
	}
	return nil
}

// OnDelete calls DeleteFunc if it's not nil.
func (r ConnectorEventHandlerFuncs) OnDelete(connectorID string) error {
	if r.DeleteFunc != nil {
		return r.DeleteFunc(connectorID)
	}
	return nil
}
