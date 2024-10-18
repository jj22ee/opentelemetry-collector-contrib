module github.com/jj22ee/opentelemetry-collector-contrib/internal/aws/awsutil

go 1.22.4

require (
	github.com/aws/aws-sdk-go v1.53.11
	github.com/stretchr/testify v1.9.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.26.0
)

require github.com/amazon-contributing/opentelemetry-collector-contrib/override/aws v0.0.0-00010101000000-000000000000

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/amazon-contributing/opentelemetry-collector-contrib/override/aws => ../../../override/aws

retract (
	v0.76.2
	v0.76.1
	v0.65.0
)
