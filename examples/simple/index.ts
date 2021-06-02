import * as miniflux from "@pulumi/miniflux";

const page = new miniflux.Service("page", {
    indexContent: "<html><body><p>Hello world!</p></body></html>",
});

export const bucket = page.bucket;
export const url = page.websiteUrl;
