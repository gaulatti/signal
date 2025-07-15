# Signal - Multi-Tenant Push Notification Service

A simple, multi-tenant push notification service built with Go, GORM, and MySQL.

## Features

- **Multi-tenant architecture** with API key-based authentication
- **Device t#### 2. Register Device T#### 3. Send Generic Push#### 4. Send APNS Push Notifi#### 5. Send FCM Push Notification

```bash
curl -X POST http://localhost:8080/push/fcm \
  -H "Authorization: Digest 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user456",
    "device_token": "android_device_token_here",
    "title": "Android Notification",
    "body": "This is an FCM notification",
    "data": {"platform": "android"}
  }'
# For custom port, replace 8080 with your port (e.g., 3000)
```sh
curl -X POST http://localhost:8080/push/apns \
  -H "Authorization: Digest 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user456",
    "device_token": "ios_device_token_here",
    "title": "iOS Notification",
    "body": "This is an APNS notification",
    "data": {"platform": "ios"}
  }'
# For custom port, replace 8080 with your port (e.g., 3000)
```

```bash
curl -X POST http://localhost:8080/push \
  -H "Authorization: Digest 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user456",
    "title": "Hello World",
    "body": "This is a test notification",
    "data": {"key": "value"}
  }'
# For custom port, replace 8080 with your port (e.g., 3000)
```h
curl -X POST http://localhost:8080/register \
  -H "Authorization: Digest 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p" \
  -H "Content-Type: application/json" \
  -d '{
    "device_token": "device123",
    "user_id": "user456",
    "platform": "ios"
  }'
# For custom port, replace 8080 with your port (e.g., 3000)
```ration** scoped to tenants
- **Real APNS push notifications** using sideshow/apns2 with S3-stored .p8 keys
- **Real FCM push notifications** using Firebase SDK with S3-stored service account JSON
- **In-memory caching** of API keys and push service clients for performance
- **Automatic database migrations** on startup
- **Daily-rotating digest authentication** using MD5 hash for security
- **CLI tool** for creating tenants and API keys
- **JSON seeding** from config files at startup
- **S3 integration** for dynamic credential fetching
- **Modular architecture** with clean separation of concerns

## Project Structure

```
signal/
├── cli/
│   ├── main.go                  # CLI tool for creating API keys
│   └── server.go                # Application entry point
├── config/
│   └── tenants.json             # Tenant seeding configuration
├── src/
│   ├── models/
│   │   ├── tenant.go            # Tenant model and functions
│   │   ├── api_key.go           # API key model
│   │   ├── device.go            # Device token model
│   │   ├── apns_config.go       # APNS configuration model
│   │   └── fcm_config.go        # FCM configuration model
│   ├── handlers/
│   │   ├── register.go          # Device registration handler
│   │   ├── push.go              # Generic push notification handler
│   │   ├── push_apns.go         # APNS-specific push handler
│   │   └── push_fcm.go          # FCM-specific push handler
│   ├── middleware/
│   │   └── auth_digest.go       # Daily-rotating digest authentication
│   ├── services/
│   │   ├── apns.go              # Apple Push Notification Service
│   │   ├── fcm.go               # Firebase Cloud Messaging Service
│   │   ├── tenant_loader.go     # Tenant management service
│   │   └── seed.go              # JSON seeding service
│   ├── storage/
│   │   └── s3.go                # Amazon S3 storage service
│   ├── config/
│   │   └── config.go            # Configuration management
│   └── database/
│       └── database.go          # Database connection and cache
├── .env                         # Environment variables (local dev)
├── .env.example                 # Environment variables template
├── Dockerfile                   # Docker configuration
├── go.mod                       # Go module dependencies
└── README.md                    # This file
```

## Setup

### Port Configuration

The service uses the `PORT` environment variable to determine which port to run on. If not specified, it defaults to port 8080.

**Setting a custom port:**
```bash
# Local development
export PORT=3000
go run ./cli/server.go

# With Air (development)
PORT=3000 air

# Docker
docker run -p 3000:3000 -e PORT=3000 signal

# In .env file
PORT=3000
```

