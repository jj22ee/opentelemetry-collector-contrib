// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package gcp // import "github.com/jj22ee/opentelemetry-collector-contrib/processor/resourcedetectionprocessor/internal/gcp"

import (
	"github.com/jj22ee/opentelemetry-collector-contrib/processor/resourcedetectionprocessor/internal/gcp/internal/metadata"
)

type Config struct {
	ResourceAttributes metadata.ResourceAttributesConfig `mapstructure:"resource_attributes"`
}

func CreateDefaultConfig() Config {
	return Config{
		ResourceAttributes: metadata.DefaultResourceAttributesConfig(),
	}
}
