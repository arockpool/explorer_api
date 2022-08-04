# 选择编译镜像
FROM golang:1.13

# 创建工作目录
WORKDIR /go/src/app

# 安装依赖
RUN go mod init \
&& GOPROXY="https://goproxy.cn" GO111MODULE=on go get -v github.com/Shopify/sarama \
&& GOPROXY="https://goproxy.cn" GO111MODULE=on go get -v github.com/gin-gonic/gin \
&& GOPROXY="https://goproxy.cn" GO111MODULE=on go get -v github.com/garyburd/redigo/redis \
&& GOPROXY="https://goproxy.cn" GO111MODULE=on go get -v github.com/dgrijalva/jwt-go \
&& GOPROXY="https://goproxy.cn" GO111MODULE=on go get -v github.com/bitly/go-simplejson 

# 修改时区
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 复制代码
COPY . .

# 安装依赖并编译
RUN go build -v -o api

EXPOSE 8000

# 启动1
CMD ./api