### Environment Variables

The service supports two modes of database configuration:

#### Local Development (using .env file)
```bash
# Set this to use local environment variables
export USE_LOCAL_DATABASE=true
export DB_HOST=localhost
export DB_PORT=3306
export DB_USERNAME=root
export DB_PASSWORD=your_password
export DB_DATABASE=signal_db
export PORT=8080
```

#### Production (using AWS Secrets Manager)
```bash
# Set this to use AWS Secrets Manager
export USE_LOCAL_DATABASE=false
export AWS_REGION=us-east-1
export DB_CREDENTIALS=arn:aws:secretsmanager:us-east-1:account:secret:name
export DB_DATABASE=signal_db
export PORT=8080
```

The AWS secret should contain JSON in this format:
```json
{
  "host": "your-rds-endpoint.region.rds.amazonaws.com",
  "port": "3306",
  "username": "admin",
  "password": "your-secure-password"
}
```

### S3 Setup

For APNS and FCM push notifications, you need to store credentials in S3:

```bash
# Upload APNS .p8 key files
aws s3 cp AuthKey_ABC123.p8 s3://your-bucket/apns/tenant-id.p8

# Upload FCM service account JSON files  
aws s3 cp service-account.json s3://your-bucket/fcm/tenant-id.json
```

The service will automatically download and cache these credentials when needed.

### Environment Variables (Additional)

```bash
# S3 bucket for storing push notification credentials
export S3_BUCKET=signal

# AWS credentials (for S3 and Secrets Manager)
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1
```

### Database Setup

1. Create a MySQL database:
```sql
CREATE DATABASE signal_db;
```

2. The application will automatically create the required tables on startup.

### Running the Application

```bash
# Install dependencies
go mod tidy

# Run the server
go run ./cli/server.go

# Or build and run the server
go build -o signal-server ./cli/server.go
./signal-server

# Create API keys using the CLI
go run ./cli/main.go -tenant-id=product-a -label="Production API Key"
```

Or using Docker:

```bash
# Build the image
docker build -t signal .

```bash
# Build the image
docker build -t signal .

# Run with environment variables (using default port 8080)
docker run -p 8080:8080 \
  -e DB_USER=root \
  -e DB_PASS=password \
  -e DB_HOST=host.docker.internal:3306 \
  -e DB_NAME=signal_db \
  signal

# Run with custom port
docker run -p 3000:3000 \
  -e PORT=3000 \
  -e DB_USER=root \
  -e DB_PASS=password \
  -e DB_HOST=host.docker.internal:3306 \
  -e DB_NAME=signal_db \
  signal
```
```

## API Usage

### Authentication

All protected endpoints require the `Authorization` header:

```
Authorization: Digest <md5_hash>
```

Where `md5_hash` is: `md5(api_key + current_utc_date)`

The current UTC date should be in format: `YYYY-MM-DD`

### Example API Key Setup

First, you need to insert an API key into the database:

```sql
INSERT INTO api_keys (tenant_id, label, api_key, disabled, created_at) 
VALUES ('tenant-123', 'Production API Key', 'my-secret-api-key', false, NOW());
```

### Generate Authentication Digest (Example)

For API key `my-secret-api-key` on date `2025-07-14`:

```bash
# Using openssl
echo -n "my-secret-api-key2025-07-14" | openssl md5
# Output: 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p

# Using Node.js
node -e "console.log(require('crypto').createHash('md5').update('my-secret-api-key2025-07-14').digest('hex'))"
```

### Endpoints

#### 1. Health Check

```bash
curl http://localhost:8080/health
# Or if using custom port:
# curl http://localhost:3000/health
```

#### 2. Register Device Token

```bash
curl -X POST http://localhost:8080/register \
  -H "Authorization: Digest 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p" \
  -H "Content-Type: application/json" \
  -d '{
    "device_token": "device123abc",
    "user_id": "user456",
    "platform": "ios"
  }'
```

