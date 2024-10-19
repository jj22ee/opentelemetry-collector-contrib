// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package filtermetric

import (
	"github.com/jj22ee/opentelemetry-collector-contrib/internal/filter/filterconfig"
	"github.com/jj22ee/opentelemetry-collector-contrib/internal/filter/filterset"
)

func createConfig(filters []string, matchType filterset.MatchType) *filterconfig.MetricMatchProperties {
	return &filterconfig.MetricMatchProperties{
		MatchType:   filterconfig.MetricMatchType(matchType),
		MetricNames: filters,
	}
}
