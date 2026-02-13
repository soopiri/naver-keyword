package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SearchAdKeywordItem struct {
	Keyword            string      `json:"keyword"`
	RelKeyword         string      `json:"relKeyword"`
	MonthlyPcQcCnt    interface{} `json:"monthlyPcQcCnt"`
	MonthlyMobileQcCnt interface{} `json:"monthlyMobileQcCnt"`
	MonthlyAvePcClkCnt interface{} `json:"monthlyAvePcClkCnt"`
	MonthlyAveMobileClkCnt interface{} `json:"monthlyAveMobileClkCnt"`
	MonthlyAvePcCtr    interface{} `json:"monthlyAvePcCtr"`
	MonthlyAveMobileCtr interface{} `json:"monthlyAveMobileCtr"`
	CompIdx            interface{} `json:"compIdx"`
	CompetitionIdx     interface{} `json:"competitionIdx"`
	CompetitionIndex   interface{} `json:"competitionIndex"`
	Competition        interface{} `json:"competition"`
	Comp               interface{} `json:"comp"`
	MonthlyAveCpc      interface{} `json:"monthlyAveCpc"`
	AvgCpc             interface{} `json:"avgCpc"`
	AvgCpcPc           interface{} `json:"avgCpcPc"`
	AvgCpcMobile       interface{} `json:"avgCpcMobile"`
	BidAmt             interface{} `json:"bidAmt"`
	Bid                interface{} `json:"bid"`
}

func toNumber(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		cleaned := strings.Map(func(r rune) rune {
			if (r >= '0' && r <= '9') || r == '.' {
				return r
			}
			return -1
		}, v)
		if cleaned == "" {
			return 0
		}
		if parsed, err := strconv.ParseFloat(cleaned, 64); err == nil {
			return parsed
		}
		return 0
	default:
		return 0
	}
}

func normalizeDevice(value interface{}) string {
	str := ""
	switch v := value.(type) {
	case string:
		str = strings.ToUpper(strings.TrimSpace(v))
	case int:
		str = fmt.Sprintf("%d", v)
	case float64:
		str = fmt.Sprintf("%.0f", v)
	default:
		return ""
	}
	str = strings.ToUpper(str)
	if strings.Contains(str, "PC") || strings.Contains(str, "DESKTOP") {
		return "PC"
	}
	if strings.Contains(str, "MOBILE") || strings.Contains(str, "MBL") {
		return "MOBILE"
	}
	return ""
}

func getKeywordLabel(item SearchAdKeywordItem) string {
	label := strings.TrimSpace(item.Keyword)
	if label == "" {
		label = strings.TrimSpace(item.RelKeyword)
	}
	return label
}

