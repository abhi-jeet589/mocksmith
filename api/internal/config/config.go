package config

import "os"

type Config struct {
	Port     string
	MongoURI string
	MongoDB  string
}

func Load() Config {
	return Config{
		Port:     getenv("PORT", "8080"),
		MongoURI: getenv("MONGO_URI", "mongodb://mongo:27017"),
		MongoDB:  getenv("MONGO_DB", "mocksmith"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
