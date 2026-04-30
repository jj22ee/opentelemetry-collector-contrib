// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/awscloudwatchlogsprovisionerextension"

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension/extensionauth"
	"go.opentelemetry.io/collector/extension/extensioncapabilities"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

const (
	defaultProvisionTimeout        = 10 * time.Second
	defaultProvisionFailureBackoff = 30 * time.Second
)

var (
	_ component.Component             = (*provisionerExtension)(nil)
	_ extensionauth.HTTPClient        = (*provisionerExtension)(nil)
	_ extensioncapabilities.Dependent = (*provisionerExtension)(nil)
)

type cacheEntry struct {
	success   bool
	expiresAt time.Time // only used for failed entries
}

// cwLogsClient abstracts the CloudWatch Logs API for testability.
type cwLogsClient interface {
	CreateLogGroup(ctx context.Context, logGroupName string) error
	CreateLogStream(ctx context.Context, logGroupName, logStreamName string) error
}

type provisionerExtension struct {
	logger *zap.Logger
	cfg    *Config

	// host is stored during Start() for lazy resolution of additional_auth.
	// Follows the same pattern as headers_setter extension.
	host           component.Host
	client         cwLogsClient
	cwLogsClientFn func(region string, timeout time.Duration) (cwLogsClient, error)

	cache          sync.Map
	sfGroup        singleflight.Group
	failureBackoff time.Duration
	provisionTimeout time.Duration
}

func newExtension(logger *zap.Logger, cfg *Config) *provisionerExtension {
	backoff := defaultProvisionFailureBackoff
	if cfg.LogsProvisionFailureBackoffSeconds > 0 {
		backoff = time.Duration(cfg.LogsProvisionFailureBackoffSeconds) * time.Second
	}
	timeout := defaultProvisionTimeout
	if cfg.LogsProvisionTimeoutSeconds > 0 {
		timeout = time.Duration(cfg.LogsProvisionTimeoutSeconds) * time.Second
	}
	return &provisionerExtension{
		logger:           logger,
		cfg:              cfg,
		failureBackoff:   backoff,
		provisionTimeout: timeout,
		cwLogsClientFn:   newDefaultCWLogsClient,
	}
}

func (e *provisionerExtension) Start(_ context.Context, host component.Host) error {
	e.host = host

	client, err := e.cwLogsClientFn(e.cfg.Region, e.provisionTimeout)
	if err != nil {
		return fmt.Errorf("failed to create CW Logs client: %w", err)
	}
	e.client = client

	e.logger.Info("awscloudwatchlogsprovisioner started",
		zap.String("region", e.cfg.Region),
	)
	return nil
}

func (e *provisionerExtension) Shutdown(_ context.Context) error {
	return nil
}

func (e *provisionerExtension) Dependencies() []component.ID {
	if e.cfg.AdditionalAuth != nil {
		return []component.ID{*e.cfg.AdditionalAuth}
	}
	return nil
}

// getAdditionalAuthExtension lazily resolves the additional_auth extension
// from the host. Follows the same pattern as headers_setter:
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.150.0/extension/headerssetterextension/extension.go
func (e *provisionerExtension) getAdditionalAuthExtension() (extensionauth.HTTPClient, error) {
	if e.cfg.AdditionalAuth == nil {
		return nil, nil
	}

	ext, ok := e.host.GetExtensions()[*e.cfg.AdditionalAuth]
	if !ok {
		return nil, fmt.Errorf("additional_auth extension %q not found", e.cfg.AdditionalAuth)
	}

	httpClient, ok := ext.(extensionauth.HTTPClient)
	if !ok {
		return nil, fmt.Errorf("additional_auth extension %q does not implement HTTPClient", e.cfg.AdditionalAuth)
	}

	return httpClient, nil
}

