import pulumi
from pulumi_aws import s3
from pulumi_miniflux import miniflux_service

config = pulumi.Config();
admin_password = config.get_secret("adminPassword")
db_password = config.get_secret("dbPassword")

# Create a new Miniflux service.
service = miniflux_service.MinifluxService("service",
        admin_password = admin_password,
        db_password = db_password
    )

# Export the URL of the service.
pulumi.export("endpoint", service.endpoint)
