package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	for {
		archive()
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
		log.Printf("next archive at %s\n", next.Format("2006-01-02 15:04:05"))
		time.Sleep(next.Sub(now))
	}
}

func archive() {
	log.Printf("start archive at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/search/square?limit=10", nil)
	if err != nil {
		log.Printf("new request failed, err:%v\n", err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.61 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("get http reponse failed, err:%v\n", err)
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read body failed, err:%v\n", err)
		return
	}

	var body body
	err = json.Unmarshal(b, &body)
	if err != nil {
		log.Printf("unmarshal body failed, err:%v\n", err)
		return
	}

	os.MkdirAll("archives/bilibili", 0755)
	name := fmt.Sprintf("archives/bilibili/%s.txt", time.Now().Format("2006-01-02"))
	b, err = os.ReadFile(name)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("read archive file failed, err:%v\n", err)
			return
		}
	}

	var words []string
	if len(b) > 0 {
		words = strings.Split(string(b), "\r\n")
	}

	n := 0
	for _, keyword := range body.Data.Trending.List {
		word := keyword.Keyword
		word = strings.TrimSpace(word)
		word = strings.ReplaceAll(word, "\r\n", "")
		has := false
		for _, w := range words {
			if w == word {
				has = true
				break
			}
		}
		if !has {
			words = append(words, word)
			n++
		}
	}

	err = os.WriteFile(name, []byte(strings.Join(words, "\r\n")), 0755)
	if err != nil {
		log.Printf("write archive file failed, err:%v\n", err)
		return
	}

	log.Printf("archived %d new words\n", n)
	log.Printf("finish archive at %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

type body struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	Data    struct {
		Trending struct {
			Title   string `json:"title,omitempty"`
			Trackid string `json:"trackid,omitempty"`
			List    []struct {
				Keyword  string `json:"keyword,omitempty"`
				ShowName string `json:"show_name,omitempty"`
				Icon     string `json:"icon,omitempty"`
				URI      string `json:"uri,omitempty"`
				Goto     string `json:"goto,omitempty"`
			} `json:"list,omitempty"`
		} `json:"trending,omitempty"`
	} `json:"data,omitempty"`
}
