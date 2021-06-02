// Copyright 2016-2021, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ecs"
	elb "github.com/pulumi/pulumi-aws/sdk/v4/go/aws/elasticloadbalancingv2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The set of arguments for creating a Service component resource.
type StaticPageArgs struct {
	// The HTML content for index.html.
	AdminPassword pulumi.StringInput `pulumi:"admin_password"`
}

// The Service component resource.
type Service struct {
	pulumi.ResourceState

	Endpoint pulumi.StringOutput `pulumi:"endpoint"`
}

// NewService creates a new Service component resource.
func NewService(ctx *pulumi.Context,
	name string, args *StaticPageArgs, opts ...pulumi.ResourceOption) (*Service, error) {
	if args == nil {
		args = &StaticPageArgs{}
	}

	component := &Service{}
	err := ctx.RegisterComponentResource("miniflux:index:Service", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// Read back the default VPC and public subnets, which we will use.
	t := true
	vpc, err := ec2.LookupVpc(ctx, &ec2.LookupVpcArgs{Default: &t})
	if err != nil {
		return nil, err
	}
	subnet, err := ec2.GetSubnetIds(ctx, &ec2.GetSubnetIdsArgs{VpcId: vpc.Id})
	if err != nil {
		return nil, err
	}

	// Create a SecurityGroup that permits HTTP ingress and unrestricted egress.
	webSg, err := ec2.NewSecurityGroup(ctx, "web-sg", &ec2.SecurityGroupArgs{
		VpcId: pulumi.String(vpc.Id),
		Egress: ec2.SecurityGroupEgressArray{
			ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Ingress: ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(80),
				ToPort:     pulumi.Int(80),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Create an ECS cluster to run a container-based service.
	cluster, err := ecs.NewCluster(ctx, "app-cluster", nil)
	if err != nil {
		return nil, err
	}

	// Create an IAM role that can be used by our service's task.
	taskExecRole, err := iam.NewRole(ctx, "task-exec-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2008-10-17",
			"Statement": [
				{
					"Sid": "",
					"Effect": "Allow",
					"Principal": {
						"Service": "ecs-tasks.amazonaws.com"
					},
					"Action": "sts:AssumeRole"
				}
			]
		}`),
	})
	if err != nil {
		return nil, err
	}
	_, err = iam.NewRolePolicyAttachment(ctx, "task-exec-policy", &iam.RolePolicyAttachmentArgs{
		Role:      taskExecRole.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
	})
	if err != nil {
		return nil, err
	}

	// Create a load balancer to listen for HTTP traffic on port 80.
	webLb, err := elb.NewLoadBalancer(ctx, "web-lb", &elb.LoadBalancerArgs{
		Subnets:        toPulumiStringArray(subnet.Ids),
		SecurityGroups: pulumi.StringArray{webSg.ID().ToStringOutput()},
	})
	if err != nil {
		return nil, err
	}
	webTg, err := elb.NewTargetGroup(ctx, "web-tg", &elb.TargetGroupArgs{
		Port:       pulumi.Int(80),
		Protocol:   pulumi.String("HTTP"),
		TargetType: pulumi.String("ip"),
		VpcId:      pulumi.String(vpc.Id),
	})
	if err != nil {
		return nil, err
	}

	webListener, err := elb.NewListener(ctx, "web-listener", &elb.ListenerArgs{
		LoadBalancerArn: webLb.Arn,
		Port:            pulumi.Int(80),
		DefaultActions: elb.ListenerDefaultActionArray{
			elb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: webTg.Arn,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// image, err := docker.NewImage(ctx, "my-image", &docker.ImageArgs{
	// 	Build: docker.DockerBuildArgs{
	// 		Context: pulumi.String("./app"),
	// 	},
	// 	ImageName: repo.RepositoryUrl,
	// 	Registry: docker.ImageRegistryArgs{
	// 		Server:   repo.RepositoryUrl,
	// 		Username: repoUser,
	// 		Password: repoPass,
	// 	},
	// })

	// containerDef := image.ImageName.ApplyT(func(name string) (string, error) {
	// 	fmtstr := `[{
	// 		"name": "my-app",
	// 		"image": "miniflux/miniflux:latest",
	// 		"portMappings": [{
	// 			"containerPort": 80,
	// 			"hostPort": 80,
	// 			"protocol": "tcp"
	// 		}]
	// 	}]`
	// 	return fmt.Sprintf(fmtstr, name), nil
	// }).(pulumi.StringOutput)

	// Spin up a load balanced service.
	appTask, err := ecs.NewTaskDefinition(ctx, "app-task", &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String("fargate-task-definition"),
		Cpu:                     pulumi.String("256"),
		Memory:                  pulumi.String("512"),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		ExecutionRoleArn:        taskExecRole.Arn,
		ContainerDefinitions: pulumi.String(`[{
			"name": "my-app",
			"image": "miniflux/miniflux:latest",
			"portMappings": [{
				"containerPort": 80,
				"hostPort": 80,
				"protocol": "tcp"
			}]
		}]`),
	})
	if err != nil {
		return nil, err
	}

	_, err = ecs.NewService(ctx, "app-svc", &ecs.ServiceArgs{
		Cluster:        cluster.Arn,
		DesiredCount:   pulumi.Int(5),
		LaunchType:     pulumi.String("FARGATE"),
		TaskDefinition: appTask.Arn,
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			AssignPublicIp: pulumi.Bool(true),
			Subnets:        toPulumiStringArray(subnet.Ids),
			SecurityGroups: pulumi.StringArray{webSg.ID().ToStringOutput()},
		},
		LoadBalancers: ecs.ServiceLoadBalancerArray{
			ecs.ServiceLoadBalancerArgs{
				TargetGroupArn: webTg.Arn,
				ContainerName:  pulumi.String("my-app"),
				ContainerPort:  pulumi.Int(80),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{webListener}))

	// Create a PostgreSQL database.

	// Create a Fargate service.

	// // Create a bucket and expose a website index document.
	// bucket, err := s3.NewBucket(ctx, name, &s3.BucketArgs{
	// 	Website: s3.BucketWebsiteArgs{
	// 		IndexDocument: pulumi.String("index.html"),
	// 	},
	// }, pulumi.Parent(component))
	// if err != nil {
	// 	return nil, err
	// }

	// // Create a bucket object for the index document.
	// if _, err := s3.NewBucketObject(ctx, name, &s3.BucketObjectArgs{
	// 	Bucket:      bucket.ID(),
	// 	Key:         pulumi.String("index.html"),
	// 	Content:     args.IndexContent,
	// 	ContentType: pulumi.String("text/html"),
	// }, pulumi.Parent(bucket)); err != nil {
	// 	return nil, err
	// }

	// // Set the access policy for the bucket so all objects are readable.
	// if _, err := s3.NewBucketPolicy(ctx, "bucketPolicy", &s3.BucketPolicyArgs{
	// 	Bucket: bucket.ID(),
	// 	Policy: pulumi.Any(map[string]interface{}{
	// 		"Version": "2012-10-17",
	// 		"Statement": []map[string]interface{}{
	// 			{
	// 				"Effect":    "Allow",
	// 				"Principal": "*",
	// 				"Action": []interface{}{
	// 					"s3:GetObject",
	// 				},
	// 				"Resource": []interface{}{
	// 					pulumi.Sprintf("arn:aws:s3:::%s/*", bucket.ID()), // policy refers to bucket name explicitly
	// 				},
	// 			},
	// 		},
	// 	}),
	// }, pulumi.Parent(bucket)); err != nil {
	// 	return nil, err
	// }

	// component.Bucket = bucket
	// component.WebsiteUrl = bucket.WebsiteEndpoint

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"url": webLb.DnsName,
	}); err != nil {
		return nil, err
	}

	return component, nil
}

func toPulumiStringArray(a []string) pulumi.StringArrayInput {
	var res []pulumi.StringInput
	for _, s := range a {
		res = append(res, pulumi.String(s))
	}
	return pulumi.StringArray(res)
}
