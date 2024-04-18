# Pulumi Template

## Introduction

Pulumi template aims to build some useful infrastructures such as data-pipeline etc. on clouds.


Every template is use yaml file like below:

``` yaml
env: 
cloud: 
template:
iam:
storage:
dwh:
stream:
api_gateway:
```

You can found the example template [here](configs/datapipeline/firehose/s3/lambda/config.yaml)




## Usage


### Install Pulumi

https://www.pulumi.com/docs/install/

---


In order to run pulumi commands you need to log in first.
```bash
pulumi login
```
---

To zip from source-code(Since cloud functions need to zipped code):

```bash
make function-zip
```

---

**Pulumi up**:

```bash
make up stack=datapipeline-firehose-s3-lambda
```

> This command output gives confirmation prompt you need to press **ENTER**. It will be created a stack according to given stack.After that **pulumi up** command will be executed.

**Available stacks**:
- datapipeline-firehose-s3-lambda
- datapipeline-firehose-s3-apigateway
- datapipeline-firehose-redshift-apigateway
- datapipeline-pubsub-bigquery
- datapipeline-pubsub-storage

---
**Pulumi destroy:**

```bash
make destroy stack=datapipeline-firehose-s3-lamdda
```

--- 

####  New Event:

**Payload**:
```json
{
    "game_name": "simulatte",
    "event_name": "user.created",
    "event_data": {
      "foo": "bar"
    }
}
```

Since AWS IAM activated and GCP Auth enabled you need to take a token for authenticated request.

**For GCP** :

```bash
make gcp-token
```

**For AWS** :

https://docs.aws.amazon.com/IAM/latest/UserGuide/create-signed-request.html

---

## Infrastructures

**AWS** :

![assets/aws_ptemplate.svg](assets/aws_ptemplate.svg)

![assets/aws_ptemplate2.svg](assets/aws_ptemplate2.svg)

---
**GCP** :

![assets/gcp_ptemplate.svg](assets/gcp_ptemplate.svg)