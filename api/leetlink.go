package handler

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

// Initiate redis client
func initRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),  // Redis server address
		Username: os.Getenv("REDIS_USER"),     // Redis username
		Password: os.Getenv("REDIS_PASSWORD"), // Redis password
		DB:       0,                           // Use default DB
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")
	return client
}

// Close the Redis client
func closeRedisClient(client *redis.Client) {
	if client != nil {
		if err := client.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		} else {
			log.Println("Redis client closed")
		}
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Log the request details
	log.Printf("Received request from IP: %s User-Agent: %s", r.RemoteAddr, r.UserAgent())

	// Fetch the problem_id from the query parameters
	problemID := r.URL.Query().Get("problem_id")
	log.Printf("Received request for problem ID: %s", problemID)
	if problemID == "" {
		http.Error(w, "problem_id is required", http.StatusBadRequest)
		return
	}

	// Initialize Redis client
	client := initRedisClient()

	// Use the problemID to get the URL from Redis
	redirectSlug, err := client.Get(context.Background(), problemID).Result()

	// Close the Redis client after use
	closeRedisClient(client)

	if err != nil || redirectSlug == "" {
		http.Error(w, "Problem URL not found", http.StatusNotFound)
		return
	}

	// Construct the redirect URL
	redirectURL := "https://leetcode.com/problems/" + redirectSlug + "/description/"
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
