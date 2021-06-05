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
