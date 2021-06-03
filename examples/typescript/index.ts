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
