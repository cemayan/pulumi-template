package aws

import (
	"fmt"
	"github.com/cemayan/pulumi-template/types"
	"github.com/pulumi/pulumi-archive/sdk/go/archive"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cognito"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/kinesis"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/redshift"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/redshiftdata"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"strings"
)

// Aws represents the AWS related resources and configs
type Aws struct {
	ctx               *pulumi.Context
	config            types.Config
	roles             map[string]*iam.Role
	s3Bucket          *s3.Bucket
	firehose          *kinesis.FirehoseDeliveryStream
	redshift          *redshift.Cluster
	redshiftStatement *redshiftdata.Statement
	restApi           *apigateway.RestApi
	userPool          *cognito.UserPool
	authorizer        *apigateway.Authorizer
}

// CreateIdentityManagement creates a identity platform on AWS with Cognito
// UserPoolClient and ManagedUserPoolClient will be created that given values
// In order to take id_token from Cognito you need to sign in first with defined user below (Click View hosted UI button on AWS or
// https://<your domain>/oauth2/authorize?response_type=code&client_id=<your app client id>&redirect_uri=<your callback url>)
// After that you are able to be request to endpoint that created by API Gateway.
func (a *Aws) CreateIdentityManagement() error {

	userPool, err := cognito.NewUserPool(a.ctx, a.config.Authorizer.UserPool.Name, &cognito.UserPoolArgs{
		Name: pulumi.String(a.config.Authorizer.UserPool.Name),
	})

	conf := config.New(a.ctx, "config")
	userPass := conf.RequireSecret("userpass")

	_, err = cognito.NewUser(a.ctx, "user", &cognito.UserArgs{
		Enabled:    pulumi.Bool(true),
		Password:   userPass,
		UserPoolId: userPool.ID(),
		Username:   pulumi.String(a.config.Authorizer.UserPool.User.Username),
	}, pulumi.DependsOn([]pulumi.Resource{userPool}))

	userPoolDomain, err := cognito.NewUserPoolDomain(a.ctx, a.config.Authorizer.UserPool.UserDomain.Name, &cognito.UserPoolDomainArgs{
		Domain:     pulumi.String(a.config.Authorizer.UserPool.UserDomain.Name),
		UserPoolId: userPool.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{userPool}))

	allowedScopes := pulumi.StringArray{}

	for _, v := range a.config.Authorizer.UserPool.UserClient.AllowedScopes {
		allowedScopes = append(allowedScopes, pulumi.String(v))
	}

	allowedFlows := pulumi.StringArray{}

	for _, v := range a.config.Authorizer.UserPool.UserClient.AllowedFlows {
		allowedFlows = append(allowedFlows, pulumi.String(v))
	}

	callbackUrls := pulumi.StringArray{}

	for _, v := range a.config.Authorizer.UserPool.UserClient.CallbackUrls {
		callbackUrls = append(callbackUrls, pulumi.String(v))
	}

	exAuthFlows := pulumi.StringArray{}

	for _, v := range a.config.Authorizer.UserPool.UserClient.ExplicitAuthFlows {
		exAuthFlows = append(exAuthFlows, pulumi.String(v))
	}

	userPoolClient, err := cognito.NewUserPoolClient(a.ctx, a.config.Authorizer.UserPool.UserClient.Name, &cognito.UserPoolClientArgs{
		Name:               pulumi.String(a.config.Authorizer.UserPool.UserClient.Name),
		UserPoolId:         userPool.ID(),
		ExplicitAuthFlows:  exAuthFlows,
		CallbackUrls:       callbackUrls,
		AllowedOauthScopes: allowedScopes,
		AllowedOauthFlows:  allowedFlows,
		GenerateSecret:     pulumi.Bool(true),
	}, pulumi.DependsOn([]pulumi.Resource{userPool}))

	a.userPool = userPool

	_, err = cognito.NewManagedUserPoolClient(a.ctx, "managed", &cognito.ManagedUserPoolClientArgs{
		NamePattern:                pulumi.String(a.config.Authorizer.UserPool.UserClient.Name),
		AllowedOauthFlows:          allowedFlows,
		AllowedOauthScopes:         allowedScopes,
		CallbackUrls:               callbackUrls,
		ExplicitAuthFlows:          exAuthFlows,
		SupportedIdentityProviders: pulumi.StringArray{pulumi.String("COGNITO")},
		UserPoolId:                 userPool.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{userPool}))

	a.ctx.Export("CognitoUserPoolClientId", userPoolClient.ID())
	a.ctx.Export("CognitoUserPoolDomain", userPoolDomain.Domain)
	return err
}

