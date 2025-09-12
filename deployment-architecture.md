# Horizon Cloud Service - Deployment Architecture

## Overview
The Horizon Cloud Service follows a modern microservices architecture with containerized deployment using Docker and Docker Compose for development environments. The system is designed to be deployed in a Kubernetes cluster for production environments with comprehensive monitoring and observability.

## Development Deployment (Docker Compose)

### Container Orchestration
- **Platform**: Docker Compose
- **Network**: Custom bridge network (horizon-cloud-net)
- **Volumes**: Named volumes for data persistence
- **Environment**: Single-node deployment for development

### Service Containers

#### Data Layer Containers
1. **MySQL**:
   - Image: mysql:8.0
   - Port: 3306
   - Volume: mysql_data
   - Purpose: User data storage

2. **PostgreSQL**:
   - Image: pgvector/pgvector:pg15
   - Port: 5432
   - Volume: postgres_data
   - Purpose: File metadata and AI embeddings

3. **Redis**:
   - Image: redis:7.2-alpine
   - Port: 6379
   - Volume: redis_data
   - Purpose: Caching, sessions, message queues

4. **MinIO**:
   - Image: minio/minio:latest
   - Ports: 9000 (API), 901 (Console)
   - Volume: minio_data
   - Purpose: Object storage for files

5. **Qdrant**:
   - Image: qdrant/qdrant:latest
   - Ports: 6333, 6334
   - Volume: qdrant_data
   - Purpose: Vector database for AI embeddings

#### Service Layer Containers
1. **User Service**:
   - Build: ./services/user-service
   - Port: 8081
   - Dependencies: MySQL, Redis
   - Purpose: User authentication and management

2. **File Service**:
   - Build: ./services/file-service
   - Port: 8083
   - Dependencies: PostgreSQL, MinIO, Redis
   - Purpose: File management and storage

3. **System Service**:
   - Build: ./services/system-service
   - Port: 803
   - Dependencies: None (uses SQLite)
   - Purpose: System statistics and settings

4. **AI Service**:
   - Build: ./services/ai-service
   - Port: 8084
   - Dependencies: PostgreSQL, Qdrant, Redis
   - Purpose: AI processing and analysis

5. **Search Service**:
   - Build: ./services/search-service
   - Port: 8086
   - Dependencies: Qdrant, PostgreSQL, Redis
   - Purpose: Semantic search capabilities

6. **Gateway Service**:
   - Build: ./services/gateway
   - Port: 8080
   - Dependencies: Redis
   - Purpose: API gateway and routing

#### Monitoring Containers
1. **Prometheus**:
   - Image: prom/prometheus:latest
   - Port: 9090
   - Volume: ./infrastructure/monitoring/prometheus
   - Purpose: Metrics collection

2. **Grafana**:
   - Image: grafana/grafana:latest
   - Port: 3000
   - Volume: ./infrastructure/monitoring/grafana
   - Purpose: Metrics visualization

#### Frontend Containers
1. **Nginx**:
   - Image: nginx:alpine
   - Ports: 80, 443
   - Volumes: 
     - ./infrastructure/nginx.conf
     - ./infrastructure/nginx/conf.d
     - ./frontend/admin-portal/dist
     - ./frontend/user-gallery/dist
   - Dependencies: Gateway
   - Purpose: Reverse proxy and static file serving

### Environment Configuration
- **Environment Files**: .env files in each service directory
- **Configuration Management**: Viper for Go services
- **Secrets Management**: Environment variables
- **Service Discovery**: Docker network DNS resolution

## Production Deployment (Kubernetes)

### Cluster Architecture
- **Orchestrator**: Kubernetes
- **Namespace**: horizon-cloud
- **Ingress**: NGINX Ingress Controller
- **Service Mesh**: (Planned) Istio for advanced traffic management
- **Secrets Management**: Kubernetes Secrets
- **Config Management**: Kubernetes ConfigMaps

### Deployment Components

#### Core Services
1. **User Service Deployment**:
   - Replicas: 3 (scalable)
   - Database: External MySQL cluster
   - Redis: External Redis cluster
   - Health Checks: Liveness and readiness probes
   - Resource Limits: CPU/Memory constraints

2. **File Service Deployment**:
   - Replicas: 3 (scalable)
   - Storage: S3-compatible object storage
   - Database: External PostgreSQL cluster
   - Redis: External Redis cluster
   - Health Checks: Liveness and readiness probes

3. **System Service Deployment**:
   - Replicas: 2
   - Database: External PostgreSQL or dedicated instance
   - Health Checks: Liveness and readiness probes

4. **AI Service Deployment**:
   - Replicas: 2
   - GPU Support: NVIDIA GPU nodes (for ML inference)
   - Model Storage: Persistent volumes for ML models
   - Health Checks: Liveness and readiness probes

5. **Gateway Service Deployment**:
   - Replicas: 3
   - Redis: External Redis cluster
   - Health Checks: Liveness and readiness probes
   - SSL Termination: TLS certificates

