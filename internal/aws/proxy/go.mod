module github.com/jj22ee/opentelemetry-collector-contrib/internal/aws/proxy

go 1.22.4

require (
	github.com/aws/aws-sdk-go v1.53.11
	github.com/jj22ee/opentelemetry-collector-contrib/internal/aws/awsutil v0.103.0
	github.com/jj22ee/opentelemetry-collector-contrib/internal/common v0.103.0
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/collector/config/confignet v0.103.0
	go.opentelemetry.io/collector/config/configtls v0.103.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/amazon-contributing/opentelemetry-collector-contrib/override/aws v0.0.0-00010101000000-000000000000 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.10.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/jj22ee/opentelemetry-collector-contrib/internal/aws/awsutil => ../../../internal/aws/awsutil

replace github.com/jj22ee/opentelemetry-collector-contrib/internal/common => ../../../internal/common

replace github.com/amazon-contributing/opentelemetry-collector-contrib/override/aws => ../../../override/aws

retract (
	v0.76.2
	v0.76.1
	v0.65.0
)
