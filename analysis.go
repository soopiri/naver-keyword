package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
)

const relatedKeywordLimit = 20

// 원본처럼 경쟁정도 문자열을 숫자로 변환
// LOW -> 1, MID/MEDIUM -> 2, HIGH -> 3
func competitionToNumber(val interface{}) *int {
	log.Printf("=== competitionToNumber 호출 ===\n입력값: %v (타입: %T)\n", val, val)
	
	if val == nil {
		log.Printf("competitionToNumber 결과: nil (입력값이 nil)\n")
		return nil
	}

	// 이미 숫자인 경우
	switch v := val.(type) {
	case int:
		log.Printf("competitionToNumber 결과: %d (int 타입 그대로 반환)\n", v)
		return &v
	case float64:
		iv := int(v)
		log.Printf("competitionToNumber 결과: %d (float64 -> int 변환)\n", iv)
		return &iv
	case string:
		upper := strings.ToUpper(strings.TrimSpace(v))
		var result int
		switch upper {
		case "LOW":
			result = 1
			log.Printf("competitionToNumber 결과: %d (LOW -> 1)\n", result)
		case "MID", "MEDIUM":
			result = 2
			log.Printf("competitionToNumber 결과: %d (%s -> 2)\n", result, upper)
		case "HIGH":
			result = 3
			log.Printf("competitionToNumber 결과: %d (HIGH -> 3)\n", result)
		default:
			// 숫자 문자열일 수 있음
			num := toNumber(val)
			if num != 0 {
				result = int(num)
				log.Printf("competitionToNumber 결과: %d (문자열 숫자 파싱: %s -> %.0f -> %d)\n", result, v, num, result)
				return &result
			}
			log.Printf("competitionToNumber 결과: nil (알 수 없는 문자열: %s)\n", v)
			return nil
		}
		return &result
	default:
		log.Printf("competitionToNumber 결과: nil (지원하지 않는 타입: %T)\n", val)
		return nil
	}
}

func normalizeTerm(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
}

func matchesAnySeed(label string, seeds []string) bool {
	if label == "" {
		return false
	}
	normalizedLabel := normalizeTerm(label)
	for _, seed := range seeds {
		normalizedSeed := normalizeTerm(seed)
		// 정확히 일치하는 키워드만 필터링
		if normalizedSeed != "" && normalizedLabel == normalizedSeed {
			return true
		}
	}
	return false
}

func chunkArray(list []string, size int) [][]string {
	if len(list) == 0 || size <= 0 {
		return [][]string{}
	}
	chunks := [][]string{}
	for i := 0; i < len(list); i += size {
		end := i + size
		if end > len(list) {
			end = len(list)
		}
		chunks = append(chunks, list[i:end])
	}
	return chunks
}

func hydrateSearchVolumes(config Config, items []RelatedDetail) ([]RelatedDetail, error) {
	keywords := []string{}
	for _, item := range items {
		if item.Keyword != "" {
			keywords = append(keywords, item.Keyword)
		}
	}
	if len(keywords) == 0 {
		return items, nil
	}

	keywordMap := make(map[string]RelatedDetail)
	for _, item := range items {
		key := normalizeTerm(item.Keyword)
		keywordMap[key] = item
	}

	chunks := chunkArray(keywords, searchAdChunkSize)
	for _, chunk := range chunks {
		keywordList, err := fetchSearchAdKeywords(config, chunk, "단일 키워드 검색")
		if err != nil {
			continue
		}

		for _, entry := range keywordList {
			label := getKeywordLabel(entry)
			if label == "" {
				continue
			}
			key := normalizeTerm(label)
			if item, ok := keywordMap[key]; ok {
				pc := int(toNumber(entry.MonthlyPcQcCnt))
				mob := int(toNumber(entry.MonthlyMobileQcCnt))
				item.SearchPc = pc
				item.SearchMobile = mob
				item.TotalSearch = pc + mob
				keywordMap[key] = item
			}
		}
	}

	result := []RelatedDetail{}
	for _, item := range items {
		key := normalizeTerm(item.Keyword)
		if updated, ok := keywordMap[key]; ok {
			result = append(result, updated)
		} else {
			result = append(result, item)
		}
	}

	return result, nil
}

