package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/bitly/go-simplejson"
	// "github.com/garyburd/redigo/redis"
	"io"
	"math"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func testCh() {
	fmt.Println("Hello from another goroutine")
	fmt.Println("Hello from main goroutine")
	time.Sleep(time.Millisecond)

	ch := make(chan string)
	go func() {
		ch <- "Hello!"
		close(ch)
	}()
	fmt.Println(<-ch)      // 输出字符串"Hello!"
	fmt.Println(<-ch)      // 输出零值 - 空字符串""，不会阻塞
	fmt.Println(<-ch, "1") // 再次打印输出空字符串""
	fmt.Println(<-ch, "2") // 再次打印输出空字符串""
	v, ok := <-ch
	fmt.Println(v, ok)

}

func testJson() {
	js, _ := simplejson.NewJson([]byte(`{
    "code": 0,
        "array": [1, "2", 3],
        "data": {"user_id":{"aa":[1,2,3]}}
    
	}`))

	ms := js.Get("code").MustInt()
	a := js.Get("data").MustMap()["user_id"]

	b := a.(map[string]interface{})
	c := b["aa"].([]interface{})

	fmt.Println(ms, a, b["aa"], c, reflect.ValueOf(c).Kind())

	var d interface{}
	var i int = 5
	s := "Hello world"
	// These are legal statements
	d = i
	fmt.Println(reflect.ValueOf(d).Kind())
	d = s
	fmt.Println(reflect.ValueOf(d).Kind())

	js.Set("msg", 1111)
	js.Get("data").MustMap()["token"] = "1111"
	fmt.Println(js)
}

func testTime() {
	b := time.Now()
	h, _ := time.ParseInLocation("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"), time.UTC)

	fmt.Println(b, b.Unix())
	fmt.Println(h, h.Unix())

}

func testToken() {
	var tp = TokenPayload{"01234567890123456789012345678922", "ch14c0ca80073b4c2c", "1.0.0", 1212}
	token := createToken(tp)
	fmt.Println(token)
	ts := token
	fmt.Println(getInfoFromToken(ts))
	// fmt.Println(needTokenInRequest("/account/api/send_sms"))
}

func testUrl() int {
	sign := "signature=ff247144bf1b71d066dac8d182b1ffbab233a56a&timestamp=1502964452&nonce=1297986261"
	secret := "8fc2fc32f20a405bb7c7c544083d030a"

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

func test() {
	testTime()
	// testToken()
	// testJson()
	// testUrl()
	// testRedis()
	// fmt.Println(CheckIsFrequently("access_count", 600, 3))
}
