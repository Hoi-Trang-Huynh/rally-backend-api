package firebase

import (
	"context"
	"log"
	"sync"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var (
	client *firebase.App
	firebaseOnce	sync.Once
)

func InitializeClient(credentialsPath string) error {
	var err error
	firebaseOnce.Do(func() {
		ctx := context.Background()

		var opt option.ClientOption
		if credentialsPath != "" {
			opt = option.WithCredentialsFile(credentialsPath)
			client, err = firebase.NewApp(ctx, nil, opt)
		} else {
			// Use Application Default Credentials on Cloud Run
			client, err = firebase.NewApp(ctx, nil)
		}

		if err != nil {
			log.Printf("Error initializing Firebase app: %v", err)
			return
		}
		log.Println("Firebase initialized successfully")
	})
	return err
}


func GetClient() *firebase.App {
	if client == nil {
		log.Fatal("Firebase client not initialized. Call InitializeClient() first")
	}
	return client
}

func MustInitialize(credentialsPath string) {
	if err := InitializeClient(credentialsPath); err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
}