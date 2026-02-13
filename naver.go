package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	fetchTimeoutMs        = 8000
	autocompleteBaseURL    = "https://ac.search.naver.com/nx/ac"
	blogRateLimitMs       = 350
	blogRetryMax          = 3
	blogRetryBaseMs       = 500
	searchAdChunkSize     = 5
	searchAdRateLimitMs   = 350
	searchAdRetryMax      = 3
	searchAdRetryBaseMs   = 600
	searchAdKeywordCacheMs = 5 * 60 * 1000
	searchAdListRecordSize = 200
)

var (
	lastBlogRequestAt        int64
	lastSearchAdRequestAt    int64
	cachedSearchAdNccKeywords []map[string]interface{}
	cachedSearchAdNccKeywordsAt int64
	rateLimitMutex           sync.Mutex
	cacheMutex              sync.Mutex
)

func waitForBlogRateLimit() {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()

	now := time.Now().UnixMilli()
	wait := lastBlogRequestAt + blogRateLimitMs - now
	if wait > 0 {
		time.Sleep(time.Duration(wait) * time.Millisecond)
	}
	lastBlogRequestAt = time.Now().UnixMilli()
}

func waitForSearchAdRateLimit() {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()

	now := time.Now().UnixMilli()
	wait := lastSearchAdRequestAt + searchAdRateLimitMs - now
	if wait > 0 {
		time.Sleep(time.Duration(wait) * time.Millisecond)
	}
	lastSearchAdRequestAt = time.Now().UnixMilli()
}

func createSignature(method, apiPath, timestamp, secretKey string) string {
	message := fmt.Sprintf("%s.%s.%s", timestamp, method, apiPath)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func createSignatureWithQuery(method, apiPath, timestamp, secretKey, query string) string {
	message := fmt.Sprintf("%s.%s.%s?%s", timestamp, method, apiPath, query)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func buildSearchAdQueries(keywords []string) []string {
	if len(keywords) == 0 {
		return []string{}
	}

	encodeStrict := func(kw string) string {
		return url.QueryEscape(kw)
	}
	encodePlus := func(kw string) string {
		return strings.ReplaceAll(url.QueryEscape(kw), "%20", "+")
	}

	var variants []string
	strictJoined := strings.Join(mapSlice(keywords, encodeStrict), ",")
	plusJoined := strings.Join(mapSlice(keywords, encodePlus), ",")

	variants = append(variants, fmt.Sprintf("hintKeywords=%s&showDetail=1", strictJoined))
	if strictJoined != plusJoined {
		variants = append(variants, fmt.Sprintf("hintKeywords=%s&showDetail=1", plusJoined))
	}

	return removeDuplicates(variants)
}

func mapSlice[T any](slice []T, fn func(T) string) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func fetchWithTimeout(client *http.Client, req *http.Request, timeoutMs int) (*http.Response, error) {
	// 원본 요청의 컨텍스트를 사용하되, 타임아웃만 추가
	// 응답을 받은 후에는 body를 읽는 동안 컨텍스트가 취소되지 않도록
	// 원본 컨텍스트를 사용 (요청이 취소되면 요청 자체가 취소됨)
	parentCtx := req.Context()
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(parentCtx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel() // 함수가 반환된 후에 cancel 호출 (응답을 받은 후 body 읽기에는 영향 없음)

	req = req.WithContext(ctx)
	return client.Do(req)
}

func fetchNaverAutocomplete(keyword string, limit int) ([]string, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []string{}, nil
	}

	params := url.Values{}
	params.Set("q", keyword)
	params.Set("r_format", "json")
	params.Set("st", "0")

	reqURL := autocompleteBaseURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: time.Duration(fetchTimeoutMs) * time.Millisecond}
	resp, err := fetchWithTimeout(client, req, fetchTimeoutMs)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("autocomplete error: %d %s", resp.StatusCode, string(body))
	}

	var data struct {
		Items [][]interface{} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var suggestions []string
	if len(data.Items) > 0 && len(data.Items[0]) > 0 {
		for _, item := range data.Items[0] {
			if itemArray, ok := item.([]interface{}); ok && len(itemArray) > 0 {
				if keywordStr, ok := itemArray[0].(string); ok && keywordStr != "" {
					suggestions = append(suggestions, keywordStr)
				}
			}
		}
	}

	unique := removeDuplicates(suggestions)
	if limit > 0 && len(unique) > limit {
		return unique[:limit], nil
	}
	return unique, nil
}

func fetchBlogCount(config Config, keyword string) (map[string]int, error) {
	if config.Naver.ClientID == "" || config.Naver.ClientSecret == "" {
		return map[string]int{"total": 0}, nil
	}

	reqURL := fmt.Sprintf("https://openapi.naver.com/v1/search/blog.json?query=%s&display=1", url.QueryEscape(keyword))

	var lastErr error
	for attempt := 0; attempt <= blogRetryMax; attempt++ {
		waitForBlogRateLimit()

		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			lastErr = err
			if attempt < blogRetryMax {
				time.Sleep(time.Duration(blogRetryBaseMs*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, err
		}

		req.Header.Set("X-Naver-Client-Id", config.Naver.ClientID)
		req.Header.Set("X-Naver-Client-Secret", config.Naver.ClientSecret)

		client := &http.Client{Timeout: time.Duration(fetchTimeoutMs) * time.Millisecond}
		resp, err := fetchWithTimeout(client, req, fetchTimeoutMs)
		if err != nil {
			lastErr = err
			if attempt < blogRetryMax {
				time.Sleep(time.Duration(blogRetryBaseMs*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, fmt.Errorf("search API timeout: %w", err)
		}

		if resp.StatusCode == 429 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if attempt < blogRetryMax {
				time.Sleep(time.Duration(blogRetryBaseMs*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, fmt.Errorf("search API error: %d %s", resp.StatusCode, string(body))
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("search API error: %d %s", resp.StatusCode, string(body))
			if attempt < blogRetryMax {
				time.Sleep(time.Duration(blogRetryBaseMs*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, lastErr
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// 네이버 검색 API 원본 응답 로그 출력
		log.Printf("=== Naver Search API 원본 응답 ===\nKeyword: %s\nResponse Body: %s\n", keyword, string(body))

		var data struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, err
		}

		log.Printf("=== Naver Search API 파싱 결과 ===\nKeyword: %s\nTotal: %d\n", keyword, data.Total)

		return map[string]int{"total": data.Total}, nil
	}

	return map[string]int{"total": 0}, lastErr
}

func testNaverSearch(config Config) *TestResult {
	clientID := strings.TrimSpace(config.Naver.ClientID)
	clientSecret := strings.TrimSpace(config.Naver.ClientSecret)

	if clientID == "" || clientSecret == "" {
		return &TestResult{
			OK:      false,
			Status:  400,
			Message: "Naver API credentials are missing.",
		}
	}

	reqURL := fmt.Sprintf("https://openapi.naver.com/v1/search/blog.json?query=%s&display=1", url.QueryEscape("test"))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return &TestResult{
			OK:      false,
			Status:  500,
			Message: err.Error(),
		}
	}

	req.Header.Set("X-Naver-Client-Id", clientID)
	req.Header.Set("X-Naver-Client-Secret", clientSecret)

	client := &http.Client{Timeout: time.Duration(fetchTimeoutMs) * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return &TestResult{
			OK:      false,
			Status:  500,
			Message: err.Error(),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return &TestResult{
		OK:      resp.StatusCode == http.StatusOK,
		Status:  resp.StatusCode,
		Message: string(body),
	}
}
