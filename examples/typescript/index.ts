import * as pulumi from "@pulumi/pulumi";
import * as miniflux from "@pulumi/aws-miniflux";

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
