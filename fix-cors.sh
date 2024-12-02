#!/bin/bash

# Set variables
API_NAME="noodlesAPI"
STAGE_NAME="prod"

# Step 1: Get the API ID for noodlesAPI
echo "Fetching API ID for $API_NAME..."
API_ID=$(aws apigateway get-rest-apis --query "items[?name=='$API_NAME'].id" --output text)

if [ -z "$API_ID" ]; then
  echo "API Gateway '$API_NAME' not found."
  exit 1
fi
echo "API ID for $API_NAME: $API_ID"

# Step 2: Get all resources for the API
echo "Fetching resources for API ID $API_ID..."
RESOURCES=$(aws apigateway get-resources --rest-api-id "$API_ID")
RESOURCE_ID=$(echo "$RESOURCES" | jq -r '.items[] | select(.path=="/search").id')

if [ -z "$RESOURCE_ID" ]; then
  echo "Resource '/search' not found in API."
  exit 1
fi
echo "Resource ID for '/search': $RESOURCE_ID"

# Step 3: Add CORS configuration to GET method (if not already set)
echo "Checking GET method CORS configuration..."
GET_METHOD=$(aws apigateway get-method --rest-api-id "$API_ID" --resource-id "$RESOURCE_ID" --http-method GET 2>/dev/null)

if [[ "$GET_METHOD" == *"method.response.header.Access-Control-Allow-Origin"* ]]; then
  echo "GET method CORS headers already configured."
else
  echo "Adding CORS headers to GET method..."
  aws apigateway put-method-response \
    --rest-api-id "$API_ID" \
    --resource-id "$RESOURCE_ID" \
    --http-method GET \
    --status-code 200 \
    --response-models '{"application/json": "Empty"}' \
    --response-parameters '{"method.response.header.Access-Control-Allow-Origin": true, "method.response.header.Access-Control-Allow-Headers": true, "method.response.header.Access-Control-Allow-Methods": true}'

  aws apigateway put-integration-response \
    --rest-api-id "$API_ID" \
    --resource-id "$RESOURCE_ID" \
    --http-method GET \
    --status-code 200 \
    --response-parameters '{"method.response.header.Access-Control-Allow-Origin": "'*'", "method.response.header.Access-Control-Allow-Headers": "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token'", "method.response.header.Access-Control-Allow-Methods": "'GET,OPTIONS'"}'
fi

# Step 4: Add OPTIONS method for preflight requests
echo "Checking OPTIONS method..."
OPTIONS_METHOD=$(aws apigateway get-method --rest-api-id "$API_ID" --resource-id "$RESOURCE_ID" --http-method OPTIONS 2>/dev/null)

if [[ "$OPTIONS_METHOD" == *"OPTIONS"* ]]; then
  echo "OPTIONS method already exists."
else
  echo "Adding OPTIONS method..."
  aws apigateway put-method \
    --rest-api-id "$API_ID" \
    --resource-id "$RESOURCE_ID" \
    --http-method OPTIONS \
    --authorization-type "NONE"

  aws apigateway put-method-response \
    --rest-api-id "$API_ID" \
    --resource-id "$RESOURCE_ID" \
    --http-method OPTIONS \
    --status-code 200 \
    --response-parameters '{"method.response.header.Access-Control-Allow-Origin": true, "method.response.header.Access-Control-Allow-Headers": true, "method.response.header.Access-Control-Allow-Methods": true}'

  aws apigateway put-integration \
    --rest-api-id "$API_ID" \
    --resource-id "$RESOURCE_ID" \
    --http-method OPTIONS \
    --type MOCK \
    --request-templates '{"application/json": "{\"statusCode\": 200}"}'

  aws apigateway put-integration-response \
    --rest-api-id "$API_ID" \
    --resource-id "$RESOURCE_ID" \
    --http-method OPTIONS \
    --status-code 200 \
    --response-parameters '{"method.response.header.Access-Control-Allow-Origin": "'*'", "method.response.header.Access-Control-Allow-Headers": "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token'", "method.response.header.Access-Control-Allow-Methods": "'GET,OPTIONS'"}'
fi

# Step 5: Deploy the changes
echo "Deploying changes to stage $STAGE_NAME..."
aws apigateway create-deployment \
  --rest-api-id "$API_ID" \
  --stage-name "$STAGE_NAME"

echo "CORS configuration complete and deployed for $API_NAME on /search."

