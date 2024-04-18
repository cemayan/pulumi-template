package cloud

// GetFunctionMap return func array that will be executed functions
// instructions are coming from yaml
func GetFunctionMap(instructions []interface{}) []func() error {
	functionArr := []func() error{}
	for _, v := range instructions {
		i := v.(string)
		switch i {
		case "configureIAM":
			functionArr = append(functionArr, CloudInstance.ConfigureIAM)
		case "createVpc":
			functionArr = append(functionArr, CloudInstance.CreateVpc)
		case "createApiGateway":
			functionArr = append(functionArr, CloudInstance.CreateApiGateway)
		case "createStorage":
			functionArr = append(functionArr, CloudInstance.CreateStorage)
		case "createDWH":
			functionArr = append(functionArr, CloudInstance.CreateDWH)
		case "createStream":
			functionArr = append(functionArr, CloudInstance.CreateStream)
		case "createFunction":
			functionArr = append(functionArr, CloudInstance.CreateFunction)
		case "createIdentityManagement":
			functionArr = append(functionArr, CloudInstance.CreateIdentityManagement)
		default:
			panic("instruction is not defined")
		}
	}

	return functionArr
}
