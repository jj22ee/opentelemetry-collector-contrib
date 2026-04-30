// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/awscloudwatchlogsprovisionerextension"

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"sync"
	"time"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension/extensionauth"
	"go.opentelemetry.io/collector/extension/extensioncapabilities"
	"go.uber.org/zap"
)

const (
	defaultProvisionTimeout        = 10 * time.Second
	defaultProvisionFailureBackoff = 30 * time.Second
)

var cwLogsEndpointPattern = regexp.MustCompile(`^https://logs\.([a-z0-9-]+)\.amazonaws\.com`)

var (
	_ component.Component             = (*provisionerExtension)(nil)
	_ extensionauth.HTTPClient        = (*provisionerExtension)(nil)
	_ extensioncapabilities.Dependent = (*provisionerExtension)(nil)
)

type provisionStatus struct {
	success   bool
	timestamp time.Time
}

type inflightEntry struct {
	done chan struct{}
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
	cwLogsClientFn func(region string, timeout time.Duration) (cwLogsClient, error)

	provisioned      sync.Map
	inflight         sync.Map
	failureBackoff   time.Duration
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
	e.logger.Info("awscloudwatchlogsprovisioner started")
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
	base       http.RoundTripper
	ext        *provisionerExtension
	clientOnce sync.Once
	client     cwLogsClient // created once from the first request URL's region
}

func (rt *provisionerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create the CW Logs client once from the first request URL. The endpoint is
	// static (configured in the otlphttp exporter), so all requests share the same
	// region and client.
	// CW OTLP endpoint URL pattern: https://logs.<region>.amazonaws.com/v1/logs
	// See: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-OTLPEndpoint.html
	rt.clientOnce.Do(func() {
		region := extractRegionFromURL(req.URL.String())
		if region == "" {
			rt.ext.logger.Warn("Cannot determine region from endpoint URL — log group creation will be skipped",
				zap.String("url", req.URL.String()),
			)
			return
		}
		var err error
		rt.client, err = rt.ext.cwLogsClientFn(region, rt.ext.provisionTimeout)
		if err != nil {
			rt.ext.logger.Error("Failed to create CW Logs client — log group creation will be skipped",
				zap.Error(err),
			)
		}
	})

	req2 := req.Clone(req.Context())

	// Apply context key overrides if configured
	rt.ext.applyContextOverrides(req2)

	// Read the final header values (may have been set by otlphttp static headers,
	// headers_setter, or overridden above from context keys)
	logGroup := req2.Header.Get("x-aws-log-group")
	logStream := req2.Header.Get("x-aws-log-stream")

	if logGroup != "" {
		if logStream == "" {
			logStream = "default"
			req2.Header.Set("x-aws-log-stream", logStream)
		}

		if rt.client != nil {
			rt.ext.ensureProvisioned(rt.client, logGroup, logStream)
		}

		resp, err := rt.base.RoundTrip(req2)

		// TODO: Add cache eviction here when the CW OTLP endpoint differentiates
		// "The specified log group does not exist" from other 400 errors with a
		// distinct status code. Currently, non-existent log groups return HTTP 400
		// — the same code used for other validation errors — making it unsafe to
		// evict based on status code alone.
		// See: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-OTLPEndpoint.html#CloudWatch-LimitsandRestrictions
		//
		// Possible workaround: read the response body and check for the string
		// "The specified log group does not exist" to distinguish from other 400s.

		return resp, err
	}

	return rt.base.RoundTrip(req)
}

// applyContextOverrides reads log group/stream from client.Metadata and sets
// them as HTTP headers, overriding any existing header values.
func (e *provisionerExtension) applyContextOverrides(req *http.Request) {
	cl := client.FromContext(req.Context())

	if e.cfg.LogGroupContextKey != "" {
		values := cl.Metadata.Get(e.cfg.LogGroupContextKey)
		if len(values) > 0 && values[0] != "" {
			req.Header.Set("x-aws-log-group", values[0])
		}
	}

	if e.cfg.LogStreamContextKey != "" {
		values := cl.Metadata.Get(e.cfg.LogStreamContextKey)
		if len(values) > 0 && values[0] != "" {
			req.Header.Set("x-aws-log-stream", values[0])
		}
	}
}

// ensureProvisioned creates the log group and stream if not already cached.
// Thread safety: provisioned and inflight are sync.Map — all operations (Load,
// Store, LoadOrStore, Delete) are safe for concurrent use without additional locks.
// The inflightEntry channel provides singleflight semantics: the first goroutine
// for a given key does the creation work while concurrent goroutines block on the
// channel. During any creation attempt (including retries after negative cache
// expiry), concurrent requests for the same key will block until creation
// completes. Each key is independent — creation of one log group does not
// block requests for a different log group.
func (e *provisionerExtension) ensureProvisioned(client cwLogsClient, logGroup, logStream string) {
	key := logGroup + "\x00" + logStream

	if val, ok := e.provisioned.Load(key); ok {
		status := val.(*provisionStatus)
		if status.success {
			return
		}
		if time.Since(status.timestamp) < e.failureBackoff {
			return
		}
	}

	entry := &inflightEntry{done: make(chan struct{})}
	if existing, loaded := e.inflight.LoadOrStore(key, entry); loaded {
		existingEntry := existing.(*inflightEntry)
		// Block until the first goroutine completes creation. This is the
		// singleflight pattern — only one API call per key, others wait.
		// The wait is bounded by jitter (0-500ms) + API timeout (10s per call, 2 calls).
		<-existingEntry.done
		return
	}

	defer func() {
		close(entry.done)
		e.inflight.Delete(key)
	}()

	// Jitter for thundering-herd mitigation
	jitter := time.Duration(rand.Int63n(int64(500 * time.Millisecond))) //nolint:gosec
	time.Sleep(jitter)

	e.logger.Debug("Creating log group/stream",
		zap.String("logGroup", logGroup),
		zap.String("logStream", logStream),
	)

	err := e.createLogGroupAndStream(client, logGroup, logStream)
	if err != nil {
		e.provisioned.Store(key, &provisionStatus{
			success:   false,
			timestamp: time.Now(),
		})
		e.logger.Warn("Failed to create log group/stream",
			zap.String("logGroup", logGroup),
			zap.String("logStream", logStream),
			zap.Duration("backoff", e.failureBackoff),
			zap.Error(err),
		)
		return
	}

	e.provisioned.Store(key, &provisionStatus{
		success:   true,
		timestamp: time.Now(),
	})
	e.logger.Debug("Successfully created log group/stream",
		zap.String("logGroup", logGroup),
		zap.String("logStream", logStream),
	)
}

func (e *provisionerExtension) createLogGroupAndStream(client cwLogsClient, logGroupName, logStreamName string) error {
	ctx := context.Background()

	if err := client.CreateLogGroup(ctx, logGroupName); err != nil {
		return fmt.Errorf("CreateLogGroup %q: %w", logGroupName, err)
	}

	if err := client.CreateLogStream(ctx, logGroupName, logStreamName); err != nil {
		return fmt.Errorf("CreateLogStream %q in %q: %w", logStreamName, logGroupName, err)
	}

	return nil
}

func extractRegionFromURL(url string) string {
	matches := cwLogsEndpointPattern.FindStringSubmatch(url)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}
