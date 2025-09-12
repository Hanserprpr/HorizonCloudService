# Horizon Cloud Service - Key Components and Responsibilities

## 1. Client Layer

### Admin Portal (React)
- **Responsibility**: Full-featured file management system for administrators
- **Status**: 90% complete
- **Technology Stack**: React 18+, TypeScript, Ant Design, React Router, TanStack Query
- **Features**:
  - User management dashboard
  - File browsing and management
  - System settings and configuration
  - Statistics and reporting
  - Responsive design for multiple devices

### User Gallery (React)
- **Responsibility**: Read-only image and video gallery browsing system with intelligent display and search
- **Status**: Not started
- **Technology Stack**: React 18+, TypeScript
- **Features**:
  - Media gallery browsing
  - Smart search and filtering
  - Responsive image display
  - User profile management

### Mobile App
- **Responsibility**: Mobile application for file access and management
- **Status**: Planned
- **Technology Stack**: React Native or Flutter (to be determined)

## 2. API Gateway Layer

### Gateway Service (Go + Gin)
- **Responsibility**: Central API gateway handling authentication, routing, rate limiting, and monitoring
- **Port**: 8080
- **Status**: Basic version in development
- **Technology Stack**: Go 1.23+, Gin, Redis
- **Features**:
  - JWT authentication middleware
  - Request routing to microservices
  - Rate limiting with Redis
  - Health monitoring and metrics collection
  - CORS configuration
  - Load balancing (planned)

## 3. Service Layer

### User Service (Go + Gin)
- **Responsibility**: User authentication, management, and quota control
- **Port**: 8001
- **Status**: 100% complete
- **Technology Stack**: Go 1.23+, Gin, GORM, bcrypt, JWT, MySQL
- **Features**:
  - JWT authentication system (HS256, 24-hour tokens)
  - User CRUD operations
  - Storage quota management
  - Asynchronous activity logging
  - bcrypt password encryption
  - Admin APIs for user management

### File Service (Go + Gin)
- **Responsibility**: File management, storage, and processing
- **Port**: 8083
- **Status**: 100% complete
- **Technology Stack**: Go 1.23+, Gin, GORM, MinIO, PostgreSQL
- **Features**:
  - Chunked upload system (5MB chunks, resumable)
  - Multi-storage backend support (Local, MinIO)
  - File deduplication with SHA-256
  - Thumbnail generation (multiple sizes)
  - Folder management with hierarchical structure
  - Integration with user service for authentication

### System Service (Go + Gin)
- **Responsibility**: System statistics, health monitoring, and administration
- **Port**: 8003
- **Status**: 100% complete
- **Technology Stack**: Go 1.23+, Gin, SQLite
- **Features**:
  - System statistics collection
  - Health status monitoring
  - System settings management
  - Cache management

### AI Service (Go + Python)
- **Responsibility**: AI-powered image analysis, semantic search, and intelligent tagging
- **Port**: 8084
- **Status**: Not started
- **Technology Stack**: Go 1.23+, Python, CLIP/BLIP models, Qdrant
- **Features**:
  - Image content recognition with CLIP
  - Image quality assessment
  - Automatic tag generation with BLIP-2
  - Vector embedding management
  - Batch processing queue with Redis Streams

### Search Service (Go)
- **Responsibility**: Semantic search and intelligent recommendations
- **Port**: 8086
- **Status**: Not started
- **Technology Stack**: Go 1.23+, Qdrant, Elasticsearch
- **Features**:
  - Vector similarity search
  - Natural language search
  - Image-based search
  - Intelligent recommendation system

### Permission Service (Go)
- **Responsibility**: Fine-grained permission control and access management
- **Port**: 8085
- **Status**: Not started
- **Technology Stack**: Go 1.23+, Casbin RBAC engine, PostgreSQL
- **Features**:
  - Role-based access control
  - Audit logging
  - Permission management

### Share Service (Go)
- **Responsibility**: File sharing and collaboration features
- **Port**: 8087
- **Status**: Not started
- **Technology Stack**: Go 1.23+
- **Features**:
  - Share link generation
  - Access control
  - Collaboration tools

### Notification Service (Go)
- **Responsibility**: Multi-channel notifications and event-driven messaging
- **Port**: 8088
- **Status**: Not started
- **Technology Stack**: Go 1.23+
- **Features**:
  - Email/SMS/push notifications
  - Event-driven architecture
  - Template engine

## 4. Data Layer

### MySQL Database
- **Responsibility**: User data storage
- **Status**: Configured and ready
- **Technology Stack**: MySQL 8.0
- **Usage**: User service

### PostgreSQL Database
- **Responsibility**: File metadata storage with pgvector support
- **Status**: Configured and ready
- **Technology Stack**: PostgreSQL with pgvector
- **Usage**: File service, AI service, Search service

### Redis
- **Responsibility**: Caching, sessions, and message queues
- **Status**: Configured and ready
- **Technology Stack**: Redis 7.2
- **Usage**: All services for caching and queuing

### MinIO
- **Responsibility**: Object storage for files
- **Status**: Configured and ready
- **Technology Stack**: MinIO S3-compatible storage
- **Usage**: File service

### Qdrant
- **Responsibility**: Vector database for AI embeddings
- **Status**: Configured and ready
- **Technology Stack**: Qdrant vector database
- **Usage**: AI service, Search service

## 5. Monitoring & Infrastructure

### Prometheus
- **Responsibility**: Metrics collection and monitoring
- **Status**: Configured and ready
- **Technology Stack**: Prometheus
- **Usage**: All services

### Grafana
- **Responsibility**: Metrics visualization and dashboard
- **Status**: Configured and ready
- **Technology Stack**: Grafana
- **Usage**: System monitoring

### Nginx
- **Responsibility**: Reverse proxy and static file serving
- **Status**: Configured and ready
- **Technology Stack**: Nginx
- **Usage**: Frontend serving