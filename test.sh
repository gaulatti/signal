#!/bin/bash

# Signal API Testing Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
API_KEY="test-api-key-123"
BASE_URL="${BASE_URL:-http://localhost:8080}"
TENANT_ID="tenant-demo"

# Function to generate digest
generate_digest() {
    local api_key=$1
    local date=$(date -u +%Y-%m-%d)
    echo -n "${api_key}${date}" | openssl md5 | awk '{print $2}'
}

# Function to make authenticated request
auth_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    local digest=$(generate_digest "$API_KEY")
    
    if [ -n "$data" ]; then
        curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "Authorization: Digest $digest" \
            -H "Content-Type: application/json" \
            -d "$data"
    else
        curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "Authorization: Digest $digest"
    fi
}

# Print usage
usage() {
    echo -e "${BLUE}Signal API Testing Script${NC}"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  health                 - Check service health"
    echo "  digest                 - Generate auth digest for today"
    echo "  register <user_id>     - Register a device token"
    echo "  push <user_id>         - Send push to specific user"
    echo "  push-apns <user_id>    - Send APNS push to specific user"
    echo "  push-fcm <user_id>     - Send FCM push to specific user"
    echo "  broadcast              - Send push to all users"
    echo "  sql                    - Generate SQL to create test tenant and API key"
    echo "  setup-tenant <name>    - Generate SQL to create new tenant"
    echo "  env                    - Show current environment configuration"
    echo ""
    echo "Environment variables:"
    echo "  API_KEY     - API key to use (default: test-api-key-123)"
    echo "  BASE_URL    - Service URL (default: http://localhost:8080)"
    echo "  TENANT_ID   - Tenant ID (default: tenant-demo)"
    echo ""
    echo "Examples:"
    echo "  # Test with different port"
    echo "  BASE_URL=http://localhost:3000 $0 health"
    echo ""
    echo "  # Test with custom API key"
    echo "  API_KEY=my-api-key $0 register user123"
}

