package database

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	clientInstance     *mongo.Client
	dbInstance         *mongo.Database
	internalDbInstance *mongo.Database
	mongoOnce          sync.Once
)

func InitializeDatabase(cfg config.DatabaseConfig) error {
	var err error
	mongoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOpts := options.Client().ApplyURI(cfg.MONGODB_URI)
		clientInstance, err = mongo.Connect(ctx, clientOpts)
		if err != nil {
			log.Printf("Failed to connect to MongoDB: %v", err)
			return
		}

		if err = clientInstance.Ping(ctx, readpref.Primary()); err != nil {
			log.Printf("MongoDB ping failed: %v", err)
			return
		}

		dbInstance = clientInstance.Database(cfg.MONGODB_DB)
		log.Printf("✅ MongoDB connected successfully (DB: %s)", cfg.MONGODB_DB)

		if cfg.MONGODB_INTERNAL_DB != "" {
			internalDbInstance = clientInstance.Database(cfg.MONGODB_INTERNAL_DB)
			log.Printf("✅ MongoDB Internal connected successfully (DB: %s)", cfg.MONGODB_INTERNAL_DB)
		}
	})

	return err
}

func GetDB() *mongo.Database {
	if dbInstance == nil {
		log.Fatal("MongoDB not initialized. Call InitializeDatabase() first.")
	}
	return dbInstance
}

func GetInternalDB() *mongo.Database {
	if internalDbInstance == nil {
		log.Fatal("MongoDB Internal not initialized or not configured.")
	}
	return internalDbInstance
}

func CloseDatabase() {
	if clientInstance != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := clientInstance.Disconnect(ctx); err != nil {
			log.Printf("⚠️ Error closing MongoDB connection: %v", err)
		} else {
			log.Println("🧹 MongoDB connection closed")
		}
	}
}
