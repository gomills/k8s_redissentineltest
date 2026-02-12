package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SentinelConfig struct {
	Address  string
	Password string
}

func main() {

	// parse sentinel config from environment variables
	sentinelConfig, err := parseSentinelConfig()
	if err != nil {
		panic(err)
	}

	fmt.Println("\n***** Redis Sentinel Test*****\n")
	fmt.Println(">>> started testing Redis")

	// create Redis client to sentinel
	ctx := context.TODO()
	Redis := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       "mymaster",
		SentinelAddrs:    []string{sentinelConfig.Address},
		Password:         sentinelConfig.Password,
		SentinelPassword: sentinelConfig.Password,
		DB:               0,
	})

	// 1. test with ping
	fmt.Println(">>> (1/4) testing ping")
	pong, err := Redis.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("	>>> (1/4) result: failed ping: %s\n", err)
		panic(err)
	}
	fmt.Printf("	>>> (1/4) result: successful ping: %s\n", pong)

	// 2. test writing a key:value (SET)
	testKey := uuid.New().String()
	testValue := "testValue"
	fmt.Printf(">>> (2/4) testing writing the key-value '%s:%s'\n", testKey, testValue)

	// first delete key to make sure it never existed
	if err := Redis.Del(ctx, testKey).Err(); err != nil {
		fmt.Printf("	>>> (2/4) result: failed deleting key: %s\n", err)
		panic(err)
	}

	err = Redis.SAdd(ctx, testKey, testValue).Err()
	if err != nil {
		fmt.Printf("	>>> (2/4) result: failed writing: %s\n", err)
		panic(err)
	}
	fmt.Println("	>>> (2/4) result: successful write")

	// 3. test popping the value of the testing key
	fmt.Println(">>> (3/4) testing popping the test key-value")
	val, err := Redis.SPop(ctx, testKey).Result()
	if err != nil {
		fmt.Printf("	>>> (3/4) result: failed SPop: %s\n", err)
		panic(err)
	}
	fmt.Printf("	>>> (3/4) result: successful SPop: %s\n", val)

	// 4. check that the key doesn't exist after popping
	fmt.Println(">>> (4/4) checking that the key doesn't exist after popping")
	exists, err := Redis.Exists(ctx, testKey).Result()
	if err != nil {
		fmt.Printf("	>>> (4/4) result: failed to check key existence: %s\n", err)
		panic(err)
	}
	if exists == 0 {
		fmt.Println("	>>> (4/4) result: as expected, key does not exist")
	} else {
		fmt.Println("	>>> (4/4) result: ERROR key still exists")
		panic(fmt.Errorf("key was not deleted properly"))
	}

	fmt.Println(">>> finished testing, exiting gracefully now.")
}

// parseSentinelConfig parses environment variables needed for the test suite. See .env.example
func parseSentinelConfig() (SentinelConfig, error) {

	sentinelConfig := SentinelConfig{}

	address := os.Getenv("SENTINEL_ADDRESS")
	if address == "" {
		return SentinelConfig{}, errors.New("SENTINEL_ADDRESS not found in environment variables")
	}
	sentinelConfig.Address = address

	password := os.Getenv("SENTINEL_PASSWORD")
	if password == "" {
		return SentinelConfig{}, errors.New("SENTINEL_PASSWORD not found in environment variables")
	}
	sentinelConfig.Password = password

	return sentinelConfig, nil
}