func getSearchAdMonthlyMetrics(item SearchAdKeywordItem, searchPc, searchMobile float64) map[string]interface{} {
	result := make(map[string]interface{})

	hasSearch := item.MonthlyPcQcCnt != nil || item.MonthlyMobileQcCnt != nil
	if hasSearch {
		result["monthlySearch"] = int(searchPc + searchMobile)
	} else {
		result["monthlySearch"] = nil
	}

	// 원본처럼 값이 null이 아닌지만 확인 (0이어도 유효한 값)
	hasClicks := item.MonthlyAvePcClkCnt != nil || item.MonthlyAveMobileClkCnt != nil
	pcClicks := toNumber(item.MonthlyAvePcClkCnt)
	mobileClicks := toNumber(item.MonthlyAveMobileClkCnt)
	if hasClicks {
		result["monthlyAvgClicks"] = pcClicks + mobileClicks
	} else {
		result["monthlyAvgClicks"] = nil
	}

	// 원본처럼 값이 null이 아닌지만 확인 (0이어도 유효한 값)
	hasCtr := item.MonthlyAvePcCtr != nil || item.MonthlyAveMobileCtr != nil
	pcCtr := toNumber(item.MonthlyAvePcCtr)
	mobileCtr := toNumber(item.MonthlyAveMobileCtr)
	if hasCtr {
		weightPc := pcClicks
		if weightPc == 0 {
			weightPc = searchPc
		}
		weightMobile := mobileClicks
		if weightMobile == 0 {
			weightMobile = searchMobile
		}
		weightTotal := weightPc + weightMobile
		if weightTotal > 0 {
			result["monthlyAvgCtr"] = (pcCtr*weightPc + mobileCtr*weightMobile) / weightTotal
		} else {
			// 원본처럼: pcCtr || mobileCtr || 0
			if pcCtr != 0 {
				result["monthlyAvgCtr"] = pcCtr
			} else if mobileCtr != 0 {
				result["monthlyAvgCtr"] = mobileCtr
			} else {
				result["monthlyAvgCtr"] = 0.0
			}
		}
	} else {
		result["monthlyAvgCtr"] = nil
	}

	// 원본처럼 여러 필드에서 경쟁정도 확인 (순서대로 확인)
	// 원본: item?.compIdx ?? item?.competitionIdx ?? item?.competitionIndex ?? item?.competition ?? item?.comp ?? null
	var competition interface{}
	if item.CompIdx != nil {
		competition = item.CompIdx
		log.Printf("=== Competition 필드 매핑 ===\nKeyword: %s\n사용된 필드: CompIdx\n값: %v (타입: %T)\n", getKeywordLabel(item), competition, competition)
	} else if item.CompetitionIdx != nil {
		competition = item.CompetitionIdx
		log.Printf("=== Competition 필드 매핑 ===\nKeyword: %s\n사용된 필드: CompetitionIdx\n값: %v (타입: %T)\n", getKeywordLabel(item), competition, competition)
	} else if item.CompetitionIndex != nil {
		competition = item.CompetitionIndex
		log.Printf("=== Competition 필드 매핑 ===\nKeyword: %s\n사용된 필드: CompetitionIndex\n값: %v (타입: %T)\n", getKeywordLabel(item), competition, competition)
	} else if item.Competition != nil {
		competition = item.Competition
		log.Printf("=== Competition 필드 매핑 ===\nKeyword: %s\n사용된 필드: Competition\n값: %v (타입: %T)\n", getKeywordLabel(item), competition, competition)
	} else if item.Comp != nil {
		competition = item.Comp
		log.Printf("=== Competition 필드 매핑 ===\nKeyword: %s\n사용된 필드: Comp\n값: %v (타입: %T)\n", getKeywordLabel(item), competition, competition)
	} else {
		competition = nil
		log.Printf("=== Competition 필드 매핑 ===\nKeyword: %s\n경쟁정도 필드 없음 (모두 nil)\nCompIdx: %v\nCompetitionIdx: %v\nCompetitionIndex: %v\nCompetition: %v\nComp: %v\n",
			getKeywordLabel(item), item.CompIdx, item.CompetitionIdx, item.CompetitionIndex, item.Competition, item.Comp)
	}
	
	// 원본처럼 값이 있으면 그대로 반환 (변환하지 않음)
	result["competition"] = competition
	log.Printf("=== getSearchAdMonthlyMetrics 결과 ===\nKeyword: %s\ncompetition (result): %v (타입: %T)\n", getKeywordLabel(item), result["competition"], result["competition"])

	return result
}

