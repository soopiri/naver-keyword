package main

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// collectCreatorAdvisor collects creator advisor data using Playwright
func collectCreatorAdvisor(payload CreatorAdvisorPayload, onProgress func(Progress)) (*CreatorAdvisorResult, error) {
	loginID := strings.TrimSpace(payload.LoginID)
	password := strings.TrimSpace(payload.Password)

	if loginID == "" || password == "" {
		return nil, fmt.Errorf("네이버 ID/PW를 입력하세요")
	}

	// Playwright 드라이버가 없으면 설치 시도
	if err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	}); err != nil {
		// 설치 실패해도 계속 진행 (이미 설치되어 있을 수 있음)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright 초기화 실패: %w", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(!payload.ShowBrowser),
	})
	if err != nil {
		return nil, fmt.Errorf("브라우저 실행 실패: %w", err)
	}
	defer browser.Close()

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1280,
			Height: 800,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("컨텍스트 생성 실패: %w", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("페이지 생성 실패: %w", err)
	}

	// Step 1: 네이버 로그인
	if onProgress != nil {
		onProgress(Progress{
			Current: 0,
			Total:   6,
			Message: "네이버 로그인 중...",
		})
	}

	if err := safeGoto(page, "https://nid.naver.com/nidlogin.login"); err != nil {
		return nil, fmt.Errorf("로그인 페이지 접속 실패: %w", err)
	}

	if err := page.Fill("#id", loginID); err != nil {
		return nil, fmt.Errorf("ID 입력 실패: %w", err)
	}

	if err := page.Fill("#pw", password); err != nil {
		return nil, fmt.Errorf("비밀번호 입력 실패: %w", err)
	}

	if err := page.Click("#log\\.login"); err != nil {
		return nil, fmt.Errorf("로그인 버튼 클릭 실패: %w", err)
	}

	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return nil, fmt.Errorf("페이지 로딩 대기 실패: %w", err)
	}

	currentURL := page.URL()

	if strings.Contains(currentURL, "nidlogin.login") {
		if onProgress != nil {
			onProgress(Progress{
				Current: 0,
				Total:   6,
				Message: "추가 인증 필요. 브라우저에서 인증 완료까지 대기 중...",
			})
		}

		// URL이 변경될 때까지 대기 (최대 10초)
		timeout := time.After(10 * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				return nil, fmt.Errorf("로그인 실패 또는 추가 인증이 필요합니다")
			case <-ticker.C:
				currentURL := page.URL()
				if !strings.Contains(currentURL, "nidlogin.login") {
					goto loginSuccess
				}
			}
		}
	}

