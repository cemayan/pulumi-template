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


You can read the related posts
- [AWS](https://cemayan.com/posts/datapipeline-on-aws-with-pulumi)
- [GCP](https://cemayan.com/posts/datapipeline-on-gcp-with-pulumi)


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
make up stack=datapipeline-firehose-redshift-apigateway
```

> This command output gives confirmation prompt you need to press **ENTER**. It will be created a stack according to given stack.After that **pulumi up** command will be executed.

**Available stacks**:
- datapipeline-firehose-s3-lambda
- datapipeline-firehose-s3-apigateway
- datapipeline-firehose-redshift-apigateway
- datapipeline-pubsub-bigquery-apigateway
- datapipeline-pubsub-bigquery-lambda
- datapipeline-pubsub-storage

---
**Pulumi destroy:**

```bash
make destroy stack=datapipeline-firehose-redshift-apigateway
```

--- 

####  New Event:

**Payload**:
```json
{
    "game_name": "amazing_game",
    "event_name": "weapon.fired",
    "event_data": {
      "weapon_name": "m4a1"
    }
}
```


---


### For internal security

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