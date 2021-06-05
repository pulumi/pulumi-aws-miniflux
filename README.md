# pulumi-miniflux

A multi-language [Pulumi](https://pulumi.com) component builder for [Miniflux](https://miniflux.app/), the excellent open-source RSS server.

This repository is used for building and publishing the binaries and language-specific SDKs that let you deploy your own Miniflux server using Pulumi and [any language Pulumi supports](https://www.pulumi.com/docs/intro/languages/). The `MinifluxService` component itself deploys a containerized Miniflux service with [AWS Fargate](https://aws.amazon.com/fargate) and a managed PostgreSQL database with [Amazon RDS](https://aws.amazon.com/rds/), and is available via common package managers, including:

* npm, for Node.js: https://www.npmjs.com/package/@cnunciato/miniflux
* PyPI, for Python: https://pypi.org/project/pulumi-miniflux/
* NuGet, for C#: https://www.nuget.org/packages/Pulumi.Miniflux/
* GitHub (in [this repo](./sdk/go)), for Go

## Using the components

All components require Pulumi, of course, along with the `pulumi-resource-miniflux` provider, which today must be installed separately. To do that:

```
pulumi plugin install resource miniflux 0.0.16 \
    --server http://cnunciato-pulumi-components.s3-website-us-west-2.amazonaws.com
```

Then, assuming you've [configured Pulumi and AWS](https://www.pulumi.com/docs/intro/cloud-providers/aws/), you can follow the instructions below to use the component in your language of choice.

Full examples in [`./examples`](./examples).

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
        var adminPassword = config.RequireSecret("adminPassword");
        var dbPassword = config.RequireSecret("dbPassword");

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

### Python

On the command line:

```
$ pulumi new python
$ pip install pulumi_miniflux
$ pulumi config set --secret adminPassword "some-secret-password"
$ pulumi config set --secret dbPassword "some-other-secret-password"
```

In `__main.py__`:

```python
import pulumi
from pulumi_aws import s3
from pulumi_miniflux import miniflux_service

config = pulumi.Config();
admin_password = config.get_secret("adminPassword")
db_password = config.get_secret("dbPassword")

service = miniflux_service.MinifluxService("service",
        admin_password = admin_password,
        db_password = db_password
    )

pulumi.export("endpoint", service.endpoint)
```

## Generating and publishing component packages

Right now, these instructions are mainly for me, but you might find them useful to refer to as well, as you build your own multi-language components:

```
$ export VERSION=0.0.16   # Sets the target package version.
$ make install generate   # Builds the provider and generates all four language SDKs.
$ make publish            # Publishes to npm, nuget, and PyPI.
```

## See also

* [Pulumi Packages](https://www.pulumi.com/docs/guides/pulumi-packages/)
* [Pulumi Package Schema](https://www.pulumi.com/docs/guides/pulumi-packages/schema/)
* [Introducing Packages and Multi-Language Components](https://www.pulumi.com/blog/pulumiup-pulumi-packages-multi-language-components/)