loginSuccess:
	// Step 2: 블로그 홈 접속
	if onProgress != nil {
		onProgress(Progress{
			Current: 1,
			Total:   6,
			Message: "블로그 홈 접속 중...",
		})
	}

	if err := safeGoto(page, "https://section.blog.naver.com/BlogHome.naver?directoryNo=0&currentPage=1&groupId=0"); err != nil {
		return nil, fmt.Errorf("블로그 홈 접속 실패: %w", err)
	}

	// Step 3: 내 블로그 접속
	if onProgress != nil {
		onProgress(Progress{
			Current: 2,
			Total:   6,
			Message: "내 블로그 접속 중...",
		})
	}

	myBlogPage, err := clickMaybePopup(page, `a[href*="MyBlog.naver"]`)
	if err != nil {
		myBlogPage = page
	}

	if err := safeGoto(myBlogPage, "https://blog.naver.com/MyBlog.naver"); err != nil {
		return nil, fmt.Errorf("내 블로그 접속 실패: %w", err)
	}

	currentURL = myBlogPage.URL()
	blogID := extractBlogIdFromURL(currentURL)

	// Step 4: 통계 접속
	if onProgress != nil {
		onProgress(Progress{
			Current: 3,
			Total:   6,
			Message: "통계 접속 중...",
		})
	}

	if blogID != "" {
		if err := safeGoto(myBlogPage, fmt.Sprintf("https://admin.blog.naver.com/%s/stat/today", blogID)); err != nil {
			return nil, fmt.Errorf("통계 페이지 접속 실패: %w", err)
		}
	} else {
		statsLink := myBlogPage.Locator(`a[href*="/stat/today"]`).First()
		count, _ := statsLink.Count()
		if count > 0 {
			href, _ := statsLink.GetAttribute("href")
			if href != "" {
				if err := safeGoto(myBlogPage, href); err != nil {
					// Try clicking instead
					statsLink.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
					myBlogPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
						State: playwright.LoadStateDomcontentloaded,
					})
				}
			} else {
				statsLink.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
				myBlogPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
					State: playwright.LoadStateDomcontentloaded,
				})
			}
		}
	}

	// Step 5: 크리에이터 어드바이저 접속
	if onProgress != nil {
		onProgress(Progress{
			Current: 4,
			Total:   6,
			Message: "크리에이터 어드바이저 접속 중...",
		})
	}

	advisorPage := myBlogPage
	if blogID != "" {
		if err := safeGoto(advisorPage, fmt.Sprintf("https://creator-advisor.naver.com/naver_blog/%s/trends", blogID)); err != nil {
			return nil, fmt.Errorf("크리에이터 어드바이저 접속 실패: %w", err)
		}
	} else {
		var err error
		advisorPage, err = clickMaybePopup(myBlogPage, `a[href*="creator-advisor.naver.com/naver_blog/"]`)
		if err != nil {
			advisorPage = myBlogPage
		}
	}

	// Step 6: 트렌드 탭 이동 및 데이터 수집
	if onProgress != nil {
		onProgress(Progress{
			Current: 5,
			Total:   6,
			Message: "트렌드 탭 이동 중...",
		})
	}

	if err := advisorPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	}); err != nil {
		return nil, fmt.Errorf("페이지 로딩 대기 실패: %w", err)
	}

	if onProgress != nil {
		onProgress(Progress{
			Current: 5,
			Total:   6,
			Message: "검색 유입 트렌드 수집 중...",
			Detail:  "슬라이드 이동 준비 중...",
		})
	}

	searchInflowPayload, err := collectSearchInflowTrends(advisorPage, func(progress Progress) {
		if onProgress != nil && progress.Step == "search-inflow" {
			onProgress(Progress{
				Current: 5,
				Total:   6,
				Message: "검색 유입 트렌드 수집 중...",
				Detail:  progress.Detail,
			})
		}
	})
	if err != nil {
		return nil, fmt.Errorf("검색 유입 트렌드 수집 실패: %w", err)
	}

	mainInflowContents, err := collectMainInflow(advisorPage)
	if err != nil {
		return nil, fmt.Errorf("메인 유입 콘텐츠 수집 실패: %w", err)
	}

	if onProgress != nil {
		onProgress(Progress{
			Current: 6,
			Total:   6,
			Message: "수집 완료",
		})
	}

	currentURL = advisorPage.URL()
	return &CreatorAdvisorResult{
		URL:                currentURL,
		SearchInflowTrends: searchInflowPayload.Results,
		SearchInflowMessage: searchInflowPayload.Message,
		MainInflowContents: mainInflowContents,
	}, nil
}

// Helper functions

func safeGoto(page playwright.Page, url string) error {
	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return err
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	return nil
}

func clickMaybePopup(page playwright.Page, selector string) (playwright.Page, error) {
	target := page.Locator(selector).First()
	count, err := target.Count()
	if err != nil || count == 0 {
		return page, fmt.Errorf("selector not found")
	}

	href, _ := target.GetAttribute("href")
	isVisible, _ := target.IsVisible()

	if !isVisible && href != "" {
		if err := safeGoto(page, href); err != nil {
			return page, err
		}
		return page, nil
	}

	// Try to click and wait for popup
	popupChan := make(chan playwright.Page, 1)
	go func() {
		popup, err := page.WaitForEvent("popup", playwright.PageWaitForEventOptions{
			Timeout: playwright.Float(3000),
		})
		if err == nil {
			popupChan <- popup.(playwright.Page)
		} else {
			popupChan <- nil
		}
	}()

	if err := target.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)}); err != nil {
		if href != "" {
			if err := safeGoto(page, href); err != nil {
				return page, err
			}
			return page, nil
		}
		return page, err
	}

	select {
	case popup := <-popupChan:
		if popup != nil {
			popup.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State: playwright.LoadStateDomcontentloaded,
			})
			return popup, nil
		}
		page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateDomcontentloaded,
		})
		return page, nil
	default:
		page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateDomcontentloaded,
		})
		return page, nil
	}
}

