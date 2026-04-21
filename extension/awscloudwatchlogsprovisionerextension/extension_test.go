// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap/zaptest"
)

// --- Mock CW Logs client ---

type mockCWLogsClient struct {
	createGroupErr  error
	createStreamErr error
	groupCalls      atomic.Int32
	streamCalls     atomic.Int32
}

func (m *mockCWLogsClient) CreateLogGroup(_ context.Context, _ string) error {
	m.groupCalls.Add(1)
	return m.createGroupErr
}

func (m *mockCWLogsClient) CreateLogStream(_ context.Context, _, _ string) error {
	m.streamCalls.Add(1)
	return m.createStreamErr
}

// --- Mock inner auth ---

type mockHTTPClient struct {
	component.StartFunc
	component.ShutdownFunc
}

func (m *mockHTTPClient) RoundTripper(base http.RoundTripper) (http.RoundTripper, error) {
	return base, nil
}

// --- Mock host ---

type mockHost struct {
	extensions map[component.ID]component.Component
}

func (h *mockHost) GetExtensions() map[component.ID]component.Component {
	return h.extensions
}

// --- Helper to build extension ---

func newTestExtension(t *testing.T, cfg *Config, mockClient *mockCWLogsClient) *provisionerExtension {
	ext := newExtension(zaptest.NewLogger(t), cfg)
	ext.cwLogsClientFn = func(_ string, _ time.Duration) (cwLogsClient, error) {
		return mockClient, nil
	}
	return ext
}

// --- Tests ---

func TestExtractRegionFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://logs.us-east-1.amazonaws.com/v1/logs", "us-east-1"},
		{"https://logs.eu-west-1.amazonaws.com/v1/logs", "eu-west-1"},
		{"https://logs.ap-southeast-2.amazonaws.com", "ap-southeast-2"},
		{"https://example.com/v1/logs", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractRegionFromURL(tt.url))
		})
	}
}

func TestResolvePlaceholders(t *testing.T) {
	ext := newExtension(zaptest.NewLogger(t), &Config{
		DefaultPlaceholderValue: "UNKNOWN",
	})

	md := client.NewMetadata(map[string][]string{
		"service.name": {"my-service"},
		"host.id":      {"i-1234"},
	})

	tests := []struct {
		template string
		expected string
	}{
		{"/test/telemetry/{service.name}", "/test/telemetry/my-service"},
		{"/logs/{host.id}/{service.name}", "/logs/i-1234/my-service"},
		{"/logs/{missing.key}", "/logs/UNKNOWN"},
		{"static-value", "static-value"},
		{"{service.name}", "my-service"},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			assert.Equal(t, tt.expected, ext.resolvePlaceholders(tt.template, md))
		})
	}
}

func TestEnsureProvisioned_Success(t *testing.T) {
	mockClient := &mockCWLogsClient{}
	ext := newTestExtension(t, &Config{Region: "us-east-1"}, mockClient)

	req := httptest.NewRequest(http.MethodPost, "https://logs.us-east-1.amazonaws.com/v1/logs", nil)
	ext.ensureProvisioned(req, "/test/group", "default")

	assert.Equal(t, int32(1), mockClient.groupCalls.Load())
	assert.Equal(t, int32(1), mockClient.streamCalls.Load())

	// Second call should hit cache
	ext.ensureProvisioned(req, "/test/group", "default")
	assert.Equal(t, int32(1), mockClient.groupCalls.Load(), "should not create again after cache hit")
}

func TestEnsureProvisioned_FailureThenBackoff(t *testing.T) {
	mockClient := &mockCWLogsClient{
		createGroupErr: fmt.Errorf("throttled"),
	}
	ext := newTestExtension(t, &Config{
		Region:                "us-east-1",
		LogsProvisionFailureBackoffSeconds: 60,
	}, mockClient)

	req := httptest.NewRequest(http.MethodPost, "https://logs.us-east-1.amazonaws.com/v1/logs", nil)
	ext.ensureProvisioned(req, "/test/group", "default")

	assert.Equal(t, int32(1), mockClient.groupCalls.Load())

	// Second call within backoff window should NOT retry
	ext.ensureProvisioned(req, "/test/group", "default")
	assert.Equal(t, int32(1), mockClient.groupCalls.Load(), "should not retry during backoff")
}

func TestEnsureProvisioned_Singleflight(t *testing.T) {
	mockClient := &mockCWLogsClient{}
	ext := newTestExtension(t, &Config{Region: "us-east-1"}, mockClient)

	req := httptest.NewRequest(http.MethodPost, "https://logs.us-east-1.amazonaws.com/v1/logs", nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ext.ensureProvisioned(req, "/test/singleflight", "default")
		}()
	}
	wg.Wait()

	// Only one CreateLogGroup call should have been made
	assert.Equal(t, int32(1), mockClient.groupCalls.Load(), "singleflight should dedup concurrent creation")
}

