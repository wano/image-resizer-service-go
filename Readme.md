# wano/image-resizer-service-go

go version of [image-resizer-service](https://serverlessrepo.aws.amazon.com/applications/arn:aws:serverlessrepo:us-east-1:526515951862:applications~image-resizer-service)

## deploy

### 1. create & update stack
```
STACK_NAME=xxxxx CODE_BUCKET=yyyy IMAGE_BUCKET=zzzz make deploy
```

### 2. test access

```
<Generated-Your-ApiGateway-Domain>/production/<Your-S3-Image-Key>?width=512&auto_rotate=true
```

### 3.  edit your cloudfront
```
Origin = <Generated-Your-ApiGateway-Domain>
OriginParh = `/production`
```

## params  option
width : int
height : int
auto_rotate : boolean

