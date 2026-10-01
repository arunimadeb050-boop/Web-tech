package main
import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var Port = ":7070"
var Limit = 5
var Window = 10 * time.Second

var mu sync.Mutex
var Counts = make(map[string]int)
var WindowStart = make(map[string]time.Time)

func ratelimitmiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		mu.Lock()

		start, exists := WindowStart[ip]
		now := time.Now()

		if !exists || now.Sub(start) > Window {
			WindowStart[ip] = now
			Counts[ip] = 0
		}

		Counts[ip]++
		count := Counts[ip]

		mu.Unlock()

		log.Printf("ip=%s count=%d limit=%d", ip, count, Limit)

		if count > Limit {
			http.Error(w, "too many requests, slow down", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

func pinghandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}

func router(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/ping":
		ratelimitmiddleware(pinghandler)(w, r)
	default:
		http.NotFound(w, r)
	}
}

func main() {
	log.Println("starting server on", Port)
	http.HandleFunc("/", router)

	er := http.ListenAndServe(Port, nil)
	if er != nil {
		log.Fatal("server failed to start ", er)
	}
}

/*for ($i=1; $i -le 8; $i++) {
    Write-Host "Request $i" -ForegroundColor Cyan
    curl.exe -s -i http://127.0.0.1:7070/ping
    Write-Host "----------------------------"
}*/