func TestRoundTripper_SetsHeaders(t *testing.T) {
	mockClient := &mockCWLogsClient{}
	ext := newTestExtension(t, &Config{
		Region:                  "us-east-1",
		LogGroupName:            "/test/{service.name}",
		LogStreamName:           "default",
		DefaultPlaceholderValue: "unknown",
	}, mockClient)

	var capturedReq *http.Request
	base := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		capturedReq = req
		return &http.Response{StatusCode: 200}, nil
	})

	rt, err := ext.RoundTripper(base)
	require.NoError(t, err)

	md := client.NewMetadata(map[string][]string{
		"service.name": {"pet-clinic"},
	})
	ctx := client.NewContext(context.Background(), client.Info{Metadata: md})
	req := httptest.NewRequest(http.MethodPost, "https://logs.us-east-1.amazonaws.com/v1/logs", nil)
	req = req.WithContext(ctx)

	_, err = rt.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, "/test/pet-clinic", capturedReq.Header.Get("x-aws-log-group"))
	assert.Equal(t, "default", capturedReq.Header.Get("x-aws-log-stream"))
}

func TestRoundTripper_NoMetadata_UsesDefault(t *testing.T) {
	mockClient := &mockCWLogsClient{}
	ext := newTestExtension(t, &Config{
		Region:                  "us-east-1",
		LogGroupName:            "/test/{service.name}",
		LogStreamName:           "default",
		DefaultPlaceholderValue: "undefined",
	}, mockClient)

	var capturedReq *http.Request
	base := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		capturedReq = req
		return &http.Response{StatusCode: 200}, nil
	})

	rt, err := ext.RoundTripper(base)
	require.NoError(t, err)

	// Empty metadata
	ctx := client.NewContext(context.Background(), client.Info{})
	req := httptest.NewRequest(http.MethodPost, "https://logs.us-east-1.amazonaws.com/v1/logs", nil)
	req = req.WithContext(ctx)

	_, err = rt.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, "/test/undefined", capturedReq.Header.Get("x-aws-log-group"))
}

func TestStart_StoresHost(t *testing.T) {
	authID := component.MustNewID("sigv4auth")
	cfg := &Config{AdditionalAuth: &authID}
	ext := newExtension(zaptest.NewLogger(t), cfg)

	mockAuth := &mockHTTPClient{}
	host := &mockHost{
		extensions: map[component.ID]component.Component{
			authID: mockAuth,
		},
	}

	err := ext.Start(context.Background(), host)
	require.NoError(t, err)
	assert.NotNil(t, ext.host)
}

func TestRoundTripper_MissingAdditionalAuth(t *testing.T) {
	authID := component.MustNewID("sigv4auth")
	cfg := &Config{AdditionalAuth: &authID}
	ext := newExtension(zaptest.NewLogger(t), cfg)

	// Start with empty host — additional_auth won't be found
	host := &mockHost{extensions: map[component.ID]component.Component{}}
	err := ext.Start(context.Background(), host)
	require.NoError(t, err)

	base := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200}, nil
	})

	_, err = ext.RoundTripper(base)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDependencies(t *testing.T) {
	authID := component.MustNewID("sigv4auth")

	ext := newExtension(zaptest.NewLogger(t), &Config{AdditionalAuth: &authID})
	deps := ext.Dependencies()
	assert.Equal(t, []component.ID{authID}, deps)

	ext2 := newExtension(zaptest.NewLogger(t), &Config{})
	assert.Nil(t, ext2.Dependencies())
}

func TestEnsureProvisioned_DifferentKeysIndependent(t *testing.T) {
	mockClient := &mockCWLogsClient{}
	ext := newTestExtension(t, &Config{Region: "us-east-1"}, mockClient)

	req := httptest.NewRequest(http.MethodPost, "https://logs.us-east-1.amazonaws.com/v1/logs", nil)

	ext.ensureProvisioned(req, "/test/service-a", "default")
	ext.ensureProvisioned(req, "/test/service-b", "default")

	assert.Equal(t, int32(2), mockClient.groupCalls.Load(), "different keys should create independently")
}

func TestEnsureProvisioned_NoRegion(t *testing.T) {
	mockClient := &mockCWLogsClient{}
	ext := newTestExtension(t, &Config{}, mockClient)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/v1/logs", nil)
	ext.ensureProvisioned(req, "/test/group", "default")

	assert.Equal(t, int32(0), mockClient.groupCalls.Load(), "should skip creation when region unknown")
}

func TestFailureBackoff_ExpiresAndRetries(t *testing.T) {
	mockClient := &mockCWLogsClient{
		createGroupErr: fmt.Errorf("throttled"),
	}
	ext := newTestExtension(t, &Config{
		Region:                "us-east-1",
		LogsProvisionFailureBackoffSeconds: 1,
	}, mockClient)

	req := httptest.NewRequest(http.MethodPost, "https://logs.us-east-1.amazonaws.com/v1/logs", nil)
	ext.ensureProvisioned(req, "/test/group", "default")
	assert.Equal(t, int32(1), mockClient.groupCalls.Load())

	// Wait for backoff to expire
	time.Sleep(1100 * time.Millisecond)

	ext.ensureProvisioned(req, "/test/group", "default")
	assert.Equal(t, int32(2), mockClient.groupCalls.Load(), "should retry after backoff expires")
}

// --- Helper ---

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
