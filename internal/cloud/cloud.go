package cloud

import (
	"github.com/cemayan/pulumi-template/internal/cloud/aws"
	"github.com/cemayan/pulumi-template/internal/cloud/gcp"
	"github.com/cemayan/pulumi-template/types"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/viper"
)

// CloudInstance gives global variable that is selected cloud
var CloudInstance Cloud

// Cloud represents the methods that needs implement on selected cloud
// Each method do same job on different cloud.
// Ex: If selected cloud is AWS, CreateStorage method will be created the S3 bucket according to given values.
// If selected cloud is GCP, CreateStorage method will be created the Cloud Storage according to given values.
type Cloud interface {
	CreateStorage() error
	CreateDWH() error
	CreateStream() error
	CreateApiGateway() error
	CreateVpc() error
	ConfigureIAM() error
	CreateFunction() error
	CreateIdentityManagement() error
}

// Build executes functions according to given instructions array in yaml.
func Build() error {
	template := viper.GetStringMap("template")
	instructions := template["instructions"].([]interface{})
	for _, f := range GetFunctionMap(instructions) {
		f()
	}
	return nil
}

// Initialize creates new instance according to given provider
func Initialize(cloudProvider types.CloudProvider, ctx *pulumi.Context, config types.Config) {

	switch cloudProvider {
	case types.Aws:
		CloudInstance = aws.New(ctx, config)
	case types.Gcp:
		CloudInstance = gcp.New(ctx, config)
	default:
		panic("unhandled default case")
	}
}
