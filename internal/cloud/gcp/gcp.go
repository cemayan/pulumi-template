package gcp

import (
	"fmt"
	"github.com/cemayan/pulumi-template/types"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/apigateway"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/bigquery"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/cloudfunctionsv2"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/iap"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/pubsub"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi-std/sdk/go/std"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	yaml "gopkg.in/yaml.v3"
	"os"
)

// Gcp represents the GCP related resources and configs
type Gcp struct {
	ctx                     *pulumi.Context
	region                  string
	project                 string
	bucket                  *storage.Bucket
	table                   *bigquery.Table
	topic                   *pubsub.Topic
	function                *cloudfunctionsv2.Function
	functionSourceBucket    *storage.Bucket
	functionSourceBucketObj *storage.BucketObject
	serviceAcc              *serviceaccount.Account
	config                  types.Config
	oauth2Client            *iap.Client
}

func (g *Gcp) CreateIdentityManagement() error {
	panic("not implemented")
}

// createServiceAccount creates a service account for cloud function
func (g *Gcp) createServiceAccount() (*serviceaccount.Account, error) {

	account, err := serviceaccount.NewAccount(g.ctx, g.config.Iam.ServiceAcc.AccountID, &serviceaccount.AccountArgs{
		AccountId:                 pulumi.String(g.config.Iam.ServiceAcc.AccountID),
		DisplayName:               pulumi.String(g.config.Iam.ServiceAcc.DisplayName),
		Project:                   pulumi.String(g.config.Iam.ServiceAcc.Project),
		CreateIgnoreAlreadyExists: pulumi.Bool(true),
	})

	g.serviceAcc = account

	return account, err
}

// createBucketForFunction creates bucket for function
// it is used to get zipped function
func (g *Gcp) createBucketForFunction() (*storage.Bucket, *storage.BucketObject, error) {

	bucket, err := storage.NewBucket(g.ctx, g.config.Function.Build.Source.Storage.Bucket.Name, &storage.BucketArgs{
		Name:                     pulumi.String(g.config.Function.Build.Source.Storage.Bucket.Name),
		Location:                 pulumi.String(g.region),
		UniformBucketLevelAccess: pulumi.Bool(true),
		ForceDestroy:             pulumi.Bool(g.config.Function.Build.Source.Storage.ForceDestroy),
	})

	g.functionSourceBucket = bucket

	object, err := storage.NewBucketObject(g.ctx, g.config.Function.Build.Source.Storage.Bucket.Object.Name, &storage.BucketObjectArgs{
		Name:   pulumi.String(g.config.Function.Build.Source.Storage.Bucket.Object.Name),
		Bucket: bucket.Name,
		Source: pulumi.NewFileAsset(g.config.Function.Build.Source.Storage.Bucket.Object.Path),
	}, pulumi.DependsOn([]pulumi.Resource{bucket}))

	g.functionSourceBucketObj = object

	return bucket, object, err
}

