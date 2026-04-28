# AWS CloudWatch Logs Provisioner Extension

The AWS CloudWatch Logs Provisioner extension creates CloudWatch log groups and streams on first encounter and sets `x-aws-log-group` and `x-aws-log-stream` HTTP headers. It implements [`extensionauth.HTTPClient`](https://pkg.go.dev/go.opentelemetry.io/collector/extension/extensionauth#HTTPClient) to participate in the HTTP auth chain.

This extension is designed for use with the `otlphttp` exporter to send logs to the [CloudWatch OTLP endpoint](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-OTLPEndpoint.html), which requires pre-existing log groups and the `x-aws-log-group` header on every request.

## How it works

1. Reads `x-aws-log-group` and `x-aws-log-stream` headers from the outgoing HTTP request
2. Optionally overrides them from [`client.Metadata`](https://pkg.go.dev/go.opentelemetry.io/collector/client#Metadata) context keys (for dynamic routing)
3. Creates the log group and stream via the AWS CloudWatch Logs API if not already cached
4. Delegates to the inner auth extension (e.g., `sigv4auth`) for request signing

The extension extracts the AWS region from the request URL (e.g., `https://logs.us-east-1.amazonaws.com` → `us-east-1`).

### Header resolution priority

1. Start with whatever `x-aws-log-group`/`x-aws-log-stream` headers are already on the request (e.g., set by otlphttp exporter static headers)
2. If `log_group_context_key` is set, override `x-aws-log-group` with the value from `client.Metadata` at that key
3. If `log_stream_context_key` is set, override `x-aws-log-stream` with the value from `client.Metadata` at that key

This means:
- **Static case**: No context keys needed. Configure `x-aws-log-group` in the otlphttp exporter headers. The extension just provisions whatever it sees.
- **Dynamic case**: Set context keys. An upstream processor (e.g., `transform` + `attributestocontext`) populates `client.Metadata` with full log group/stream names, and the extension overrides the headers and provisions.

## Configuration

| Field | Default | Description |
|---|---|---|
| `additional_auth` | (none) | Inner auth extension for request signing (typically `sigv4auth`) |
| `log_group_context_key` | `""` | Optional `client.Metadata` key to read the log group name from. Overrides the `x-aws-log-group` header. |
| `log_stream_context_key` | `""` | Optional `client.Metadata` key to read the log stream name from. Overrides the `x-aws-log-stream` header. |
| `logs_provision_timeout_seconds` | `10` | HTTP timeout per CreateLogGroup/CreateLogStream API call (seconds) |
| `logs_provision_failure_backoff_seconds` | `30` | TTL for negative cache entries after a creation failure (seconds) |

## Examples

### Dynamic routing (per-service log groups)

Routes OTLP logs to per-service CloudWatch log groups based on the `service.name` resource attribute. The `transform` processor builds the full log group name, the `attributestocontext` processor copies it to `client.Metadata`, and this extension reads it from metadata, creates the log group, and sets the header.

```yaml
extensions:
  sigv4auth/logs:
    region: us-east-1
    service: logs

  awscloudwatchlogsprovisioner:
    additional_auth: sigv4auth/logs
    log_group_context_key: cwlogs.log_group
    log_stream_context_key: cwlogs.log_stream

receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  transform:
    log_statements:
      - context: resource
        statements:
          - set(resource.attributes["cwlogs.log_group"], Concat(["/aws/telemetry/", resource.attributes["service.name"]], ""))
          - set(resource.attributes["cwlogs.log_stream"], "default")

  attributestocontext:
    actions:
      - key: cwlogs.log_group
        action: upsert
        from_resource_attribute: cwlogs.log_group
      - key: cwlogs.log_stream
        action: upsert
        from_resource_attribute: cwlogs.log_stream

  batch:
    metadata_keys:
      - cwlogs.log_group
      - cwlogs.log_stream

exporters:
  otlphttp/cw-logs:
    endpoint: https://logs.us-east-1.amazonaws.com
    logs_endpoint: https://logs.us-east-1.amazonaws.com/v1/logs
    auth:
      authenticator: awscloudwatchlogsprovisioner
    compression: gzip

service:
  extensions: [sigv4auth/logs, awscloudwatchlogsprovisioner]
  pipelines:
    logs:
      receivers: [otlp]
      processors: [transform, attributestocontext, batch]
      exporters: [otlphttp/cw-logs]
```

### Static routing (single log group)

All logs go to a single, pre-defined log group. No `transform` or `attributestocontext` processors needed — the otlphttp exporter sets the headers directly.

```yaml
extensions:
  sigv4auth/logs:
    region: us-east-1
    service: logs

  awscloudwatchlogsprovisioner:
    additional_auth: sigv4auth/logs

receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:

exporters:
  otlphttp/cw-logs:
    endpoint: https://logs.us-east-1.amazonaws.com
    logs_endpoint: https://logs.us-east-1.amazonaws.com/v1/logs
    headers:
      x-aws-log-group: /my-app/logs
      x-aws-log-stream: default
    auth:
      authenticator: awscloudwatchlogsprovisioner
    compression: gzip

service:
  extensions: [sigv4auth/logs, awscloudwatchlogsprovisioner]
  pipelines:
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/cw-logs]
```

## Provisioning behavior

- **Singleflight**: Only one API call per (log group, stream) pair. Concurrent requests for the same key block until the first goroutine completes.
- **Negative cache**: Failed creation attempts are cached for `logs_provision_failure_backoff_seconds`. During this period, the extension skips retries for that key.
- **Jitter**: A random 0–500ms delay before each creation call mitigates thundering-herd scenarios.
- **AlreadyExists**: `ResourceAlreadyExistsException` from the API is treated as success.
