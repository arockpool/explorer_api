package main

import (
	// "fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

// IP returns request client ip.
// if in proxy, return first proxy id.
// if error, return 127.0.0.1.
func GetIp(c *gin.Context) string {
	ips := Proxy(c)
	if len(ips) > 0 && ips[0] != "" {
		rip := strings.Split(ips[0], ":")
		return rip[0]
	}
	ip := strings.Split(c.Request.RemoteAddr, ":")
	if len(ip) > 0 {
		if ip[0] != "[" {
			return ip[0]
		}
	}
	return "127.0.0.1"
}

// Proxy returns proxy client ips slice.
func Proxy(c *gin.Context) []string {
	// fmt.Println("X-Forwarded-For-->", c.Request.Header.Get("X-Forwarded-For"))
	// fmt.Println("c.Request.RemoteAddr-->", c.Request.RemoteAddr)
	// fmt.Println("X-Real-IP-->", c.Request.Header.Get("X-Real-IP"))
	if ips := c.Request.Header.Get("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	return []string{}
}
