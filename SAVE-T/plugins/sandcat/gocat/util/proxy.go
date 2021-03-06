package util 

import (
	"fmt"
	"net"
	"net/http"
	"time"
	"io/ioutil"
	"bytes"
	"math/rand"
)

//StartProxy creates an HTTP listener to forward traffic to server
func StartProxy(server string) {
	 http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpClient := http.Client{}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		url := server + r.RequestURI

		proxyReq, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		proxyReq.Header = make(http.Header)
		for h, val := range r.Header {
			proxyReq.Header[h] = val
		}
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		bites, _ := ioutil.ReadAll(resp.Body)
		w.Write(bites)
	})
	http.ListenAndServe(":8889", nil)
}

//FindProxy locates a useable host for coms
func FindProxy(masters []string) string {
	shuffle(masters)
	for _, m := range masters {
		connected := testConnection(m)
		if connected {
			proxy := fmt.Sprintf("http://%s", m)
			fmt.Println("Located available proxy server", proxy)
			return proxy
		}
	}
	return ""
}

func testConnection(master string) bool {
	conn, _ := net.DialTimeout("tcp", master, time.Second)
	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}

func shuffle(slice []string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(slice); n > 0; n-- {
	   randIndex := r.Intn(n)
	   slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
	}
 }