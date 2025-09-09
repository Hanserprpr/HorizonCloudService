package main

import (
	"fmt"
	"time"

	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/services"
	"file-service/internal/storage"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("🚀 Testing File Service Core Components...")

	// 1. Initialize database
	fmt.Println("1. Initializing database...")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// 2. Run migrations
	fmt.Println("2. Running database migrations...")
	err = db.AutoMigrate(
		&models.File{},
		&models.Folder{},
		&models.Thumbnail{},
		&models.UploadSession{},
		&models.UploadChunk{},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}

	// 3. Initialize storage
	fmt.Println("3. Initializing storage...")
	localStorage, err := storage.NewLocalStorage(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: "/tmp/file-service-test",
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize storage: %v", err))
	}

	// 4. Initialize repository
	fmt.Println("4. Initializing repository...")
	repo := repository.NewRepository(db)

	// 5. Initialize services
	fmt.Println("5. Initializing services...")
	config := &services.ServicesConfig{
		DefaultChunkSize:    1024 * 1024, // 1MB
		ThumbnailSizes:      []string{"small", "medium", "large"},
		ThumbnailQuality:    85,
		ThumbnailTimeout:    30 * time.Second,
		MaxBatchSize:        100,
		BatchConcurrency:    5,
		SearchLimit:         100,
	}

	serviceContainer := services.NewServices(repo, localStorage, config)

	// 6. Test service initialization
	fmt.Println("6. Testing service availability...")
	if serviceContainer.File == nil {
		panic("File service not initialized")
	}
	if serviceContainer.Folder == nil {
		panic("Folder service not initialized")
	}
	if serviceContainer.Upload == nil {
		panic("Upload service not initialized")
	}
	if serviceContainer.Thumbnail == nil {
		panic("Thumbnail service not initialized")
	}

	fmt.Println("✅ All core components initialized successfully!")
	fmt.Println("📋 Service Components:")
	fmt.Println("   - File Service: Ready")
	fmt.Println("   - Folder Service: Ready")
	fmt.Println("   - Upload Service: Ready")
	fmt.Println("   - Thumbnail Service: Ready")
	fmt.Println("   - Repository Layer: Ready")
	fmt.Println("   - Storage Layer: Ready (Local)")
	fmt.Println("   - Database: Ready (SQLite In-Memory)")
	fmt.Println("")
	fmt.Println("🎉 File Service is ready for deployment!")
}