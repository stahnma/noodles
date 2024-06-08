

all: tidy fmt
	go build .

fmt:
	go fmt .

tidy:
	go mod tidy

lambda: tidy fmt
	GOOS=linux GOARCH=arm64 go build -o bootstrap .
	zip function.zip bootstrap

# Note that this will only update the lambda code.
aws: lambda
	aws lambda update-function-code --function-name noodles --zip-file fileb://function.zip


# Note you'll still need something like an API Gateway setup to trigger the
# lambda. This is just the lambda setup.
aws-init: lambda
	aws lambda create-function \
    --function-name noodles \
    --runtime provided.al2023 \
    --role $(AWS_IAM_ROLE) \
    --architectures arm64 \
    --environment Variables={GITHUB_TOKEN=$(GITHUB_TOKEN)} \
    --handler bootstrap \
    --zip-file fileb://function.zip

clean:
	rm -f bootstrap function.zip noodles
