default: help

arch?=$(shell go env GOARCH)
os?=$(shell go env GOOS)
ARCH=$(arch)
OS=$(os)
CGO=1
INFRA_MAIN=cmd/infra/main.go
FUNCTION_MAIN=cmd/function/main.go
PROJECT_FOLDER=.
ASSETS_FOLDER=assets
BIN_FOLDER=bin
INFRA_BIN_NAME=ptemplate
FUNCTION_BIN_NAME=function
STACK_NAME=${stack}
CONFIG_PATH=${cfg}
SECRET=${secret}
CONFIG_FILE="stacks/Pulumi.${STACK_NAME}.yaml"
GCP_PROJECT_NAME=pulumi-template
GCP_REGION=europe-west3



init:  function-zip set-config

infra-build:
	@echo "  >  Building binary for ${OS}-${ARCH}"
	CGO_ENABLED=${CGO} GOOS=${OS} GOARCH=${ARCH} go build -C ${PROJECT_FOLDER} -o ${BIN_FOLDER}/${INFRA_BIN_NAME} "${INFRA_MAIN}"

function-zip:
	cd ${PROJECT_FOLDER}/functions/gcp/pubsubproducer && zip -r ../../../${ASSETS_FOLDER}/cloudfunction/${FUNCTION_BIN_NAME}.zip * &
	cd ${PROJECT_FOLDER}/functions/aws/firehoseproducer && zip -r ../../../${ASSETS_FOLDER}/lambda/${FUNCTION_BIN_NAME}.zip *

remove-stack:
	pulumi stack rm  ${STACK_NAME} -f
select-stack:
	pulumi stack select  ${STACK_NAME}
set-config:
	pulumi config set --path 'config:path' ${CONFIG_PATH} -s ${STACK_NAME}
set-userpass:
	pulumi config set --secret 'config:userpass' ${SECRET} -s ${STACK_NAME}
set-gcp:
	pulumi config set gcp:project  ${GCP_PROJECT_NAME}
	pulumi config set functions/region  ${GCP_REGION}
up: infra-build
	pulumi up -s ${STACK_NAME} --config-file  ${CONFIG_FILE}
destroy: infra-build
	pulumi destroy -s ${STACK_NAME}
delete-state:
	pulumi state delete urn:pulumi:${STACK_NAME}::pulumi-template::pulumi:pulumi:Stack::pulumi-template-${STACK_NAME} -y --target-dependents
gcp-token:
	 gcloud auth print-identity-token