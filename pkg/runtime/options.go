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

package runtime

type ConnectorOption func(opt *connectorOptions)

type connectorOptions struct {
	labelSelector string
	handler       ConnectorEventHandler
}

func newConnectorOptions(options ...ConnectorOption) connectorOptions {
	opts := defaultConnectorOptions()

	for _, apply := range options {
		apply(&opts)
	}
	return opts
}

func defaultConnectorOptions() connectorOptions {
	return connectorOptions{}
}

func WithFilter(filter string) ConnectorOption {
	return func(opt *connectorOptions) {
		opt.labelSelector = filter
	}
}

func WithEventHandler(handler ConnectorEventHandler) ConnectorOption {
	return func(opt *connectorOptions) {
		opt.handler = handler
	}
}
