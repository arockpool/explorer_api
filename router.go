package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"time"
)

func initRouter() *gin.Engine {
	// gin.SetMode(gin.ReleaseMode) //设置线上模式

	router := gin.Default()
	router.Use(CorsMiddleware())
	router.Use(RunTimeMonitorMiddleware())
	router.Use(TokenMiddleware())
	router.Use(CheckIsFrequentlyMiddleware())

	router.POST("*url", MainHandle)
	router.GET("*url", MainHandle)

	return router
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}

func RunTimeMonitorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		duration := time.Since(t).Seconds()
		if duration > 2 {
			var post_data = strings.Replace(c.Request.Header.Get("post_data"), "\"", "\\\"", -1)
			var total_data = fmt.Sprintf("%v----------%v,%v,%v,%v", post_data, c.Request.Header.Get("UserId"), c.Request.Header.Get("AppID"),
				c.Request.Header.Get("AppVersion"), c.Request.Header.Get("DeviceId"))
			var value = fmt.Sprintf("{\"duration\":\"%v\", \"url\":\"%v\", \"ip\":\"%v\", \"post_data\":\"%v\"}", duration,
				c.Param("url"), GetIp(c), total_data)
			log.Println(value)
			SendMq(MQ_TOPIC_SLOW_REQUEST, value)
		}
	}
}

func TokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if needTokenInRequest(c.Param("url")) && c.Request.Method == "POST" {
			tokenString := c.Request.Header.Get("Authorization")
			// fmt.Println("token is:", tokenString)
			if s := strings.Split(tokenString, " "); len(s) == 2 {
				tokenString = s[1]
			}
			code, userId, appID, _ := getInfoFromToken(tokenString)
			if code != 0 {
				c.JSON(http.StatusForbidden, GetErrInfo(code))
				c.Abort()
			} else {
				c.Request.Header.Set("UserId", userId)
				c.Request.Header.Set("AppID", appID)
				// c.Request.Header.Set("AppVersion", appVersion) # 解决appVersion变动问题
			}
		}
		c.Next()
	}
}

func CheckIsFrequentlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var LimitUrlList = map[string][]int{
			"/account/api/": {600, 100}, //10分钟100次
		}

		url := c.Param("url")
		for limitUrl, limitParams := range LimitUrlList {
			if strings.HasPrefix(url, limitUrl) {
				ip := GetIp(c)
				cache_key := fmt.Sprintf("%v_%v", url, ip)
				expires := limitParams[0]
				limit := limitParams[1]

				code := CheckIsFrequently(cache_key, expires, limit)
				// fmt.Println("CheckIsFrequently info is:", cache_key, code, ip)
				if code != 0 {
					c.JSON(http.StatusOK, GetErrInfo(code))
					c.Abort()
					break
				}
			}
		}
		c.Next()
	}
}
