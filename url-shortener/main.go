package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/metrosystems-cpe/GopherLab/url-shortener/storage"
	"github.com/metrosystems-cpe/GopherLab/url-shortener/utils"
)

var (
	// StoreConfig is the backend service config
	StoreConfig storage.Config
)

func init() {
	// in a cloud app read them from ENV
	StoreConfig.Addr = "http://localhost:8080"
	StoreConfig.Set = "/set-key"
	StoreConfig.Get = "/get-key/"
}

func shortHandler(wr http.ResponseWriter, req *http.Request) {

	urls, ok := req.URL.Query()["url"] // Get a copy of the queried value.
	if !ok || len(urls[0]) < 1 {
		http.Error(wr, utils.ReturnError("missing url"), http.StatusBadRequest)
		return
	}

	url, err := url.ParseRequestURI(urls[0])
	if err != nil {
		http.Error(wr, utils.ReturnError("failed to parse URL"), http.StatusBadRequest)
		return
	}

	urlHash := utils.DataHash(url.String())

	ssJSON, err := StoreConfig.NewStorageKey(urlHash, url.String())
	if err != nil {
		log.Printf(err.Error())
		http.Error(wr, utils.ReturnError("Oops... JSONs"), http.StatusInternalServerError)
		return
	}
	log.Printf("%v", string(ssJSON))
	ok, err = StoreConfig.StorageSet(ssJSON)
	if err != nil {
		log.Printf(err.Error())
		http.Error(wr, utils.ReturnError("Oops... could not contact backing service"), http.StatusInternalServerError)
		return
	}

	if ok {
		wr.WriteHeader(http.StatusOK)
		wr.Write(utils.ReturnURL(req.Host + "/r/" + urlHash))
	}
}

func redirectHandler(wr http.ResponseWriter, req *http.Request) {
	// fmt.Println(req.URL.Path)
	p := strings.Split(req.URL.Path, "/")[1:] // get the keys from 1 to n

	if len(p) < 2 {
		http.Error(wr, "missing key", http.StatusNotFound)
		log.Printf("Key not found in url path")
		return
	}
	key := p[1]
	storageData, err := StoreConfig.StorageGet(key)
	if err != nil {
		log.Printf(err.Error())
		http.Error(wr, utils.ReturnError("Oops... Backing services"), http.StatusInternalServerError)
	}
	redirectURL, _ := StoreConfig.DecodeStorageData(storageData)
	if err != nil {
		log.Printf(err.Error())
		http.Error(wr, utils.ReturnError("Oops... url not in our DB"), http.StatusBadRequest)
	}

	http.Redirect(wr, req, redirectURL, http.StatusMovedPermanently)
}

func healthHandler(wr http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		wr.Header().Set("Access-Control-Allow-Origin", origin)
	}
	wr.Header().Set("Access-Control-Allow-Methods", "GET")
	wr.Header().Set("Content-Type", "application/json")
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	wr.WriteHeader(http.StatusOK)
	wr.Write([]byte(`{"Status": "OK"}`))
	// io.WriteString(wr, `{"Status": OK}`)
}

func main() {

	var addr = flag.String("addr", ":8081", "The addr of the application.")

	// we will use default go package: https://golang.org/pkg/net/http/#ServeMux
	// and crete a ne multiplexer http server
	mux := http.NewServeMux()

	// root handler will redirect user to www folder and act as an http file server
	mux.Handle("/", http.FileServer(http.Dir("./www")))

	// metrics handler
	mux.Handle("/metrics", utils.OCPrometheusExporter())

	// shortener handler will not allow child routes as it does not have a final '/'
	// wrapped by a middleware that measures response times
	mux.HandleFunc("/s", utils.WithMetrics(shortHandler))
	// mux.HandleFunc("/s", shortHandler)

	// redirect handler will allow child routes as it does have a final '/'
	mux.HandleFunc("/r/", redirectHandler)

	// health checks handlers - dummy implementation
	// https://reliability.metrosystems.net/docs/dev-best-practices/kubernetes/health-checks/
	mux.HandleFunc("/.well-known/live", healthHandler)
	mux.HandleFunc("/.well-known/ready", healthHandler)

	// start the http server
	log.Println("Starting application on", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