// CreateFunction creates Cloud function/Cloud Run according to given values
// You can upload zip to lambda
// If lambda function creation operation is successful URL will be exported.
// cloudfunctionsv2 is new version for cloudfunctions You can check the link below
// https://cloud.google.com/functions/docs/concepts/version-comparison
func (g *Gcp) CreateFunction() error {

	if g.config.Iam.ServiceAcc != nil {
		_, _ = g.createServiceAccount()
	}

	buildArgs := &cloudfunctionsv2.FunctionBuildConfigArgs{
		Runtime:    pulumi.String(g.config.Function.Build.Runtime),
		EntryPoint: pulumi.String(g.config.Function.Build.EntryPoint),
	}

	funcArgs := &cloudfunctionsv2.FunctionArgs{
		Name:        pulumi.String(g.config.Function.Name),
		Location:    pulumi.String(g.region),
		Project:     pulumi.String(g.project),
		BuildConfig: buildArgs,

		ServiceConfig: &cloudfunctionsv2.FunctionServiceConfigArgs{
			MaxInstanceCount:    pulumi.Int(g.config.Function.ServiceConf.MaxInstance),
			AvailableMemory:     pulumi.String(g.config.Function.ServiceConf.AvailableMem),
			TimeoutSeconds:      pulumi.Int(g.config.Function.ServiceConf.Timeout),
			ServiceAccountEmail: g.serviceAcc.Email,
			EnvironmentVariables: pulumi.StringMap{
				"PROJECT_ID": pulumi.String(g.project),
				"TOPIC_ID":   pulumi.String(g.config.Stream.PubSubConf.Topic.Name)},
		},
	}

	if g.config.Function.Trigger != nil {

		triggerArgs := &cloudfunctionsv2.FunctionEventTriggerArgs{
			EventType:     pulumi.String(g.config.Function.Trigger.EventType),
			RetryPolicy:   pulumi.String("RETRY_POLICY_RETRY"),
			TriggerRegion: pulumi.String(g.config.Function.Trigger.Region),
		}

		if g.config.Template.Name == "data-pipeline" {
			triggerArgs.PubsubTopic = pulumi.String(g.config.Function.Trigger.EventType)
		}

		funcArgs.EventTrigger = triggerArgs
	}

	if g.config.Function.Build.Source != nil {
		bucket, object, _ := g.createBucketForFunction()

		buildArgs.Source = &cloudfunctionsv2.FunctionBuildConfigSourceArgs{
			StorageSource: &cloudfunctionsv2.FunctionBuildConfigSourceStorageSourceArgs{
				Bucket: bucket.Name,
				Object: object.Name,
			},
		}
	}

	function, err := cloudfunctionsv2.NewFunction(g.ctx, g.config.Function.Name, funcArgs, pulumi.DependsOn([]pulumi.Resource{g.functionSourceBucket, g.functionSourceBucketObj, g.serviceAcc}))

	g.function = function

	g.configureRolesForFunction()

	g.ctx.Export("function_url", function.Url)

	return err
}

// CreateDWH creates BigQuery Dataset and Table according to given values
// Initial schema will be created
func (g *Gcp) CreateDWH() error {

	dataset, err := bigquery.NewDataset(g.ctx, g.config.Dwh.BigQuery.Dataset, &bigquery.DatasetArgs{
		DatasetId: pulumi.String(g.config.Dwh.BigQuery.Dataset),
		Location:  pulumi.String(g.region),
	})

	table, err := bigquery.NewTable(g.ctx, g.config.Dwh.BigQuery.TableId, &bigquery.TableArgs{
		DeletionProtection: pulumi.Bool(g.config.Dwh.BigQuery.DeletionProtection),
		TableId:            pulumi.String(g.config.Dwh.BigQuery.TableId),
		DatasetId:          dataset.DatasetId,
		Schema:             pulumi.String(g.config.Dwh.BigQuery.Schema),
	})

	g.table = table

	return err
}

// CreateStorage created Cloud Storage according to given values
// ForceDestroy may set the false if files are important.
func (g *Gcp) CreateStorage() error {

	bucket, err := storage.NewBucket(g.ctx, g.config.Storage.Name, &storage.BucketArgs{
		Name:                     pulumi.String(g.config.Storage.Name),
		Location:                 pulumi.String(g.region),
		ForceDestroy:             pulumi.Bool(g.config.Storage.ForceDestroy),
		UniformBucketLevelAccess: pulumi.Bool(true),
	})

	g.bucket = bucket

	return err
}

// CreateStream create Pubsub topic and subscription according to given values
// You can set the destination such as "cloudstorage,bigquery"
func (g *Gcp) CreateStream() error {

	var err error

	topic, err := pubsub.NewTopic(g.ctx, g.config.Stream.PubSubConf.Topic.Name, &pubsub.TopicArgs{
		Name: pulumi.String(g.config.Stream.PubSubConf.Topic.Name),
	})

	g.topic = topic

	subsArgs := &pubsub.SubscriptionArgs{
		Name:               pulumi.String(g.config.Stream.PubSubConf.Subscription.Name),
		Topic:              topic.ID(),
		AckDeadlineSeconds: pulumi.Int(20),
	}

	resources := []pulumi.Resource{}

	if g.config.Stream.Destination == "cloudstorage" {
		subsArgs.CloudStorageConfig = &pubsub.SubscriptionCloudStorageConfigArgs{
			Bucket:      g.bucket.ID(),
			MaxDuration: pulumi.String(g.config.Stream.PubSubConf.Subscription.CloudStorageConf.Duration),
		}
	} else if g.config.Stream.Destination == "bigquery" {
		subsArgs.BigqueryConfig = &pubsub.SubscriptionBigqueryConfigArgs{
			Table: pulumi.All(g.table.Project, g.table.DatasetId, g.table.TableId).ApplyT(func(_args []interface{}) (string, error) {
				project := _args[0].(string)
				datasetId := _args[1].(string)
				tableId := _args[2].(string)
				return fmt.Sprintf("%v.%v.%v", project, datasetId, tableId), nil
			}).(pulumi.StringOutput),
			UseTableSchema: pulumi.Bool(true),
		}

		resources = append(resources, g.table)
	}

	_, err = pubsub.NewSubscription(g.ctx, g.config.Stream.PubSubConf.Subscription.Name, subsArgs, pulumi.DependsOn(resources))

	return err
}

