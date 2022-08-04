package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TokenPayload struct {
	userId     string
	appID      string
	appVersion string
	iat        int64 //签发时间戳
}

func GetErrInfo(errCode int) gin.H {
	return gin.H{"code": errCode, "msg": ErrInfo[errCode], "data": make(map[string]string)}
}

func needTokenInResp(url string) bool {
	// 根据url判断是否需要返回token，注册和登陆接口需要返回token
	var nrurlList = []string{
		"/account/api/register",
		"/account/api/login",
		"/account/api/get_user_info_by_miner_no",
		"/account/api/login_with_vercode",
		"/pool/api/machine/get_order_detail",
		"/pool/api/company/authentication_user",
		"/account/api/sms_login",
		"/account/api/sms_login_explorer",
	}

	for _, nrurl := range nrurlList {
		if strings.Contains(url, nrurl) {
			return true
		}
	}
	return false
}

func needTokenInRequest(url string) bool {
	// 根据url判断是否需要token，非匿名访问接口均需要传入token
	var notNeedList = []string{
		"/account/",

		"/profile/admin/",
		"/profile/api/get_user_profile_by_mobile",

		"/system/",

		"/activity/",

		"/pool/api/miner/",
		"/pool/api/machine/",
		"/pool/api/browser/",
		"/pool/api/cluster/",
		"/pool/api/company/v3/get_system_by_uuid",
		"/pool/admin/",

		"/explorer/",
		"/explorer_v2/",

		"/data/",

		"/sector/",
	}
	var needList = []string{
		"/account/api/change_password",
		"/account/api/get_change_password_vercode",
		"/account/api/set_mobile",
		"/pool/admin/machine/v3/add_order",
		// "/activity/api/calculator/get_calculate_sum",
		// "/activity/api/calculator/get_calculate_detail",
		"/explorer/api/pro/",
		"/explorer/api/admin/",
	}

	var flag = true

	if is_transparent_url(url) {
		flag = false
	} else {
		for _, nnu := range notNeedList {
			if strings.HasPrefix(url, nnu) {
				flag = false
			}
		}
	}

	for _, nnu := range needList {
		if strings.HasPrefix(url, nnu) {
			flag = true
		}
	}
	return flag
}

func createToken(tp TokenPayload) string {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":     tp.userId,
		"appID":      tp.appID,
		"appVersion": tp.appVersion,
		"iat":        tp.iat,
	})

	tokenString, err := token.SignedString([]byte(TokenSecret))
	if err != nil {
		fmt.Println("get token error, token is:", tokenString, err)
	}
	return tokenString
}

func getInfoFromToken(ts string) (int, string, string, string) {

	token, err := jwt.Parse(ts, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(TokenSecret), nil
	})
	if err != nil {
		fmt.Println("get token err is:", err)
	} else {
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userId := claims["userId"].(string)
			appID := claims["appID"].(string)
			appVersion := claims["appVersion"].(string)
			if userId != "" {
				return 0, userId, appID, appVersion
			}
		} else {
			fmt.Println("get token err is:", err)
		}
	}

	return 80007, "", "", ""
}

func checkSign(sign string, secret string) int {
	// 验证签名
	rawQuery, err := url.ParseQuery(sign)
	if err != nil {
		return 80005
	}

	nonces := rawQuery["nonce"]
	signatures := rawQuery["signature"]
	timestamps := rawQuery["timestamp"]

	if len(nonces) == 0 || len(signatures) == 0 || len(timestamps) == 0 {
		return 80005
	}
	nonce := nonces[0]
	signature := signatures[0]
	timestamp := timestamps[0]

	timestampInt, _ := strconv.ParseInt(timestamp, 0, 64)

	//时间戳允许十分钟的缓冲期
	if math.Abs(float64(time.Now().Unix()-timestampInt)) > 600 {
		return 80006
	}
	paramList := []string{secret, nonce, timestamp}
	sort.Sort(sort.StringSlice(paramList))
	params := strings.Join(paramList, "")

	h := sha1.New()
	io.WriteString(h, params)
	signature_f := fmt.Sprintf("%x", h.Sum(nil))
	if signature_f != signature {
		return 80005
	}
	return 0
}

func get_app_by_id(appId string) App {
	// 通过appId获取App信息
	for _, appInfo := range Apps {
		if appInfo.appId == appId {
			return appInfo
		}
	}
	// 查询缓存
	return getAppFromCache("APP_" + appId)
}

