package test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"file-service/internal/models"
	"file-service/internal/repository"
	"file-service/internal/services"
	"file-service/internal/storage"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// BenchmarkConfig 基准测试配置
type BenchmarkConfig struct {
	DB      *gorm.DB
	Storage storage.Storage
	Repos   *repository.Repositories
	Service *services.Services
}

// setupBenchmarkConfig 设置基准测试环境
func setupBenchmarkConfig(b *testing.B) *BenchmarkConfig {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(b, err)

	// 运行迁移
	err = db.AutoMigrate(
		&models.File{},
		&models.Folder{},
		&models.Thumbnail{},
		&models.UploadSession{},
	)
	require.NoError(b, err)

	// 创建存储和仓库
	stor := storage.NewMemoryStorage()
	repos := &repository.Repositories{
		File:   repository.NewFileRepository(db),
		Folder: repository.NewFolderRepository(db),
		Upload: repository.NewUploadRepository(db),
	}

	// 创建服务
	config := &services.ServiceConfig{
		Storage:          stor,
		UserService:      services.NewMockUserServiceClient(),
		DefaultChunkSize: 1024 * 1024, // 1MB
	}
	svc := services.NewServices(repos, config)

	return &BenchmarkConfig{
		DB:      db,
		Storage: stor,
		Repos:   repos,
		Service: svc,
	}
}

// 文件服务基准测试