func fetchSearchAdKeywords(config Config, keywords []string) ([]SearchAdKeywordItem, error) {
	baseURL := config.SearchAd.BaseURL
	if baseURL == "" {
		baseURL = "https://api.searchad.naver.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	customerID := strings.TrimSpace(config.SearchAd.CustomerID)
	accessKey := strings.TrimSpace(config.SearchAd.AccessKey)
	secretKey := strings.TrimSpace(config.SearchAd.SecretKey)

	if customerID == "" || accessKey == "" || secretKey == "" {
		return []SearchAdKeywordItem{}, nil
	}

	cleaned := []string{}
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw != "" {
			cleaned = append(cleaned, kw)
		}
	}
	if len(cleaned) == 0 {
		return []SearchAdKeywordItem{}, nil
	}

	keywordSets := [][]string{cleaned}
	hasWhitespace := false
	for _, kw := range cleaned {
		if strings.Contains(kw, " ") {
			hasWhitespace = true
			break
		}
	}

	if hasWhitespace {
		compacted := []string{}
		for _, kw := range cleaned {
			compacted = append(compacted, strings.ReplaceAll(kw, " ", ""))
		}
		compactedJoined := strings.Join(compacted, ",")
		cleanedJoined := strings.Join(cleaned, ",")
		if compactedJoined != cleanedJoined {
			keywordSets = append(keywordSets, compacted)
		}
	}

	apiPath := "/keywordstool"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	method := "GET"

	requestWithRetry := func(encodedQuery string) (*http.Response, []byte, error) {
		var lastResp *http.Response
		var lastBody []byte
		var lastErr error

		for attempt := 0; attempt <= searchAdRetryMax; attempt++ {
			waitForSearchAdRateLimit()

			reqURL := fmt.Sprintf("%s%s?%s", baseURL, apiPath, encodedQuery)
			signature := createSignature(method, apiPath, timestamp, secretKey)

			req, err := http.NewRequest(method, reqURL, nil)
			if err != nil {
				lastErr = err
				if attempt < searchAdRetryMax {
					time.Sleep(time.Duration(searchAdRetryBaseMs*(1<<attempt)) * time.Millisecond)
					continue
				}
				return nil, nil, err
			}

			req.Header.Set("Content-Type", "application/json; charset=UTF-8")
			req.Header.Set("X-Timestamp", timestamp)
			req.Header.Set("X-API-KEY", accessKey)
			req.Header.Set("X-Customer", customerID)
			req.Header.Set("X-Signature", signature)

			client := &http.Client{Timeout: time.Duration(fetchTimeoutMs) * time.Millisecond}
			resp, err := fetchWithTimeout(client, req, fetchTimeoutMs)
			if err != nil {
				lastErr = err
				if attempt < searchAdRetryMax {
					time.Sleep(time.Duration(searchAdRetryBaseMs*(1<<attempt)) * time.Millisecond)
					continue
				}
				return nil, nil, err
			}

			// 응답을 받은 후 즉시 body를 읽어서 컨텍스트 취소 문제 방지
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = err
				if attempt < searchAdRetryMax {
					time.Sleep(time.Duration(searchAdRetryBaseMs*(1<<attempt)) * time.Millisecond)
					continue
				}
				return nil, nil, err
			}

			if resp.StatusCode == http.StatusOK {
				// body를 이미 읽었으므로, 새로운 Response 객체를 만들어서 반환
				// 하지만 실제로는 body만 반환하면 되므로, resp와 body를 함께 반환
				return resp, body, nil
			}

			if resp.StatusCode == 403 && strings.Contains(string(body), "invalid-signature") {
				signature = createSignatureWithQuery(method, apiPath, timestamp, secretKey, encodedQuery)
				req.Header.Set("X-Signature", signature)
				resp2, err2 := fetchWithTimeout(client, req, fetchTimeoutMs)
				if err2 == nil {
					body2, err2 := io.ReadAll(resp2.Body)
					resp2.Body.Close()
					if err2 == nil && resp2.StatusCode == http.StatusOK {
						return resp2, body2, nil
					}
				}
			}

			if resp.StatusCode == 429 {
				logLine("searchad.log", fmt.Sprintf("status=%d url=%s query=%s body=%s", resp.StatusCode, reqURL, encodedQuery, string(body)))
				if attempt < searchAdRetryMax {
					time.Sleep(time.Duration(searchAdRetryBaseMs*(1<<attempt)) * time.Millisecond)
					continue
				}
			}

			lastResp = resp
			lastBody = body
			logLine("searchad.log", fmt.Sprintf("status=%d url=%s query=%s body=%s", resp.StatusCode, reqURL, encodedQuery, string(body)))
		}

		if lastResp != nil {
			return lastResp, lastBody, nil
		}
		return nil, nil, lastErr
	}

	for _, keywordSet := range keywordSets {
		queryVariants := buildSearchAdQueries(keywordSet)
		log.Printf("=== fetchSearchAdKeywords: keywordSet 처리 시작 ===\nKeywords: %v\nkeywordSet: %v\nqueryVariants 개수: %d\n", keywords, keywordSet, len(queryVariants))
		for idx, encodedQuery := range queryVariants {
			log.Printf("=== fetchSearchAdKeywords: API 호출 시도[%d] ===\nKeywords: %v\nencodedQuery: %s\n", idx, keywords, encodedQuery)
			resp, body, err := requestWithRetry(encodedQuery)
			if err != nil {
				log.Printf("=== fetchSearchAdKeywords: requestWithRetry 실패 ===\nKeywords: %v\nError: %v\n", keywords, err)
				continue
			}
			if resp == nil {
				log.Printf("=== fetchSearchAdKeywords: resp가 nil ===\nKeywords: %v\n", keywords)
				continue
			}

			log.Printf("=== fetchSearchAdKeywords: 응답 받음 ===\nKeywords: %v\nStatusCode: %d\n", keywords, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				// body는 이미 requestWithRetry에서 읽었음
				if body == nil {
					log.Printf("=== fetchSearchAdKeywords: body가 nil ===\nKeywords: %v\n", keywords)
					continue
				}

				// API 응답 원본 로그 출력
				log.Printf("=== SearchAd API 원본 응답 ===\nKeywords: %v\nResponse Body: %s\n", keywords, string(body))

				var data struct {
					KeywordList []SearchAdKeywordItem `json:"keywordList"`
				}
				if err := json.Unmarshal(body, &data); err != nil {
					log.Printf("=== fetchSearchAdKeywords: JSON 파싱 실패 ===\nKeywords: %v\nError: %v\nBody: %s\n", keywords, err, string(body))
					continue
				}

				// 파싱된 데이터 로그 출력 (특히 competition 필드)
				for i, item := range data.KeywordList {
					log.Printf("=== SearchAd Item[%d] ===\nKeyword: %s\nRelKeyword: %s\nCompIdx: %v\nCompetitionIdx: %v\nCompetitionIndex: %v\nCompetition: %v\nComp: %v\n",
						i, item.Keyword, item.RelKeyword, item.CompIdx, item.CompetitionIdx, item.CompetitionIndex, item.Competition, item.Comp)
				}

				log.Printf("=== fetchSearchAdKeywords: 성공적으로 반환 ===\nKeywords: %v\n결과 개수: %d\n", keywords, len(data.KeywordList))
				return data.KeywordList, nil
			} else {
				log.Printf("=== fetchSearchAdKeywords: HTTP 상태 코드 오류 ===\nKeywords: %v\nStatusCode: %d\nResponse Body: %s\n", keywords, resp.StatusCode, string(body))
			}
		}
	}

	log.Printf("=== fetchSearchAdKeywords: 모든 시도 실패, 빈 배열 반환 ===\nKeywords: %v\n", keywords)
	return []SearchAdKeywordItem{}, nil
}