func analyzeKeyword(config Config, keyword string, onProgress func(Progress)) (*KeywordAnalysisResult, error) {
	target := strings.TrimSpace(keyword)
	if target == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	searchTarget := target
	keywordList, err := fetchSearchAdKeywords(config, []string{searchTarget}, "단일 키워드 검색")
	if err != nil {
		return nil, err
	}

	// 원본처럼 정확히 일치하는 것을 찾고, 없으면 첫 번째 항목 사용
	var match SearchAdKeywordItem
	if len(keywordList) > 0 {
		// 정확히 일치하는 것을 찾음
		for _, item := range keywordList {
			label := getKeywordLabel(item)
			if label == searchTarget {
				match = item
				break
			}
		}
		// 정확히 일치하는 것이 없으면 첫 번째 항목 사용 (원본과 동일)
		if match.Keyword == "" && match.RelKeyword == "" {
			match = keywordList[0]
		}
	}

	searchPc := int(toNumber(match.MonthlyPcQcCnt))
	searchMobile := int(toNumber(match.MonthlyMobileQcCnt))
	hasSearchVolume := searchPc > 0 || searchMobile > 0

	if !hasSearchVolume && strings.Contains(searchTarget, " ") {
		compact := strings.ReplaceAll(searchTarget, " ", "")
		if compact != "" && compact != searchTarget {
			fallbackList, err := fetchSearchAdKeywords(config, []string{compact}, "단일 키워드 검색")
			if err == nil && len(fallbackList) > 0 {
				searchTarget = compact
				keywordList = fallbackList
				// 원본처럼 정확히 일치하는 것을 찾고, 없으면 첫 번째 항목 사용
				// 원본: keywordList.find((k) => getKeywordLabel(k) === searchTarget) || keywordList[0] || {}
				match = SearchAdKeywordItem{}
				found := false
				for _, item := range fallbackList {
					label := getKeywordLabel(item)
					if label == searchTarget {
						match = item
						found = true
						break
					}
				}
				// 정확히 일치하는 것이 없으면 첫 번째 항목 사용 (원본과 동일)
				if !found {
					match = fallbackList[0]
				}
				searchPc = int(toNumber(match.MonthlyPcQcCnt))
				searchMobile = int(toNumber(match.MonthlyMobileQcCnt))
			}
		}
	}

	totalSearch := searchPc + searchMobile
	searchAdMetrics := getSearchAdMonthlyMetrics(match, float64(searchPc), float64(searchMobile))

	relatedMap := make(map[string]RelatedDetail)
	normalizedTarget := normalizeTerm(target)
	for _, item := range keywordList {
		label := getKeywordLabel(item)
		if label == "" {
			continue
		}
		// 대소문자 무시 비교로 타겟 키워드 제외
		if normalizeTerm(label) == normalizedTarget {
			continue
		}
		// 같은 키워드가 이미 있으면 스킵 (대소문자 구분하여 저장)
		if _, exists := relatedMap[label]; exists {
			continue
		}
		rSearchPc := int(toNumber(item.MonthlyPcQcCnt))
		rSearchMobile := int(toNumber(item.MonthlyMobileQcCnt))
		rTotalSearch := rSearchPc + rSearchMobile
		relatedMap[label] = RelatedDetail{
			Keyword:      label,
			SearchPc:     rSearchPc,
			SearchMobile: rSearchMobile,
			TotalSearch:  rTotalSearch,
		}
	}

	relatedBase := []RelatedDetail{}
	for _, item := range relatedMap {
		relatedBase = append(relatedBase, item)
	}

	sort.Slice(relatedBase, func(i, j int) bool {
		return relatedBase[i].TotalSearch > relatedBase[j].TotalSearch
	})

	if len(relatedBase) > relatedKeywordLimit {
		relatedBase = relatedBase[:relatedKeywordLimit]
	}

	if len(relatedBase) < relatedKeywordLimit {
		suggestions, err := fetchNaverAutocomplete(target, relatedKeywordLimit, "단일 키워드 검색")
		if err == nil {
			existing := make(map[string]bool)
			existingNormalized := make(map[string]bool)
			for _, item := range relatedBase {
				existing[item.Keyword] = true
				existingNormalized[normalizeTerm(item.Keyword)] = true
			}

			missing := []string{}
			normalizedTarget := normalizeTerm(target)
			for _, sug := range suggestions {
				if sug == "" {
					continue
				}
				// 정확히 일치하는 경우 제외
				if existing[sug] {
					continue
				}
				// 대소문자 무시하여 타겟과 일치하는 경우 제외
				if normalizeTerm(sug) == normalizedTarget {
					continue
				}
				// 대소문자 무시하여 이미 있는 경우 제외
				if existingNormalized[normalizeTerm(sug)] {
					continue
				}
				missing = append(missing, sug)
			}

			if len(missing) > 0 {
				searchadFallback, err := fetchSearchAdKeywords(config, missing, "단일 키워드 검색")
				if err == nil {
					byLabel := make(map[string]RelatedDetail)
					for _, item := range searchadFallback {
						label := getKeywordLabel(item)
						if label == "" {
							continue
						}
						found := false
						for _, m := range missing {
							if label == m {
								found = true
								break
							}
						}
						if !found {
							continue
						}
						byLabel[label] = RelatedDetail{
							Keyword:      label,
							SearchPc:     int(toNumber(item.MonthlyPcQcCnt)),
							SearchMobile: int(toNumber(item.MonthlyMobileQcCnt)),
							TotalSearch:  int(toNumber(item.MonthlyPcQcCnt) + toNumber(item.MonthlyMobileQcCnt)),
						}
					}

					for _, label := range missing {
						if item, ok := byLabel[label]; ok {
							relatedBase = append(relatedBase, item)
						} else {
							relatedBase = append(relatedBase, RelatedDetail{
								Keyword:      label,
								SearchPc:     0,
								SearchMobile: 0,
								TotalSearch:  0,
							})
						}
					}
				}
			}
		}
	}

	relatedBase, _ = hydrateSearchVolumes(config, relatedBase)

	sort.Slice(relatedBase, func(i, j int) bool {
		return relatedBase[i].TotalSearch > relatedBase[j].TotalSearch
	})

	if len(relatedBase) > relatedKeywordLimit {
		relatedBase = relatedBase[:relatedKeywordLimit]
	}

	totalSteps := len(relatedBase) + 1
	completed := 0

	if onProgress != nil {
		onProgress(Progress{
			Current: completed,
			Total:   totalSteps,
			Keyword: target,
			Stage:   "start",
		})
	}

	relatedDetails := []RelatedDetail{}
	for _, item := range relatedBase {
		stats, err := fetchBlogCount(config, item.Keyword, "단일 키워드 검색")
		if err != nil {
			stats = map[string]int{"total": 0}
		}
		item.DocCount = stats["total"]
		relatedDetails = append(relatedDetails, item)
		completed++

		if onProgress != nil {
			onProgress(Progress{
				Current: completed,
				Total:   totalSteps,
				Keyword: item.Keyword,
				Stage:   "related",
			})
		}
	}

	docCount := 0
	stats, err := fetchBlogCount(config, target, "단일 키워드 검색")
	if err == nil {
		docCount = stats["total"]
	}
	completed++

	if onProgress != nil {
		onProgress(Progress{
			Current: completed,
			Total:   totalSteps,
			Keyword: target,
			Stage:   "done",
		})
	}

	volumeGroup := getVolumeGroup(totalSearch)
	groupLabel := "1천 미만"
	if volumeGroup != nil {
		groupLabel = volumeGroup.Label
	}

	scoreParams := ScoreParams{
		SearchPc:       float64(searchPc),
		SearchMobile:   float64(searchMobile),
		DocCount:       float64(docCount),
		RecentDocCount: 0,
		TrendSeries:    []TrendSeriesItem{},
		Params: struct {
			WPC   float64
			WMob  float64
			K     float64
			Alpha float64
			T     float64
			R     float64
		}{
			WPC:   config.Scoring.WPC,
			WMob:  config.Scoring.WMob,
			K:     config.Scoring.K,
			Alpha: config.Scoring.Alpha,
			T:     config.Scoring.T,
			R:     config.Scoring.R,
		},
	}

	scoring := CalculateScore(scoreParams)

	// 원본처럼 타입 변환 (interface{}에서 적절한 타입으로 변환)
	var monthlySearch *int
	if val, ok := searchAdMetrics["monthlySearch"]; ok && val != nil {
		switch v := val.(type) {
		case int:
			monthlySearch = &v
		case float64:
			iv := int(v)
			monthlySearch = &iv
		case float32:
			iv := int(v)
			monthlySearch = &iv
		}
	}

	var monthlyAvgClicks *float64
	if val, ok := searchAdMetrics["monthlyAvgClicks"]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			monthlyAvgClicks = &v
		case float32:
			fv := float64(v)
			monthlyAvgClicks = &fv
		case int:
			fv := float64(v)
			monthlyAvgClicks = &fv
		}
	}

	var monthlyAvgCtr *float64
	if val, ok := searchAdMetrics["monthlyAvgCtr"]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			monthlyAvgCtr = &v
		case float32:
			fv := float64(v)
			monthlyAvgCtr = &fv
		case int:
			fv := float64(v)
			monthlyAvgCtr = &fv
		}
	}

	// 원본처럼 competition 값을 그대로 사용 (변환하지 않음)
	var competition interface{} = nil
	if val, ok := searchAdMetrics["competition"]; ok && val != nil {
		competition = val
	}

	result := &KeywordAnalysisResult{
		Keyword:         target,
		SearchPc:        searchPc,
		SearchMobile:    searchMobile,
		TotalSearch:     totalSearch,
		MonthlySearch:   monthlySearch,
		MonthlyAvgClicks: monthlyAvgClicks,
		MonthlyAvgCtr:   monthlyAvgCtr,
		Competition:     competition,
		DocCount:        docCount,
		Group:           groupLabel,
		RelatedDetails:  relatedDetails,
		Score:           scoring.Score,
	}

	return result, nil
}

