package provider

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ecs"
	elb "github.com/pulumi/pulumi-aws/sdk/v4/go/aws/elasticloadbalancingv2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The set of arguments for creating a Service component resource.
type MinifluxServiceArgs struct {
	DbName        pulumi.StringInput `pulumi:"dbName"`
	DbUsername    pulumi.StringInput `pulumi:"dbUsername"`
	DbPassword    pulumi.StringInput `pulumi:"dbPassword"`
	AdminUsername pulumi.StringInput `pulumi:"adminUsername"`
	AdminPassword pulumi.StringInput `pulumi:"adminPassword"`
}

// The MinifluxService component resource.
type MinifluxService struct {
	pulumi.ResourceState

	Endpoint pulumi.StringOutput `pulumi:"endpoint"`
}

func NewMinifluxService(ctx *pulumi.Context,
	name string, args *MinifluxServiceArgs, opts ...pulumi.ResourceOption) (*MinifluxService, error) {

	if args == nil {
		args = &MinifluxServiceArgs{}
	}

	component := &MinifluxService{}

	serviceName := "miniflux-service"

	err := ctx.RegisterComponentResource("miniflux:service:MinifluxService", name, component, opts...)
	if err != nil {
		return nil, err
	}

	region, err := aws.GetRegion(ctx, nil, nil)
	if err != nil {
		return nil, err
	}

	t := true
	vpc, err := ec2.LookupVpc(ctx, &ec2.LookupVpcArgs{Default: &t})
	if err != nil {
		return nil, err
	}

	subnets, err := ec2.GetSubnetIds(ctx, &ec2.GetSubnetIdsArgs{VpcId: vpc.Id})
	if err != nil {
		return nil, err
	}

	logGroup, err := cloudwatch.NewLogGroup(ctx, "logGroup", &cloudwatch.LogGroupArgs{}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Create a SecurityGroup that permits HTTP ingress and unrestricted egress.
	webSecurityGroup, err := ec2.NewSecurityGroup(ctx, "webSecurityGroup", &ec2.SecurityGroupArgs{
		VpcId: pulumi.String(vpc.Id),
		Ingress: ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(80),
				ToPort:     pulumi.Int(80),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
			ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(8080),
				ToPort:     pulumi.Int(8080),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	dbSecurityGroup, err := ec2.NewSecurityGroup(ctx, "dbSecurityGroup", &ec2.SecurityGroupArgs{
		VpcId: pulumi.String(vpc.Id),
		Ingress: ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(5432),
				ToPort:     pulumi.Int(5432),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Create an ECS cluster to run the service.
	cluster, err := ecs.NewCluster(ctx, "cluster", nil, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	// Create an IAM role that can be assumed by the service task.
	taskRole, err := iam.NewRole(ctx, "taskRole", &iam.RoleArgs{
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
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	if _, err := iam.NewRolePolicyAttachment(ctx, "taskRolePolicy", &iam.RolePolicyAttachmentArgs{
		Role:      taskRole.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
	}, pulumi.Parent(component)); err != nil {
		return nil, err
	}

	// Create a load balancer to listen for HTTP traffic on port 80.
	loadBalancer, err := elb.NewLoadBalancer(ctx, "loadBalancer", &elb.LoadBalancerArgs{
		Subnets:        pulumi.ToStringArray(subnets.Ids),
		SecurityGroups: pulumi.StringArray{webSecurityGroup.ID().ToStringOutput()},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	targetGroup, err := elb.NewTargetGroup(ctx, "targetGroup", &elb.TargetGroupArgs{
		Port:       pulumi.Int(8080),
		Protocol:   pulumi.String("HTTP"),
		TargetType: pulumi.String("ip"),
		VpcId:      pulumi.String(vpc.Id),
		HealthCheck: elb.TargetGroupHealthCheckArgs{
			HealthyThreshold: pulumi.Int(2),
			Interval:         pulumi.Int(5),
			Timeout:          pulumi.Int(4),
			Protocol:         pulumi.String("HTTP"),
			Matcher:          pulumi.String("200-399"),
		},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	webListener, err := elb.NewListener(ctx, "webListener", &elb.ListenerArgs{
		LoadBalancerArn: loadBalancer.Arn,
		Port:            pulumi.Int(80),
		DefaultActions: elb.ListenerDefaultActionArray{
			elb.ListenerDefaultActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: targetGroup.Arn,
			},
		},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	dbSubnets, err := rds.NewSubnetGroup(ctx, "dbsubnets", &rds.SubnetGroupArgs{
		SubnetIds: pulumi.ToStringArray(subnets.Ids),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	db, err := rds.NewInstance(ctx, "db", &rds.InstanceArgs{
		InstanceClass:       pulumi.String("db.t3.micro"),
		Engine:              pulumi.String("postgres"),
		AllocatedStorage:    pulumi.Int(10),
		Name:                args.DbName,
		Username:            args.DbUsername,
		Password:            args.DbPassword,
		SkipFinalSnapshot:   pulumi.Bool(true),
		DbSubnetGroupName:   dbSubnets.ID().ToStringOutput(),
		VpcSecurityGroupIds: pulumi.StringArray{dbSecurityGroup.ID().ToStringOutput()},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	containerDefinitions := pulumi.Sprintf(`[
		{
			"name": "%s",
			"image": "miniflux/miniflux:latest",
			"portMappings": [
				{
					"containerPort": 8080,
					"hostPort": 8080,
					"protocol": "tcp"
				}
			],
			"environment": [
				{
					"name": "DATABASE_URL",
					"value": "postgres://%s:%s@%s/miniflux?sslmode=disable"
				},
				{
					"name": "RUN_MIGRATIONS",
					"value": "1"
				},
				{
					"name": "CREATE_ADMIN",
					"value": "1"
				},
				{
					"name": "ADMIN_USERNAME",
					"value": "%s"
				},
				{
					"name": "ADMIN_PASSWORD",
					"value": "%s"
				}
			],
			"logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "%s",
					"awslogs-region": "%s",
                    "awslogs-stream-prefix": "%s"
                }
            }
		}
	]`, serviceName, args.DbUsername, args.DbPassword, db.Endpoint, args.AdminUsername, args.AdminPassword, logGroup.Name, region.Name, serviceName)

	taskDefinition, err := ecs.NewTaskDefinition(ctx, "taskDefinition", &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String("fargate-task-definition"),
		Cpu:                     pulumi.String("256"),
		Memory:                  pulumi.String("512"),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
		ExecutionRoleArn:        taskRole.Arn,
		ContainerDefinitions:    containerDefinitions,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	_, err = ecs.NewService(ctx, "fargateService", &ecs.ServiceArgs{
		Cluster:        cluster.Arn,
		DesiredCount:   pulumi.Int(1),
		LaunchType:     pulumi.String("FARGATE"),
		TaskDefinition: taskDefinition.Arn,

		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			AssignPublicIp: pulumi.Bool(true),
			Subnets:        pulumi.ToStringArray(subnets.Ids),
			SecurityGroups: pulumi.StringArray{webSecurityGroup.ID().ToStringOutput()},
		},
		LoadBalancers: ecs.ServiceLoadBalancerArray{
			ecs.ServiceLoadBalancerArgs{
				TargetGroupArn: targetGroup.Arn,
				ContainerName:  pulumi.String(serviceName),
				ContainerPort:  pulumi.Int(8080),
			},
		},
	}, pulumi.Parent(component), pulumi.DependsOn([]pulumi.Resource{webListener}))
	if err != nil {
		return nil, err
	}

	component.Endpoint = loadBalancer.DnsName

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"endpoint": loadBalancer.DnsName,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
