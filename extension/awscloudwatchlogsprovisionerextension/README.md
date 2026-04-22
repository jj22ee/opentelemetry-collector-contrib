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

```yaml
extensions:
  sigv4auth/logs:
    region: us-east-1
    service: logs

  awscloudwatchlogsprovisioner:
    # Inner auth extension for request signing (typically sigv4auth).
    additional_auth: sigv4auth/logs

    # Log group name template (required). Placeholders are resolved from client.Metadata.
    log_group_name: "/aws/telemetry/{service.name}"

    # Log stream name (static or with placeholders).
    log_stream_name: "default"

    # Fallback value for unresolved placeholders. Default: "undefined"
    default_placeholder_value: "undefined"

    # AWS region override. If empty, extracted from the endpoint URL.
    region: "us-east-1"

    # HTTP timeout per CreateLogGroup/CreateLogStream API call (seconds). Default: 10
    logs_provision_timeout_seconds: 10

    # TTL for negative cache entries after a creation failure (seconds). Default: 30
    logs_provision_failure_backoff_seconds: 30

exporters:
  otlphttp/cw-logs:
    endpoint: https://logs.us-east-1.amazonaws.com
    logs_endpoint: https://logs.us-east-1.amazonaws.com/v1/logs
    auth:
      authenticator: awscloudwatchlogsprovisioner

service:
  extensions: [sigv4auth/logs, awscloudwatchlogsprovisioner]
  pipelines:
    logs:
      receivers: [otlp]
      processors: [attributestocontext, batch]
      exporters: [otlphttp/cw-logs]
```

## Placeholder resolution

Placeholders in `log_group_name` and `log_stream_name` are resolved from `client.Metadata` keys set by the [Attributes to Context processor](../../processor/attributestocontextprocessor)..

For example, with the processor configured as:

```yaml
processors:
  attributestocontext:
    actions:
      - key: service.name
        action: upsert
        from_resource_attribute: service.name
```

The template `{service.name}` resolves to the value of the `service.name` resource attribute.
