package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

const (
	ServerAddress = "localhost:8080"
	FilesDir      = "./files"
)

var (
	requestCounter int
	wordSearches   map[string]int
	wordSearchesMu sync.Mutex
	words          []string
)

func incrementRequestCounter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCounter++
		next.ServeHTTP(w, r)
	})
}

type WordCount struct {
	Word          string `json:"word"`
	TF            int    `json:"tf"`
	DF            int    `json:"df"`
	LastTF        int    `json:"last_tf"`
	LastDF        int    `json:"last_df"`
	TotalSearches int    `json:"total_searches"`
}

type WordCountMap map[string]WordCount

func init() {
	wordSearches = make(map[string]int)
}



func main() {
	router := setupRouter()
	err := http.ListenAndServe(ServerAddress, router)
	if err != nil {
		panic(err)
	}
}

func setupRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(incrementRequestCounter)
	router.HandleFunc("/search", searchHandler).Methods("POST")
	router.HandleFunc("/", wellcome).Methods("GET")

	return router
}

func wellcome(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode("wellcome")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func searchHandler(w http.ResponseWriter, r *http.Request) {
	err := json.NewDecoder(r.Body).Decode(&words)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wordCounts := make(WordCountMap)
	files, err := ioutil.ReadDir(FilesDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var wg sync.WaitGroup
	for _, word := range words {
		if val, ok := wordSearches[word]; ok {
			wordSearches[word] = val + 1
		} else {
			wordSearches[word] = 1
		}
		wg.Add(1)
		go func(word string) {
			defer wg.Done()
			wordCounts[word] = countWord(word, files)
		}(word)
	}

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(wordCounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func countWord(word string, files []os.FileInfo) WordCount {
	var wordCount WordCount

	for _, file := range files {
		if !file.IsDir() {
			content, err := ioutil.ReadFile(filepath.Join(FilesDir, file.Name()))
			if err != nil {
				continue
			}
			wc := countWordInFile(string(content), word)
			wordCount.TF += wc.TF
			if wc.TF > 0 {
				wordCount.DF++
			}
		}
	}

	wordCount.Word = word
	wordCount.LastTF = wordCount.TF
	wordCount.LastDF = wordCount.DF

	// Update total searches
	wordSearchesMu.Lock()
	wordCount.TotalSearches = wordSearches[word]
	//wordSearches[word]++
	wordSearchesMu.Unlock()

	return wordCount
}

func countWordInFile(content string, word string) WordCount {
	var wordCount WordCount

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		words := strings.Fields(line)
		for _, w := range words {
			if strings.EqualFold(w, word) {
				wordCount.TF++
				wordCount.LastTF++
			}
		}
	}

	wordCount.Word = word
	wordCount.DF = 1
	wordCount.LastDF = 1

	// Update total searches
	wordSearchesMu.Lock()
	wordCount.TotalSearches = wordSearches[word]
	//wordSearches[word]++
	wordSearchesMu.Unlock()

	return wordCount
}