#### Data Layer (External Services)
1. **Database Cluster**:
   - MySQL: Managed service or self-hosted cluster
   - PostgreSQL: Managed service or self-hosted cluster with pgvector
   - Backup Strategy: Automated backups with point-in-time recovery

2. **Cache Cluster**:
   - Redis: Managed service or Redis cluster
   - Persistence: AOF and RDB snapshots

3. **Object Storage**:
   - S3: AWS S3 or compatible service
   - Buckets: Separate buckets for different data types

4. **Vector Database**:
   - Qdrant: Managed service or self-hosted cluster
   - Replication: Clustered deployment for high availability

#### Monitoring Stack
1. **Prometheus**:
   - Operator: Prometheus Operator
   - Storage: Persistent volumes
   - Scraping: Service discovery for metrics endpoints

2. **Grafana**:
   - Dashboards: Pre-configured dashboards
   - Authentication: OAuth or internal user management
   - Storage: Persistent volumes for dashboard data

3. **Logging**:
   - Stack: ELK (Elasticsearch, Logstash, Kibana) or EFK (Elasticsearch, Fluentd, Kibana)
   - Collection: Fluentd or Filebeat agents
   - Retention: Configurable log retention policies

4. **Tracing**:
   - Solution: Jaeger or OpenTelemetry
   - Integration: Application-level tracing instrumentation

#### Ingress and Load Balancing
1. **Ingress Controller**:
   - Type: NGINX Ingress Controller
   - SSL: TLS certificates with Let's Encrypt
   - Routing: Path-based and host-based routing rules

2. **Load Balancer**:
   - External: Cloud load balancer (AWS ELB, GCP Load Balancer)
   - Internal: Kubernetes Service with ClusterIP or NodePort

#### Security
1. **Network Policies**:
   - Pod-to-Pod communication restrictions
   - Ingress/Egress rules for services

2. **RBAC**:
   - Kubernetes RBAC for cluster access
   - Service accounts with minimal permissions

3. **Image Security**:
   - Image scanning for vulnerabilities
   - Private registry with image signing

## CI/CD Pipeline

### Development Workflow
1. **Source Control**: Git with GitHub/GitLab
2. **Branching Strategy**: GitFlow or GitHub Flow
3. **Code Review**: Pull requests with mandatory reviews
4. **Testing**: Automated unit, integration, and E2E tests

### Build Process
1. **Container Images**:
   - Multi-stage Docker builds
   - Base image security scanning
   - Tagging strategy: semantic versioning

2. **Artifact Storage**:
   - Container registry (Docker Hub, AWS ECR, GCP GCR)
   - Helm charts repository

### Deployment Process
1. **Environments**:
   - Development: Feature branches
   - Staging: Develop branch
   - Production: Main/Master branch

2. **Deployment Strategy**:
   - Blue/Green deployments
   - Rolling updates
   - Canary deployments (for critical services)

3. **Automation Tools**:
   - GitHub Actions/GitLab CI
   - ArgoCD for GitOps deployments
   - Helm for Kubernetes package management

## Backup and Disaster Recovery

### Data Backup Strategy
1. **Database Backups**:
   - Daily full backups
   - Hourly incremental backups
   - Point-in-time recovery capability

2. **File Storage Backups**:
   - Cross-region replication
   - Versioning for object storage
   - Regular consistency checks

3. **Configuration Backups**:
   - Git repository for infrastructure as code
   - Regular snapshots of Kubernetes resources

### Disaster Recovery Plan
1. **Recovery Time Objective (RTO)**: < 4 hours
2. **Recovery Point Objective (RPO)**: < 1 hour
3. **Failover Process**: Automated failover to secondary region
4. **Testing**: Quarterly disaster recovery drills

## Scaling Strategy

### Horizontal Scaling
1. **Auto Scaling**:
   - Kubernetes Horizontal Pod Autoscaler (HPA)
   - Custom metrics-based scaling
   - Node auto-scaling based on resource demand

2. **Load Distribution**:
   - Round-robin load balancing
   - Session affinity where required
   - Geographic load balancing (multi-region)

### Vertical Scaling
1. **Resource Optimization**:
   - Container resource requests and limits
   - Database connection pooling
   - Caching strategies to reduce database load

## High Availability

### Service Redundancy
1. **Multi-AZ Deployment**: Services deployed across multiple availability zones
2. **Database Replication**: Master-slave replication with automatic failover
3. **Load Balancer Redundancy**: Multiple load balancer instances
4. **Node Redundancy**: Kubernetes worker nodes across multiple zones

### Monitoring and Alerting
1. **Health Checks**: Liveness and readiness probes for all services
2. **Alerting**: Slack, email, and SMS notifications for critical issues
3. **Dashboard**: Real-time status dashboard for system health
4. **Log Analysis**: Automated log analysis for anomaly detection

This deployment architecture provides a robust, scalable, and maintainable foundation for the Horizon Cloud Service, supporting both development and production environments with comprehensive monitoring and disaster recovery capabilities.