func extractBlogIdFromURL(pageURL string) string {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		// Try regex fallback
		re := regexp.MustCompile(`blogId=([^&]+)`)
		matches := re.FindStringSubmatch(pageURL)
		if len(matches) > 1 {
			return matches[1]
		}
		return ""
	}

	blogID := parsedURL.Query().Get("blogId")
	if blogID != "" {
		return blogID
	}

	if parsedURL.Hostname() == "blog.naver.com" {
		parts := strings.FieldsFunc(parsedURL.Path, func(c rune) bool {
			return c == '/'
		})
		if len(parts) > 0 && parts[0] != "PostList.naver" {
			return parts[0]
		}
	}

	re := regexp.MustCompile(`blogId=([^&]+)`)
	matches := re.FindStringSubmatch(pageURL)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// collectSearchInflowTrends collects search inflow trends
func collectSearchInflowTrends(page playwright.Page, onProgress func(Progress)) (*struct {
	Results []SearchInflowTrend `json:"results"`
	Message string              `json:"message"`
	NoData  bool                `json:"noData"`
}, error) {
	if err := clickTrendTab(page, "검색 유입 트렌드"); err != nil {
		return nil, fmt.Errorf("검색 유입 트렌드 탭 클릭 실패: %w", err)
	}

	// Collect topic trends
	topicClicked, err := clickTrendMenu(page, []string{"주제별 인기유입검색어", "주제별 인기 유입 검색어"})
	if err != nil || !topicClicked {
		return nil, fmt.Errorf("주제별 인기유입검색어 메뉴를 찾지 못했습니다")
	}

	waitForTrendMenuActive(page, []string{"주제별 인기유입검색어", "주제별 인기 유입 검색어"})

	hasNoData, _ := hasTrendNoDataMessage(page)
	if hasNoData {
		return &struct {
			Results []SearchInflowTrend `json:"results"`
			Message string              `json:"message"`
			NoData  bool                `json:"noData"`
		}{
			Results: []SearchInflowTrend{},
			Message: "00시 기준 데이터 집계 중으로, 현재 수집된 데이터는 없습니다.",
			NoData:  true,
		}, nil
	}

	topicSwiperIndex, _ := findTrendSwiperIndex(page, "topic")
	prevTitle, _ := getFirstTrendTitle(page, topicSwiperIndex)

	clickTrendButton(page, "설정순 보기")
	waitForTrendTitleChange(page, prevTitle, topicSwiperIndex)
	resetTrendSwiper(page, topicSwiperIndex)

	if onProgress != nil {
		onProgress(Progress{
			Step:   "search-inflow",
			Detail: "주제별 슬라이드 이동 준비 중...",
		})
	}

	topicResults, _ := collectTrendSwiper(page, "주제별 인기유입검색어", 6, topicSwiperIndex, false)

	// Collect demographic trends
	log.Printf("[collectSearchInflowTrends] Starting demographic trends collection")
	demoClicked, err := clickTrendMenu(page, []string{"성별, 연령별 인기유입검색어", "성별,연령별 인기유입검색어"})
	if err != nil || !demoClicked {
		log.Printf("[collectSearchInflowTrends] Failed to click demographic menu: err=%v, clicked=%v", err, demoClicked)
		return nil, fmt.Errorf("성별, 연령별 인기유입검색어 메뉴를 찾지 못했습니다")
	}
	log.Printf("[collectSearchInflowTrends] Demographic menu clicked successfully")

	waitForTrendMenuActive(page, []string{"성별, 연령별 인기유입검색어", "성별,연령별 인기유입검색어"})
	
	// Wait for page to update after menu click
	page.WaitForTimeout(500)
	log.Printf("[collectSearchInflowTrends] Waited for menu activation")

	hasNoData, _ = hasTrendNoDataMessage(page)
	if hasNoData {
		log.Printf("[collectSearchInflowTrends] Demographic trends: no data message found")
		return &struct {
			Results []SearchInflowTrend `json:"results"`
			Message string              `json:"message"`
			NoData  bool                `json:"noData"`
		}{
			Results: []SearchInflowTrend{},
			Message: "00시 기준 데이터 집계 중으로, 현재 수집된 데이터는 없습니다.",
			NoData:  true,
		}, nil
	}

	scrollToTrendSection(page, "성별, 연령별 인기유입검색어")
	ensureDemographicVisible(page)
	
	// Additional wait after scrolling
	page.WaitForTimeout(500)
	log.Printf("[collectSearchInflowTrends] Scrolled to demographic section")

	// Debug: Check all swipers
	swiperCount, _ := page.Evaluate(`() => {
		return document.querySelectorAll(".u_ni_search_swiper").length;
	}`, nil)
	log.Printf("[collectSearchInflowTrends] Total swipers found: %v", swiperCount)
	
	// Debug: Check all titles
	titlesResult, _ := page.Evaluate(`() => {
		const titles = Array.from(document.querySelectorAll("h3.u_ni_trend_title"));
		return titles.map(el => el.textContent?.trim() || "").filter(Boolean);
	}`, nil)
	log.Printf("[collectSearchInflowTrends] All titles found: %v", titlesResult)

	demoSwiperIndex, err := findTrendSwiperIndex(page, "demographic")
	if err != nil {
		log.Printf("[collectSearchInflowTrends] Error finding demographic swiper index: %v", err)
	}
	log.Printf("[collectSearchInflowTrends] Demographic swiper index: %d", demoSwiperIndex)
	
	// Verify the swiper index by checking its first title
	if demoSwiperIndex >= 0 {
		firstTitle, _ := getFirstTrendTitle(page, demoSwiperIndex)
		log.Printf("[collectSearchInflowTrends] First title in swiper[%d]: '%s'", demoSwiperIndex, firstTitle)
	}
	resetTrendSwiper(page, demoSwiperIndex)

	if onProgress != nil {
		onProgress(Progress{
			Step:   "search-inflow",
			Detail: "성별/연령 슬라이드 이동 준비 중...",
		})
	}

	log.Printf("[collectSearchInflowTrends] Calling collectTrendSwiper for demographic trends: maxCategories=0 (collect all), swiperIndex=%d", demoSwiperIndex)
	demoResults, err := collectTrendSwiper(page, "성별,연령별 인기유입검색어", 0, demoSwiperIndex, false)
	if err != nil {
		log.Printf("[collectSearchInflowTrends] Error collecting demographic trends: %v", err)
	} else {
		log.Printf("[collectSearchInflowTrends] Demographic trends collected: %d results", len(demoResults))
		for i, trend := range demoResults {
			log.Printf("[collectSearchInflowTrends] Demographic result[%d]: category='%s', items=%d", 
				i, trend.Category, len(trend.Items))
		}
	}

	allResults := append(topicResults, demoResults...)
	log.Printf("[collectSearchInflowTrends] Combined results: topic=%d, demographic=%d, total=%d", 
		len(topicResults), len(demoResults), len(allResults))

	// No filtering - include all demographic results
	filteredResults := []SearchInflowTrend{}
	demoCount := 0
	for _, group := range allResults {
		source := strings.ReplaceAll(group.Source, " ", "")
		if source == "성별,연령별인기유입검색어" {
			demoCount++
			log.Printf("[collectSearchInflowTrends] Demographic group[%d]: category='%s', items=%d", 
				demoCount, group.Category, len(group.Items))
		}
		filteredResults = append(filteredResults, group)
	}
	log.Printf("[collectSearchInflowTrends] Filtered results: total=%d (demographic count=%d)", 
		len(filteredResults), demoCount)

	hasItems := false
	for _, group := range filteredResults {
		if len(group.Items) > 0 {
			hasItems = true
			break
		}
	}

	if !hasItems {
		hasNoData, _ = hasTrendNoDataMessage(page)
		if hasNoData {
			return &struct {
				Results []SearchInflowTrend `json:"results"`
				Message string              `json:"message"`
				NoData  bool                `json:"noData"`
			}{
				Results: []SearchInflowTrend{},
				Message: "00시 기준 데이터 집계 중으로, 현재 수집된 데이터는 없습니다.",
				NoData:  true,
			}, nil
		}
	}

	return &struct {
		Results []SearchInflowTrend `json:"results"`
		Message string              `json:"message"`
		NoData  bool                `json:"noData"`
	}{
		Results: filteredResults,
		Message: "",
		NoData:  false,
	}, nil
}

