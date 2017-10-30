package main

import (
	"flag"
	"gopkg.in/redis.v3"
	"net/http"
	"github.com/kusora/dlog"
	"encoding/json"
)

const (
	REDIS_RESP_KEY = "fake_resps"
	FAIL_RESP = "fail"
)



type Server struct {
	RedisClient *redis.Client
}

func (s *Server) Start() {
	http.HandleFunc("/fake/query", s.Query)
	http.HandleFunc("/fake/submit", s.Submit)
}

func (s *Server) Query(w http.ResponseWriter, r *http.Request) {
	// 返回所有的设置
	result := s.RedisClient.HGetAllMap(REDIS_RESP_KEY)
	if result.Err() != nil {
		dlog.Error("failed to get redis config, %+v", result.Err())
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(result.Err().Error()))
		return
	}

	data, _ := json.Marshal(result.Val())
	w.Write(data)
	return
}

func (s *Server) Submit(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
}

func (s *Server) Call(w *http.ResponseWriter, r *http.Request) {

}


func main() {
	redisHost := flag.String("redis-host", "127.0.0.1:6379", "redis地址: 127.0.0.1:6379")
	client := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})


}



