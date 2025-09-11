# Horizon Cloud Service - System Architecture

## Overview
AI智能媒体云存储系统是一个企业级文件管理、智能分析和语义搜索平台，采用微服务架构和事件驱动设计，提供类似百度网盘、Google Drive的功能，并集成AI增强能力。

## System Architecture Diagram

```mermaid
graph TD
    subgraph "Client Layer"
        A["Admin Portal (React)"] -->|"API Calls"| G
        B["User Gallery (React)"] -->|"API Calls"| G
        C["Mobile App (Planned)"] -->|"API Calls"| G
        D["API Documentation"] -->|"Swagger"| G
    end

    subgraph "API Gateway Layer"
        G["🛡️ Gateway (8080)"]
        G -->|"JWT Auth"| G1
        G -->|"Rate Limiting"| G1
        G -->|"Routing"| G1
        G -->|"Monitoring"| G1
        G1["Load Balancing & Proxy"]
    end

    subgraph "Service Layer"
        G1 -->|"User Management"| US
        G1 -->|"File Management"| FS
        G1 -->|"System Management"| SS
        G1 -->|"AI Processing"| AIS
        G1 -->|"Search"| SES
        G1 -->|"Permissions"| PS
        G1 -->|"Sharing"| SHS
        G1 -->|"Notifications"| NS
        
        US["User Service (8001)"]
        FS["File Service (8083)"]
        SS["System Service (8003)"]
        AIS["AI Service (8084)"]
        SES["Search Service (8086)"]
        PS["Permission Service (8085)"]
        SHS["Share Service (8087)"]
        NS["Notification Service (8088)"]
    end

    subgraph "Data Layer"
        US -->|"User Data"| DB1[(MySQL)]
        FS -->|"File Metadata"| DB2[(PostgreSQL)]
        SS -->|"System Settings"| DB3[(SQLite)]
        AIS -->|"AI Models"| DB4[(PostgreSQL)]
        SES -->|"Search Index"| DB5[(Qdrant)]
        
        FS -->|"File Storage"| S1[(MinIO/S3)]
        FS -->|"Cache"| R1[(Redis)]
        AIS -->|"Vector DB"| DB5
        AIS -->|"Task Queue"| R1
        NS -->|"Notifications"| R1
        
        R1 -->|"Sessions"| R1
        R1 -->|"Caching"| R1
    end

    subgraph "Monitoring & Infrastructure"
        M1["Prometheus"]
        M2["Grafana"]
        N1["Nginx"]
        M1 -->|"Metrics"| M2
        G -->|"Metrics"| M1
        US -->|"Metrics"| M1
        FS -->|"Metrics"| M1
        SS -->|"Metrics"| M1
        N1 -->|"Static Files"| A
        N1 -->|"Static Files"| B
    end

    style A fill:#4CAF50,stroke:#388E3C
    style B fill:#2196F3,stroke:#0D47A1
    style G fill:#FF9800,stroke:#E65100
    style US fill:#9C27B0,stroke:#4A148C
    style FS fill:#9C27B0,stroke:#4A148C
    style SS fill:#9C27B0,stroke:#4A148C
    style AIS fill:#F44336,stroke:#B71C1C
    style SES fill:#F44336,stroke:#B71C1C
    style PS fill:#F44336,stroke:#B71C1C
    style SHS fill:#F44336,stroke:#B71C1C
    style NS fill:#F44336,stroke:#B71C1C
    style DB1 fill:#009688,stroke:#04D40
    style DB2 fill:#009688,stroke:#004D40
    style DB3 fill:#009688,stroke:#004D40
    style DB4 fill:#009688,stroke:#004D40
    style DB5 fill:#009688,stroke:#004D40
    style S1 fill:#FFC107,stroke:#FF6F00
    style R1 fill:#E91E63,stroke:#880E4F
    style M1 fill:#3F51B5,stroke:#1A237E
    style M2 fill:#3F51B5,stroke:#1A237E
    style N1 fill:#607D8B,stroke:#263238