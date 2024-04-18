package types

type CloudProvider int64

const (
	Aws CloudProvider = iota + 1
	Gcp
	Azure
)

func (c CloudProvider) String() string {
	switch c {
	case Aws:
		return "aws"
	case Gcp:
		return "gcp"
	case Azure:
		return "azure"
	}
	return ""
}

var CloudMap = map[string]CloudProvider{"aws": Aws, "gcp": Gcp, "azure": Azure}

type AwsIamStruct struct {
	AWS []IamPolicies `mapstructure:"aws" yaml:"aws"`
}

type IamPolicies struct {
	Name         interface{} `mapstructure:"name"`
	InlinePolicy interface{} `mapstructure:"inlinePolicy" yaml:"inlinePolicy"`
	AssumePolicy interface{} `mapstructure:"assumePolicy" yaml:"assumePolicy"`
}
