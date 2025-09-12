# 用户画廊API设计方案

基于用户画廊的特殊需求，需要在现有文件服务基础上扩展专门的画廊API。

## 📋 画廊API需求分析

### 核心功能需求
1. **每日推荐** - 智能推荐高质量图片/视频
2. **全局搜索** - 跨文件夹的媒体内容搜索  
3. **丰富元数据** - 格式化的媒体信息展示
4. **智能筛选** - 按地点、设备、标签等维度过滤
5. **相关推荐** - 基于当前查看内容的相关推荐

## 🚀 新增画廊API设计

### 1. 每日推荐API
```http
GET /api/v1/gallery/recommendations/daily
GET /api/v1/gallery/recommendations/trending
GET /api/v1/gallery/recommendations/similar/{file_id}
```

**响应格式:**
```json
{
  "success": true,
  "data": {
    "title": "今日推荐",
    "description": "基于质量评分和多样性的智能推荐",
    "items": [
      {
        "id": 123,
        "name": "sunset_beach.jpg",
        "thumbnail_url": "/api/v1/files/123/thumbnail?size=medium",
        "original_url": "/api/v1/files/123/download",
        "metadata": {
          "width": 4032,
          "height": 3024,
          "camera_model": "iPhone 13 Pro",
          "location": "Santorini, Greece",
          "taken_at": "2024-08-15T18:30:00Z"
        },
        "quality_score": 0.92,
        "view_count": 15,
        "is_favorite": false
      }
    ],
    "total": 20,
    "algorithm": "quality_diversity_v1"
  }
}
```

### 2. 画廊专用搜索API
```http
POST /api/v1/gallery/search/global
POST /api/v1/gallery/search/visual
POST /api/v1/gallery/search/advanced
```

**全局搜索请求:**
```json
{
  "query": "sunset beach photo",
  "filters": {
    "media_types": ["image", "video"],
    "date_range": {
      "from": "2024-01-01",
      "to": "2024-12-31"
    },
    "location": "Santorini",
    "camera_model": "iPhone",
    "quality_min": 0.7
  },
  "sort": "relevance",
  "limit": 50
}
```

**视觉搜索请求:**
```json
{
  "reference_file_id": 123,
  "similarity_threshold": 0.8,
  "search_scope": "all_images",
  "limit": 20
}
```

### 3. 画廊文件详情API
```http
GET /api/v1/gallery/files/{id}/details
GET /api/v1/gallery/files/{id}/related
```

**详情响应:**
```json
{
  "success": true,
  "data": {
    "file": {
      "id": 123,
      "name": "sunset_beach.jpg",
      "original_url": "/api/v1/files/123/download",
      "thumbnails": [
        {"size": "small", "url": "/api/v1/files/123/thumbnail?size=small"},
        {"size": "medium", "url": "/api/v1/files/123/thumbnail?size=medium"},
        {"size": "large", "url": "/api/v1/files/123/thumbnail?size=large"}
      ]
    },
    "media_info": {
      "type": "image",
      "format": "JPEG",
      "dimensions": {"width": 4032, "height": 3024},
      "file_size": "3.2 MB",
      "color_space": "sRGB"
    },
    "capture_info": {
      "taken_at": "2024-08-15T18:30:00Z",
      "camera": {
        "make": "Apple",
        "model": "iPhone 13 Pro",
        "lens": "iPhone 13 Pro back camera 5.7mm f/1.5"
      },
      "settings": {
        "iso": 64,
        "aperture": "f/1.5",
        "shutter_speed": "1/120s",
        "focal_length": "5.7mm"
      }
    },
    "location_info": {
      "latitude": 36.3932,
      "longitude": 25.4615,
      "address": "Santorini, South Aegean, Greece",
      "city": "Santorini",
      "country": "Greece"
    },
    "interaction": {
      "view_count": 15,
      "is_favorite": false,
      "tags": ["sunset", "beach", "vacation", "greece"],
      "quality_score": 0.92
    },
    "related_files": [
      {
        "id": 124,
        "name": "beach_morning.jpg",
        "thumbnail_url": "/api/v1/files/124/thumbnail?size=small",
        "similarity_score": 0.85
      }
    ]
  }
}
```

### 4. 画廊统计API
```http
GET /api/v1/gallery/stats/overview
GET /api/v1/gallery/stats/timeline
GET /api/v1/gallery/stats/locations
```

