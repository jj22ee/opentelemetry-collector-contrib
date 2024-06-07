// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package awsapplicationsignalsprocessor

import (
	"errors"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/awsapplicationsignalsprocessor/config"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/awsapplicationsignalsprocessor/rules"
)

type Config struct {
	Resolvers []config.Resolver     `mapstructure:"resolvers"`
	Rules     []rules.Rule          `mapstructure:"rules"`
	Limiter   *config.LimiterConfig `mapstructure:"limiter"`
}

func (cfg *Config) Validate() error {
	if len(cfg.Resolvers) == 0 {
		return errors.New("resolvers must not be empty")
	}
	for _, resolver := range cfg.Resolvers {
		switch resolver.Platform {
		case config.PlatformEKS:
			if resolver.Name == "" {
				return errors.New("name must not be empty for eks resolver")
			}
		case config.PlatformK8s:
			if resolver.Name == "" {
				return errors.New("name must not be empty for k8s resolver")
			}
		case config.PlatformEC2, config.PlatformGeneric:
		case config.PlatformECS:
			return errors.New("ecs resolver is not supported")
		default:
			return errors.New("unknown resolver")
		}
	}

	if cfg.Limiter != nil {
		cfg.Limiter.Validate()
	}
	return nil
}
