package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const url = "https://leetcode.com/api/problems/all/"

func Cron(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	cronSecret := os.Getenv("CRON_SECRET")

	if authHeader != "Bearer "+cronSecret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	manager := NewRedisManager()

	// Total count and memory before
	totalCount, usedMemory := manager.GetTotalCountAndMemory()
	log.Printf("Total count before: %d, Used memory: %s", totalCount, usedMemory)

	// Fetch and push problems to Redis
	manager.FetchAndPushProblems()

	// Total count and memory after
	totalCount, usedMemory = manager.GetTotalCountAndMemory()
	log.Printf("Total count after: %d, Used memory: %s", totalCount, usedMemory)

	manager.Close()

	response := map[string]string{
		"status":  "success",
		"message": "Problems fetched and pushed to Redis successfully",
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

type RedisManager struct {
	redisClient *redis.Client
}

func NewRedisManager() *RedisManager {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),  // Redis server address
		Username: os.Getenv("REDIS_USER"),     // Redis username
		Password: os.Getenv("REDIS_PASSWORD"), // Redis password
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")
	return &RedisManager{
		redisClient: redisClient,
	}
}

func (rm *RedisManager) Close() {
	if rm.redisClient != nil {
		if err := rm.redisClient.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		} else {
			log.Println("Redis client closed")
		}
	}
}

func (rm *RedisManager) Clear() {
	if rm.redisClient != nil {
		if err := rm.redisClient.FlushAll(context.Background()).Err(); err != nil {
			log.Printf("Error clearing Redis: %v", err)
		} else {
			log.Println("Redis cleared")
		}
	}
}

func (rm *RedisManager) GetTotalCountAndMemory() (int64, string) {
	if rm.redisClient != nil {
		count, err := rm.redisClient.DBSize(context.Background()).Result()
		if err != nil {
			log.Printf("Error getting Redis size: %v", err)
			return 0, ""
		}
		// Fetch only "used_memory_human" from the memory info
		memoryInfo, err := rm.redisClient.Info(context.Background(), "memory").Result()
		if err != nil {
			log.Printf("Error getting Redis memory info: %v", err)
			return count, ""
		}
		usedMemory := extractUsedMemory(memoryInfo)
		return count, usedMemory
	}
	return 0, ""
}

func (rm *RedisManager) FetchAndPushProblems() {
	// Create an HTTP client with timeout
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return
	}

	response, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch problems: %v", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code: %d", response.StatusCode)
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	problemsData := struct {
		StatStatusPairs []struct {
			Stat struct {
				FrontendQuestionID int    `json:"frontend_question_id"`
				QuestionTitleSlug  string `json:"question__title_slug"`
			} `json:"stat"`
		} `json:"stat_status_pairs"`
	}{}

	err = json.Unmarshal(body, &problemsData)
	if err != nil {
		log.Printf("Error unmarshalling problems data: %v", err)
		return
	}
	log.Printf("Fetched %d problems from LeetCode", len(problemsData.StatStatusPairs))

	// Build key-value pairs for MSet
	keyValues := make([]interface{}, 0, len(problemsData.StatStatusPairs)*2)
	for _, problem := range problemsData.StatStatusPairs {
		key := strconv.Itoa(problem.Stat.FrontendQuestionID)
		value := problem.Stat.QuestionTitleSlug
		keyValues = append(keyValues, key, value)
	}

	// Push to Redis using MSet
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = rm.redisClient.MSet(ctx, keyValues...).Err()
	if err != nil {
		log.Printf("Error performing MSet: %v", err)
		return
	}

	log.Printf("Successfully pushed %d problems to Redis", len(problemsData.StatStatusPairs))
}

func extractUsedMemory(info string) string {
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "used_memory_human:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