// generateSpecFile generates a yaml file according to given values
// Since oauth2 client cannot be created by Gcloud CLI we need to create Oauth2 Client first
// With this yaml you are able to use Google Authentication while using the function URL
func (g *Gcp) generateSpecFile() error {

	spec := types.OpenApiSpec{}
	spec.Swagger = "2.0"
	spec.Info.Title = g.config.APIGateway.Name
	spec.Info.Version = "1.0.0"

	spec.SecurityDefinitions.GoogleIDToken.AuthorizationURL = ""
	spec.SecurityDefinitions.GoogleIDToken.Flow = "implicit"
	spec.SecurityDefinitions.GoogleIDToken.Type = "oauth2"
	spec.SecurityDefinitions.GoogleIDToken.XGoogleIssuer = "https://accounts.google.com"
	spec.SecurityDefinitions.GoogleIDToken.XGoogleJwksURI = "https://www.googleapis.com/oauth2/v3/certs"
	spec.SecurityDefinitions.GoogleIDToken.XGoogleAudiences = g.config.Idp.ClientId

	spec.Schemes = []string{"https"}
	spec.Produces = []string{"application/json"}

	spec.Paths.Event.Post.OperationID = fmt.Sprintf("%v-op", g.config.APIGateway.Name)

	spec.Paths.Event.Post.XGoogleBackend.Address = fmt.Sprintf("https://%v-%v.cloudfunctions.net/%v", g.config.Function.Region, g.config.Iam.ServiceAcc.Project, g.config.Function.Name)

	securities := []types.Security{}

	securities = append(securities, types.Security{GoogleIDToken: make([]interface{}, 0)})

	spec.Paths.Event.Post.Security = securities
	spec.Paths.Event.Post.Responses.Num200.Description = "OK"

	yamlFile, _ := yaml.Marshal(&spec)

	err := os.WriteFile("api/gcp/api.yaml", yamlFile, 0644)
	return err
}

