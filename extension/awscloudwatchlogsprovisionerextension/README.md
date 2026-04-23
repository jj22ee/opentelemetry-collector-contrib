# AWS CloudWatch Logs Provisioner Extension

The AWS CloudWatch Logs Provisioner extension dynamically sets `x-aws-log-group` and `x-aws-log-stream` HTTP headers and lazily creates CloudWatch log groups and streams on first encounter. It implements [`extensionauth.HTTPClient`](https://pkg.go.dev/go.opentelemetry.io/collector/extension/extensionauth#HTTPClient) to participate in the HTTP auth chain.

This extension is designed for use with the `otlphttp` exporter to send logs to the [CloudWatch OTLP endpoint](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-OTLPEndpoint.html), which requires pre-existing log groups and the `x-aws-log-group` header on every request.

## How it works

1. Reads placeholder values from [`client.Metadata`](https://pkg.go.dev/go.opentelemetry.io/collector/client#Metadata) on the request context (populated by the [Attributes to Context processor](../../processor/attributestocontextprocessor))
2. Resolves `{placeholder}` templates in log group/stream names (e.g., `{service.name}` → `my-service`)
3. Creates the log group and stream via the AWS CloudWatch Logs API if not already cached
4. Sets `x-aws-log-group` and `x-aws-log-stream` headers on the HTTP request
5. Delegates to the inner auth extension (e.g., `sigv4auth`) for request signing

## Configuration

### Extension options

| Field | Default | Description |
|---|---|---|
| `additional_auth` | (none) | Inner auth extension for request signing (typically `sigv4auth`) |
| `log_group_name` | (required) | Log group name template. Placeholders like `{service.name}` are resolved from `client.Metadata` |
| `log_stream_name` | `"default"` | Log stream name template. Placeholders like `{xyz}` are resolved from `client.Metadata` |
| `default_placeholder_value` | `"undefined"` | Fallback value for unresolved placeholders |
| `logs_provision_timeout_seconds` | `10` | HTTP timeout per CreateLogGroup/CreateLogStream API call (seconds) |
| `logs_provision_failure_backoff_seconds` | `30` | TTL for negative cache entries after a creation failure (seconds) |

### Full example

This example routes OTLP logs to per-service CloudWatch log groups based on the `service.name` resource attribute. The [Attributes to Context processor](../../processor/attributestocontextprocessor) bridges the resource attribute into `client.Metadata`, and this extension resolves the `{service.name}` placeholder and creates the log group if needed.

```yaml
extensions:
  sigv4auth/logs:
    region: us-east-1
    service: logs

  awscloudwatchlogsprovisioner:
    additional_auth: sigv4auth/logs
    log_group_name: "/aws/telemetry/{service.name}"
    log_stream_name: "default"

receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  attributestocontext:
    actions:
      - key: service.name
        action: upsert
        from_resource_attribute: service.name

  batch:
    metadata_keys:
      - service.name

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
      processors: [attributestocontext, batch]
      exporters: [otlphttp/cw-logs]
```

## Placeholder resolution

Placeholders in `log_group_name` and `log_stream_name` are resolved from `client.Metadata` keys set by the [Attributes to Context processor](../../processor/attributestocontextprocessor).

The processor config controls which resource attributes are available as placeholders:

```yaml
processors:
  attributestocontext:
    actions:
      - key: service.name
        action: upsert
        from_resource_attribute: service.name
      - key: k8s.pod.name
        action: upsert
        from_resource_attribute: k8s.pod.name
```

With this config, both `{service.name}` and `{k8s.pod.name}` can be used in templates.
