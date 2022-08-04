package main

import (
	"os"
)

const TokenSecret = "ccd9f89ba6074db0a1db2d63ca8844bf"

var ServicesConfig = map[string]string{
	"account":     os.Getenv("SERVER_ACCOUNT"),
	"profile":     os.Getenv("SERVER_PROFILE"),
	"system":      os.Getenv("SERVER_SYSTEM"),
	"activity":    os.Getenv("SERVER_ACTIVITY"),
	"pool":        os.Getenv("SERVER_POOL"),
	"explorer":    os.Getenv("SERVER_EXPLORER"),
	"explorer_v2": os.Getenv("SERVER_EXPLORER_V2"),
	"data":        os.Getenv("SERVER_DATA"),
	"sector":      os.Getenv("SERVER_SECTOR"),
}

var ErrInfo = map[int]string{
	80000: "url未找到匹配服务，请检查url是否正确",
	80001: "服务器接口请求异常",
	80002: "the method of request not support",
	80003: "need AppId and AppVersion in header",
	80004: "invalid appId",
	80005: "invalid signature",
	80006: "请检查系统时间是否正确后重试",
	80007: "无效的token",

	80008: "redis cache error",
	80009: "访问过于频繁，请稍后再试",
}

type App struct {
	appId     string
	appSecret string
	appName   string
	appType   int //0:mobile 1:小程序 2:网页版 3:pc
}

var Apps = map[string]App{
	"h5":    App{"h5b9557f39e2f84d", "80c9772b3e634602a7596969ef617fbf", "h5网页", 2},
	"rmd":   App{"rmd3c8a2b3451214", "0f3c8f7307814b7b9475cda0e45fd1aa", "MINING POOL", 2},
	"fam":   App{"fam6cea0d84f3904", "2e11a2fa348044639a1cc2b8541c068f", "fam", 2},
	"chain": App{"chainb904a05a418", "fde60d67a1c042d59a6281656e8f8422", "链服务器", 2},
	"gpu":   App{"gpu13c7e53952b54", "d825e01626c8459c8dea45a2166cc10e", "云GPU小程序", 1},
	"llq":   App{"llq59184577f4c34", "b91153d8252f4af1b39227338b8932d8", "浏览器小程序", 1},
	"stc":   App{"stc542e2c61df554", "e28743837c824a1fad05686946d09105", "MTXSTORAGE企业版", 1},
	"bhpay": App{"bhpay86a66ece077", "736b095b501a425fa06a9c53d85515da", "bhpay", 1},
	"pl":    App{"pl0fbebc29bd5646", "cf814b9e16ca4f5a907e1c5431164184", "pl", 1},
	"cps":   App{"cps76572f0101414", "e5a96de8083144a5be7e4882c611a0ee", "cps", 1},
}

type RedisConfig struct {
	host     string
	port     string
	password string
	index    string
}
