package main

import (
	"flag"
	"gopkg.in/redis.v3"
	"net/http"
	"encoding/json"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
)

const (
	REDIS_RESP_KEY = "fake_resps"
	FAIL_RESP = "fail"
	SUCC_RESP = "ok"

	PARAM_PATH = "__path"
)

type Server struct {
	RedisClient *redis.Client
	config map[string]string
}

func (s *Server) Start() {
	http.HandleFunc("/fake/query", MakeHandler(s.Query))
	http.HandleFunc("/fake/submit", MakeHandler(s.Submit))


}

// 从redis加载
func (s *Server) InitConfig() {
	result := s.RedisClient.HGetAllMap(REDIS_RESP_KEY)
	if result.Err() != nil && result.Err() != redis.Nil {
		fmt.Println("failed to fetch redis config", result.Err())
		os.Exit(1)
	}

	m := make(map[string]string, 0)
	if result.Err() == redis.Nil {
		s.config = m
	} else {
		s.config = result.Val()
	}
	// 这里打印一个log， 保存一下redis里面的值
	fmt.Println("successfully init config from redis")
}


// 从内存写到redis
func (s *Server) SyncConfig() {

}

type Response struct {
	ReturnCode string `json:"return_code"`
	ReturnMessage string `json:"return_message"`
	Data       interface{}  `json:"data"`
}

func MakeHandler(f func(*http.Request) interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result := f(r)
		data, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write(data)
	}
}

func (s *Server) Query(r *http.Request) interface{} {
	// 返回所有的设置
	result := s.RedisClient.HGetAllMap(REDIS_RESP_KEY)
	if result.Err() != nil {
		return &Response{
			ReturnCode:FAIL_RESP,
			ReturnMessage:result.Err(),
		}
	}

	return &Response{
		ReturnCode: SUCC_RESP,
		Data:result.Val(),
	}
}

func (s *Server) Submit(r *http.Request) interface{} {
	path := r.FormValue(PARAM_PATH)
	if path == "" {
		return &Response{
			ReturnCode:FAIL_RESP,
			ReturnMessage:"缺少"+PARAM_PATH+"参数",
		}
	}
	params := make(map[string]string, 0)
	for key, value := range r.Form {
		if len(value) > 0 && key != PARAM_PATH {
			params[key] = value[0]
		}
	}

	result := s.RedisClient.HGet(REDIS_RESP_KEY, path)
	if result.Err() != nil && result.Err() != redis.Nil {
		return &Response{
			ReturnCode: FAIL_RESP,
			ReturnMessage:result.Err(),
		}
	}

	m := make(map[string]string, 0)
	data := result.Val()
	if data != "" {
		err := json.Unmarshal([]byte(data), &m)
		if err != nil {
			return &Response{
				ReturnCode: FAIL_RESP,
				ReturnMessage: err.Error(),
			}
		}
	}

	key, value := genPair(params)
	m[key] = value

	data, _ = json.Marshal(m)
	boolResult := s.RedisClient.HSet(REDIS_RESP_KEY, path, string(data))
	if boolResult.Err() != nil {
		return &Response{
			ReturnCode: FAIL_RESP,
			ReturnMessage:  boolResult.Err().Error(),
		}
	}
	return &Response{
		ReturnCode:SUCC_RESP,
		ReturnMessage:"ok",
	}
}

func (s *Server) Call(w *http.ResponseWriter, r *http.Request) {

}


// 对于参数就暗处
func genPair(params map[string]string) (string, string) {
	data, _ := json.Marshal(params)
	return base64.StdEncoding.EncodeToString(md5.Sum(data)[:]), string(data)
}

func main() {
	redisHost := flag.String("redis-host", "127.0.0.1:6379", "redis地址: 127.0.0.1:6379")
	client := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

}