func (e *provisionerExtension) RoundTripper(base http.RoundTripper) (http.RoundTripper, error) {
	transport := base

	// Chain with additional_auth (e.g., sigv4auth) if configured.
	additionalAuth, err := e.getAdditionalAuthExtension()
	if err != nil {
		return nil, err
	}
	if additionalAuth != nil {
		transport, err = additionalAuth.RoundTripper(base)
		if err != nil {
			return nil, fmt.Errorf("failed to get additional_auth RoundTripper: %w", err)
		}
	}

	return &provisionerRoundTripper{
		base: transport,
		ext:  e,
	}, nil
}

type provisionerRoundTripper struct {
	base http.RoundTripper
	ext  *provisionerExtension
}

func (rt *provisionerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read headers set by the otlphttp exporter (static) or headers_setter (dynamic).
	logGroup := req.Header.Get("x-aws-log-group")
	logStream := req.Header.Get("x-aws-log-stream")

	if logGroup != "" && logStream != "" {
		rt.ext.ensure(req.Context(), logGroup, logStream)

		resp, err := rt.base.RoundTrip(req)
		if err != nil {
			return resp, err
		}

		// If the CW OTLP endpoint returns 400 with "does not exist", evict the
		// cache entry so the next request re-provisions.
		if resp.StatusCode == http.StatusBadRequest {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr == nil && strings.Contains(string(body), "does not exist") {
				rt.ext.evict(logGroup, logStream)
				rt.ext.ensure(req.Context(), logGroup, logStream)
			}
			resp.Body = io.NopCloser(strings.NewReader(string(body)))
		}

		return resp, nil
	}

	return rt.base.RoundTrip(req)
}

func cacheKey(logGroup, logStream string) string {
	return logGroup + "\x00" + logStream
}

// ensure creates the log group and stream if not already cached.
// Uses singleflight to deduplicate concurrent creation attempts for the same key.
func (e *provisionerExtension) ensure(ctx context.Context, logGroup, logStream string) {
	key := cacheKey(logGroup, logStream)

	if entry, ok := e.cache.Load(key); ok {
		ce := entry.(cacheEntry)
		if ce.success || time.Now().Before(ce.expiresAt) {
			return
		}
	}

	_, _, _ = e.sfGroup.Do(key, func() (any, error) {
		// Double-check cache after acquiring singleflight.
		if entry, ok := e.cache.Load(key); ok {
			ce := entry.(cacheEntry)
			if ce.success || time.Now().Before(ce.expiresAt) {
				return nil, nil
			}
		}

		err := e.provision(ctx, logGroup, logStream)
		if err != nil {
			e.cache.Store(key, cacheEntry{expiresAt: time.Now().Add(e.failureBackoff)})
			e.logger.Warn("Failed to create log group/stream",
				zap.String("logGroup", logGroup),
				zap.String("logStream", logStream),
				zap.Duration("backoff", e.failureBackoff),
				zap.Error(err),
			)
		} else {
			e.cache.Store(key, cacheEntry{success: true})
			e.logger.Debug("Successfully provisioned log group/stream",
				zap.String("logGroup", logGroup),
				zap.String("logStream", logStream),
			)
		}
		return nil, nil
	})
}

// provision creates the log stream (and log group if needed).
// Tries stream first — if the group doesn't exist, creates it and retries.
func (e *provisionerExtension) provision(ctx context.Context, logGroup, logStream string) error {
	err := e.client.CreateLogStream(ctx, logGroup, logStream)
	if err == nil {
		return nil
	}

	if !isNotFound(err) {
		return fmt.Errorf("CreateLogStream %q in %q: %w", logStream, logGroup, err)
	}

	e.logger.Debug("Log group not found, creating",
		zap.String("logGroup", logGroup),
	)
	if grpErr := e.client.CreateLogGroup(ctx, logGroup); grpErr != nil {
		return fmt.Errorf("CreateLogGroup %q: %w", logGroup, grpErr)
	}

	if retryErr := e.client.CreateLogStream(ctx, logGroup, logStream); retryErr != nil {
		return fmt.Errorf("CreateLogStream %q in %q (retry): %w", logStream, logGroup, retryErr)
	}

	return nil
}

func (e *provisionerExtension) evict(logGroup, logStream string) {
	e.cache.Delete(cacheKey(logGroup, logStream))
}
