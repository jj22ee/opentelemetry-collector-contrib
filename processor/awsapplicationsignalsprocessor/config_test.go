// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package awsapplicationsignalsprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePassed(t *testing.T) {
	config := Config{}
	assert.Nil(t, config.Validate())
}
