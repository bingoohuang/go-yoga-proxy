package main

import (
	"flag"
	"fmt"
	"github.com/bingoohuang/go-utils"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	contextPath string
	port        string
	redisServer RedisServer
)

func init() {
	contextPathArg := flag.String("contextPath", "", "context path")
	redisAddrArg := flag.String("redisAddr", "127.0.0.1:6379", "context path")
	portArg := flag.Int("port", 8082, "Port to serve.")

	flag.Parse()

	contextPath = *contextPathArg
	if contextPath != "" && strings.Index(contextPath, "/") < 0 {
		contextPath = "/" + contextPath
	}

	port = strconv.Itoa(*portArg)
	redisServer = parseServerItem(*redisAddrArg)
}

func main() {
	r := mux.NewRouter()

	handleFunc(r, "/clearCache", clearCache, false)
	handleFunc(r, "/getCache", getCache, false)
	handleFunc(r, "/setCache", setCache, false)
	http.Handle("/", r)

	fmt.Println("start to listen at ", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleFunc(r *mux.Router, path string, f func(http.ResponseWriter, *http.Request), requiredGzip bool) {
	wrap := go_utils.DumpRequest(f)

	if requiredGzip {
		wrap = go_utils.GzipHandlerFunc(wrap)
	}

	r.HandleFunc(contextPath+path, wrap)
}

func setCache(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.FormValue("key"))
	value := strings.TrimSpace(r.FormValue("value"))
	ttl := strings.TrimSpace(r.FormValue("ttl"))

	val, err := setRedisContent(key, value, ttl)
	if err != nil {
		http.Error(w, err.Error(), 405)
		return
	}
	w.Write([]byte(val))
}

func getCache(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.FormValue("key"))

	val, err := getRedisContent(key)
	if err == redis.Nil {
		w.Write([]byte(""))
		return
	}

	if err != nil {
		http.Error(w, err.Error(), 405)
		return
	}

	w.Write([]byte(val))
}

func clearCache(w http.ResponseWriter, r *http.Request) {
	keys := strings.TrimSpace(r.FormValue("keys"))
	log.Println("clear cache for keys:", keys)

	result, err := deleteMultiKeys(strings.Split(keys, ","))
	if err != nil {
		http.Error(w, err.Error(), 405)
		return
	}

	w.Write([]byte(strconv.FormatInt(result, 10)))
}

type RedisServer struct {
	Addr      string
	Password  string
	DefaultDb int
}

func splitTrim(str, sep string) []string {
	subs := strings.Split(str, sep)
	ret := make([]string, 0)
	for i, v := range subs {
		v := strings.TrimSpace(v)
		if len(subs[i]) > 0 {
			ret = append(ret, v)
		}
	}

	return ret
}

// password2/localhost:6388/0

func parseServerItem(serverConfig string) RedisServer {
	serverItems := splitTrim(serverConfig, "/")
	len := len(serverItems)
	if len == 1 {
		return RedisServer{
			Addr:      serverItems[0],
			Password:  "",
			DefaultDb: 0,
		}
	} else if len == 2 {
		dbIndex, _ := strconv.Atoi(serverItems[1])
		return RedisServer{
			Addr:      serverItems[0],
			Password:  "",
			DefaultDb: dbIndex,
		}
	} else if len == 3 {
		dbIndex, _ := strconv.Atoi(serverItems[2])
		return RedisServer{
			Addr:      serverItems[1],
			Password:  serverItems[0],
			DefaultDb: dbIndex,
		}
	} else {
		panic("invalid servers argument")
	}
}

func newRedisClient(server RedisServer) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     server.Addr,
		Password: server.Password,  // no password set
		DB:       server.DefaultDb, // use default DB
	})
}

func getRedisContent(key string) (string, error) {
	client := newRedisClient(redisServer)
	defer client.Close()

	return client.Get(key).Result()
}

func setRedisContent(key, value, ttl string) (string, error) {
	client := newRedisClient(redisServer)
	defer client.Close()

	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return "", err
	}

	return client.Set(key, value, duration).Result()
}

func deleteMultiKeys(keys []string) (int64, error) {
	client := newRedisClient(redisServer)
	defer client.Close()

	return client.Del(keys...).Result()
}