func is_transparent_url(url string) bool {
	// 特殊透传
	var transparentUrls = []string{
		"/parking/api/order/alipaynotify",
	}

	var flag = false
	for _, tu := range transparentUrls {
		if strings.HasPrefix(url, tu) {
			flag = true
		}
	}
	return flag
	// return strings.HasPrefix(url, "/mall/api/alipaynotify") || strings.HasPrefix(url, "/mall/api/weixinpaynotify")
}

func httpPost(urlIn string, c *gin.Context) {
	req := c.Request
	result, _ := ioutil.ReadAll(req.Body)
	data := bytes.NewBuffer(result).String()
	c.Request.Header.Set("post_data", data) // 给runtime中间件用

	appId := c.Request.Header.Get("AppId")
	appVersion := c.Request.Header.Get("AppVersion")

	if !is_transparent_url(c.Param("url")) {
		if appId == "" || appVersion == "" {
			c.JSON(http.StatusOK, GetErrInfo(80003))
			return
		}
		app := get_app_by_id(appId)
		if app.appId == "" {
			c.JSON(http.StatusOK, GetErrInfo(80004))
			return
		}

		code := checkSign(c.Request.Header.Get("Signature"), app.appSecret)
		fmt.Println("checkSign code is:", code)
		if code > 0 && os.Getenv("DEVCODE") == "prod" {
			c.JSON(http.StatusOK, GetErrInfo(code))
			return
		}
	}

	request, err := http.NewRequest("POST", urlIn, strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println("user_id is", c.Request.Header.Get("UserId"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("UserId", c.Request.Header.Get("UserId"))
	request.Header.Set("AppID", c.Request.Header.Get("AppID"))
	request.Header.Set("AppVersion", c.Request.Header.Get("AppVersion"))
	request.Header.Set("DeviceId", c.Request.Header.Get("DeviceId"))
	request.Header.Set("REAL-IP-FROM-API", GetIp(c))
	request.Header.Set("Lang", c.Request.Header.Get("Lang"))

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, GetErrInfo(80001))
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("err is:", err)
	}
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		fmt.Println("post data is:", data, "\nreturn data is:", string(body),
			"\nuser_id:", c.Request.Header.Get("UserId"), "app_id:", c.Request.Header.Get("AppID"), "app_version:", c.Request.Header.Get("AppVersion"),
			"device_id:", c.Request.Header.Get("DeviceId"), time.Now().Format("2006-01-02 15:04:05"), c.Param("url"), "lang:", c.Request.Header.Get("Lang"))
	} else {
		fmt.Println("post data is:", data,
			"\nuser_id:", c.Request.Header.Get("UserId"), "app_id:", c.Request.Header.Get("AppID"), "app_version:", c.Request.Header.Get("AppVersion"),
			"device_id:", c.Request.Header.Get("DeviceId"), time.Now().Format("2006-01-02 15:04:05"), c.Param("url"), "lang:", c.Request.Header.Get("Lang"))
	}

	// fmt.Println("return data is:", string(body))

	// 设置token
	if needTokenInResp(urlIn) {
		jsonData, err := simplejson.NewJson(body)
		if err == nil {
			if jsonData.Get("code").MustInt() == 0 {
				userIdObj := jsonData.Get("data").MustMap()["user_id"]
				if userIdObj != nil {
					userId := userIdObj.(string)
					iat := time.Now().Unix()
					var tp = TokenPayload{userId, appId, appVersion, iat}
					token := createToken(tp)
					jsonData.Get("data").MustMap()["token"] = token
					c.JSON(http.StatusOK, jsonData) //token放在body的data字典中返回
					return
				}
			}
		}
	}
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)

}

func MainHandle(c *gin.Context) {
	// fmt.Println("header is:", c.Request.Header)
	// fmt.Println("ip is:", c.Request.RemoteAddr)

	url := c.Param("url")
	serviceDomain := ServicesConfig[strings.SplitN(url, "/", 3)[1]]
	if serviceDomain == "" {
		c.JSON(http.StatusNotFound, GetErrInfo(80000))
		return
	}

	urlIn := serviceDomain + url
	if c.Request.Method == "POST" {
		httpPost(urlIn, c)
	} else if c.Request.Method == "GET" {
		if is_transparent_url(url) {
			httpPost(urlIn, c)
		} else {
			c.JSON(http.StatusOK, GetErrInfo(80002))
		}
	}
}
