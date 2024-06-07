// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package config

import (
	"testing"
)

func TestCreateDefaultLimiterConfig(t *testing.T) {
	_ = NewDefaultLimiterConfig()
}
