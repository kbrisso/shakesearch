package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/patrickmn/go-cache"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

var c *cache.Cache

func main() {
	//setup cache
	c = cache.New(30*time.Minute, 60*time.Minute)
	ctx, cancel := context.WithCancel(context.Background())

	searcher := Searcher{}

	err := searcher.Load("completeworks.txt")

	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/search", handleSearch(searcher))
    //added some params to server
	srv := &http.Server{
		Addr:         ":3001",
		IdleTimeout:  time.Minute * time.Duration(1),
		ReadTimeout:  time.Second * time.Duration(5),
		WriteTimeout: time.Second * time.Duration(10),
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
	}

    //register context for graceful shut down
	srv.RegisterOnShutdown(cancel)

	//run server
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			//It is fine to use Fatal here because it is not the main gorutine
			fmt.Println("HTTP server ListenAndServe: %v" + err.Error())
		}
	}()

	//sig hook for shutdown
	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	<-signalChan

	fmt.Println("Os.Interrupt - shutting down...")

	go func() {
		<-signalChan
		fmt.Println("Os.Kill - terminating...")
	}()

	gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancelShutdown()

	if err := srv.Shutdown(gracefullCtx); err == nil {
		fmt.Println("Server gracefully stopped")
	} else {
		fmt.Println("Shutdown error: " + err.Error())
		defer os.Exit(1)
		return
	}

	//manually cancel context if not using httpServer.RegisterOnShutdown(cancel)
	cancel()

	defer os.Exit(0)

	return
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		results := searcher.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	return nil
}

func (s *Searcher) Search(query string) []string {
	result := []string{}
	startIdx := 0
	endIdx := 0
	//check cache
	v, found := c.Get(strings.ToLower(query))
    //if found return results
	if found {
		result = v.([]string)
		return result
	}
	//this allows for case insensitive search
	r := regexp.MustCompile("(?i)" + query)
	//find all returns all locations by index - start and stop
	matches := s.SuffixArray.FindAllIndex(r, -1)
	//loop through match indexes
	for _, idxRange := range matches {
		for i, idx := range idxRange {
			//i = 0 is start position - i = 1 stop position , idx is the char sequence of start and stop
			if i == 0 {
				startIdx = idx
			} else {
				endIdx = idx
			}
			//only enter this loop when startIdx and stopIdx are found
			if i == 1 {
				//get founds search terms
				found := r.FindAllString(s.CompleteWorks[startIdx:endIdx], -1)
				//if a list of like terms are found remove the dupes
				list := removeDup(found)
				for _, v := range list {
					re := regexp.MustCompile(v)
					//find exact results by case return 100 chars before - 100 chars after term - add css to highlight term
					result = append(result, re.ReplaceAllString(s.CompleteWorks[(startIdx-100):(endIdx+100)], "<strong class=\"sb\">"+v+"</strong>"))
				}
			}

		}
	}
    //add result to cache if not found
	if !found {
		c.Set(strings.ToLower(query), result, cache.DefaultExpiration)
	}
	return result
}

func removeDup(s []string) []string {
	//removes duplicate search terms
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
