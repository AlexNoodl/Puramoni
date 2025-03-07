package routes

import (
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/logger"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

// @title Go MongoDB API with JWT
// @version 1.0
// @description Это API с авторизацией через JWT, MongoDB и Gin
// @host localhost:8080
// @BasePath /
func SetupAuthRoutes(r *gin.Engine, db *db.MongoClient, cfg config.Config, log *logger.Logger) {
	authService := services.NewAuthService(db, cfg.JWTKey, log)

	r.POST("/register", registerHandler(authService, log))
	r.POST("/login", loginHandler(authService, log))

	protected := r.Group("/protected")
	protected.Use(middleware.AuthMiddleware(cfg.JWTKey, log))
	{
		protected.GET("/user", func(c *gin.Context) {
			log.Info("Accessed protected/user endpoint")
			c.JSON(http.StatusOK, gin.H{"message": "Доступно всем авторизованным"})
		})
		protected.GET("/admin", func(c *gin.Context) {
			log.Info("Accessed protected/admin endpoint")
			c.JSON(http.StatusOK, gin.H{"message": "Доступно только админам"})
		})
	}
}

// registerHandler регистрация пользователя
// @Summary Регистрация пользователя
// @Description Создаёт нового пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "Данные пользователя"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
func registerHandler(svc *services.AuthService, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Error("Invalid input for registration", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := svc.Register(&user)
		if err == services.ErrUserExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered", "id": id})
	}
}

// loginHandler вход пользователя
// @Summary Вход пользователя
// @Description Возвращает JWT токен, вход по username или email
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body map[string]string true "Логин (username или email) и пароль"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
func loginHandler(svc *services.AuthService, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var creds struct {
			Login    string `json:"login" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&creds); err != nil {
			log.Error("Invalid input for login", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := svc.Login(creds.Login, creds.Password)
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}
