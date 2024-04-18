package main

import (
	_cloud "github.com/cemayan/pulumi-template/internal/cloud"
	"github.com/cemayan/pulumi-template/types"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/spf13/viper"
)

func readConfigFromFile(path string) error {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName(path)   // path to look for the configs file in
	err := viper.ReadInConfig() // Find and read the configs file
	return err
}

func main() {

	pulumi.Run(func(ctx *pulumi.Context) error {

		conf := config.New(ctx, "config") // conf gives "config" value in pulumi config file.
		path := conf.Require("path")      // path gives "config:path" value in pulumi config file.

		err := readConfigFromFile(path)
		if err != nil {
			ctx.Log.Error("config file read error", nil)
			return err
		}

		// selectedCloud gives cloud that is given in yaml file
		selectedCloud := viper.GetString("cloud")

		ctx.Log.Info("selected cloud is "+selectedCloud, nil)

		appConfigs := types.Config{}
		_ = viper.Unmarshal(&appConfigs)

		// It will be initialized according the given cloud provider
		_cloud.Initialize(types.CloudMap[selectedCloud], ctx, appConfigs)

		return _cloud.Build()
	})
}
