package main

import (
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var tpl = template.Must(template.ParseFiles("index.html"))

type Items struct {
	Title       string `json:"title"`
	HTMLTitle   string `json:"htmlTitle"`
	Link        string `json:"link"`
	DisplayLink string `json:"displayLink"`
	Snippet     string `json:"snippet"`
	HTMLSnippet string `json:"htmlSnippet"`
	Mime        string `json:"mime"`
	FileFormat  string `json:"fileFormat"`
	Pagemap     struct {
		CseThumbnail []struct {
			Src    string `json:"src"`
			Width  string `json:"width"`
			Height string `json:"height"`
		} `json:"cse_thumbnail"`
	} `json:"pagemap,omitempty"`
}

type Results struct {
	Queries struct {
		Request []struct {
			TotalResults string `json:"totalResults"`
			Count        int    `json:"count"`
			StartIndex   int    `json:"startIndex"`
		} `json:"request"`
		NextPage []struct {
			Count      int `json:"count"`
			StartIndex int `json:"startIndex"`
		} `json:"nextPage"`
		PreviousPage []struct {
			Count      int `json:"count"`
			StartIndex int `json:"startIndex"`
		} `json:"previousPage"`
	} `json:"queries"`
	Items []Items `json:"items"`
}

type Search struct {
	SearchKey    string
	TotalPages   int
	TotalResults int
	Start        int
	Results      Results
}
type Cred struct {
	ApiKey string `envconfig:"api_key"`
	Cx     string
}

func (s *Search) CurrentPage() int {
	return s.Results.Queries.Request[0].StartIndex/10 + 1
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.Execute(w, nil)
	if err != nil {
		return
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		return
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	var c Cred
	err := envconfig.Process("goo", &c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		log.Print(err)
		return
	}

	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		log.Print(err)
		return
	}

	search := &Search{}
	params := u.Query()

	search.SearchKey = params.Get("q")

	search.Start, err = strconv.Atoi(params.Get("start"))
	if err != nil {
		search.Start = 0
	}

	endpoint := fmt.Sprintf(
		"https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=query=%s+filetype%%3Apdf&start=%d",
		c.ApiKey,
		c.Cx,
		url.QueryEscape(search.SearchKey),
		search.Start)

	resp, err := http.Get(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		log.Fatal(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		log.Fatal(err)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		log.Fatal(err)
		return
	}

	search.TotalResults, _ = strconv.Atoi(search.Results.Queries.Request[0].TotalResults)

	search.TotalPages = search.TotalResults/10 + 1
	err = tpl.Execute(w, search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		log.Fatal(err)
	}

}
