package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"os"
)

type Cache struct {
	// index int
}

func (self *Cache) GetConnection() (int, redis.Conn) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDb := os.Getenv("REDIS_DB")
	redisConfig := RedisConfig{redisHost, redisPort, redisPassword, redisDb}
	// fmt.Println("redisConfig-->:", redisConfig)

	c, err := redis.Dial("tcp", fmt.Sprintf("%v:%v", redisConfig.host, redisConfig.port))
	if err != nil {
		fmt.Println("Connect to redis error:", err)
		return -1, c
	}

	if redisConfig.password != "" {
		if _, err := c.Do("AUTH", redisConfig.password); err != nil {
			fmt.Println("redis auth failed:", err)
			return -1, c
		}
	}

	if _, err := c.Do("SELECT", redisConfig.index); err != nil {
		fmt.Println("redis select db failed:", err)
		return -1, c
	}
	return 0, c
}

func CheckIsFrequently(cache_key string, expires int, limit int) int {
	cache := Cache{}
	code, c := cache.GetConnection()
	defer c.Close()
	if code < 0 {
		return 80008
	}

	access_count, err := redis.Int(c.Do("GET", cache_key))
	// fmt.Println("access_count is:", access_count)
	if err != nil {
		_, err = c.Do("SET", cache_key, 1, "EX", expires)
	} else {
		// fmt.Printf("Got cache_key %v\n", access_count)
		access_count += 1
		if access_count > limit {
			return 80009
		}
		c.Do("INCRBY", cache_key, 1)
	}
	return 0
}

// 从缓存中获取对应的appsecret
func getAppFromCache(appId string) App {

	app := App{}

	cache := Cache{}
	code, c := cache.GetConnection()
	defer c.Close()
	if code < 0 {
		return app
	}

	appSecret, err := redis.String(c.Do("GET", appId))
	if err == nil {
		app = App{appId, appSecret, "", 2}
	}
	return app
}