func autoExtract(config Config, seeds []string, onProgress func(Progress)) ([]AutoExtractResult, error) {
	if len(seeds) == 0 {
		return []AutoExtractResult{}, nil
	}

	keywordList, err := fetchSearchAdKeywords(config, seeds, "황금 키워드 검색")
	if err != nil {
		return nil, err
	}

	keywordBidMap := make(map[string]float64)
	for _, item := range keywordList {
		label := getKeywordLabel(item)
		if label == "" {
			continue
		}
		bidValue := toNumber(item.MonthlyAveCpc)
		if bidValue == 0 {
			bidValue = toNumber(item.AvgCpc)
		}
		if bidValue == 0 {
			bidValue = toNumber(item.AvgCpcPc)
		}
		if bidValue == 0 {
			bidValue = toNumber(item.AvgCpcMobile)
		}
		if bidValue == 0 {
			bidValue = toNumber(item.BidAmt)
		}
		if bidValue == 0 {
			bidValue = toNumber(item.Bid)
		}
		if bidValue > 0 {
			keywordBidMap[label] = bidValue
		}
	}

	candidates := []struct {
		Item        SearchAdKeywordItem
		Label       string
		SearchPc    int
		SearchMobile int
		TotalSearch int
		Group       *VolumeGroup
	}{}

	for _, item := range keywordList {
		label := getKeywordLabel(item)
		if label == "" {
			continue
		}
		searchPc := int(toNumber(item.MonthlyPcQcCnt))
		searchMobile := int(toNumber(item.MonthlyMobileQcCnt))
		totalSearch := searchPc + searchMobile
		volumeGroup := getVolumeGroup(totalSearch)
		if volumeGroup == nil {
			continue
		}
		if !matchesAnySeed(label, seeds) {
			continue
		}
		candidates = append(candidates, struct {
			Item        SearchAdKeywordItem
			Label       string
			SearchPc    int
			SearchMobile int
			TotalSearch int
			Group       *VolumeGroup
		}{
			Item:        item,
			Label:       label,
			SearchPc:    searchPc,
			SearchMobile: searchMobile,
			TotalSearch:  totalSearch,
			Group:       volumeGroup,
		})
	}

	grouped := map[string][]struct {
		Item        SearchAdKeywordItem
		Label       string
		SearchPc    int
		SearchMobile int
		TotalSearch int
		Group       *VolumeGroup
	}{
		"g1": {},
		"g2": {},
		"g3": {},
		"g4": {},
	}

	for _, entry := range candidates {
		if entry.Group != nil {
			grouped[entry.Group.ID] = append(grouped[entry.Group.ID], entry)
		}
	}

	limited := []struct {
		Item        SearchAdKeywordItem
		Label       string
		SearchPc    int
		SearchMobile int
		TotalSearch int
		Group       *VolumeGroup
	}{}

	for _, groupID := range []string{"g1", "g2", "g3", "g4"} {
		entries := grouped[groupID]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].TotalSearch > entries[j].TotalSearch
		})
		maxCount := 25
		if len(entries) > maxCount {
			entries = entries[:maxCount]
		}
		limited = append(limited, entries...)
	}

	// Limit to 100 total
	if len(limited) > 100 {
		limited = limited[:100]
	}

	if onProgress != nil {
		onProgress(Progress{
			Current: 0,
			Total:   len(limited),
			Keyword: "",
			Stage:   "extract",
		})
	}

	// Collect CPC keywords
	cpcKeywords := []string{}
	for _, entry := range limited {
		cpcKeywords = append(cpcKeywords, entry.Label)
	}

	// Fetch CPC stats and estimates
	cpcStatsMap := make(map[string]KeywordStats)
	if len(cpcKeywords) > 0 {
		stats, err := fetchSearchAdKeywordStats(config, cpcKeywords, "황금 키워드 검색")
		if err == nil {
			cpcStatsMap = stats
		}
	}

	cpcMapPc := make(map[string]CpcEstimate)
	cpcMapMobile := make(map[string]CpcEstimate)
	if len(cpcKeywords) > 0 {
		bidMap := make(map[string]float64)
		for keyword, bid := range keywordBidMap {
			bidMap[keyword] = bid
		}
		pcEst, _ := fetchSearchAdCpcEstimates(config, cpcKeywords, 70, "PC", bidMap, "황금 키워드 검색")
		mobileEst, _ := fetchSearchAdCpcEstimates(config, cpcKeywords, 70, "MOBILE", bidMap, "황금 키워드 검색")
		cpcMapPc = pcEst
		cpcMapMobile = mobileEst
	}

	results := []AutoExtractResult{}
	totalSteps := len(limited)
	completed := 0

	for _, entry := range limited {
		keyword := entry.Label
		searchPc := entry.SearchPc
		searchMobile := entry.SearchMobile
		totalSearch := entry.TotalSearch
		searchAdMetrics := getSearchAdMonthlyMetrics(entry.Item, float64(searchPc), float64(searchMobile))

		docCount := 0
		stats, err := fetchBlogCount(config, keyword, "황금 키워드 검색")
		if err == nil {
			docCount = stats["total"]
		}

		scoreParams := ScoreParams{
			SearchPc:       float64(searchPc),
			SearchMobile:   float64(searchMobile),
			DocCount:       float64(docCount),
			RecentDocCount: 0,
			TrendSeries:    []TrendSeriesItem{},
			Params: struct {
				WPC   float64
				WMob  float64
				K     float64
				Alpha float64
				T     float64
				R     float64
			}{
				WPC:   config.Scoring.WPC,
				WMob:  config.Scoring.WMob,
				K:     config.Scoring.K,
				Alpha: config.Scoring.Alpha,
				T:     config.Scoring.T,
				R:     config.Scoring.R,
			},
		}

		scoring := CalculateScore(scoreParams)

		// 원본처럼 타입 변환 (interface{}에서 적절한 타입으로 변환)
		var monthlySearch *int
		if val, ok := searchAdMetrics["monthlySearch"]; ok && val != nil {
			switch v := val.(type) {
			case int:
				monthlySearch = &v
			case float64:
				iv := int(v)
				monthlySearch = &iv
			case float32:
				iv := int(v)
				monthlySearch = &iv
			}
		}

		var monthlyAvgClicks *float64
		if val, ok := searchAdMetrics["monthlyAvgClicks"]; ok && val != nil {
			switch v := val.(type) {
			case float64:
				monthlyAvgClicks = &v
			case float32:
				fv := float64(v)
				monthlyAvgClicks = &fv
			case int:
				fv := float64(v)
				monthlyAvgClicks = &fv
			}
		}

		var monthlyAvgCtr *float64
		if val, ok := searchAdMetrics["monthlyAvgCtr"]; ok && val != nil {
			switch v := val.(type) {
			case float64:
				monthlyAvgCtr = &v
			case float32:
				fv := float64(v)
				monthlyAvgCtr = &fv
			case int:
				fv := float64(v)
				monthlyAvgCtr = &fv
			}
		}

		// 원본처럼 competition 값을 그대로 사용 (변환하지 않음)
		var competition interface{} = nil
		if val, ok := searchAdMetrics["competition"]; ok && val != nil {
			competition = val
			log.Printf("=== autoExtract: competition ===\nKeyword: %s\n값: %v (타입: %T)\n", keyword, competition, competition)
		} else {
			log.Printf("=== autoExtract: competition 없음 ===\nKeyword: %s\nsearchAdMetrics[\"competition\"] 존재: %v\n", keyword, ok)
		}

		// Get CPC stats
		statsInfo, hasStats := cpcStatsMap[keyword]
		statsPcClicks := 0.0
		statsPcCost := 0.0
		statsMobileClicks := 0.0
		statsMobileCost := 0.0
		statsTotalClicks := 0.0
		statsTotalCost := 0.0
		if hasStats {
			statsPcClicks = statsInfo.PC.Clicks
			statsPcCost = statsInfo.PC.Cost
			statsMobileClicks = statsInfo.Mobile.Clicks
			statsMobileCost = statsInfo.Mobile.Cost
			statsTotalClicks = statsInfo.Total.Clicks
			statsTotalCost = statsInfo.Total.Cost
		}

		// Get CPC estimates
		estPcInfo, hasEstPc := cpcMapPc[keyword]
		estMobileInfo, hasEstMobile := cpcMapMobile[keyword]
		estPcClicks := 0.0
		estPcCost := 0.0
		estMobileClicks := 0.0
		estMobileCost := 0.0
		if hasEstPc {
			estPcClicks = estPcInfo.Clicks
			estPcCost = estPcInfo.Cost
		}
		if hasEstMobile {
			estMobileClicks = estMobileInfo.Clicks
			estMobileCost = estMobileInfo.Cost
		}

		hasStatsDevice := statsPcClicks+statsMobileClicks > 0
		hasStatsTotal := statsTotalClicks > 0
		hasEstDevice := estPcClicks+estMobileClicks > 0

		pcClicks := statsPcClicks
		pcCost := statsPcCost
		if pcClicks == 0 && hasEstDevice {
			pcClicks = estPcClicks
			pcCost = estPcCost
		}

		mobileClicks := statsMobileClicks
		mobileCost := statsMobileCost
		if mobileClicks == 0 && hasEstDevice {
			mobileClicks = estMobileClicks
			mobileCost = estMobileCost
		}

		totalClicks := 0.0
		totalCost := 0.0
		if hasStatsDevice {
			totalClicks = statsPcClicks + statsMobileClicks
			totalCost = statsPcCost + statsMobileCost
		} else if hasStatsTotal {
			totalClicks = statsTotalClicks
			totalCost = statsTotalCost
		} else if hasEstDevice {
			totalClicks = estPcClicks + estMobileClicks
			totalCost = estPcCost + estMobileCost
		}

		pcCpc := (*float64)(nil)
		if pcClicks > 0 {
			val := pcCost / pcClicks
			pcCpc = &val
		}

		mobileCpc := (*float64)(nil)
		if mobileClicks > 0 {
			val := mobileCost / mobileClicks
			mobileCpc = &val
		}

		weightedCpc := (*float64)(nil)
		if totalClicks > 0 {
			val := totalCost / totalClicks
			weightedCpc = &val
		}

		groupLabel := "1천 미만"
		if entry.Group != nil {
			groupLabel = entry.Group.Label
		}

		result := AutoExtractResult{
			Keyword:         keyword,
			SearchPc:        searchPc,
			SearchMobile:    searchMobile,
			TotalSearch:     totalSearch,
			MonthlySearch:   monthlySearch,
			MonthlyAvgClicks: monthlyAvgClicks,
			MonthlyAvgCtr:   monthlyAvgCtr,
			Competition:     competition,
			DocCount:        docCount,
			Group:           groupLabel,
			Score:           scoring.Score,
			Cpc:             weightedCpc,
			CpcPc:           pcCpc,
			CpcMobile:       mobileCpc,
			PcClicks:        pcClicks,
			PcCost:          pcCost,
			MobileClicks:    mobileClicks,
			MobileCost:      mobileCost,
			TotalClicks:     totalClicks,
			TotalCost:       totalCost,
		}

		// 결과 로그 출력
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		log.Printf("=== autoExtract 결과 항목 ===\nKeyword: %s\n%s\n", keyword, string(resultJSON))

		results = append(results, result)

		completed++
		if onProgress != nil {
			onProgress(Progress{
				Current: completed,
				Total:   totalSteps,
				Keyword: keyword,
				Stage:   "extract",
			})
		}
	}

	return results, nil
}