**统计响应:**
```json
{
  "success": true,
  "data": {
    "overview": {
      "total_images": 2458,
      "total_videos": 387,
      "total_size": "15.2 GB",
      "date_range": {
        "earliest": "2020-01-01",
        "latest": "2024-12-31"
      }
    },
    "by_year": [
      {"year": 2024, "images": 856, "videos": 124},
      {"year": 2023, "images": 742, "videos": 98}
    ],
    "top_locations": [
      {"name": "Santorini, Greece", "count": 156},
      {"name": "Tokyo, Japan", "count": 89}
    ],
    "top_cameras": [
      {"model": "iPhone 13 Pro", "count": 1234},
      {"model": "Canon EOS R5", "count": 567}
    ]
  }
}
```

### 5. 画廊交互API
```http
POST /api/v1/gallery/files/{id}/favorite
DELETE /api/v1/gallery/files/{id}/favorite
GET /api/v1/gallery/favorites
POST /api/v1/gallery/files/{id}/view
```

## 🛠️ 实现方案

### 方案1: 扩展现有file-service
**优点:**
- 复用现有数据模型和存储逻辑
- 减少服务间调用开销
- 维护成本较低

**缺点:**
- 文件服务职责过重
- 推荐算法逻辑复杂，影响核心文件功能

### 方案2: 创建独立gallery-service ⭐ (推荐)
**优点:**
- 职责分离，专注画廊功能
- 可以独立优化推荐算法
- 支持未来AI功能扩展
- 更好的可测试性和维护性

**缺点:**
- 需要额外的服务部署
- 增加系统复杂度

### 推荐实现架构

```
Gallery Service (Port: 8085)
├── internal/
│   ├── handlers/
│   │   ├── recommendation_handler.go  # 推荐API
│   │   ├── search_handler.go         # 搜索API
│   │   ├── detail_handler.go         # 详情API
│   │   └── interaction_handler.go    # 交互API
│   ├── services/
│   │   ├── recommendation_service.go # 推荐算法
│   │   ├── search_service.go        # 搜索逻辑
│   │   └── analytics_service.go     # 统计分析
│   ├── models/
│   │   ├── gallery_models.go        # 画廊专用模型
│   │   └── interaction_models.go    # 交互数据模型
│   └── external/
│       ├── file_service_client.go   # 文件服务客户端
│       └── ai_service_client.go     # AI服务客户端(未来)
```

## 🔄 数据流设计

### 每日推荐流程
```
1. Gallery Service 调用 File Service 获取用户所有媒体文件
2. 应用质量评分算法（基于分辨率、文件大小、EXIF数据）
3. 应用多样性算法（时间分布、地点分布、内容类型）
4. 基于用户历史行为调整权重
5. 返回推荐结果
```

### 全局搜索流程
```
1. 接收用户搜索请求
2. 解析查询条件和过滤器
3. 调用 File Service 进行数据库查询
4. 对结果进行排序和相关性评分
5. 返回格式化的搜索结果
```

## 📊 性能考虑

### 缓存策略
- **推荐结果缓存**: Redis缓存每日推荐，有效期24小时
- **搜索结果缓存**: 热门搜索词结果缓存，有效期1小时  
- **元数据缓存**: 文件详情信息缓存，有效期6小时
- **统计数据缓存**: 概览统计缓存，有效期30分钟

### 数据库优化
- 为推荐算法创建复合索引
- 分离读写操作，推荐系统使用只读副本
- 对大量历史数据进行分表

## 🚧 开发优先级

### MVP阶段 (2周)
1. ✅ 基础画廊详情API - 格式化元数据展示
2. ✅ 全局搜索API - 跨文件夹媒体搜索  
3. ✅ 简单推荐API - 基于时间和质量的推荐
4. ✅ 基础统计API - 媒体数量和分布统计

### 增强阶段 (4周)
1. 🔄 智能推荐算法 - 基于用户行为的个性化推荐
2. 🔄 视觉相似搜索 - 以图搜图功能(需要AI服务)
3. 🔄 交互功能API - 收藏、标签、评分
4. 🔄 高级筛选 - 地理位置、设备型号等维度

### 完整阶段 (6周)  
1. ⏳ AI驱动推荐 - 基于图像内容的智能推荐
2. ⏳ 实时搜索建议 - 搜索自动补全和拼写纠错
3. ⏳ 社交功能API - 分享、评论、点赞
4. ⏳ 高级统计 - 查看热力图、使用趋势分析

## 🔚 总结

现有文件服务API确实无法满足用户画廊的完整需求。建议：

1. **短期方案**: 在file-service中快速添加画廊相关的API端点
2. **长期方案**: 创建独立的gallery-service，专门处理画廊功能
3. **核心缺失功能**: 推荐算法、全局搜索、丰富元数据、智能筛选

这样的设计既能满足画廊的特殊需求，又保持了系统架构的清晰和可扩展性。