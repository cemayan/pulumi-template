package types

// Config represents the config yaml
type Config struct {
	Env        string     `mapstructure:"env"`
	Cloud      string     `mapstructure:"cloud"`
	Template   Template   `mapstructure:"template"`
	Iam        Iam        `mapstructure:"iam"`
	Storage    Storage    `mapstructure:"storage"`
	Stream     Stream     `mapstructure:"stream"`
	Dwh        Dwh        `mapstructure:"dwh"`
	APIGateway APIGateway `mapstructure:"api_gateway"`
	Function   Function   `mapstructure:"function"`
	Authorizer Authorizer `mapstructure:"authorizer"`
	Idp        Idp        `mapstructure:"idp"`
}

type Idp struct {
	Enabled      bool   `mapstructure:"enabled"`
	IdpId        string `mapstructure:"idp_id"`
	SupportEmail string `mapstructure:"support_email"`
	DisplayName  string `mapstructure:"display_name"`
	ClientId     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type UserClient struct {
	Name              string   `mapstructure:"name"`
	AllowedScopes     []string `mapstructure:"allowed_scopes"`
	AllowedFlows      []string `mapstructure:"allowed_flows"`
	ExplicitAuthFlows []string `mapstructure:"ex_auth_flows"`
	CallbackUrls      []string `mapstructure:"callback_urls"`
}

type UserDomain struct {
	Name string `mapstructure:"name"`
}

type User struct {
	Email    string `mapstructure:"email"`
	Username string `mapstructure:"username"`
}

type UserPool struct {
	Name       string     `mapstructure:"name"`
	User       User       `mapstructure:"user"`
	UserClient UserClient `mapstructure:"user_client"`
	UserDomain UserDomain `mapstructure:"user_domain"`
}

type Authorizer struct {
	UserPool UserPool `mapstructure:"user_pool"`
	Name     string   `mapstructure:"name"`
	Type     string   `mapstructure:"type"`
}

type ServiceAcc struct {
	AccountID   string   `mapstructure:"account_id"`
	DisplayName string   `mapstructure:"display_name"`
	Project     string   `mapstructure:"project"`
	Location    string   `mapstructure:"location"`
	Role        string   `mapstructure:"role"`
	Members     []string `mapstructure:"members"`
}

type ServiceConf struct {
	MaxInstance  int    `mapstructure:"max_instance"`
	AvailableMem string `mapstructure:"available_mem"`
	Timeout      int    `mapstructure:"timeout"`
	Ingress      string `mapstructure:"ingress"`
}

type Source struct {
	Storage    Storage `mapstructure:"storage"`
	Zip        string  `mapstructure:"zip"`
	OutputPath string  `mapstructure:"output_path"`
}
type Build struct {
	Runtime    string            `mapstructure:"runtime"`
	Handler    string            `mapstructure:"handler"`
	EntryPoint string            `mapstructure:"entry_point"`
	DockerRepo string            `mapstructure:"docker_repo"`
	Source     *Source           `mapstructure:"source"`
	Envs       map[string]string `mapstructure:"envs"`
}
type Trigger struct {
	EventType string `mapstructure:"event_type"`
	Region    string `mapstructure:"region"`
}
type Function struct {
	Name        string      `mapstructure:"name"`
	Auth        string      `mapstructure:"auth"`
	Region      string      `mapstructure:"region"`
	Build       Build       `mapstructure:"build"`
	Trigger     *Trigger    `mapstructure:"trigger"`
	ServiceConf ServiceConf `mapstructure:"service_conf"`
}

type BigQuery struct {
	Dataset            string `mapstructure:"dataset"`
	TableId            string `mapstructure:"table_id"`
	DeletionProtection bool   `mapstructure:"delete_protection"`
	Schema             string `mapstructure:"schema"`
}

type Redshift struct {
	Identifier    string `mapstructure:"identifier"`
	DbName        string `mapstructure:"db_name"`
	MasterUser    string `mapstructure:"master_user"`
	MasterPass    string `mapstructure:"master_pass"`
	NodeType      string `mapstructure:"node_type"`
	NumberOfNodes int    `mapstructure:"number_of_nodes"`
	ClusterType   string `mapstructure:"cluster_type"`
	SkipSnapshot  bool   `mapstructure:"skip_snapshot"`
	Sql           string `mapstructure:"sql"`
	PublicAccess  bool   `mapstructure:"public_access"`
}

type Dwh struct {
	BigQuery BigQuery `mapstructure:"bq"`
	Redshift Redshift `mapstructure:"redshift"`
}

type Template struct {
	Name         string   `mapstructure:"name"`
	Instructions []string `mapstructure:"instructions"`
}

type Roles struct {
	Name                string `mapstructure:"name"`
	Role                string `mapstructure:"role"`
	Type                string `mapstructure:"type"`
	Member              string `mapstructure:"member"`
	ForceDetachPolicies bool   `mapstructure:"force_detach_policies"`
	AssumePolicy        string `mapstructure:"assume_policy"`
	InlinePolicy        string `mapstructure:"inline_policy"`
}
type Iam struct {
	ServiceAcc *ServiceAcc `mapstructure:"service_acc"`
	Roles      []Roles     `mapstructure:"roles"`
}

type BucketObject struct {
	Path string `mapstructure:"path"`
	Name string `mapstructure:"name"`
}

type Bucket struct {
	Name   string       `mapstructure:"name"`
	Object BucketObject `mapstructure:"object"`
}

type Storage struct {
	Name         string `mapstructure:"name"`
	Location     string `mapstructure:"location"`
	Bucket       Bucket `mapstructure:"bucket"`
	ForceDestroy bool   `mapstructure:"force_destroy"`
}
type S3Conf struct {
	BufferingSize     int    `mapstructure:"buffering_size"`
	BufferingInterval int    `mapstructure:"buffering_interval"`
	PartitionEnabled  bool   `mapstructure:"partition_enabled"`
	S3Prefix          string `mapstructure:"s3_prefix"`
}

type RedshiftConf struct {
	Username         string `mapstructure:"username"`
	Password         string `mapstructure:"password"`
	CopyOptions      string `mapstructure:"copy_options"`
	DataTableName    string `mapstructure:"data_table_name"`
	DataTableColumns string `mapstructure:"data_table_columns"`
}

type Topic struct {
	Name string `mapstructure:"name"`
}

type CloudStorageConf struct {
	Name       string `mapstructure:"name"`
	Duration   string `mapstructure:"duration"`
	FilePrefix string `mapstructure:"file_prefix"`
}

type Subscription struct {
	Name             string           `mapstructure:"name"`
	CloudStorageConf CloudStorageConf `mapstructure:"cloud_storage_conf"`
}

type PubSubConf struct {
	Topic        Topic        `mapstructure:"topic"`
	Subscription Subscription `mapstructure:"subscription"`
}

type Stream struct {
	Name         string       `mapstructure:"name"`
	Destination  string       `mapstructure:"destination"`
	PubSubConf   PubSubConf   `mapstructure:"pubsub_conf"`
	S3Conf       S3Conf       `mapstructure:"s3Config"`
	RedshiftConf RedshiftConf `mapstructure:"redshift_conf"`
}
type ResponseParams struct {
	Key string `mapstructure:"key"`
	Val bool   `mapstructure:"val"`
}
type Response struct {
	StatusCode     string           `mapstructure:"status_code"`
	ResponseParams []ResponseParams `mapstructure:"response_params"`
}
type Method struct {
	Name     string   `mapstructure:"name"`
	Type     string   `mapstructure:"type"`
	Auth     string   `mapstructure:"auth"`
	Response Response `mapstructure:"response"`
}
type ReqParams struct {
	Key string `mapstructure:"key"`
	Val string `mapstructure:"val"`
}
type ReqTemplate struct {
	Key string `mapstructure:"key"`
	Val string `mapstructure:"val"`
}
type ResTemplate struct {
	Key string `mapstructure:"key"`
	Val string `mapstructure:"val"`
}
type ResParams struct {
	Key string `mapstructure:"key"`
	Val string `mapstructure:"val"`
}

type Integrations struct {
	Name        string        `mapstructure:"name"`
	Type        string        `mapstructure:"type"`
	HTTPMethod  string        `mapstructure:"http_method"`
	URI         string        `mapstructure:"uri"`
	Method      Method        `mapstructure:"method"`
	ResParams   []ResParams   `mapstructure:"res_params"`
	ReqParams   []ReqParams   `mapstructure:"req_params"`
	ReqTemplate []ReqTemplate `mapstructure:"req_template"`
	ResTemplate []ResTemplate `mapstructure:"res_template"`
}
type Routes struct {
	Name         string         `mapstructure:"name"`
	Integrations []Integrations `mapstructure:"integrations"`
}
type APIGateway struct {
	Name         string   `mapstructure:"name"`
	Stage        string   `mapstructure:"stage"`
	DeploymentId int      `mapstructure:"deployment_id"`
	Region       string   `mapstructure:"region"`
	Routes       []Routes `mapstructure:"routes"`
	OpenApiSpec  string   `mapstructure:"open_api_spec"`
}
