package types

type OpenApiSpec struct {
	Swagger             string              `json:"swagger"`
	Info                Info                `json:"info"`
	SecurityDefinitions SecurityDefinitions `json:"securityDefinitions" yaml:"securityDefinitions"`
	Schemes             []string            `json:"schemes"`
	Produces            []string            `json:"produces"`
	Paths               Paths               `json:"paths"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Paths struct {
	Event Event `json:"/event" yaml:"/event"`
}

type Event struct {
	Post Post `json:"post"`
}

type Post struct {
	Summary        string         `json:"summary"`
	OperationID    string         `json:"operationId" yaml:"operationId"`
	XGoogleBackend XGoogleBackend `json:"x-google-backend" yaml:"x-google-backend"`
	Security       []Security     `json:"security"`
	Responses      Responses      `json:"responses"`
}

type Responses struct {
	Num200 Num200 `json:"200" yaml:"200"`
}

type Num200 struct {
	Description string `json:"description"`
}

type Security struct {
	GoogleIDToken []interface{} `json:"google_id_token" yaml:"google_id_token"`
}

type XGoogleBackend struct {
	Address string `json:"address"`
}

type SecurityDefinitions struct {
	GoogleIDToken GoogleIDToken `json:"google-id-token" yaml:"google-id-token"`
}

type GoogleIDToken struct {
	AuthorizationURL string `json:"authorizationUrl" yaml:"authorizationUrl"`
	Flow             string `json:"flow"`
	Type             string `json:"type"`
	XGoogleIssuer    string `json:"x-google-issuer" yaml:"x-google-issuer"`
	XGoogleJwksURI   string `json:"x-google-jwks_uri" yaml:"x-google-jwks_uri"`
	XGoogleAudiences string `json:"x-google-audiences" yaml:"x-google-audiences"`
}
