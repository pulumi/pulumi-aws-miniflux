# miniflux

A multi-language [Pulumi](https://pulumi.com) component for [Miniflux](https://miniflux.app/), the excellent open-source RSS server. This component deploys a containerized Miniflux service to the AWS cloud using [AWS Fargate](https://aws.amazon.com/fargate) and [Amazon RDS](https://aws.amazon.com/rds/).

## Usage

To use this component, you'll need:

* An AWS account, with credentials (`AWS_ACCESS_KEY_ID`, etc.) properly set.
* A relatively recent (>= 3.3.1) version of [Pulumi](https://pulumi.com) installed
* Node.js, of course

### Create a new Pulumi project

```
$ pulumi new typescript
```

### Install this package and its associated plugin

The plugin is currently hosted on AWS.

```
$ npm install --save @cnunciato/miniflux
$ pulumi plugin install resource miniflux ${VERSION} \
    --server http://cnunciato-pulumi-components.s3-website-us-west-2.amazonaws.com
```

### Provide passwords for the Miniflux admin and database users

```
$ pulumi config set --secret adminPassword "some-secret-password"
$ pulumi config set --secret dbPassword "some-other-secret-password"
```

### Write the Pulumi program

Here's a complete example. In `index.ts`:

```
import * as pulumi from "@pulumi/pulumi";
import * as miniflux from "@cnunciato/miniflux";

const config = new pulumi.Config();
const adminPassword = config.requireSecret("adminPassword");
const dbPassword = config.requireSecret("adminPassword");

const service = new miniflux.MinifluxService("service", {
    adminPassword,
    dbPassword,
});

export const endpoint = pulumi.interpolate`http://${service.endpoint}`;
```

### Deploy with Pulumi!

```
$ pulumi up

...
Updating (dev)

...
     Type                                           Name                 Status
 +   pulumi:pulumi:Stack                            miniflux-typescript  created
 +   └─ miniflux:service:MinifluxService            service              created
 +      ├─ aws:cloudwatch:LogGroup                  logGroup             created
 +      ├─ aws:iam:Role                             taskRole             created
 +      ├─ aws:ecs:Cluster                          cluster              created
 +      ├─ aws:ec2:SecurityGroup                    dbSecurityGroup      created
 +      ├─ aws:rds:SubnetGroup                      dbsubnets            created
 +      ├─ aws:ec2:SecurityGroup                    webSecurityGroup     created
 +      ├─ aws:elasticloadbalancingv2:TargetGroup   targetGroup          created
 +      ├─ aws:iam:RolePolicyAttachment             taskRolePolicy       created
 +      ├─ aws:rds:Instance                         db                   created
 +      ├─ aws:elasticloadbalancingv2:LoadBalancer  loadBalancer         created
 +      ├─ aws:elasticloadbalancingv2:Listener      webListener          created
 +      ├─ aws:ecs:TaskDefinition                   taskDefinition       created
 +      └─ aws:ecs:Service                          fargateService       created

Resources:
    + 15 created

Duration: 3m13s
```
