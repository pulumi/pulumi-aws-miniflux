# miniflux

A multi-language [Pulumi](https://pulumi.com) component builder for [Miniflux](https://miniflux.app/), the excellent open-source RSS server. The resulting component deploys a containerized Miniflux service to the AWS cloud using [AWS Fargate](https://aws.amazon.com/fargate) and a managed PostgreSQL database with [Amazon RDS](https://aws.amazon.com/rds/).

This repository is used for building and publishing the binaries and language-specific SDKs that let work with `MinifluxService` instances in any Pulumi-supported language. The base component, written in Go, produces installable packages for Node.js, Python, Go, and .NET.

## Using the components

All components require Pulumi, of course, along with the `pulumi-resource-miniflux` provider, which today must be installed separately. To do that:

```
pulumi plugin install resource miniflux 0.0.16 \
    --server http://cnunciato-pulumi-components.s3-website-us-west-2.amazonaws.com
```

### Node.js

On the command line:

```
$ pulumi new typescript
$ npm install --save @cnunciato/miniflux
$ pulumi config set --secret adminPassword "some-secret-password"
$ pulumi config set --secret dbPassword "some-other-secret-password"
```

In `index.ts`:

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as miniflux from "@cnunciato/miniflux";

const config = new pulumi.Config();
const adminPassword = config.requireSecret("adminPassword");
const dbPassword = config.requireSecret("adminPassword");

// Create a new Miniflux service.
const service = new miniflux.MinifluxService("service", {
    adminPassword,
    dbPassword,
});

// Export the URL of the service.
export const endpoint = pulumi.interpolate`http://${service.endpoint}`;
```

### Go

On the command line:

```
$ pulumi new go
$ go get github.com/cnunciato/pulumi-miniflux/sdk/go/miniflux
$ pulumi config set --secret adminPassword "some-secret-password"
$ pulumi config set --secret dbPassword "some-other-secret-password"
```

In `main.go`:

```go
package main

import (
	"github.com/cnunciato/pulumi-miniflux/sdk/go/miniflux"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		conf := config.New(ctx, "")
		adminPassword := conf.RequireSecret("adminPassword")
		dbPassword := conf.RequireSecret("dbPassword")

		// Create a new Miniflux service.
		service, err := miniflux.NewMinifluxService(ctx, "service", &miniflux.MinifluxServiceArgs{
			AdminPassword: adminPassword,
			DbPassword:    dbPassword,
		})
		if err != nil {
			return nil
		}

		// Export the URL of the service.
		ctx.Export("endpoint", pulumi.Sprintf("http://%s", service.Endpoint))
		return nil
	})
}
```

### C#

On the command line:

```
$ pulumi new csharp
$ dotnet add package Pulumi.Miniflux
$ pulumi config set --secret adminPassword "some-secret-password"
$ pulumi config set --secret dbPassword "some-other-secret-password"
```

In `MyStack.cs`:

```csharp
using Pulumi;
using Pulumi.Aws.S3;
using Pulumi.Miniflux;

class MyStack : Stack
{
    public MyStack()
    {
        var config = new Pulumi.Config();
        var adminPassword = config.Require("adminPassword");
        var dbPassword = config.Require("dbPassword");

        // Create a new Miniflux service.
        var service = new Pulumi.Miniflux.MinifluxService("service", new Pulumi.Miniflux.MinifluxServiceArgs{
            AdminPassword = adminPassword,
            DbPassword = dbPassword,
        });

        // Export the URL of the service.
        this.Endpoint = Output.Format($"http://{service.Endpoint}");
    }

    [Output]
    public Output<string> Endpoint { get; set; }
}
```

## Generating and publishing component packages

Right now, these instructions are mainly for me, but you might find them useful to refer to as well, as you build your own multi-language components:

```
$ export VERSION=0.0.16   # Sets the target package version.
$ make install generate   # To build the provider and generate all four language SDKs.
$ make publish            # To publish to npm, nuget, and (soon) PyPi.
```

## See also

* [Pulumi Packages](https://www.pulumi.com/docs/guides/pulumi-packages/)
* [Pulumi Package Schema](https://www.pulumi.com/docs/guides/pulumi-packages/schema/)
* [Introducing Packages and Multi-Language Components](https://www.pulumi.com/blog/pulumiup-pulumi-packages-multi-language-components/)
