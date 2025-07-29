# Leeta Golang Exercise

A location management service built with Go, featuring geographic location storage, nearest location finding, and RESTful API endpoints.

## 🚀 Features

- **Location Management**: Create, read, update, and delete locations
- **Geographic Search**: Find the nearest location using PostGIS
- **RESTful API**: Clean HTTP endpoints with JSON responses
- **Database Integration**: PostgreSQL with PostGIS extension for geographic data
- **Swagger Documentation**: Auto-generated API documentation
- **Docker Support**: Containerized development environment
- **Comprehensive Testing**: Unit and integration tests

## 📋 Prerequisites

- **Go 1.24+**: [Download Go](https://golang.org/dl/)
- **Docker & Docker Compose**: [Install Docker](https://docs.docker.com/get-docker/)
- **yq**: For YAML configuration parsing (only necessary in dev)
  ```bash
  brew install yq  # macOS
  ```
- **make**: Makefile for cli command management

## 🛠️ Setup

### 1. Clone the Repository
```bash
git clone <repository-url>
cd leeta-exercise
```

### 2. Configuration Setup
Copy the sample configuration file and customize it for your environment:

```bash
cp config-sample.yml config.yml
```

Edit `config.yml` with your preferred settings:

```yaml
database:
  protocol: "postgresql"
  host: "127.0.0.1"
  port: "5433"
  name: "leeta"
  user: "postgres"
  password: "postgres"
server:
  httpUrl: "0.0.0.0"
  httpPort: "8080"
  httpAllowedOrigins: "http://127.0.0.1:3000,http://127.0.0.1:8080"
app: 
  name: "leeta"
  env: "development"
```

Ensure your settings match with the `environment` variables of the posgres service in `docker-compose.yml`

```yaml
    postgres:
        .
        .
        .
        environment:
            POSTGRES_USER: "postgres" 
            POSTGRES_PASSWORD: "postgres"
            POSTGRES_DB: "leeta"
```

If you are on a mac and yq is installed, set your environment to have this format so docker compose can pick your configs from `config.yml`
```yaml
    postgres:
        .
        .
        .
        environment:
            POSTGRES_USER: "${DB_USER}"
            POSTGRES_PASSWORD: "${DB_PASSWORD}"
            POSTGRES_DB: "${DB_NAME}"
```

### 3. Start the Service
Build and start all services (database, migrations, and application):

```bash
make service-up
```

OR

```bash 
DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) DB_USER=$(DB_USER) docker-compose up -d
```

You can ignore `DB_PASSWORD`, `DB_NAME` and `DB_USER` if they are already set in `docker-compose.yml`

This command will:
- Build the Docker image
- Start PostgreSQL with PostGIS
- Run database migrations
- Start the application server

### 4. Access the Service
The service will be available at: **http://localhost:8081**

## 📚 API Documentation

### Swagger UI
Access the interactive API documentation at: **http://localhost:8081/swagger/**

### Postman
You can import the postman documentation for this API using the json file in the root directory:
`Leeta.postman_collection.json`

### API Endpoints

#### Health Check
- `GET /v1/health/` - Health check endpoint
- `POST /v1/health/` - Health check with POST method

#### Location Management

##### Create Location
```http
POST /v1/locations/
Content-Type: application/json

{
  "name": "New York",
  "latitude": 40.7128,
  "longitude": -74.0060
}
```

**Response:**
```json
{
  "success": true,
  "message": "Location created successfully",
  "data": {
    "id": "uuid",
    "name": "New York",
    "slug": "new-york",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

##### Get Location
```http
GET /v1/locations/{name}
```

**Response:**
```json
{
  "success": true,
  "message": "Success",
  "data": {
    "id": "uuid",
    "name": "New York",
    "slug": "new-york",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

##### List All Locations
```http
GET /v1/locations/
```

**Response:**
```json
{
  "success": true,
  "message": "Success",
  "data": [
    {
      "id": "uuid1",
      "name": "New York",
      "slug": "new-york",
      "latitude": 40.7128,
      "longitude": -74.0060,
      "created_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "uuid2",
      "name": "Los Angeles",
      "slug": "los-angeles",
      "latitude": 34.0522,
      "longitude": -118.2437,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

##### Delete Location
```http
DELETE /v1/locations/{name}
```

**Response:**
```json
{
  "success": true,
  "message": "Deleted location successfully"
}
```

##### Find Nearest Location
```http
GET /v1/locations/nearest?lat=40.7589&lng=-73.9851
```

**Response:**
```json
{
  "success": true,
  "message": "Success",
  "data": {
    "id": "uuid",
    "name": "New York",
    "slug": "new-york",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "created_at": "2024-01-01T00:00:00Z",
    "distance": "1.50 kilometers"
  }
}
```

## 🧪 Testing

### Run All Tests

To run all tests, ensure that your database is running on the specified port in your `config.yml`. A quick feat to 
achieve this is to make sure your service is running on one terminal, then you run a test on the other.
```bash
make test
```

### Run Tests with Race Detection
```bash
go test -v ./... -race
```

## 🛠️ Development

### Available Make Commands

| Command | Description |
|---------|-------------|
| `make service-build` | Build and start all services |
| `make service-up` | Start services in background |
| `make service-down` | Stop services |
| `make service-down-add` | Stop services and remove volumes |
| `make dev` | Start development server with hot reload |
| `make build` | Build the binary |
| `make start` | Build and start the binary |
| `make test` | Run all tests |
| `make lint` | Run linter |
| `make swag` | Generate Swagger documentation |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback database migrations |

### Database Management

#### Create Migration
```bash
make create-migration NAME=migration_name
```

#### Run Migrations
```bash
make migrate-up
```

#### Rollback Migrations
```bash
make migrate-down
```

## 🏗️ Project Structure

```
leeta-exercise/
├── cmd/http/                    # Application entry point
├── internal/
│   ├── adapter/                 # External adapters
│   │   ├── config/             # Configuration management
│   │   ├── handler/http/       # HTTP handlers
│   │   ├── logger/             # Logging
│   │   └── storage/postgres/   # Database layer
│   ├── core/                   # Business logic
│   │   ├── domain/             # Domain models
│   │   ├── port/               # Interfaces
│   │   └── service/            # Business services
│   └── util/                   # Utilities
├── docs/                       # Swagger documentation
├── migrations/                 # Database migrations
├── config-sample.yml           # Sample configuration
├── docker-compose.yml          # Docker services
├── Dockerfile                  # Application container
├── Makefile                    # Build and development commands
└── README.md                   # This file
```

## 🔧 Configuration

### Environment Variables
The application uses a YAML configuration file (`config.yml`) with the following sections:

- **database**: PostgreSQL connection settings
- **server**: HTTP server configuration
- **app**: Application metadata

### Database Configuration
- **protocol**: Database protocol (postgresql)
- **host**: Database host
- **port**: Database port
- **name**: Database name
- **user**: Database username
- **password**: Database password

### Server Configuration
- **httpUrl**: Server bind address
- **httpPort**: Server port
- **httpAllowedOrigins**: CORS allowed origins

## 🐳 Docker

### Services
- **postgres**: PostgreSQL with PostGIS extension
- **app**: Go application server

### Environment Variables
Docker Compose uses the following environment variables:
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name

## 📝 API Examples

### Using curl

#### Create a location
```bash
curl -X POST http://localhost:8081/v1/locations/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Times Square",
    "latitude": 40.7589,
    "longitude": -73.9851
  }'
```

#### Get a location
```bash
curl http://localhost:8081/v1/locations/times-square
```

#### Find nearest location
```bash
curl "http://localhost:8081/v1/locations/nearest?lat=40.7589&lng=-73.9851"
```

#### List all locations
```bash
curl http://localhost:8081/v1/locations/
```

#### Delete a location
```bash
curl -X DELETE http://localhost:8081/v1/locations/times-square
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🆘 Support

For support and questions, please open an issue in the repository.