func testSearchAd(config Config) *TestResult {
	baseURL := config.SearchAd.BaseURL
	if baseURL == "" {
		baseURL = "https://api.searchad.naver.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	customerID := strings.TrimSpace(config.SearchAd.CustomerID)
	accessKey := strings.TrimSpace(config.SearchAd.AccessKey)
	secretKey := strings.TrimSpace(config.SearchAd.SecretKey)

	if customerID == "" || accessKey == "" || secretKey == "" {
		return &TestResult{
			OK:      false,
			Status:  400,
			Message: "SearchAd credentials are missing.",
		}
	}

	apiPath := "/keywordstool"
	queries := buildSearchAdQueries([]string{"test"})
	if len(queries) == 0 {
		return &TestResult{
			OK:      false,
			Status:  400,
			Message: "Failed to build query",
		}
	}
	encodedQuery := queries[0]
	reqURL := fmt.Sprintf("%s%s?%s", baseURL, apiPath, encodedQuery)
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	method := "GET"

	signature := createSignature(method, apiPath, timestamp, secretKey)
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return &TestResult{
			OK:      false,
			Status:  500,
			Message: err.Error(),
		}
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-API-KEY", accessKey)
	req.Header.Set("X-Customer", customerID)
	req.Header.Set("X-Signature", signature)

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

	if resp.StatusCode == 403 && strings.Contains(string(body), "invalid-signature") {
		signature = createSignatureWithQuery(method, apiPath, timestamp, secretKey, encodedQuery)
		req.Header.Set("X-Signature", signature)
		resp2, err2 := client.Do(req)
		if err2 == nil {
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)
			return &TestResult{
				OK:      resp2.StatusCode == http.StatusOK,
				Status:  resp2.StatusCode,
				Message: string(body2),
			}
		}
	}

	return &TestResult{
		OK:      resp.StatusCode == http.StatusOK,
		Status:  resp.StatusCode,
		Message: string(body),
	}
}

const (
	cpcDefaultBid            = 70
	cpcDefaultDevice         = "PC"
	searchAdStatsDatePreset  = "last7days"
	searchAdStatsBreakdown    = "pcMblTp"
)

type NccKeywordItem struct {
	NccKeywordId string      `json:"nccKeywordId"`
	Keyword      string      `json:"keyword"`
	Label        string      `json:"label"`
	Status       string      `json:"status"`
	StatusType   string      `json:"statusType"`
	ID           string      `json:"id"`
}

type KeywordStats struct {
	PC     DeviceStats `json:"pc"`
	Mobile DeviceStats `json:"mobile"`
	Total  DeviceStats `json:"total"`
}

type DeviceStats struct {
	Clicks float64 `json:"clicks"`
	Cost   float64 `json:"cost"`
}

type CpcEstimate struct {
	Clicks      float64 `json:"clicks"`
	Cost        float64 `json:"cost"`
	Impressions float64 `json:"impressions"`
}

func normalizeStatsRows(payload interface{}) []map[string]interface{} {
	switch v := payload.(type) {
	case []interface{}:
		result := []map[string]interface{}{}
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		return result
	case map[string]interface{}:
		if data, ok := v["data"].([]interface{}); ok {
			return normalizeStatsRows(data)
		}
		if rows, ok := v["rows"].([]interface{}); ok {
			return normalizeStatsRows(rows)
		}
		if items, ok := v["items"].([]interface{}); ok {
			return normalizeStatsRows(items)
		}
	}
	return []map[string]interface{}{}
}