func BenchmarkFileUpload(b *testing.B) {
	config := setupBenchmarkConfig(b)
	ctx := context.Background()
	userID := uint(1)

	sizes := []int{
		1024,        // 1KB
		1024 * 10,   // 10KB
		1024 * 100,  // 100KB
		1024 * 1024, // 1MB
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%dB", size), func(b *testing.B) {
			content := strings.Repeat("x", size)
			b.SetBytes(int64(size))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := &services.UploadFileRequest{
					FileName:    fmt.Sprintf("bench-%d.txt", i),
					Size:        int64(size),
					ContentType: "text/plain",
					UserID:      userID,
					Reader:      strings.NewReader(content),
				}

				_, err := config.Service.File.UploadFile(ctx, req)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkFileList(b *testing.B) {
	config := setupBenchmarkConfig(b)
	ctx := context.Background()
	userID := uint(1)

	// 预创建文件
	fileCounts := []int{10, 100, 1000}

	for _, count := range fileCounts {
		b.Run(fmt.Sprintf("Files%d", count), func(b *testing.B) {
			// 创建测试文件
			for i := 0; i < count; i++ {
				file := &models.File{
					Name:        fmt.Sprintf("file-%d.txt", i),
					Size:        1024,
					ContentType: "text/plain",
					UserID:      userID,
					Hash:        fmt.Sprintf("hash-%d", i),
					StoragePath: fmt.Sprintf("/test/file-%d.txt", i),
					Status:      models.FileStatusActive,
					Category:    "document",
				}
				config.DB.Create(file)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, err := config.Service.File.ListFiles(ctx, userID, nil, 0, 50, nil)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkFileSearch(b *testing.B) {
	config := setupBenchmarkConfig(b)
	ctx := context.Background()
	userID := uint(1)

	// 创建测试文件
	for i := 0; i < 1000; i++ {
		file := &models.File{
			Name:        fmt.Sprintf("document-%d.txt", i),
			Size:        1024,
			ContentType: "text/plain",
			UserID:      userID,
			Hash:        fmt.Sprintf("hash-%d", i),
			StoragePath: fmt.Sprintf("/test/document-%d.txt", i),
			Status:      models.FileStatusActive,
			Category:    "document",
		}
		config.DB.Create(file)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := config.Service.File.SearchFiles(ctx, userID, "document", 0, 20, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFileDelete(b *testing.B) {
	config := setupBenchmarkConfig(b)
	ctx := context.Background()
	userID := uint(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 创建文件
		file := &models.File{
			Name:        fmt.Sprintf("delete-%d.txt", i),
			Size:        1024,
			ContentType: "text/plain",
			UserID:      userID,
			Hash:        fmt.Sprintf("delete-hash-%d", i),
			StoragePath: fmt.Sprintf("/test/delete-%d.txt", i),
			Status:      models.FileStatusActive,
		}
		config.DB.Create(file)

		// 删除文件
		err := config.Service.File.DeleteFile(ctx, file.ID, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 上传服务基准测试

func BenchmarkChunkedUpload(b *testing.B) {
	config := setupBenchmarkConfig(b)
	ctx := context.Background()
	userID := uint(1)

	chunkSizes := []int64{
		256 * 1024,   // 256KB
		1024 * 1024,  // 1MB
		5 * 1024 * 1024, // 5MB
	}

	for _, chunkSize := range chunkSizes {
		b.Run(fmt.Sprintf("ChunkSize%dB", chunkSize), func(b *testing.B) {
			fileSize := chunkSize * 4 // 4个分片
			b.SetBytes(fileSize)

			for i := 0; i < b.N; i++ {
				// 初始化上传
				initReq := &services.InitiateUploadRequest{
					FileName:    fmt.Sprintf("chunk-bench-%d.txt", i),
					Size:        fileSize,
					ContentType: "text/plain",
					UserID:      userID,
					ChunkSize:   chunkSize,
				}

				result, err := config.Service.Upload.InitiateUpload(ctx, initReq)
				if err != nil {
					b.Fatal(err)
				}

				// 上传分片
				for chunkIndex := 0; chunkIndex < 4; chunkIndex++ {
					chunkContent := strings.Repeat("x", int(chunkSize))
					chunkReq := &services.UploadChunkRequest{
						SessionID:  result.SessionID,
						ChunkIndex: chunkIndex,
						ChunkSize:  chunkSize,
						ChunkData:  strings.NewReader(chunkContent),
						UserID:     userID,
					}

					_, err := config.Service.Upload.UploadChunk(ctx, chunkReq)
					if err != nil {
						b.Fatal(err)
					}
				}

				// 完成上传
				_, err = config.Service.Upload.CompleteUpload(ctx, result.SessionID, userID)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkConcurrentUploads(b *testing.B) {
	config := setupBenchmarkConfig(b)
	ctx := context.Background()

	concurrency := []int{1, 5, 10, 20}

	for _, concurrent := range concurrency {
		b.Run(fmt.Sprintf("Concurrent%d", concurrent), func(b *testing.B) {
			content := strings.Repeat("x", 1024) // 1KB files
			b.SetBytes(1024)

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(concurrent)

				for j := 0; j < concurrent; j++ {
					go func(index int) {
						defer wg.Done()

						userID := uint(index%10 + 1) // 10个不同用户
						req := &services.UploadFileRequest{
							FileName:    fmt.Sprintf("concurrent-%d-%d.txt", i, index),
							Size:        1024,
							ContentType: "text/plain",
							UserID:      userID,
							Reader:      strings.NewReader(content),
						}

						_, err := config.Service.File.UploadFile(ctx, req)
						if err != nil {
							b.Error(err)
						}
					}(j)
				}

				wg.Wait()
			}
		})
	}
}

// 存储基准测试

func BenchmarkStorageOperations(b *testing.B) {
	stor := storage.NewMemoryStorage()
	ctx := context.Background()

	sizes := []int{
		1024,        // 1KB
		1024 * 10,   // 10KB
		1024 * 100,  // 100KB
		1024 * 1024, // 1MB
	}

	for _, size := range sizes {
		content := strings.Repeat("x", size)
		b.Run(fmt.Sprintf("Upload%dB", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				path := fmt.Sprintf("/bench/upload-%d.txt", i)
				err := stor.Upload(ctx, path, strings.NewReader(content), int64(size))
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		// 先上传一些文件用于下载测试
		for i := 0; i < 100; i++ {
			path := fmt.Sprintf("/bench/download-%d.txt", i)
			stor.Upload(ctx, path, strings.NewReader(content), int64(size))
		}

		b.Run(fmt.Sprintf("Download%dB", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				path := fmt.Sprintf("/bench/download-%d.txt", i%100)
				reader, err := stor.Download(ctx, path)
				if err != nil {
					b.Fatal(err)
				}
				io.Copy(io.Discard, reader)
				reader.Close()
			}
		})
	}
}

// 数据库操作基准测试

func BenchmarkDatabaseOperations(b *testing.B) {
	config := setupBenchmarkConfig(b)
	ctx := context.Background()

	b.Run("FileCreate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			file := &models.File{
				Name:        fmt.Sprintf("bench-create-%d.txt", i),
				Size:        1024,
				ContentType: "text/plain",
				UserID:      uint(i%100 + 1), // 100个不同用户
				Hash:        fmt.Sprintf("hash-%d", i),
				StoragePath: fmt.Sprintf("/bench/create-%d.txt", i),
				Status:      models.FileStatusActive,
				Category:    "document",
			}

			err := config.Repos.File.Create(ctx, file)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 预创建文件用于查询测试
	for i := 0; i < 10000; i++ {
		file := &models.File{
			Name:        fmt.Sprintf("bench-query-%d.txt", i),
			Size:        1024,
			ContentType: "text/plain",
			UserID:      uint(i%100 + 1),
			Hash:        fmt.Sprintf("query-hash-%d", i),
			StoragePath: fmt.Sprintf("/bench/query-%d.txt", i),
			Status:      models.FileStatusActive,
			Category:    "document",
		}
		config.DB.Create(file)
	}

	b.Run("FileQuery", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userID := uint(i%100 + 1)
			_, _, err := config.Repos.File.List(ctx, userID, nil, 0, 50, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("FileSearch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			userID := uint(i%100 + 1)
			_, _, err := config.Repos.File.Search(ctx, userID, "bench", 0, 20, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 内存使用基准测试

func BenchmarkMemoryUsage(b *testing.B) {
	config := setupBenchmarkConfig(b)
	ctx := context.Background()

	b.Run("LargeFileUpload", func(b *testing.B) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		for i := 0; i < b.N; i++ {
			content := strings.Repeat("x", 10*1024*1024) // 10MB
			req := &services.UploadFileRequest{
				FileName:    fmt.Sprintf("large-%d.txt", i),
				Size:        int64(len(content)),
				ContentType: "text/plain",
				UserID:      1,
				Reader:      strings.NewReader(content),
			}

			_, err := config.Service.File.UploadFile(ctx, req)
			if err != nil {
				b.Fatal(err)
			}
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)
		b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB/op")
	})
}

// 压力测试

func BenchmarkStressTest(b *testing.B) {
	if os.Getenv("STRESS_TESTS") == "" {
		b.Skip("设置 STRESS_TESTS=1 环境变量来运行压力测试")
	}

	config := setupBenchmarkConfig(b)
	ctx := context.Background()

	b.Run("HighConcurrency", func(b *testing.B) {
		const concurrency = 100
		const opsPerGoroutine = 10

		b.SetParallelism(concurrency)
		b.RunParallel(func(pb *testing.PB) {
			userID := uint(1)
			i := 0
			for pb.Next() {
				content := fmt.Sprintf("stress-test-content-%d", i)
				req := &services.UploadFileRequest{
					FileName:    fmt.Sprintf("stress-%d.txt", i),
					Size:        int64(len(content)),
					ContentType: "text/plain",
					UserID:      userID,
					Reader:      strings.NewReader(content),
				}

				_, err := config.Service.File.UploadFile(ctx, req)
				if err != nil {
					b.Error(err)
				}
				i++
			}
		})
	})
}

// 性能回归测试

func BenchmarkPerformanceRegression(b *testing.B) {
	config := setupBenchmarkConfig(b)
	ctx := context.Background()
	userID := uint(1)

	// 基准场景：上传1KB文件
	b.Run("Baseline1KB", func(b *testing.B) {
		content := strings.Repeat("x", 1024)
		b.SetBytes(1024)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := &services.UploadFileRequest{
				FileName:    fmt.Sprintf("baseline-%d.txt", i),
				Size:        1024,
				ContentType: "text/plain",
				UserID:      userID,
				Reader:      strings.NewReader(content),
			}

			_, err := config.Service.File.UploadFile(ctx, req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 文件列表性能
	b.Run("FileList50", func(b *testing.B) {
		// 预创建文件
		for i := 0; i < 200; i++ {
			file := &models.File{
				Name:        fmt.Sprintf("list-test-%d.txt", i),
				Size:        1024,
				ContentType: "text/plain",
				UserID:      userID,
				Hash:        fmt.Sprintf("list-hash-%d", i),
				StoragePath: fmt.Sprintf("/list/test-%d.txt", i),
				Status:      models.FileStatusActive,
			}
			config.DB.Create(file)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := config.Service.File.ListFiles(ctx, userID, nil, 0, 50, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 资源利用率基准测试

func BenchmarkResourceUtilization(b *testing.B) {
	config := setupBenchmarkConfig(b)

	b.Run("CPUIntensive", func(b *testing.B) {
		ctx := context.Background()
		userID := uint(1)

		for i := 0; i < b.N; i++ {
			// 模拟CPU密集型操作
			content := strings.Repeat("x", 1024*100) // 100KB
			var buf bytes.Buffer
			for j := 0; j < 100; j++ {
				buf.WriteString(content)
			}

			req := &services.UploadFileRequest{
				FileName:    fmt.Sprintf("cpu-intensive-%d.txt", i),
				Size:        int64(buf.Len()),
				ContentType: "text/plain",
				UserID:      userID,
				Reader:      &buf,
			}

			_, err := config.Service.File.UploadFile(ctx, req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MemoryIntensive", func(b *testing.B) {
		ctx := context.Background()
		userID := uint(1)

		for i := 0; i < b.N; i++ {
			// 模拟内存密集型操作
			content := strings.Repeat("x", 1024*1024*5) // 5MB
			req := &services.UploadFileRequest{
				FileName:    fmt.Sprintf("memory-intensive-%d.txt", i),
				Size:        int64(len(content)),
				ContentType: "text/plain",
				UserID:      userID,
				Reader:      strings.NewReader(content),
			}

			_, err := config.Service.File.UploadFile(ctx, req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 运行所有基准测试的助手函数
func BenchmarkAll(b *testing.B) {
	if os.Getenv("BENCHMARK_ALL") == "" {
		b.Skip("设置 BENCHMARK_ALL=1 环境变量来运行所有基准测试")
	}

	benchmarks := []struct {
		name string
		fn   func(*testing.B)
	}{
		{"FileUpload", BenchmarkFileUpload},
		{"FileList", BenchmarkFileList},
		{"FileSearch", BenchmarkFileSearch},
		{"ChunkedUpload", BenchmarkChunkedUpload},
		{"ConcurrentUploads", BenchmarkConcurrentUploads},
		{"StorageOperations", BenchmarkStorageOperations},
		{"DatabaseOperations", BenchmarkDatabaseOperations},
		{"MemoryUsage", BenchmarkMemoryUsage},
	}

	for _, bench := range benchmarks {
		b.Run(bench.name, bench.fn)
	}
}