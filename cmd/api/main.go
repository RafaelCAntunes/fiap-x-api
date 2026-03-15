package main

import (
	"context"
	"fiap-x-api/internal/config"
	"fiap-x-api/internal/handler"
	"fiap-x-api/internal/infra/aws"
	"fiap-x-api/internal/infra/repository"
	"fiap-x-api/internal/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.LoadConfig()
	pgConfig := repository.PostgresConfig{
		Host:     cfg.DBHost,
		Port:     "5432",
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  "required",
	}

	db := repository.NewPostgresConnection(pgConfig)

	videoRepo := repository.NewPostgresVideoRepository(db)

	isLocal := cfg.EnvVariable == "local" || cfg.EnvVariable == ""

	awsCfg, err := aws.NewAWSConfig(isLocal)
	if err != nil {
		log.Fatal("Erro ao configurar AWS:", err)
	}

	s3Adapter := aws.NewS3Adapter(awsCfg, cfg.S3Bucket)
	sqsAdapter := aws.NewSQSAdapter(awsCfg, cfg.SQSURL)

	err = s3Adapter.EnsureBucketExists(context.Background())
	if err != nil {
		log.Printf("Aviso: Não foi possível garantir o bucket: %v", err)
	}

	authService := &service.AuthService{DB: db}
	videoHandler := &handler.VideoHandler{
		Repo:       videoRepo,
		S3Adapter:  s3Adapter,
		SQSAdapter: sqsAdapter,
	}

	r := gin.Default()

	r.LoadHTMLGlob("web/*.html")

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/login-page", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.POST("/login", func(c *gin.Context) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{"error": "Dados inválidos"})
			return
		}

		token, err := authService.Login(body.Username, body.Password)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"token": token})
	})

	// Rotas Protegidas
	authorized := r.Group("/")
	authorized.Use(handler.AuthMiddleware())
	{

		authorized.POST("/upload", videoHandler.Upload)
		authorized.GET("/status", videoHandler.ListStatus)
		authorized.GET("/download/:id", videoHandler.DownloadZip)
	}

	r.Run(":8080")
}
