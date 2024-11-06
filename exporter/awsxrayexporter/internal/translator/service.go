// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package translator // import "github.com/amazon-contributing/opentelemetry-collector-contrib/exporter/awsxrayexporter/internal/translator"

import (
	awsxray "github.com/amazon-contributing/opentelemetry-collector-contrib/internal/aws/xray"
	"go.opentelemetry.io/collector/pdata/pcommon"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
)

func makeService(resource pcommon.Resource) *awsxray.ServiceData {
	var service *awsxray.ServiceData

	verStr, ok := resource.Attributes().Get(conventions.AttributeServiceVersion)
	if !ok {
		verStr, ok = resource.Attributes().Get(conventions.AttributeContainerImageTag)
	}
	if ok {
		service = &awsxray.ServiceData{
			Version: awsxray.String(verStr.Str()),
		}
	}
	return service
}