// CreateFunction creates Lambda function according to given values
// You can upload zip to lambda
// If lambda function creation operation is successful URL will be exported.
func (a *Aws) CreateFunction() error {
	var err error

	// With lookup file it will be captured the changes.
	arch, err := archive.LookupFile(a.ctx, &archive.LookupFileArgs{
		Type:       "zip",
		SourceDir:  pulumi.StringRef(a.config.Function.Build.Source.Zip),
		OutputPath: a.config.Function.Build.Source.OutputPath,
	}, nil)

	envMap := pulumi.StringMap{}
	envMap["firehose_name"] = a.firehose.Name

	_func, err := lambda.NewFunction(a.ctx, a.config.Function.Name, &lambda.FunctionArgs{
		Code:           pulumi.NewFileArchive(a.config.Function.Build.Source.OutputPath),
		Name:           pulumi.String(a.config.Function.Name),
		Role:           a.roles["lambdafirehose"].Arn,
		Handler:        pulumi.String(a.config.Function.Build.Handler),
		Runtime:        pulumi.String(a.config.Function.Build.Runtime),
		SourceCodeHash: pulumi.String(arch.OutputBase64sha256),
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: envMap,
		},
	}, pulumi.DependsOn([]pulumi.Resource{a.firehose}))

	functionUrl, err := lambda.NewFunctionUrl(a.ctx, fmt.Sprintf("%v-url", a.config.Function.Name), &lambda.FunctionUrlArgs{
		FunctionName:      pulumi.String(a.config.Function.Name),
		AuthorizationType: pulumi.String(a.config.Function.Auth),
		Cors: &lambda.FunctionUrlCorsArgs{
			AllowCredentials: pulumi.Bool(true),
			AllowOrigins: pulumi.StringArray{
				pulumi.String("*"),
			},
			AllowMethods: pulumi.StringArray{
				pulumi.String("*"),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{_func}))

	a.ctx.Export("lambda_function_url", functionUrl.FunctionUrl)

	return err
}

// CreateDWH creates Redshift cluster according to given values
// Initial SQL will be executed
func (a *Aws) CreateDWH() error {

	cluster, err := redshift.NewCluster(a.ctx, a.config.Dwh.Redshift.Identifier, &redshift.ClusterArgs{
		ClusterIdentifier:  pulumi.String(a.config.Dwh.Redshift.Identifier),
		DatabaseName:       pulumi.String(a.config.Dwh.Redshift.DbName),
		MasterUsername:     pulumi.String(a.config.Dwh.Redshift.MasterUser),
		MasterPassword:     pulumi.String(a.config.Dwh.Redshift.MasterPass),
		NodeType:           pulumi.String(a.config.Dwh.Redshift.NodeType),
		NumberOfNodes:      pulumi.Int(a.config.Dwh.Redshift.NumberOfNodes),
		ClusterType:        pulumi.String(a.config.Dwh.Redshift.ClusterType),
		SkipFinalSnapshot:  pulumi.Bool(a.config.Dwh.Redshift.SkipSnapshot),
		PubliclyAccessible: pulumi.Bool(a.config.Dwh.Redshift.PublicAccess),
		IamRoles: pulumi.StringArray{
			a.roles["redshift"].Arn,
		},
	}, pulumi.DependsOn([]pulumi.Resource{a.roles["redshift"]}))

	a.redshift = cluster

	statement := &redshiftdata.StatementArgs{
		ClusterIdentifier: pulumi.String(a.config.Dwh.Redshift.Identifier),
		Database:          pulumi.String(a.config.Dwh.Redshift.DbName),
		DbUser:            pulumi.String(a.config.Dwh.Redshift.MasterUser),
		Sql:               pulumi.String(a.config.Dwh.Redshift.Sql),
	}

	newStatement, err := redshiftdata.NewStatement(a.ctx, "statement", statement, pulumi.DependsOn([]pulumi.Resource{cluster}))
	a.redshiftStatement = newStatement
	return err
}

// CreateStorage created S3 according to given values
// ForceDestroy may set the false if files are important.
func (a *Aws) CreateStorage() error {

	s3Bucket, err := s3.NewBucket(a.ctx, a.config.Storage.Name, &s3.BucketArgs{
		Bucket:       pulumi.String(a.config.Storage.Name),
		ForceDestroy: pulumi.Bool(a.config.Storage.ForceDestroy),
	})

	a.s3Bucket = s3Bucket

	return err
}

// CreateStream creates Kinesis Firehose according to given values
// You can set the destination such as "s3,redshift"
// Also you can set partition enabled config for S3.(You can set the prefix)
func (a *Aws) CreateStream() error {

	args := &kinesis.FirehoseDeliveryStreamArgs{
		Name: pulumi.String(a.config.Stream.Name),
	}

	resources := []pulumi.Resource{}

	if a.config.Stream.Destination == "s3" {
		s3ConfArgs := &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationArgs{
			RoleArn:           a.roles["firehose"].Arn,
			BucketArn:         a.s3Bucket.Arn,
			BufferingSize:     pulumi.IntPtr(a.config.Stream.S3Conf.BufferingSize),
			BufferingInterval: pulumi.IntPtr(a.config.Stream.S3Conf.BufferingInterval),
			CloudwatchLoggingOptions: kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationCloudwatchLoggingOptionsArgs{
				Enabled:       pulumi.BoolPtr(true),
				LogGroupName:  pulumi.String(a.config.Stream.Name),
				LogStreamName: pulumi.String(fmt.Sprintf("%v-stream", a.config.Stream.Name)),
			},

			ErrorOutputPrefix: pulumi.String("errors/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/hour=!{timestamp:HH}/!{firehose:error-output-type}/"),
		}

		if a.config.Stream.S3Conf.PartitionEnabled {
			s3ConfArgs.Prefix = pulumi.String(a.config.Stream.S3Conf.S3Prefix)
			s3ConfArgs.DynamicPartitioningConfiguration = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationDynamicPartitioningConfigurationArgs{
				Enabled: pulumi.Bool(true),
			}
			s3ConfArgs.ProcessingConfiguration = &kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationArgs{
				Enabled: pulumi.Bool(true),
				Processors: kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorArray{
					&kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorArgs{
						Type: pulumi.String("RecordDeAggregation"),
						Parameters: kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArray{
							&kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArgs{
								ParameterName:  pulumi.String("SubRecordType"),
								ParameterValue: pulumi.String("JSON"),
							},
						},
					},
					&kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorArgs{
						Type: pulumi.String("AppendDelimiterToRecord"),
					},
					&kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorArgs{
						Type: pulumi.String("MetadataExtraction"),
						Parameters: kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArray{
							&kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArgs{
								ParameterName:  pulumi.String("JsonParsingEngine"),
								ParameterValue: pulumi.String("JQ-1.6"),
							},
							&kinesis.FirehoseDeliveryStreamExtendedS3ConfigurationProcessingConfigurationProcessorParameterArgs{
								ParameterName:  pulumi.String("MetadataExtractionQuery"),
								ParameterValue: pulumi.String("{game_name:.game_name}"),
							},
						},
					},
				},
			}
		}

		args.Destination = pulumi.String("extended_s3")
		args.ExtendedS3Configuration = s3ConfArgs
		resources = append(resources, a.s3Bucket)
	} else if a.config.Stream.Destination == "redshift" {

		redshiftConf := &kinesis.FirehoseDeliveryStreamRedshiftConfigurationArgs{
			RoleArn: a.roles["firehose"].Arn,
			ClusterJdbcurl: pulumi.All(a.redshift.Endpoint, a.redshift.DatabaseName).ApplyT(func(_args []interface{}) (string, error) {
				endpoint := _args[0].(string)
				databaseName := _args[1].(string)
				return fmt.Sprintf("jdbc:redshift://%v/%v", endpoint, databaseName), nil
			}).(pulumi.StringOutput),
			Username: pulumi.String(a.config.Stream.RedshiftConf.Username),
			CloudwatchLoggingOptions: &kinesis.FirehoseDeliveryStreamRedshiftConfigurationCloudwatchLoggingOptionsArgs{
				Enabled:       pulumi.Bool(true),
				LogStreamName: pulumi.String(fmt.Sprintf("%v-kinesis-stream", a.config.Stream.Name)),
				LogGroupName:  pulumi.String(fmt.Sprintf("%v-kinesis-loggroup", a.config.Stream.Name)),
			},
			Password:      pulumi.String(a.config.Stream.RedshiftConf.Password),
			DataTableName: pulumi.String(a.config.Stream.RedshiftConf.DataTableName),
			CopyOptions:   pulumi.String(a.config.Stream.RedshiftConf.CopyOptions),
			S3Configuration: &kinesis.FirehoseDeliveryStreamRedshiftConfigurationS3ConfigurationArgs{
				RoleArn:           a.roles["firehose"].Arn,
				BucketArn:         a.s3Bucket.Arn,
				BufferingSize:     pulumi.Int(10),
				BufferingInterval: pulumi.Int(0),
			},
		}

		args.RedshiftConfiguration = redshiftConf
		args.Destination = pulumi.String("redshift")
		resources = append(resources, a.redshift)
	}

	firehose_, err := kinesis.NewFirehoseDeliveryStream(a.ctx, a.config.Stream.Name, args, pulumi.DependsOn(resources))
	a.firehose = firehose_

	return err
}

// CreateApiGateway create API Gateway according to given values
// You can create multiple Routes
// Example can be found in configs/datapipeline/redshift/apigateway/config.yaml
func (a *Aws) CreateApiGateway() error {

	restApi, err := apigateway.NewRestApi(a.ctx, a.config.APIGateway.Name, &apigateway.RestApiArgs{
		Name: pulumi.String(a.config.APIGateway.Name),
	})

	a.restApi = restApi

	if a.userPool != nil {
		authorizer, _ := apigateway.NewAuthorizer(a.ctx, a.config.Authorizer.Name, &apigateway.AuthorizerArgs{
			RestApi: restApi,
			Name:    pulumi.String(a.config.Authorizer.Name),
			ProviderArns: pulumi.StringArray{
				a.userPool.Arn,
			},
			Type: pulumi.String(a.config.Authorizer.Type),
		}, pulumi.DependsOn([]pulumi.Resource{restApi, a.userPool}))

		a.authorizer = authorizer
	}

	resources := []pulumi.Resource{}
	resources = append(resources, restApi)

	for _, route := range a.config.APIGateway.Routes {

		resource, _ := apigateway.NewResource(a.ctx, route.Name, &apigateway.ResourceArgs{
			RestApi:  restApi.ID(),
			ParentId: restApi.RootResourceId,
			PathPart: pulumi.String(route.Name),
		}, pulumi.DependsOn([]pulumi.Resource{restApi}))

		for _, integration := range route.Integrations {
			var err error

			methodArgs := &apigateway.MethodArgs{
				RestApi:    restApi.ID(),
				ResourceId: resource.ID(),
				HttpMethod: pulumi.String(integration.Method.Type),
			}

			if integration.Method.Auth == "COGNITO_USER_POOLS" {
				methodArgs.AuthorizerId = a.authorizer.ID()
			}

			methodArgs.Authorization = pulumi.String(integration.Method.Auth)

			method, err := apigateway.NewMethod(a.ctx, integration.Method.Name, methodArgs, pulumi.DependsOn([]pulumi.Resource{restApi, resource, a.authorizer}))

			resources = append(resources, method)

			respParamMap := pulumi.BoolMap{}

			for _, respPar := range integration.Method.Response.ResponseParams {
				respParamMap[respPar.Key] = pulumi.Bool(respPar.Val)
			}

			methodResp, err := apigateway.NewMethodResponse(a.ctx, fmt.Sprintf("response_%v", integration.Method.Name), &apigateway.MethodResponseArgs{
				RestApi:            restApi.ID(),
				ResourceId:         resource.ID(),
				HttpMethod:         method.HttpMethod,
				StatusCode:         pulumi.String(integration.Method.Response.StatusCode),
				ResponseParameters: respParamMap,
			}, pulumi.DependsOn([]pulumi.Resource{restApi, resource}))

			integrationArgs := &apigateway.IntegrationArgs{
				RestApi:               restApi.ID(),
				ResourceId:            resource.ID(),
				HttpMethod:            method.HttpMethod,
				Type:                  pulumi.String(integration.Type),
				IntegrationHttpMethod: pulumi.String(integration.HTTPMethod),
				Credentials:           a.roles["apigateway"].Arn,
				Uri:                   pulumi.String(integration.URI),
			}

			if len(integration.ReqParams) > 0 {
				reqParamMap := pulumi.StringMap{}

				for _, respPar := range integration.ReqParams {
					reqParamMap[respPar.Key] = pulumi.String(respPar.Val)
				}
				integrationArgs.RequestParameters = reqParamMap
			}

			if len(integration.ReqTemplate) > 0 {
				reqTemplateMap := pulumi.StringMap{}

				for _, reqTemp := range integration.ReqTemplate {
					reqTemplateMap[reqTemp.Key] = pulumi.String(reqTemp.Val)
				}
				integrationArgs.RequestTemplates = reqTemplateMap
			}

			integrationDependsOn := []pulumi.Resource{}

			if strings.Contains(integration.URI, "firehose:action") {
				integrationDependsOn = append(integrationDependsOn, a.firehose)
			}

			_integration, err := apigateway.NewIntegration(a.ctx, integration.Name, integrationArgs,
				pulumi.DependsOn(integrationDependsOn))

			resources = append(resources, _integration)

			integrationResponseArgs := &apigateway.IntegrationResponseArgs{
				RestApi:    restApi.ID(),
				ResourceId: resource.ID(),
				HttpMethod: methodResp.HttpMethod,
				StatusCode: methodResp.StatusCode,
			}

			if len(integration.ResTemplate) > 0 {
				resTemplateMap := pulumi.StringMap{}

				for _, respTemp := range integration.ResTemplate {
					resTemplateMap[respTemp.Key] = pulumi.String(respTemp.Val)
				}

				integrationResponseArgs.ResponseTemplates = resTemplateMap
			}

			if len(integration.ResParams) > 0 {
				resParamMap := pulumi.StringMap{}

				for _, respParam := range integration.ResParams {
					resParamMap[respParam.Key] = pulumi.String(respParam.Val)
				}

				integrationResponseArgs.ResponseParameters = resParamMap
			}

			integrationResp, err := apigateway.NewIntegrationResponse(a.ctx,
				fmt.Sprintf("integration_%v_response", integration.Method.Name),
				integrationResponseArgs, pulumi.DependsOn([]pulumi.Resource{_integration}))

			resources = append(resources, integrationResp)

			if err != nil {
				a.ctx.Log.Error(err.Error(), nil)
			}
		}
	}

	deployment, err := apigateway.NewDeployment(a.ctx, fmt.Sprintf("deploymentResource%v", a.config.APIGateway.DeploymentId), &apigateway.DeploymentArgs{
		RestApi:   restApi.ID(),
		StageName: pulumi.String(a.config.APIGateway.Stage),
	}, pulumi.DependsOn(resources))

	a.ctx.Export("apiGatewayUrl", deployment.InvokeUrl)

	return err
}

func (a *Aws) CreateVpc() error {
	return nil
}

// ConfigureIAM configures the IAM role according to given values.
// You can create the multiple role.
func (a *Aws) ConfigureIAM() error {

	a.roles = map[string]*iam.Role{}

	for _, role := range a.config.Iam.Roles {

		args := &iam.RoleArgs{
			Name:                pulumi.String(role.Name),
			ForceDetachPolicies: pulumi.Bool(role.ForceDetachPolicies),
		}

		if role.AssumePolicy != "" {
			args.AssumeRolePolicy = pulumi.String(role.AssumePolicy)
		}

		if role.InlinePolicy != "" {
			args.InlinePolicies = iam.RoleInlinePolicyArray{
				&iam.RoleInlinePolicyArgs{
					Name:   pulumi.String(fmt.Sprintf("%s-inline-role", role.Name)),
					Policy: pulumi.String(role.InlinePolicy),
				},
			}
		}

		iamRole, err := iam.NewRole(a.ctx, role.Name, args)

		if strings.HasPrefix(strings.ToLower(role.Name), "api_gateway") {
			a.roles["apigateway"] = iamRole
		} else if strings.HasPrefix(strings.ToLower(role.Name), "kinesis_firehose") {
			a.roles["firehose"] = iamRole
		} else if strings.HasPrefix(strings.ToLower(role.Name), "redshift_service") {
			a.roles["redshift"] = iamRole
		} else if strings.HasPrefix(strings.ToLower(role.Name), "lambda_firehose") {
			a.roles["lambdafirehose"] = iamRole
		}

		if err != nil {
			a.ctx.Log.Error(err.Error(), nil)
			continue
		}

	}

	return nil
}

// New returns Aws struct
func New(ctx *pulumi.Context, config types.Config) *Aws {
	return &Aws{ctx: ctx, config: config}
}
