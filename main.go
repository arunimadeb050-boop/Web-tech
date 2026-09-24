package main

import(
	"encoding/json"
	"net/http"
	"log"
	"math/rand"
	"strings"
	"sync"
)

var store = make(map[string]string)
var mu sync.Mutex

var Port = ":8080"
var Baseurl = "http://localhost:8080"
var Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Shortenrequest struct{
	Url string `json:"url"`
}

type Shortenresponse struct{
	Code string `json:"code"`
	ShortUrl string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

func randomcode(n int) string{
	b := make([]byte, n)
	for i := range b{
		b[i] = Charset[rand.Intn(len(Charset))]
	}
	return string(b)
}

func shortenhandler (w http.ResponseWriter, r *http.Request){

	if r.Method!=http.MethodPost{
		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return
	}

	var req Shortenrequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err!=nil{
		http.Error(w,"Invalid Request", http.StatusBadRequest)
		return
	}

	req.Url = strings.TrimSpace(req.Url)

	if !strings.HasPrefix(req.Url,"http://") && !strings.HasPrefix(req.Url,"https://"){
		http.Error(w,"url must start with http:// or https://",http.StatusBadRequest)
		return
	}

	mu.Lock()

	var code string
	for{
		code = randomcode(6)
		_, taken := store[code]
		if !taken{
			break
		}
	}
	store[code] = req.Url

	mu.Unlock()

	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(Shortenresponse{Code : code, ShortUrl : Baseurl+"/"+code, OriginalUrl : req.Url})

}

func redirecthandler (w http.ResponseWriter, r *http.Request){

	if r.Method!=http.MethodGet{
		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return
	}

	code := strings.TrimPrefix(r.URL.Path,"/")

	if code==""{
		http.Error(w,"no code given",http.StatusBadRequest)
		return
	}

	mu.Lock()
	target, found := store[code]
	mu.Unlock()

	if !found{
		http.Error(w,"short link not found",http.StatusNotFound)
		return
	}

	http.Redirect(w,r,target,http.StatusFound)

}

func healthhandler (w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(map[string]string{"status":"ok"})
}

func router (w http.ResponseWriter, r *http.Request){
	switch{
	case r.URL.Path=="/health":
		healthhandler(w,r)
	case r.URL.Path=="/shorten":
		shortenhandler(w,r)
	default:
		redirecthandler(w,r)
	}
}

func main(){
	http.HandleFunc("/", router)

	er := http.ListenAndServe(Port,nil)
	if er!=nil{
		log.Fatal("server failed to start ", er)
	}
}


//curl.exe -X POST localhost:8080/shorten -d '{\"url\":\"https://go.dev\"}' -H "Content-Type: application/json"

//curl.exe -iL localhost:8080/FiTp7M