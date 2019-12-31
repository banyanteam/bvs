
package main

import (
	"testing"
	"fmt"

	"github.com/go-redis/redis/v7"
)

var redisdb *redis.ClusterClient

func ExampleNewClusterClient() {
	// See http://redis.io/topics/cluster-tutorial for instructions
	// how to setup Redis Cluster.
	redisdb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{":9001", ":9002", ":9003", ":9004", ":9005", ":9006"},
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

	t.Log("ExampleNewClusterClient")
	ExampleNewClusterClient()

	t.Log("ExampleClient")
	ExampleClient()
}