func buildKeywordIdMap(items []map[string]interface{}) map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, item := range items {
		label := ""
		if kw, ok := item["keyword"].(string); ok {
			label = strings.TrimSpace(kw)
		} else if lbl, ok := item["label"].(string); ok {
			label = strings.TrimSpace(lbl)
		}
		if label == "" {
			continue
		}

		id := ""
		if nccId, ok := item["nccKeywordId"].(string); ok {
			id = strings.TrimSpace(nccId)
		} else if itemId, ok := item["id"].(string); ok {
			id = strings.TrimSpace(itemId)
		}
		if id == "" {
			continue
		}

		status := ""
		if s, ok := item["status"].(string); ok {
			status = strings.ToUpper(strings.TrimSpace(s))
		} else if st, ok := item["statusType"].(string); ok {
			status = strings.ToUpper(strings.TrimSpace(st))
		}

		key := normalizeTerm(label)
		existing, exists := result[key]
		if !exists {
			result[key] = map[string]string{
				"id":     id,
				"label":  label,
				"status": status,
			}
			continue
		}
		if existing["status"] != "ON" && status == "ON" {
			result[key] = map[string]string{
				"id":     id,
				"label":  label,
				"status": status,
			}
		}
	}
	return result
}

func fetchSearchAdList(config Config, apiPath string, query map[string]string) ([]map[string]interface{}, error) {
	normalizedBase := strings.TrimSuffix(config.SearchAd.BaseURL, "/")
	if normalizedBase == "" {
		normalizedBase = "https://api.searchad.naver.com"
	}
	trimmed := SearchAdConfig{
		CustomerID: strings.TrimSpace(config.SearchAd.CustomerID),
		AccessKey:  strings.TrimSpace(config.SearchAd.AccessKey),
		SecretKey:  strings.TrimSpace(config.SearchAd.SecretKey),
	}
	if trimmed.CustomerID == "" || trimmed.AccessKey == "" || trimmed.SecretKey == "" {
		return []map[string]interface{}{}, nil
	}

	params := make([]string, 0)
	for k, v := range query {
		if v != "" {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
	}
	queryString := strings.Join(params, "&")
	url := normalizedBase + apiPath
	if queryString != "" {
		url += "?" + queryString
	}
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"

	requestFunc := func(signature string) (*http.Response, error) {
		waitForSearchAdRateLimit()
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Set("X-Timestamp", timestamp)
		req.Header.Set("X-API-KEY", trimmed.AccessKey)
		req.Header.Set("X-Customer", trimmed.CustomerID)
		req.Header.Set("X-Signature", signature)

		client := &http.Client{Timeout: 8 * time.Second}
		return client.Do(req)
	}

	requestWithSignature := func() (*http.Response, error) {
		sig1 := createSignature(method, apiPath, timestamp, trimmed.SecretKey)
		resp, err := requestFunc(sig1)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 403 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if strings.Contains(string(body), "invalid-signature") && queryString != "" {
				sig2 := createSignatureWithQuery(method, apiPath, timestamp, trimmed.SecretKey, queryString)
				return requestFunc(sig2)
			}
			return resp, nil
		}
		return resp, nil
	}

	requestWithRetry := func() ([]map[string]interface{}, error) {
		for attempt := 0; attempt <= 3; attempt++ {
			resp, err := requestWithSignature()
			if err != nil {
				if attempt < 3 {
					time.Sleep(time.Duration(600*(1<<attempt)) * time.Millisecond)
					continue
				}
				return nil, err
			}
			if resp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				var data interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					return []map[string]interface{}{}, nil
				}
				if arr, ok := data.([]interface{}); ok {
					result := []map[string]interface{}{}
					for _, item := range arr {
						if m, ok := item.(map[string]interface{}); ok {
							result = append(result, m)
						}
					}
					return result, nil
				}
				if obj, ok := data.(map[string]interface{}); ok {
					if arr, ok := obj["data"].([]interface{}); ok {
						result := []map[string]interface{}{}
						for _, item := range arr {
							if m, ok := item.(map[string]interface{}); ok {
								result = append(result, m)
							}
						}
						return result, nil
					}
				}
				return []map[string]interface{}{}, nil
			}
			if resp.StatusCode == 429 && attempt < 3 {
				resp.Body.Close()
				logLine("searchad.log", fmt.Sprintf("status=%d url=%s", resp.StatusCode, url))
				time.Sleep(time.Duration(600*(1<<attempt)) * time.Millisecond)
				continue
			}
			resp.Body.Close()
		}
		return []map[string]interface{}{}, nil
	}

	return requestWithRetry()
}

