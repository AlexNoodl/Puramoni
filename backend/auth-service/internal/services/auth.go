package services

import (
	"backend/internal/db"
	"backend/internal/logger"
	"backend/internal/models"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthService struct {
	db     *db.MongoClient
	secret []byte
	log    *logger.Logger
}

func NewAuthService(db *db.MongoClient, secret []byte, log *logger.Logger) *AuthService {
	return &AuthService{db: db, secret: secret, log: log}
}

func (s *AuthService) Register(user *models.User) (string, error) {
	collection := s.db.Database("library").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"$or": []bson.M{
		{"username": user.Username},
		{"email": user.Email},
	}}).Decode(&existingUser)
	if err == nil {
		s.log.Info("User registration failed: username or email already exists", zap.String("username", user.Username))
		return "", ErrUserExists
	}
	if err != mongo.ErrNoDocuments {
		s.log.Error("Failed to check user existence", zap.Error(err))
		return "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("Failed to hash password", zap.Error(err))
		return "", err
	}
	user.Password = string(hashedPassword)

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		s.log.Error("Failed to register user", zap.Error(err))
		return "", err
	}

	s.log.Info("User registered successfully", zap.String("username", user.Username))
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (s *AuthService) Login(login, password string) (string, error) {
	collection := s.db.Database("library").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"$or": []bson.M{
		{"username": login},
		{"email": login},
	}}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		s.log.Info("Login failed: invalid credentials", zap.String("username", login))
		return "", ErrInvalidCredentials
	}
	if err != nil {
		s.log.Error("Failed to fetch user", zap.Error(err))
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.log.Info("Login failed: invalid password", zap.String("username", login))
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   user.ID.Hex(),
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	})

	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		s.log.Error("Failed to generate token", zap.Error(err))
		return "", err
	}

	s.log.Info("User logged in successfully", zap.String("login", login), zap.String("id", user.ID.Hex()))
	return tokenString, nil
}

var (
	ErrInvalidCredentials = mongo.ErrNoDocuments
	ErrUserExists         = mongo.ErrNoDocuments
)
