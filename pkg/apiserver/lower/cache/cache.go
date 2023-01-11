// Copyright 2017 EasyStack, Inc.
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