// collectMainInflow collects main inflow content
func collectMainInflow(page playwright.Page) ([]MainInflowContent, error) {
	if err := clickTrendTab(page, "메인 유입 트렌드"); err != nil {
		return nil, fmt.Errorf("메인 유입 트렌드 탭 클릭 실패: %w", err)
	}

	page.WaitForSelector("ul.u_ni_realtime_list", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(10000),
	})

	list := page.Locator("ul.u_ni_realtime_list li.u_ni_realtime_item")
	prevCount := -1

	// Click "더보기" button up to 6 times
	for i := 0; i < 6; i++ {
		count, _ := list.Count()
		moreBtn := page.Locator("button.u_ni_btn_more").First()
		moreBtnCount, _ := moreBtn.Count()

		if count == prevCount || moreBtnCount == 0 {
			break
		}

		prevCount = count
		isVisible, _ := moreBtn.IsVisible()
		if isVisible {
			moreBtn.Click()
			page.WaitForTimeout(600)
		} else {
			break
		}
	}

	results := []MainInflowContent{}
	finalCount, _ := list.Count()

	for i := 0; i < finalCount; i++ {
		item := list.Nth(i)
		link := item.Locator("a.u_ni_realtime_link").First()
		linkURL := ""
		if linkCount, _ := link.Count(); linkCount > 0 {
			linkURL, _ = link.GetAttribute("href")
		}

		rankEl := item.Locator(".u_ni_rank_num").First()
		rankText := ""
		if rankCount, _ := rankEl.Count(); rankCount > 0 {
			rankText, _ = rankEl.InnerText()
			rankText = strings.TrimSpace(rankText)
		}

		titleEl := item.Locator(".u_ni_rank_text").First()
		title := ""
		if titleCount, _ := titleEl.Count(); titleCount > 0 {
			title, _ = titleEl.InnerText()
			title = strings.TrimSpace(title)
		}

		if title == "" {
			continue
		}

		rank := 0
		if rankText != "" {
			if parsedRank, err := strconv.Atoi(rankText); err == nil {
				rank = parsedRank
			} else {
				rank = i + 1
			}
		} else {
			rank = i + 1
		}

		results = append(results, MainInflowContent{
			Rank:  rank,
			Title: title,
			URL:   linkURL,
		})
	}

	return results, nil
}

func clickTrendTab(page playwright.Page, label string) error {
	tab := page.Locator("div.u_ni_link").Filter(playwright.LocatorFilterOptions{
		HasText: label,
	}).First()

	count, err := tab.Count()
	if err != nil || count == 0 {
		return fmt.Errorf("탭을 찾을 수 없습니다: %s", label)
	}

	tab.ScrollIntoViewIfNeeded()
	if err := tab.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)}); err != nil {
		// Try force click
		tab.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(1000),
			Force:   playwright.Bool(true),
		})
	}

	page.WaitForTimeout(300)
	return nil
}