// CreateApiGateway creates Api and Gateway according to given values
// Before create a Gateway you need to generate spec file
func (g *Gcp) CreateApiGateway() error {

	err := g.generateSpecFile()

	apiGw, err := apigateway.NewApi(g.ctx, g.config.APIGateway.Name, &apigateway.ApiArgs{
		ApiId: pulumi.String(g.config.APIGateway.Name),
	}, pulumi.DependsOn([]pulumi.Resource{g.function}))
	if err != nil {
		return err
	}
	invokeFilebase64, err := std.Filebase64(g.ctx, &std.Filebase64Args{
		Input: g.config.APIGateway.OpenApiSpec,
	}, nil)

	apiGwApiConfig, err := apigateway.NewApiConfig(g.ctx, fmt.Sprintf("%v-config", g.config.APIGateway.Name), &apigateway.ApiConfigArgs{
		Api:         apiGw.ApiId,
		ApiConfigId: pulumi.String(fmt.Sprintf("%v-config", g.config.APIGateway.Name)),
		OpenapiDocuments: apigateway.ApiConfigOpenapiDocumentArray{
			&apigateway.ApiConfigOpenapiDocumentArgs{
				Document: &apigateway.ApiConfigOpenapiDocumentDocumentArgs{
					Path:     pulumi.String(g.config.APIGateway.OpenApiSpec),
					Contents: pulumi.String(invokeFilebase64.Result),
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{apiGw}))

	_, err = apigateway.NewGateway(g.ctx, fmt.Sprintf("%v-gw", g.config.APIGateway.Name), &apigateway.GatewayArgs{
		ApiConfig: apiGwApiConfig.ID(),
		GatewayId: pulumi.String(fmt.Sprintf("%v-gw", g.config.APIGateway.Name)),
		Region:    pulumi.String(g.config.APIGateway.Region),
	}, pulumi.DependsOn([]pulumi.Resource{apiGwApiConfig}))

	return err
}

func (g *Gcp) CreateVpc() error {
	//TODO implement me
	panic("implement me")
}

// configureRolesForFunction creates/binds a member according to given values.
func (g *Gcp) configureRolesForFunction() error {

	var err error

	for _, role := range g.config.Iam.Roles {
		if role.Type == "cloudfuncv2member" {
			_, err = cloudfunctionsv2.NewFunctionIamMember(g.ctx, role.Name, &cloudfunctionsv2.FunctionIamMemberArgs{
				Project:       pulumi.String(g.project),
				Location:      pulumi.String(g.region),
				CloudFunction: pulumi.String(g.config.Function.Name),
				Role:          pulumi.String(role.Role),
				Member: g.serviceAcc.Email.ApplyT(func(email string) (string, error) {
					return fmt.Sprintf("serviceAccount:%v", email), nil
				}).(pulumi.StringOutput),
			}, pulumi.DependsOn([]pulumi.Resource{g.function}))
		} else if role.Type == "pubsubmember" {
			_, err = pubsub.NewTopicIAMMember(g.ctx, role.Name, &pubsub.TopicIAMMemberArgs{
				Project: pulumi.String(g.project),
				Topic:   pulumi.String(g.config.Stream.PubSubConf.Topic.Name),
				Role:    pulumi.String(role.Role),
				Member: g.serviceAcc.Email.ApplyT(func(email string) (string, error) {
					return fmt.Sprintf("serviceAccount:%v", email), nil
				}).(pulumi.StringOutput),
			}, pulumi.DependsOn([]pulumi.Resource{g.topic}))
		} else if role.Type == "cloudrunbinding" {
			_, err = cloudrun.NewIamBinding(g.ctx, role.Name, &cloudrun.IamBindingArgs{
				Project:  pulumi.String(g.project),
				Service:  pulumi.String(g.config.Function.Name),
				Location: pulumi.String(g.region),
				Role:     pulumi.String(role.Role),
				Members: pulumi.StringArray{
					g.serviceAcc.Email.ApplyT(func(email string) (string, error) {
						return fmt.Sprintf("serviceAccount:%v", email), nil
					}).(pulumi.StringOutput),
				},
			}, pulumi.DependsOn([]pulumi.Resource{g.function}))
		}
	}

	return err
}

// ConfigureIAM configures the IAM member according to given values.
// You can create the multiple role.
func (g *Gcp) ConfigureIAM() error {
	project, _ := organizations.LookupProject(g.ctx, &organizations.LookupProjectArgs{ProjectId: pulumi.StringRef(g.config.Iam.ServiceAcc.Project)}, nil)

	var err error

	for _, role := range g.config.Iam.Roles {

		if role.Type == "bucketmember" {
			_, _ = storage.NewBucketIAMMember(g.ctx, role.Name, &storage.BucketIAMMemberArgs{
				Bucket: g.bucket.Name,
				Role:   pulumi.String(role.Role),
				Member: pulumi.String(fmt.Sprintf(role.Member, project.Number)),
			}, pulumi.DependsOn([]pulumi.Resource{g.bucket}))
		} else if role.Type == "projectmember" {
			_, _ = projects.NewIAMMember(g.ctx, role.Name, &projects.IAMMemberArgs{
				Project: pulumi.String(*project.ProjectId),
				Role:    pulumi.String(role.Role),
				Member:  pulumi.String(fmt.Sprintf(role.Member, project.Number)),
			})
		}
	}

	return err
}

// New returns Gcp struct
func New(ctx *pulumi.Context, yamlConf types.Config) *Gcp {
	conf := config.New(ctx, "gcp")
	project := conf.Require("project")
	region := conf.Require("region")

	return &Gcp{ctx: ctx, project: project, region: region, config: yamlConf}
}
