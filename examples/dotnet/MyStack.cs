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
