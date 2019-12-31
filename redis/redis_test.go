package main

import (
	"testing"
	"fmt"

	"github.com/go-redis/redis/v7"
)


var redisdb *redis.Client

func ExampleNewClient(){
	redisdb = redis.NewClient(&redis.Options{
		Addr:     "172.20.226.192:6379", // use default Addr
		Password: "Occ2018",        // no password set
		DB:       0,                // use default DB
	})

	if redisdb == nil{
		fmt.Println("redisdb is nil")
		return
	}

	pong, err := redisdb.Ping().Result()
	fmt.Println(pong, err)
}

func ExampleNewFailoverClient() {
	redisdb = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		Password: "Occ2018",
		SentinelAddrs: []string{":26379", ":26380", ":26381"},
	})

	pong, err := redisdb.Ping().Result()
	fmt.Println(pong, err)
}

func ExampleClient() {
	if redisdb == nil{
		fmt.Println("redisdb is nil")
		return
	}

	err := redisdb.Set("h1", "10", 0).Err()
	if err != nil {
		fmt.Println("Set", err)
		return
	}

	val, err := redisdb.Get("h1").Result()
	if err != nil {
		fmt.Println("Get", err)
		return
	}
	fmt.Println("h1", val)

	val2, err := redisdb.Get("missing_key").Result()
	if err == redis.Nil {
		fmt.Println("missing_key does not exist")
	} else if err != nil {
		fmt.Println("Get val2", err)
		return
	} else {
		fmt.Println("missing_key", val2)
	}
}

func TestRedis(t *testing.T){
	// t.Log("ExampleNewClient")
	// ExampleNewClient()
	
	t.Log("ExampleNewFailoverClient")
	ExampleNewFailoverClient()

	t.Log("ExampleClient")
	ExampleClient()
}
