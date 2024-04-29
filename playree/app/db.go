package app

import (
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func createSQLDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(os.Getenv("SQL_DB_ADDRESS")), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func createRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		PoolSize: 10,
	})
}
