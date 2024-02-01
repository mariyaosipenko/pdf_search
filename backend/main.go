package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/kelseyhightower/envconfig"
)

var tpl = template.Must(template.ParseFiles("index.html"))

const googleURL = "https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=query=%s+filetype%%3Apdf&start=%d"
const defaultPort = "3000"
const linksOnPage = 10

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
		Request      []Request `json:"request"`
		NextPage     []Page    `json:"nextPage"`
		PreviousPage []Page    `json:"previousPage"`
	} `json:"queries"`
	Items []Items `json:"items"`
}

type Page struct {
	Count      int `json:"count"`
	StartIndex int `json:"startIndex"`
}

type Request struct {
	TotalResults string `json:"totalResults"`
	Count        int    `json:"count"`
	StartIndex   int    `json:"startIndex"`
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
	if s.Results.Queries.Request == nil || len(s.Results.Queries.Request) < 1 {
		return 1
	}
	return s.Results.Queries.Request[0].StartIndex/linksOnPage + 1
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.Execute(w, nil)
	if err != nil {
		return
	}
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = defaultPort
	}

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/search", SearchHandler)
	mux.HandleFunc("/", indexHandler)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		return
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {

	var c Cred
	err := envconfig.Process("goo", &c)
	if err != nil {
		handleError(err, w)
	}

	u, err := url.Parse(r.URL.String())
	if err != nil {
		handleError(err, w)
	}

	search := &Search{}
	params := u.Query()

	search.SearchKey = params.Get("q")

	search.Start, err = strconv.Atoi(params.Get("start"))
	if err != nil {
		search.Start = 0
	}

	endpoint := fmt.Sprintf(
		googleURL,
		c.ApiKey,
		c.Cx,
		url.QueryEscape(search.SearchKey),
		search.Start)

	resp, err := http.Get(endpoint)
	if err != nil {
		handleError(err, w)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		handleError(err, w)
	}

	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		handleError(err, w)
	}

	search.TotalResults, _ = strconv.Atoi(search.Results.Queries.Request[0].TotalResults)

	search.TotalPages = search.TotalResults/linksOnPage + 1
	err = tpl.Execute(w, search)
	if err != nil {
		handleError(err, w)
	}

}

func handleError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
	log.Fatal(err)
}
