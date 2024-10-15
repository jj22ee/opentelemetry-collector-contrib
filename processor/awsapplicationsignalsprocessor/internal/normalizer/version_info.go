// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package normalizer

import (
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
)

var (
	once          sync.Once
	cachedVersion string
)

func GetCollectorVersion() string {
	once.Do(func() {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			cachedVersion = "UNKNOWN"
			return
		}

		for _, mod := range info.Deps {
			if mod.Path == "go.opentelemetry.io/collector" {
				cachedVersion = mod.Version
				return
			}
		}

		cachedVersion = "UNKNOWN"
	})

	return cachedVersion
}
