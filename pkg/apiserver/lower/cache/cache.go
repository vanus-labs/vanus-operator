// Copyright 2022 Linkall Inc.
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

package cache

import (
	"time"

	"github.com/karlseguin/ccache/v2"
)

type Dofn func(key string) (interface{}, error)

type Cache struct {
	max time.Duration
	cc  *ccache.Cache
}

func NewCache(liveDu time.Duration) *Cache {
	return &Cache{
		cc:  ccache.New(ccache.Configure()),
		max: liveDu,
	}
}

func (c *Cache) Do(key string, fn Dofn) (interface{}, error) {
	item := c.cc.Get(key)
	if item == nil {
		value, err := fn(key)
		if err != nil {
			return nil, err
		}
		c.cc.Set(key, value, c.max)
		return value, nil
	}
	return item.Value(), nil
}
