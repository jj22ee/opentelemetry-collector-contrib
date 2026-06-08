// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awscloudwatchlogsprovisionerextension

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestNewDefaultCWLogsClient_WithCABundle reproduces the ISO/ADC/ITAR startup
// crash. When AWS_CA_BUNDLE is set, the SDK injects custom root CAs into the
// HTTP transport, which requires a client implementing WithTransportOptions.
// A plain *http.Client fails with "unable to add custom RootCAs HTTPClient,
// has no WithTransportOptions, *http.Client".
func TestNewDefaultCWLogsClient_WithCABundle(t *testing.T) {
	caPath := writeTestCABundle(t)
	t.Setenv("AWS_CA_BUNDLE", caPath)

	client, err := newDefaultCWLogsClient(context.Background(), "us-east-1", 10*time.Second)
	require.NoError(t, err)
	require.NotNil(t, client)
}

// writeTestCABundle writes a self-signed CA cert to a temp PEM file and returns
// its path.
func writeTestCABundle(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Unix(0, 0),
		NotAfter:              time.Unix(1<<31-1, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "ca.pem")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: der}))

	return path
}
