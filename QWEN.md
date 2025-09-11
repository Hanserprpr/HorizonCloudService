# Horizon Cloud Service - Project Context for Qwen Code

## Project Overview

This is a Go-based microservices architecture for an AI-driven media cloud storage platform called "Horizon Cloud Service". The system provides intelligent file management with AI-enhanced features like semantic search, auto-tagging, and smart recommendations.

### Core Features
- **Admin Portal**: Complete file management system (similar to Google Drive/Baidu Netdisk)
- **User Gallery**: Read-only image/video gallery browsing experience
- **AI-Driven Capabilities**: Image recognition, semantic search, auto-tagging
- **Smart File Management**: Chunked uploads, deduplication, multi-resolution thumbnails

### Architecture
The system follows a microservices architecture with 8 planned services:
- Completed: User Service, File Service, System Service (3/8)
- In Progress: API Gateway (basic version)
- Planned: AI Service, Search Service, Permission Service, Share Service, Notification Service

### Technology Stack

#### Backend
- **Languages**: Go 1.23+, Python 3.11+ (for AI models)
- **Frameworks**: Gin (HTTP), GORM (ORM)
- **Databases**: PostgreSQL 15+ (main), Redis 7+ (cache/queues), MySQL 8.0 (user data)
- **Storage**: MinIO (S3-compatible object storage)
- **Vector DB**: Qdrant (for AI embeddings and search)
- **Message Queue**: Redis Streams

#### AI/ML Stack
- **Models**: CLIP, BLIP-2, YOLOv8
- **Frameworks**: PyTorch, HuggingFace Transformers
- **Inference**: ONNX Runtime

#### Frontend
- **Framework**: React 18+, TypeScript
- **UI Library**: Ant Design
- **State Management**: Zustand
- **Data Fetching**: TanStack Query (React Query)
- **Build Tool**: Vite 4+

#### Infrastructure
- **Containerization**: Docker, Kubernetes
- **Monitoring**: Prometheus + Grafana
- **Deployment**: Docker Compose (dev), Kubernetes (prod)

## Project Structure

```
HorizonCloudService/
├── services/                    # Microservices
│   ├── gateway/                # API Gateway
│   ├── user-service/           # User Management
│   ├── file-service/           # File Management
│   ├── system-service/         # System Statistics
│   ├── ai-service/             # AI Processing
│   ├── search-service/         # Search Service
│   └── ...                     # Other services
├── shared/                     # Shared Components
│   ├── pkg/                    # Shared Go packages
│   ├── database/               # Database migrations
│   └── monitoring/             # Monitoring configuration
├── frontend/                   # Frontend Applications
│   ├── admin-portal/           # Admin Dashboard (React)
│   └── user-gallery/           # User Gallery (React)
├── infrastructure/             # Infrastructure
│   ├── kubernetes/             # K8s deployment
│   ├── monitoring/             # Monitoring configuration
│   └── nginx/                  # Reverse proxy
├── docker-compose.yml          # Development environment
├── Makefile                    # Development commands
└── go.work                     # Go workspace
```

## Building and Running

### Development Environment Setup
```bash
# Check environment requirements
go version     # Go 1.23+
node --version # Node.js 18+
docker --version # Docker 20.10+

# Install dependencies for each service
cd services/user-service && go mod tidy
cd ../file-service && go mod tidy
cd ../system-service && go mod tidy
cd ../gateway && go mod tidy

# Install frontend dependencies
cd frontend/admin-portal && npm install
```

### Starting Services

#### Manual Start (Recommended for Development)
```bash
# Terminal 1: Start user service
cd services/user-service && go run cmd/main.go

# Terminal 2: Start file service
cd services/file-service && go run test_main_simple.go

# Terminal 3: Start system service
cd services/system-service && go run cmd/main.go

# Terminal 4: Start admin frontend
cd frontend/admin-portal && npm run dev

# Terminal 5: (Optional) Start API gateway
cd services/gateway && go run cmd/main.go
```

#### Docker Compose Start
```bash
# Start all services with Docker Compose
make dev-up

# View logs
make dev-logs

# Stop services
make dev-down
```

### Service Endpoints
- **User Service**: http://localhost:8001 (API), http://localhost:8001/health (health check)
- **File Service**: http://localhost:8083 (API), http://localhost:8083/health (health check)
- **System Service**: http://localhost:8003 (API), http://localhost:8003/health (health check)
- **API Gateway**: http://localhost:8080 (API), http://localhost:8080/health (health check)
- **Admin Portal**: http://localhost:5173 (default login: admin/password123)

### Testing and Quality Assurance
```bash
# Run tests for all services
make test

# Run tests for a specific service
make test-service SERVICE=user-service

# Code formatting
make format

# Code linting
make lint
```

### Database Management
```bash
# Run database migrations
make db-migrate

# Seed test data
make db-seed

# Backup databases
make db-backup
```

## Development Conventions

### Backend Development
- **Project Structure**: Standard Go project layout with cmd, internal, pkg directories
- **API Design**: RESTful HTTP/JSON with URL path versioning (/api/v1/)
- **Error Handling**: Unified error response format
- **Authentication**: JWT-based with HS256 signing
- **Configuration**: Environment-based with .env files and Viper
- **Testing**: Three-layer approach (unit, integration, E2E)

### Frontend Development
- **Component Organization**: Feature-based structure
- **State Management**: Zustand for global state
- **Data Fetching**: TanStack Query for server state
- **Styling**: Ant Design components with CSS Modules/styled-components
- **Routing**: React Router v7

### API Standards
- **Response Format**: Consistent JSON structure with code, message, data
- **Status Codes**: Standard HTTP status codes
- **Pagination**: Standard pagination with page, page_size, total, pages
- **Versioning**: URL path versioning (/api/v1/)

### Security Practices
- **Secrets Management**: Environment variables
- **Password Security**: bcrypt encryption (cost=12)
- **Transport Security**: TLS in production
- **Input Validation**: Request binding and validation
- **CORS**: Configured CORS middleware