#### 3. Send Generic Push Notification

```bash
curl -X POST http://localhost:8080/push \
  -H "Authorization: Digest 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user456",
    "title": "Hello!",
    "body": "This is a test notification",
    "data": {"custom": "payload"}
  }'
```

#### 4. Send APNS Push Notification

```bash
curl -X POST http://localhost:8080/push/apns \
  -H "Authorization: Digest 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user456",
    "device_token": "ios-device-token",
    "title": "iOS Notification",
    "body": "This is sent via APNS",
    "data": {"platform": "ios"}
  }'
```

#### 5. Send FCM Push Notification

```bash
curl -X POST http://localhost:8080/push/fcm \
  -H "Authorization: Digest 1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user456", 
    "device_token": "android-device-token",
    "title": "Android Notification",
    "body": "This is sent via FCM",
    "data": {"platform": "android"}
  }'
```

## Architecture

The application follows a clean, modular architecture:

### Core Components

- **cli/server.go** - Application entry point and HTTP server setup
- **src/models/** - GORM model definitions for all entities
- **src/database/** - Database connection, migrations, and caching
- **src/handlers/** - HTTP request handlers for API endpoints
- **src/middleware/** - Authentication and other middleware
- **src/config/** - Configuration management (env vars, AWS Secrets)
- **src/services/** - Business logic services (APNS, FCM, tenant management)
- **src/storage/** - External storage integrations (S3)

### Data Flow

1. **Authentication**: Requests pass through digest authentication middleware
2. **Tenant Resolution**: Middleware resolves tenant from API key
3. **Request Handling**: Handlers process requests with tenant context
4. **Database Operations**: Models handle data persistence with tenant scoping
5. **External Services**: Services handle push notifications and storage

### Database Schema

#### tenants table (Parent)
- `id` - Primary key (auto-increment)
- `tenant_id` - Unique tenant identifier (varchar 255)
- `name` - Human-readable tenant name
- `description` - Optional tenant description
- `active` - Boolean flag to enable/disable tenant
- `created_at` - Timestamp when created
- `updated_at` - Timestamp when last updated

#### api_keys table (Child of tenants)
- `id` - Primary key (auto-increment)
- `tenant_id` - Foreign key to tenants.tenant_id
- `label` - Optional label for the API key
- `api_key` - The actual API key string (unique)
- `disabled` - Boolean flag to disable keys
- `created_at` - Timestamp when created

#### device_tokens table (Child of tenants)
- `id` - Primary key (auto-increment)
- `tenant_id` - Foreign key to tenants.tenant_id
- `device_token` - Device push token
- `user_id` - User identifier
- `platform` - Platform (ios, android, web, etc.)
- `created_at` - When first registered
- `updated_at` - When last updated
- `UNIQUE(tenant_id, user_id, platform)` - One device per user per platform per tenant

#### apns_configs table (Configuration for Apple Push)
- `id` - Primary key (auto-increment)
- `tenant_id` - Foreign key to tenants.tenant_id
- `team_id` - Apple Developer Team ID
- `key_id` - Apple Push Key ID
- `bundle_id` - iOS App Bundle ID
- `private_key` - P8 certificate content
- `environment` - 'production' or 'sandbox'
- `active` - Boolean flag
- `created_at` / `updated_at` - Timestamps

#### fcm_configs table (Configuration for Firebase)
- `id` - Primary key (auto-increment)
- `tenant_id` - Foreign key to tenants.tenant_id
- `project_id` - Firebase Project ID
- `service_account` - JSON service account key
- `active` - Boolean flag
- `created_at` / `updated_at` - Timestamps

## Security Notes

- API keys are cached in memory for performance
- Digest authentication prevents replay attacks (date-based)
- Each tenant is isolated by `tenant_id`
- Database queries are scoped to the authenticated tenant

## Next Steps

- Replace simulated push with actual push notification services (FCM, APNs)
- Add rate limiting
- Add logging and monitoring
- Add API key management endpoints
- Add webhook support for delivery receipts