func fetchAll(config Config, apiPath string, idField string, baseQuery map[string]string) ([]map[string]interface{}, error) {
	collected := []map[string]interface{}{}
	baseSearchID := ""
	for {
		query := make(map[string]string)
		for k, v := range baseQuery {
			query[k] = v
		}
		query["recordSize"] = fmt.Sprintf("%d", searchAdListRecordSize)
		query["selector"] = "NEXT"
		if baseSearchID != "" {
			query["baseSearchId"] = baseSearchID
		}
		chunk, err := fetchSearchAdList(config, apiPath, query)
		if err != nil || len(chunk) == 0 {
			break
		}
		collected = append(collected, chunk...)
		last := chunk[len(chunk)-1]
		nextID := ""
		if id, ok := last[idField].(string); ok {
			nextID = strings.TrimSpace(id)
		}
		if nextID == "" || len(chunk) < searchAdListRecordSize {
			break
		}
		baseSearchID = nextID
	}
	return collected, nil
}

func fetchSearchAdNccKeywords(config Config) ([]map[string]interface{}, error) {
	now := time.Now().UnixMilli()
	if cachedSearchAdNccKeywords != nil && now-cachedSearchAdNccKeywordsAt < searchAdKeywordCacheMs {
		return cachedSearchAdNccKeywords, nil
	}

	campaigns, err := fetchAll(config, "/ncc/campaigns", "nccCampaignId", map[string]string{})
	if err != nil {
		return []map[string]interface{}{}, err
	}
	logSearchAdCpc(fmt.Sprintf("ncc campaigns loaded count=%d", len(campaigns)))

	adgroups := []map[string]interface{}{}
	for _, campaign := range campaigns {
		campaignID := ""
		if id, ok := campaign["nccCampaignId"].(string); ok {
			campaignID = strings.TrimSpace(id)
		}
		if campaignID == "" {
			continue
		}
		groups, err := fetchAll(config, "/ncc/adgroups", "nccAdgroupId", map[string]string{
			"nccCampaignId": campaignID,
		})
		if err == nil {
			adgroups = append(adgroups, groups...)
		}
	}
	logSearchAdCpc(fmt.Sprintf("ncc adgroups loaded count=%d", len(adgroups)))

	keywords := []map[string]interface{}{}
	for _, adgroup := range adgroups {
		adgroupID := ""
		if id, ok := adgroup["nccAdgroupId"].(string); ok {
			adgroupID = strings.TrimSpace(id)
		}
		if adgroupID == "" {
			continue
		}
		groupKeywords, err := fetchAll(config, "/ncc/keywords", "nccKeywordId", map[string]string{
			"nccAdgroupId": adgroupID,
		})
		if err == nil {
			keywords = append(keywords, groupKeywords...)
		}
	}
	logSearchAdCpc(fmt.Sprintf("ncc keywords loaded count=%d", len(keywords)))

	cachedSearchAdNccKeywords = keywords
	cachedSearchAdNccKeywordsAt = now
	return keywords, nil
}

func fetchSearchAdStats(config Config, ids []string) ([]map[string]interface{}, error) {
	normalizedBase := strings.TrimSuffix(config.SearchAd.BaseURL, "/")
	if normalizedBase == "" {
		normalizedBase = "https://api.searchad.naver.com"
	}
	trimmed := SearchAdConfig{
		CustomerID: strings.TrimSpace(config.SearchAd.CustomerID),
		AccessKey:  strings.TrimSpace(config.SearchAd.AccessKey),
		SecretKey:  strings.TrimSpace(config.SearchAd.SecretKey),
	}
	if trimmed.CustomerID == "" || trimmed.AccessKey == "" || trimmed.SecretKey == "" {
		return []map[string]interface{}{}, nil
	}

	cleanedIDs := []string{}
	for _, id := range ids {
		trimmedID := strings.TrimSpace(id)
		if trimmedID != "" {
			cleanedIDs = append(cleanedIDs, trimmedID)
		}
	}
	if len(cleanedIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	apiPath := "/stats"
	params := []string{
		fmt.Sprintf("ids=%s", strings.Join(cleanedIDs, ",")),
		fmt.Sprintf(`fields=["clkCnt","salesAmt"]`),
		fmt.Sprintf("datePreset=%s", searchAdStatsDatePreset),
		fmt.Sprintf("breakdown=%s", searchAdStatsBreakdown),
	}
	queryString := strings.Join(params, "&")
	url := normalizedBase + apiPath + "?" + queryString
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"

	logSearchAdCpc(fmt.Sprintf("stats request start ids=%d datePreset=%s breakdown=%s", len(cleanedIDs), searchAdStatsDatePreset, searchAdStatsBreakdown))

	requestFunc := func(signature string) (*http.Response, error) {
		waitForSearchAdRateLimit()
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Set("X-Timestamp", timestamp)
		req.Header.Set("X-API-KEY", trimmed.AccessKey)
		req.Header.Set("X-Customer", trimmed.CustomerID)
		req.Header.Set("X-Signature", signature)

		client := &http.Client{Timeout: 8 * time.Second}
		return client.Do(req)
	}

	requestWithSignature := func() (*http.Response, error) {
		sig1 := createSignature(method, apiPath, timestamp, trimmed.SecretKey)
		resp, err := requestFunc(sig1)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 403 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if strings.Contains(string(body), "invalid-signature") {
				sig2 := createSignatureWithQuery(method, apiPath, timestamp, trimmed.SecretKey, queryString)
				return requestFunc(sig2)
			}
			return resp, nil
		}
		return resp, nil
	}

	requestWithRetry := func() ([]map[string]interface{}, error) {
		for attempt := 0; attempt <= 3; attempt++ {
			resp, err := requestWithSignature()
			if err != nil {
				if attempt < 3 {
					time.Sleep(time.Duration(600*(1<<attempt)) * time.Millisecond)
					continue
				}
				return nil, err
			}
			if resp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				var data interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					return []map[string]interface{}{}, nil
				}
				rows := normalizeStatsRows(data)
				logSearchAdCpc(fmt.Sprintf("stats response ok rows=%d", len(rows)))
				return rows, nil
			}
			if resp.StatusCode == 429 && attempt < 3 {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				logLine("searchad.log", fmt.Sprintf("status=%d url=%s body=%s", resp.StatusCode, url, string(body)))
				time.Sleep(time.Duration(600*(1<<attempt)) * time.Millisecond)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			logLine("searchad.log", fmt.Sprintf("status=%d url=%s body=%s", resp.StatusCode, url, string(body)))
			logSearchAdCpc(fmt.Sprintf("stats request failed status=%d", resp.StatusCode))
		}
		return []map[string]interface{}{}, nil
	}

	return requestWithRetry()
}

