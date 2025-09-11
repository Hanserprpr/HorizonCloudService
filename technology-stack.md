# Horizon Cloud Service - Technology Stack and Frameworks

## Backend Technology Stack

### Core Languages and Frameworks
- **Primary Language**: Go (1.23+)
- **Web Framework**: Gin
- **ORM**: GORM
- **Database Migrations**: GORM AutoMigrate
- **Authentication**: JWT (HS256)
- **Password Security**: bcrypt
- **Configuration Management**: Viper
- **Logging**: Zap (implied by code structure)
- **Testing**: testify, mock

### Database Technologies
- **Relational Databases**:
  - MySQL 8.0 (User data)
  - PostgreSQL with pgvector (File metadata and AI embeddings)
  - SQLite (System settings for development)
- **NoSQL Databases**:
  - Redis 7.2 (Caching, sessions, message queues)
  - Qdrant (Vector database for AI embeddings and search)

### Storage Technologies
- **Object Storage**: MinIO (S3-compatible)
- **Local Storage**: File system (development and fallback)

### AI and Machine Learning
- **Models**: CLIP, BLIP-2
- **Python Runtime**: For ML model inference
- **Vector Search**: Qdrant

### Search Technologies
- **Vector Search**: Qdrant
- **Text Search**: Elasticsearch (planned)

### Monitoring and Observability
- **Metrics Collection**: Prometheus
- **Visualization**: Grafana
- **Health Checks**: Custom Gin endpoints
- **Logging**: Standard Go logging with potential for structured logging

### API and Communication
- **API Protocol**: RESTful HTTP/JSON
- **API Documentation**: Swagger/OpenAPI (planned)
- **Service Communication**: HTTP/REST (synchronous)
- **Message Queue**: Redis Streams
- **Reverse Proxy**: httputil.ReverseProxy (Gateway)

### Security
- **Authentication**: JWT tokens
- **Authorization**: Custom middleware (planned: Casbin RBAC)
- **Password Encryption**: bcrypt
- **Transport Security**: TLS (implied by production deployment)

## Frontend Technology Stack

### Core Frameworks
- **Primary Framework**: React 18+
- **Language**: TypeScript
- **State Management**: Zustand
- **Data Fetching**: TanStack Query (React Query)
- **Routing**: React Router v7
- **UI Components**: Ant Design v5
- **Icons**: Ant Design Icons
- **Date/Time**: Day.js

### Development Tools
- **Build Tool**: Vite
- **Package Manager**: npm
- **Linting**: ESLint with TypeScript ESLint
- **Testing**:
  - Unit Testing: Vitest
  - UI Testing: @testing-library/react
  - E2E Testing: Playwright
- **Type Definitions**: DefinitelyTyped (@types packages)

### Styling
- **CSS Framework**: Ant Design built-in styling
- **CSS Reset**: Ant Design reset.css
- **Custom Styling**: CSS Modules or styled-components (implied)

## Infrastructure and Deployment

### Containerization
- **Container Runtime**: Docker
- **Container Orchestration**: Kubernetes
- **Container Build**: Multi-stage Dockerfiles
- **Base Images**: Alpine Linux for minimal footprint

### Development Environment
- **Configuration Management**: Docker Compose
- **Environment Variables**: .env files with Viper
- **Service Discovery**: Docker network DNS

### Reverse Proxy and Load Balancing
- **Web Server**: Nginx
- **SSL Termination**: Nginx
- **Static File Serving**: Nginx

### Monitoring Stack
- **Metrics Collection**: Prometheus
- **Metrics Visualization**: Grafana
- **Log Aggregation**: (Not explicitly configured, but implied)

## Development and Build Tools

### Backend Development
- **Dependency Management**: Go modules
- **Code Generation**: Go generate (implied)
- **Linting**: golint, go vet
- **Formatting**: go fmt

### Frontend Development
- **Development Server**: Vite dev server
- **Hot Module Replacement**: Vite
- **Type Checking**: TypeScript compiler
- **Bundle Analysis**: (Not explicitly configured)

### CI/CD and DevOps
- **Container Registry**: (Not specified)
- **Deployment**: Docker Compose, Kubernetes
- **Infrastructure as Code**: Docker Compose files
- **Environment Management**: .env files

## Third-Party Libraries and Services

### Backend Libraries
- **Gin-Gonic/Gin**: HTTP web framework
- **GORM**: ORM library
- **dgrijalva/jwt-go**: JWT implementation
- **joho/godotenv**: Environment variable loading
- **spf13/viper**: Configuration solution
- **stretchr/testify**: Testing toolkit
- **go-redis/redis**: Redis client
- **minio/minio-go**: MinIO client

### Frontend Libraries
- **Ant Design**: UI component library
- **TanStack Query**: Server state management
- **React Router**: Routing library
- **Zustand**: State management
- **Axios**: HTTP client
- **Day.js**: Date/time manipulation

## Standards and Conventions

### API Design
- **Versioning**: URL path versioning (/api/v1/)
- **Status Codes**: Standard HTTP status codes
- **Response Format**: Consistent JSON structure with code, message, data
- **Error Handling**: Unified error response format
- **Pagination**: Standard pagination with page, page_size, total, pages

### Code Organization
- **Backend**: Standard Go project structure with cmd, internal, pkg directories
- **Frontend**: Feature-based component organization
- **Configuration**: Environment-based configuration with .env files
- **Testing**: Three-layer testing approach (unit, integration, E2E)

### Security Practices
- **Secrets Management**: Environment variables
- **Input Validation**: Request binding and validation
- **CORS**: Configured CORS middleware
- **Rate Limiting**: Redis-based rate limiting (planned)
- **SQL Injection Prevention**: ORM usage