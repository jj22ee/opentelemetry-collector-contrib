// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package config

import (
	"context"
	"time"
)

type LimiterConfig struct {
	Threshold                 int             `mapstructure:"drop_threshold"`
	Disabled                  bool            `mapstructure:"disabled"`
	LogDroppedMetrics         bool            `mapstructure:"log_dropped_metrics"`
	RotationInterval          time.Duration   `mapstructure:"rotation_interval"`
	GarbageCollectionInterval time.Duration   `mapstructure:"garbage_collection_interval"`
	ParentContext             context.Context `mapstructure:"-"`
}

const (
	DefaultThreshold        = 500
	DefaultRotationInterval = 1 * time.Hour
	DefaultGCInterval       = 10 * time.Minute
)

func NewDefaultLimiterConfig() *LimiterConfig {
	return &LimiterConfig{
		Threshold:                 DefaultThreshold,
		Disabled:                  false,
		LogDroppedMetrics:         false,
		RotationInterval:          DefaultRotationInterval,
		GarbageCollectionInterval: DefaultGCInterval,
	}
}

func (lc *LimiterConfig) Validate() {
	if lc.GarbageCollectionInterval == 0 {
		lc.GarbageCollectionInterval = DefaultGCInterval
	}
}