func fetchSearchAdKeywordStats(config Config, keywords []string) (map[string]KeywordStats, error) {
	cleaned := []string{}
	seen := make(map[string]bool)
	for _, kw := range keywords {
		trimmed := strings.TrimSpace(kw)
		if trimmed != "" && !seen[trimmed] {
			cleaned = append(cleaned, trimmed)
			seen[trimmed] = true
		}
	}
	if len(cleaned) == 0 {
		return make(map[string]KeywordStats), nil
	}

	nccKeywords, err := fetchSearchAdNccKeywords(config)
	if err != nil {
		return make(map[string]KeywordStats), err
	}
	logSearchAdCpc(fmt.Sprintf("ncc keywords loaded count=%d", len(nccKeywords)))

	keywordIDMap := buildKeywordIdMap(nccKeywords)
	idToKeyword := make(map[string]string)
	ids := []string{}
	for _, keyword := range cleaned {
		key := normalizeTerm(keyword)
		if entry, ok := keywordIDMap[key]; ok {
			if id := entry["id"]; id != "" {
				ids = append(ids, id)
				idToKeyword[id] = keyword
			}
		}
	}

	statsMap := make(map[string]KeywordStats)
	for _, id := range ids {
		rows, err := fetchSearchAdStats(config, []string{id})
		if err != nil || len(rows) == 0 {
			continue
		}
		keyword := idToKeyword[id]
		if keyword == "" {
			continue
		}
		if _, exists := statsMap[keyword]; !exists {
			statsMap[keyword] = KeywordStats{
				PC:     DeviceStats{},
				Mobile: DeviceStats{},
				Total:  DeviceStats{},
			}
		}
		entry := statsMap[keyword]
		for _, row := range rows {
			device := normalizeDevice(getString(row, "pcMblTp", "device", "pcMobileType", "deviceType"))
			clicks := toNumber(row["clkCnt"])
			cost := toNumber(row["salesAmt"])
			if device == "PC" {
				entry.PC.Clicks += clicks
				entry.PC.Cost += cost
			} else if device == "MOBILE" {
				entry.Mobile.Clicks += clicks
				entry.Mobile.Cost += cost
			} else {
				entry.Total.Clicks += clicks
				entry.Total.Cost += cost
			}
		}
		statsMap[keyword] = entry
	}
	return statsMap, nil
}