case "$1" in
    "health")
        echo -e "${BLUE}Checking service health...${NC}"
        response=$(curl -s "$BASE_URL/health")
        if [ "$response" = "OK" ]; then
            echo -e "${GREEN}✅ Service is healthy${NC}"
        else
            echo -e "${RED}❌ Service is not healthy${NC}"
            exit 1
        fi
        ;;
    
    "digest")
        digest=$(generate_digest "$API_KEY")
        date=$(date -u +%Y-%m-%d)
        echo -e "${BLUE}API Key:${NC} $API_KEY"
        echo -e "${BLUE}Date:${NC} $date"
        echo -e "${BLUE}Digest:${NC} $digest"
        ;;
    
    "register")
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Please provide user_id${NC}"
            echo "Usage: $0 register <user_id>"
            exit 1
        fi
        
        user_id=$2
        device_token="device_${user_id}_$(date +%s)"
        
        echo -e "${BLUE}Registering device for user: $user_id${NC}"
        
        data="{
            \"device_token\": \"$device_token\",
            \"user_id\": \"$user_id\",
            \"platform\": \"ios\"
        }"
        
        response=$(auth_request "POST" "/register" "$data")
        echo -e "${GREEN}Response:${NC} $response"
        ;;
    
    "push")
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Please provide user_id${NC}"
            echo "Usage: $0 push <user_id>"
            exit 1
        fi
        
        user_id=$2
        
        echo -e "${BLUE}Sending push notification to user: $user_id${NC}"
        
        data="{
            \"user_id\": \"$user_id\",
            \"title\": \"Test Notification\",
            \"body\": \"Hello from Signal API at $(date)\",
            \"data\": {\"test\": true, \"timestamp\": $(date +%s)}
        }"
        
        response=$(auth_request "POST" "/push" "$data")
        echo -e "${GREEN}Response:${NC} $response"
        ;;
    
    "push-apns")
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Please provide user_id${NC}"
            echo "Usage: $0 push-apns <user_id>"
            exit 1
        fi
        
        user_id=$2
        device_token="ios_device_${user_id}_$(date +%s)"
        
        echo -e "${BLUE}Sending APNS push notification to user: $user_id${NC}"
        
        data="{
            \"user_id\": \"$user_id\",
            \"device_token\": \"$device_token\",
            \"title\": \"APNS Test\",
            \"body\": \"Hello from APNS at $(date)\",
            \"data\": {\"platform\": \"ios\", \"timestamp\": $(date +%s)}
        }"
        
        response=$(auth_request "POST" "/push/apns" "$data")
        echo -e "${GREEN}Response:${NC} $response"
        ;;
    
    "push-fcm")
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Please provide user_id${NC}"
            echo "Usage: $0 push-fcm <user_id>"
            exit 1
        fi
        
        user_id=$2
        device_token="android_device_${user_id}_$(date +%s)"
        
        echo -e "${BLUE}Sending FCM push notification to user: $user_id${NC}"
        
        data="{
            \"user_id\": \"$user_id\",
            \"device_token\": \"$device_token\",
            \"title\": \"FCM Test\",
            \"body\": \"Hello from FCM at $(date)\",
            \"data\": {\"platform\": \"android\", \"timestamp\": $(date +%s)}
        }"
        
        response=$(auth_request "POST" "/push/fcm" "$data")
        echo -e "${GREEN}Response:${NC} $response"
        ;;
    
    "broadcast")
        echo -e "${BLUE}Sending broadcast notification...${NC}"
        
        data="{
            \"title\": \"Broadcast Message\",
            \"body\": \"This is a broadcast to all users at $(date)\",
            \"data\": {\"type\": \"broadcast\", \"timestamp\": $(date +%s)}
        }"
        
        response=$(auth_request "POST" "/push" "$data")
        echo -e "${GREEN}Response:${NC} $response"
        ;;
    
    "sql")
        echo -e "${BLUE}SQL to create test tenant and API key:${NC}"
        echo ""
        echo "-- Create test tenant"
        echo "INSERT IGNORE INTO tenants (tenant_id, name, description, active, created_at)"
        echo "VALUES ('$TENANT_ID', 'Demo Tenant', 'Demo tenant for testing', true, NOW());"
        echo ""
        echo "-- Create test API key"
        echo "INSERT IGNORE INTO api_keys (tenant_id, label, api_key, disabled, created_at)"
        echo "VALUES ('$TENANT_ID', 'Test API Key', '$API_KEY', false, NOW());"
        echo ""
        echo "-- Create test APNS config (optional)"
        echo "INSERT IGNORE INTO apns_configs (tenant_id, team_id, key_id, bundle_id, environment, active, created_at)"
        echo "VALUES ('$TENANT_ID', 'ABCD1234', 'XYZ987', 'com.example.app', 'sandbox', true, NOW());"
        echo ""
        echo "-- Create test FCM config (optional)"
        echo "INSERT IGNORE INTO fcm_configs (tenant_id, project_id, active, created_at)"
        echo "VALUES ('$TENANT_ID', 'firebase-project-id', true, NOW());"
        echo ""
        echo -e "${YELLOW}Note: Run these SQL commands in your database before testing${NC}"
        echo -e "${YELLOW}Also upload .p8 and .json files to S3: s3://signal/apns/$TENANT_ID.p8 and s3://signal/fcm/$TENANT_ID.json${NC}"
        echo ""
        echo -e "${BLUE}Example database connection (if using local .env):${NC}"
        echo "mysql -h localhost -P 3306 -u root -p signal"
        ;;
    
    "setup-tenant")
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Please provide tenant name${NC}"
            echo "Usage: $0 setup-tenant \"Tenant Name\""
            exit 1
        fi
        
        tenant_name="$2"
        custom_tenant_id="${3:-tenant-$(date +%s)}"
        custom_api_key="${4:-api-key-$(date +%s)}"
        
        echo -e "${BLUE}Setting up new tenant: $tenant_name${NC}"
        echo ""
        echo "-- Create tenant"
        echo "INSERT INTO tenants (tenant_id, name, description, active, created_at)"
        echo "VALUES ('$custom_tenant_id', '$tenant_name', 'Created via test script', true, NOW());"
        echo ""
        echo "-- Create API key for tenant"
        echo "INSERT INTO api_keys (tenant_id, label, api_key, disabled, created_at)"
        echo "VALUES ('$custom_tenant_id', 'Primary API Key', '$custom_api_key', false, NOW());"
        echo ""
        echo -e "${GREEN}Tenant ID: $custom_tenant_id${NC}"
        echo -e "${GREEN}API Key: $custom_api_key${NC}"
        echo ""
        echo -e "${YELLOW}Set these in your test environment:${NC}"
        echo "export TENANT_ID=\"$custom_tenant_id\""
        echo "export API_KEY=\"$custom_api_key\""
        ;;
    
    "env")
        echo -e "${BLUE}Current environment configuration:${NC}"
        echo ""
        if [ "$USE_LOCAL_DATABASE" = "true" ]; then
            echo -e "${GREEN}Mode: Local Database${NC}"
            echo "DB_HOST: ${DB_HOST:-localhost}"
            echo "DB_PORT: ${DB_PORT:-3306}"
            echo "DB_USERNAME: ${DB_USERNAME:-root}"
            echo "DB_DATABASE: ${DB_DATABASE:-signal_db}"
        else
            echo -e "${GREEN}Mode: AWS Secrets Manager${NC}"
            echo "AWS_REGION: ${AWS_REGION:-not set}"
            echo "DB_CREDENTIALS: ${DB_CREDENTIALS:-not set}"
            echo "DB_DATABASE: ${DB_DATABASE:-not set}"
        fi
        ;;
    
    *)
        usage
        ;;
esac
