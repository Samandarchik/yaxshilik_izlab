package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type UploadHandler struct {
	cfg    *Config
	client *minio.Client
	bucket string
	public string // public host prefix (optional)
}

func NewUploadHandler(cfg *Config) *UploadHandler {
	if cfg.MinioEndpoint == "" {
		log.Println("upload: MinIO not configured")
		return &UploadHandler{cfg: cfg}
	}
	endpoint := cfg.MinioEndpoint
	useSSL := strings.HasPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")

	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Printf("upload: minio init failed: %v", err)
		return &UploadHandler{cfg: cfg}
	}

	// Bucket borligini tekshirish (yo'q bo'lsa yaratish)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	exists, err := cli.BucketExists(ctx, cfg.MinioBucketName)
	if err != nil {
		log.Printf("upload: bucket check failed: %v", err)
	} else if !exists {
		if err := cli.MakeBucket(ctx, cfg.MinioBucketName, minio.MakeBucketOptions{}); err != nil {
			log.Printf("upload: make bucket failed: %v", err)
		}
	}

	pub := cfg.MinioPublicHost
	if pub == "" {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		pub = fmt.Sprintf("%s://%s", scheme, endpoint)
	}

	return &UploadHandler{cfg: cfg, client: cli, bucket: cfg.MinioBucketName, public: pub}
}

const maxUploadBytes = 8 << 20 // 8 MB

// Faqat rasm fayllariga ruxsat (kengaytma + haqiqiy MIME)
var allowedImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
}
var allowedImageMIME = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/webp": true, "image/gif": true,
}

// POST /api/admin/upload  (multipart, field 'file')
func (h *UploadHandler) Upload(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MinIO not configured"})
		return
	}

	// Hajmni o'qishdan oldin cheklash (DoS himoyasi)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required (max 8MB)"})
		return
	}
	defer file.Close()

	if header.Size <= 0 || header.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fayl hajmi 0..8MB bo'lishi kerak"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImageExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "faqat rasm: jpg, jpeg, png, webp, gif"})
		return
	}

	// Haqiqiy mazmun turini tekshirish (kengaytmaga ishonib bo'lmaydi)
	sniff := make([]byte, 512)
	n, _ := io.ReadFull(file, sniff)
	contentType := http.DetectContentType(sniff[:n])
	if !allowedImageMIME[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fayl mazmuni rasm emas"})
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read error"})
		return
	}

	objName := fmt.Sprintf("uploads/%d-%s%s", time.Now().UnixNano(), secureRandHex(8), ext)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	_, err = h.client.PutObject(ctx, h.bucket, objName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed: " + err.Error()})
		return
	}

	publicURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(h.public, "/"), h.bucket, objName)
	c.JSON(http.StatusOK, gin.H{
		"url":    publicURL,
		"object": objName,
	})
}
