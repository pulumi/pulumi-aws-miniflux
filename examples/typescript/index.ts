import * as pulumi from "@pulumi/pulumi";
import * as miniflux from "@cnunciato/miniflux";

const config = new pulumi.Config();
const dbPassword = config.requireSecret("dbPassword");
const adminPassword = config.requireSecret("adminPassword");

const service = new miniflux.MinifluxService("miniflux", {
    dbPassword,
    adminPassword,
});

export const endpoint = pulumi.interpolate`http://${service.endpoint}`;
