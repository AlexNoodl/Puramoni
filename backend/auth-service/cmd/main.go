package main

import (
	_ "backend/docs" // Swagger
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/logger"
	"backend/internal/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.NewLogger()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting application")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("Failed to load config", zap.Error(err))
		panic("Failed to load config: " + err.Error())
	}
	log.Info("Config loaded successfully")

	mongoClient, err := db.ConnectMongo(cfg.MongoURI, log)
	if err != nil {
		log.Error("Failed to connect to MongoDB", zap.Error(err))
		panic("Failed to connect to MongoDB: " + err.Error())
	}
	defer mongoClient.Close()

	r := gin.Default()

	routes.SetupAuthRoutes(r, mongoClient, cfg, log)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Info("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Error("Failed to run server", zap.Error(err))
		panic(err)
	}

}
