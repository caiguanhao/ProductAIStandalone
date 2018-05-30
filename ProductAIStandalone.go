package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	netUrl "net/url"
	"os"
	"strings"
	"time"
)

var (
	listenAddr  string
	serviceId   string
	accessKeyId string
	urlPrefix   string
)

type (
	searchResponse struct {
		DetectedObjs []struct {
			Loc []float64 `json:"loc"`
		} `json:"detected_objs"`
		Results []struct {
			MetaData string  `json:"metadata"`
			Score    float64 `json:"score"`
			URL      string  `json:"url"`
		} `json:"results"`
		Time string   `json:"time"`
		Type []string `json:"type"`
	}

	processedSearchResponseResult struct {
		ImageUrl string `json:"image_url"`
		HtmlUrl  string `json:"html_url"`
	}

	processedSearchResponse struct {
		Coordinates [][]float64                     `json:"coordinates"`
		Results     []processedSearchResponseResult `json:"results"`
	}

	responseWriter struct {
		http.ResponseWriter
		statusCode int
	}
)

func (w *responseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func search(w http.ResponseWriter, url string, coords []string) error {
	v := netUrl.Values{}
	v.Set("ret_detected_objs", "1")
	v.Set("url", url)
	if len(coords) == 4 {
		v.Add("loc", strings.Join(coords, "-"))
	}
	req, err := http.NewRequest("POST", "https://api.productai.cn/search/"+serviceId, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CA-Version", "1.0")
	req.Header.Set("X-CA-AccessKeyId", accessKeyId)
	client := http.Client{
		Timeout: time.Duration(3 * time.Second),
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	var response searchResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	resp.Body.Close()
	if err != nil {
		return err
	}

	var processed processedSearchResponse
	for _, result := range response.DetectedObjs {
		processed.Coordinates = append(processed.Coordinates, result.Loc)
	}
	for _, result := range response.Results {
		processed.Results = append(processed.Results, processedSearchResponseResult{
			ImageUrl: result.URL,
			HtmlUrl:  urlPrefix + result.MetaData,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(&processed)
}

func errorJson(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"message":"%s"}`, msg)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		url := strings.TrimSpace(r.PostFormValue("url"))
		if url == "" {
			errorJson(w, "Please provide url", http.StatusBadRequest)
			return
		}
		if err := search(w, url, r.PostForm["coords[]"]); err != nil {
			errorJson(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	errorJson(w, "Please use POST", http.StatusMethodNotAllowed)
}

func log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rW := &responseWriter{w, http.StatusOK}
		handler.ServeHTTP(rW, r)
		queries, _ := json.Marshal(r.PostForm)
		end := time.Now()
		fmt.Fprintf(os.Stderr,
			"%s [%s] [%s] %d %s %s %s\n",
			end.Format(time.RFC3339),
			realIp(r),
			end.Sub(start),
			rW.statusCode,
			r.Method,
			r.URL.String(),
			queries,
		)
	})
}

func realIp(r *http.Request) string {
	return r.Header.Get("X-Real-Ip")
}

func init() {
	flag.StringVar(&listenAddr, "listen", "127.0.0.1:8080", "Listen To Address")
	flag.StringVar(&serviceId, "service-id", "", "Service ID")
	flag.StringVar(&accessKeyId, "access-key-id", "", "Access Key ID")
	flag.StringVar(&urlPrefix, "url-prefix", "", "URL Prefix")
}

func main() {
	flag.Parse()

	http.HandleFunc("/SearchImageByURL", handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		errorJson(w, "No such route", http.StatusNotFound)
	})
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Fprintf(os.Stderr, "Started listening on %s\n", listenAddr)
		fmt.Fprintln(os.Stderr, http.Serve(listener, log(http.DefaultServeMux)))
	}
	os.Exit(1)
}
