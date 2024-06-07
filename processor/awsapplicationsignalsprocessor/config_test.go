// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package awsapplicationsignalsprocessor

import (
	"testing"

	resolver "github.com/open-telemetry/opentelemetry-collector-contrib/processor/awsapplicationsignalsprocessor/config"
	"github.com/stretchr/testify/assert"
)

func TestValidatePassed(t *testing.T) {
	config := Config{
		Resolvers: []resolver.Resolver{resolver.NewEKSResolver("test"), resolver.NewGenericResolver("")},
		Rules:     nil,
	}
	assert.Nil(t, config.Validate())

	config = Config{
		Resolvers: []resolver.Resolver{resolver.NewK8sResolver("test"), resolver.NewGenericResolver("")},
		Rules:     nil,
	}
	assert.Nil(t, config.Validate())

	config = Config{
		Resolvers: []resolver.Resolver{resolver.NewEC2Resolver("test"), resolver.NewGenericResolver("")},
		Rules:     nil,
	}
	assert.Nil(t, config.Validate())
}

func TestValidateFailedOnEmptyResolver(t *testing.T) {
	config := Config{
		Resolvers: []resolver.Resolver{},
		Rules:     nil,
	}
	assert.NotNil(t, config.Validate())
}

func TestValidateFailedOnEmptyResolverName(t *testing.T) {
	config := Config{
		Resolvers: []resolver.Resolver{resolver.NewEKSResolver("")},
		Rules:     nil,
	}
	assert.NotNil(t, config.Validate())

	config = Config{
		Resolvers: []resolver.Resolver{resolver.NewK8sResolver("")},
		Rules:     nil,
	}
	assert.NotNil(t, config.Validate())
}
