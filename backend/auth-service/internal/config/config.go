package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	MongoURI string
	JWTKey   []byte
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	jwtKey := []byte(os.Getenv("JWT_KEY"))
	if jwtKey == nil {
		jwtKey = []byte("mysecretkey")
	}

	return Config{
		MongoURI: mongoURI,
		JWTKey:   jwtKey,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
