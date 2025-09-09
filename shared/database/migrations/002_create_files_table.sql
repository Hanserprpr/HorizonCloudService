-- 创建文件表
CREATE TABLE IF NOT EXISTS files (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    folder_id BIGINT UNSIGNED DEFAULT NULL,
    filename VARCHAR(255) NOT NULL COMMENT '文件名',
    original_name VARCHAR(255) NOT NULL COMMENT '原始文件名',
    file_path VARCHAR(1000) NOT NULL COMMENT '文件存储路径',
    file_size BIGINT NOT NULL COMMENT '文件大小(字节)',
    file_hash VARCHAR(64) NOT NULL COMMENT 'SHA-256文件哈希',
    mime_type VARCHAR(100) NOT NULL COMMENT 'MIME类型',
    file_extension VARCHAR(20) COMMENT '文件扩展名',
    storage_type ENUM('local', 's3', 'minio') NOT NULL DEFAULT 'minio',
    storage_tier ENUM('hot', 'warm', 'cold', 'archive') NOT NULL DEFAULT 'hot',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    download_count INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    parent_file_id BIGINT UNSIGNED DEFAULT NULL COMMENT '父版本文件ID',
    metadata JSON COMMENT '文件元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_folder_id (folder_id),
    INDEX idx_file_hash (file_hash),
    INDEX idx_mime_type (mime_type),
    INDEX idx_storage_tier (storage_tier),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件表';

-- 创建文件夹表
CREATE TABLE IF NOT EXISTS folders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    parent_id BIGINT UNSIGNED DEFAULT NULL,
    name VARCHAR(255) NOT NULL COMMENT '文件夹名称',
    path VARCHAR(1000) NOT NULL COMMENT '文件夹路径',
    level INT NOT NULL DEFAULT 0 COMMENT '层级深度',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_parent_id (parent_id),
    INDEX idx_path (path(255)),
    INDEX idx_level (level),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件夹表';

-- 创建缩略图表
CREATE TABLE IF NOT EXISTS thumbnails (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    file_id BIGINT UNSIGNED NOT NULL,
    size VARCHAR(20) NOT NULL COMMENT '尺寸规格(如: small, medium, large)',
    width INT NOT NULL,
    height INT NOT NULL,
    file_path VARCHAR(1000) NOT NULL,
    file_size BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_file_id (file_id),
    INDEX idx_size (size),
    UNIQUE KEY uk_file_size (file_id, size),
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='缩略图表';