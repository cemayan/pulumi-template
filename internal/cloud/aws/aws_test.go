package aws

import (
	"github.com/cemayan/pulumi-template/types"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
)

type testSuite struct {
	suite.Suite
	config                types.Config
	awsConfigureIamPolicy string
}

func (ts *testSuite) SetupSuite() {

	policy := `
					{
						"Version": "2012-10-17",
						"Statement": [
							{
								"Sid": "",
								"Effect": "Allow",
								"Principal": {
									"Service": [
										"apigateway.amazonaws.com",
										"firehose.amazonaws.com"
									]
								},
								"Action": "sts:AssumeRole"
							}
						]
					}
 				`

	ts.awsConfigureIamPolicy = policy

	config := types.Config{
		Storage: types.Storage{
			Name: "test-bucket",
		},
		Iam: types.Iam{
			Roles: []types.Roles{{
				Name:         "api_gateway_kinesis_proxy_policy_pulumi-s3-lambda",
				AssumePolicy: policy,
			}},
		}}

	ts.config = config
}

type mocks int

func (m mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func (m mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	return args.Name + "_id", args.Inputs, nil
}

func (ts *testSuite) TestCreateStorage() {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {

		aws := New(ctx, ts.config)
		err := aws.CreateStorage()

		ts.NoError(err)

		var wg sync.WaitGroup
		wg.Add(1)

		// Test if the service has tags and a name tag.
		pulumi.Any(aws.s3Bucket.Bucket).ApplyT(func(data interface{}) error {
			ts.Equal("test-bucket", data)
			wg.Done()
			return nil
		})

		wg.Wait()
		return nil
	}, pulumi.WithMocks("project", "stack", mocks(0)))
	ts.NoError(err)
}
func (ts *testSuite) TestConfigureIAM() {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {

		aws := New(ctx, ts.config)
		err := aws.ConfigureIAM()

		ts.NoError(err)

		var wg sync.WaitGroup
		wg.Add(1)

		// Test if the service has tags and a name tag.
		pulumi.All(aws.roles["apigateway"].Name, aws.roles["apigateway"].AssumeRolePolicy).ApplyT(func(data []interface{}) error {
			name := data[0]
			assumePolicy := data[1]
			ts.Equal("api_gateway_kinesis_proxy_policy_pulumi-s3-lambda", name)
			ts.Equal(ts.awsConfigureIamPolicy, assumePolicy)
			wg.Done()
			return nil
		})

		wg.Wait()
		return nil
	}, pulumi.WithMocks("project", "stack", mocks(0)))
	ts.NoError(err)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
