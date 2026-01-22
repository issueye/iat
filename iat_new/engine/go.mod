module iat/engine

go 1.21

require (
	github.com/cloudwego/eino v0.7.24
	github.com/cloudwego/eino-ext/components/model/openai v0.1.8
	github.com/dop251/goja v0.0.0-20241024094426-79f3a7efcdbd
	github.com/eino-contrib/jsonschema v1.0.3
	github.com/glebarez/sqlite v1.11.0
	github.com/google/uuid v1.6.0
	github.com/mark3labs/mcp-go v0.43.2
	github.com/sergi/go-diff v1.3.1
	github.com/syndtr/goleveldb v1.0.0
	gorm.io/gorm v1.25.12
)

replace iat/common => ../common