func fetchSearchAdCpcEstimates(config Config, keywords []string, bid float64, device string, bidMap map[string]float64) (map[string]CpcEstimate, error) {
	normalizedBase := strings.TrimSuffix(config.SearchAd.BaseURL, "/")
	if normalizedBase == "" {
		normalizedBase = "https://api.searchad.naver.com"
	}
	trimmed := SearchAdConfig{
		CustomerID: strings.TrimSpace(config.SearchAd.CustomerID),
		AccessKey:  strings.TrimSpace(config.SearchAd.AccessKey),
		SecretKey:  strings.TrimSpace(config.SearchAd.SecretKey),
	}
	if trimmed.CustomerID == "" || trimmed.AccessKey == "" || trimmed.SecretKey == "" {
		logSearchAdCpc("credentials missing")
		return make(map[string]CpcEstimate), nil
	}

	cleaned := []string{}
	seen := make(map[string]bool)
	for _, kw := range keywords {
		trimmedKW := strings.TrimSpace(kw)
		if trimmedKW != "" && !seen[trimmedKW] {
			cleaned = append(cleaned, trimmedKW)
			seen[trimmedKW] = true
		}
	}
	if len(cleaned) == 0 {
		logSearchAdCpc("no keywords")
		return make(map[string]CpcEstimate), nil
	}

	if bid == 0 {
		bid = cpcDefaultBid
	}
	if device == "" {
		device = cpcDefaultDevice
	}

	resolveBid := func(keyword string) float64 {
		if bidMap != nil {
			if mapped, ok := bidMap[keyword]; ok && mapped > 0 {
				return mapped
			}
		}
		return bid
	}

	items := []map[string]interface{}{}
	for _, keyword := range cleaned {
		items = append(items, map[string]interface{}{
			"keyword":     keyword,
			"bid":         resolveBid(keyword),
			"device":      device,
			"keywordplus": false,
		})
	}

	payload := map[string]interface{}{
		"items": items,
	}
	payloadJSON, _ := json.Marshal(payload)

	apiPath := "/estimate/performance-bulk"
	url := normalizedBase + apiPath
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "POST"

	logSearchAdCpc(fmt.Sprintf("request start path=%s device=%s bid=%.0f count=%d", apiPath, device, bid, len(items)))

	requestFunc := func(signature string) (*http.Response, error) {
		waitForSearchAdRateLimit()
		req, err := http.NewRequest(method, url, strings.NewReader(string(payloadJSON)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Set("X-Timestamp", timestamp)
		req.Header.Set("X-API-KEY", trimmed.AccessKey)
		req.Header.Set("X-Customer", trimmed.CustomerID)
		req.Header.Set("X-Signature", signature)

		client := &http.Client{Timeout: 8 * time.Second}
		return client.Do(req)
	}

	requestWithRetry := func() (map[string]CpcEstimate, error) {
		for attempt := 0; attempt <= 3; attempt++ {
			sig := createSignature(method, apiPath, timestamp, trimmed.SecretKey)
			resp, err := requestFunc(sig)
			if err != nil {
				if attempt < 3 {
					time.Sleep(time.Duration(600*(1<<attempt)) * time.Millisecond)
					continue
				}
				return make(map[string]CpcEstimate), err
			}
			if resp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				var data map[string]interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					return make(map[string]CpcEstimate), nil
				}
				result := make(map[string]CpcEstimate)
				if items, ok := data["items"].([]interface{}); ok {
					logSearchAdCpc(fmt.Sprintf("response items=%d device=%s bid=%.0f", len(items), device, bid))
					for _, item := range items {
						if m, ok := item.(map[string]interface{}); ok {
							keyword := getString(m, "keyword")
							if keyword == "" {
								continue
							}
							result[keyword] = CpcEstimate{
								Clicks:      toNumber(m["clicks"]),
								Cost:        toNumber(m["cost"]),
								Impressions: toNumber(m["impressions"]),
							}
						}
					}
				} else {
					logLine("searchad.log", fmt.Sprintf("status=200 url=%s path=%s device=%s bid=%.0f items=0", url, apiPath, device, bid))
					logSearchAdCpc(fmt.Sprintf("response items=0 device=%s bid=%.0f", device, bid))
				}
				logSearchAdCpc(fmt.Sprintf("response ok status=%d device=%s bid=%.0f", resp.StatusCode, device, bid))
				return result, nil
			}
			if resp.StatusCode == 429 && attempt < 3 {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				logLine("searchad.log", fmt.Sprintf("status=%d url=%s body=%s", resp.StatusCode, url, string(body)))
				time.Sleep(time.Duration(600*(1<<attempt)) * time.Millisecond)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			logLine("searchad.log", fmt.Sprintf("status=%d url=%s path=%s device=%s bid=%.0f body=%s", resp.StatusCode, url, apiPath, device, bid, string(body)))
			logSearchAdCpc(fmt.Sprintf("request failed status=%d device=%s bid=%.0f", resp.StatusCode, device, bid))
		}
		return make(map[string]CpcEstimate), nil
	}

	return requestWithRetry()
}

func getString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key].(string); ok {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func logSearchAdCpc(message string) {
	logLine("searchad.log", fmt.Sprintf("CPC: %s", message))
}
