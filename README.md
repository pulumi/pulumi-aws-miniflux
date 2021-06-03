# miniflux

A multi-language [Pulumi](https://pulumi.com) component builder for [Miniflux](https://miniflux.app/), the excellent open-source RSS server. The resulting component deploys a containerized Miniflux service to the AWS cloud using [AWS Fargate](https://aws.amazon.com/fargate) and [Amazon RDS](https://aws.amazon.com/rds/).

## Generate and publish!

Publish currently works only for npm.

```
VERSION=0.0.11 make install generate publish
```

See https://www.npmjs.com/package/@cnunciato/miniflux for component usage.
