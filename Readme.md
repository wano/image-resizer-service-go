# wano/image-resizer-service-go

go version of [image-resizer-service](https://serverlessrepo.aws.amazon.com/applications/arn:aws:serverlessrepo:us-east-1:526515951862:applications~image-resizer-service)

## 1. deploy
```
STACK_NAME=xxxxx CODE_BUCKET=yyyy IMAGE_BUCKET=zzzz make deploy
```

## 2.  edit your cloudfront
```
Origin = <Generated-Your-ApiGateway-Domain>
OriginParh = `/production`
```