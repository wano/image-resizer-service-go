
build:
	GOOS=linux GOARCH=amd64 go build -o ./dist/build/main ./lambda/...
	zip -r ./dist/bundle.zip ./dist/build/
package:
	aws cloudformation package --template-file template.yaml --output-template-file dist/template-packaged.yml --s3-bucket ${CODE_BUCKET}
install-cf:
	aws cloudformation deploy --template-file dist/template-packaged.yml --capabilities CAPABILITY_IAM --stack-name ${STACK_NAME} --parameter-overrides ImageBucket=${IMAGE_BUCKET}

deploy:
	make build && make package && make install-cf
#    "build": "webpack --mode production",
#    "package": "npm run build && aws cloudformation package --template-file template.yaml --output-template-file dist/template-packaged.yml --s3-bucket $CODE_BUCKET",
#    "install-cf": "aws cloudformation deploy --template-file dist/template-packaged.yml --capabilities CAPABILITY_IAM --stack-name $STACK_NAME --parameter-overrides ImageBucket=$IMAGE_BUCKET",
#    "deploy": "npm run package && npm run install-cf",
