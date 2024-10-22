// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package prune

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/jj22ee/opentelemetry-collector-contrib/processor/awsapplicationsignalsprocessor/common"
)

type MetricPruner struct {
}

func (p *MetricPruner) ShouldBeDropped(attributes pcommon.Map) (bool, error) {
	for _, attributeKey := range common.CWMetricAttributes {
		if val, ok := attributes.Get(attributeKey); ok {
			if !isAsciiPrintable(val.Str()) {
				return true, errors.New("Metric attribute " + attributeKey + " must contain only ASCII characters.")
			}
		}
		if _, ok := attributes.Get(common.MetricAttributeTelemetrySource); !ok {
			return true, errors.New(fmt.Sprintf("Metric must contain %s.", common.MetricAttributeTelemetrySource))
		}
	}
	return false, nil
}

func NewPruner() *MetricPruner {
	return &MetricPruner{}
}

func isAsciiPrintable(val string) bool {
	nonWhitespaceFound := false
	for _, c := range val {
		if c < 32 || c > 126 {
			return false
		} else if !nonWhitespaceFound && c != 32 {
			nonWhitespaceFound = true
		}
	}
	return nonWhitespaceFound
}