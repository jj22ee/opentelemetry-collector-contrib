# Client Metadata Processor

The Client Metadata processor extracts attributes from telemetry data (traces, metrics, logs) and inserts them into the [client.Metadata](https://pkg.go.dev/go.opentelemetry.io/collector/client#Metadata) stored in the context. Can be used in combination with [Headers Setter extension](../../extension/headerssetterextension)'s `from_context` field to configure dynamic headers.

## Configuration

The processor supports the following configuration:

```yaml
processors:
  attributestocontext:
    actions:
      - key: "key1"
        action: insert
        from_resource_attribute: "resource.attribute1"
      - key: "key2" 
        action: update
        from_attribute: "attribute1"
      - key: "key3"
        action: upsert
        value: "static-value"
      - key: "old-key"
        action: delete
```

### Configuration Options

- `actions`: List of actions to perform on client metadata
  - `key`: The key to use in the client metadata (required)
  - `action`: The action to perform (required)
    - `insert`: Add key/value when key doesn't exist
    - `update`: Update key/value when key exists  
    - `upsert`: Insert or update key/value
    - `delete`: Remove key from client metadata
  - `from_resource_attribute`: Extract value from a resource attribute
  - `from_attribute`: Extract value from span/log/metric attributes  
  - `value`: Set a static value

Note: For `insert`, `update`, and `upsert` actions, exactly one of `value`, `from_attribute`, or `from_resource_attribute` must be specified. The `delete` action should not specify any value source.

## Example

```yaml
processors:
  resourcedetection:
    detectors: [ec2]
    ec2:
      resource_attributes:
        host.id:
          enabled: true
  attributestocontext:
    actions:
      - key: host.id
        action: insert
        from_resource_attribute: host.id

extensions:
  headers_setter:
    headers:
      - key: x-aws-log-group
        from_context: host.id
      - key: x-aws-log-stream
        from_context: host.id
```
