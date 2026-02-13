package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Helper functions for Playwright automation

func clickTrendMenu(page playwright.Page, labels []string) (bool, error) {
	normalizedTargets := make([]string, len(labels))
	for i, label := range labels {
		normalizedTargets[i] = strings.ReplaceAll(strings.TrimSpace(label), " ", "")
	}

	clicked, err := page.Evaluate(`(targets) => {
		const items = Array.from(document.querySelectorAll("span.u_ni_menu_txt, a.u_ni_link"));
		const normalize = (value) => String(value || "").replace(/\s/g, "");
		for (const el of items) {
			const text = normalize(el.textContent || "");
			if (targets.includes(text)) {
				const clickable = el.closest("a.u_ni_link") || el;
				(clickable instanceof HTMLElement ? clickable : clickable.parentElement)?.click();
				return true;
			}
		}
		return false;
	}`, normalizedTargets)

	if err != nil {
		return false, err
	}

	if clicked, ok := clicked.(bool); ok && clicked {
		page.WaitForTimeout(250)
		return true, nil
	}

	return false, nil
}

func waitForTrendMenuActive(page playwright.Page, labels []string) error {
	normalizedTargets := make([]string, len(labels))
	for i, label := range labels {
		normalizedTargets[i] = strings.ReplaceAll(strings.TrimSpace(label), " ", "")
	}

	_, err := page.WaitForFunction(`(targets) => {
		const items = Array.from(document.querySelectorAll("span.u_ni_menu_txt, a.u_ni_link"));
		const normalize = (value) => String(value || "").replace(/\s/g, "");
		return items.some((el) => {
			const text = normalize(el.textContent || "");
			if (!targets.includes(text)) return false;
			const candidate =
				el.closest("a.u_ni_link")?.classList ||
				el.classList ||
				el.closest(".u_ni_active, .u_ni_on, .u_ni_selected")?.classList;
			if (!candidate) return false;
			return candidate.contains("u_ni_active") || candidate.contains("u_ni_on");
		});
	}`, normalizedTargets, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(4000),
	})

	return err
}

func hasTrendNoDataMessage(page playwright.Page) (bool, error) {
	locator := page.Locator(".u_ni_nodata_section .u_ni_desc").Filter(playwright.LocatorFilterOptions{
		HasText: "조회한 기간의 데이터가 없습니다",
	})
	count, err := locator.Count()
	return count > 0, err
}

func findTrendSwiperIndex(page playwright.Page, kind string) (int, error) {
	normalizedKind := strings.ToLower(kind)
	log.Printf("[findTrendSwiperIndex] Looking for kind='%s'", normalizedKind)

	result, err := page.Evaluate(`(targetKind) => {
		const isDemographicTitle = (text) => {
			const value = String(text || "");
			return /\d+\s*-\s*\d+\s*세/.test(value) || /\d+\s*세/.test(value) || /세\s*(남자|여자)/.test(value);
		};
		const swipers = Array.from(document.querySelectorAll(".u_ni_search_swiper"));
		console.log("Total swipers found:", swipers.length);

		const findByTitle = (wantDemographic) => {
			const titles = Array.from(document.querySelectorAll("h3.u_ni_trend_title"));
			console.log("Total titles found:", titles.length);
			for (const titleEl of titles) {
				const text = titleEl.textContent || "";
				const isDemo = isDemographicTitle(text);
				console.log("Title:", text, "isDemographic:", isDemo, "wantDemographic:", wantDemographic);
				if (wantDemographic !== isDemo) continue;
				const swiper = titleEl.closest(".u_ni_search_swiper");
				if (!swiper) continue;
				const index = swipers.indexOf(swiper);
				if (index >= 0) {
					console.log("Found swiper index:", index);
					return index;
				}
			}
			return -1;
		};

		if (targetKind === "demographic") {
			const index = findByTitle(true);
			if (index >= 0) return index;
		}
		if (targetKind === "topic") {
			const index = findByTitle(false);
			if (index >= 0) return index;
		}

		for (let i = 0; i < swipers.length; i += 1) {
			const swiper = swipers[i];
			const titles = Array.from(swiper.querySelectorAll("h3.u_ni_trend_title"))
				.map((el) => el.textContent || "")
				.filter(Boolean);
			if (titles.length === 0) continue;
			const hasDemographic = titles.some((title) => isDemographicTitle(title));
			console.log("Swiper", i, "titles:", titles, "hasDemographic:", hasDemographic);
			if (targetKind === "demographic" && hasDemographic) return i;
			if (targetKind === "topic" && !hasDemographic) return i;
		}

		console.log("Returning default index 0");
		return 0;
	}`, normalizedKind)

	if err != nil {
		log.Printf("[findTrendSwiperIndex] Error evaluating: %v", err)
		return 0, err
	}

	// Handle different numeric types from JavaScript
	var idx int
	switch v := result.(type) {
	case float64:
		idx = int(v)
		log.Printf("[findTrendSwiperIndex] Found swiper index (float64): %d", idx)
	case int:
		idx = v
		log.Printf("[findTrendSwiperIndex] Found swiper index (int): %d", idx)
	case int64:
		idx = int(v)
		log.Printf("[findTrendSwiperIndex] Found swiper index (int64): %d", idx)
	default:
		log.Printf("[findTrendSwiperIndex] Result type: %T, value: %v", result, result)
		return 0, nil
	}

	return idx, nil
}

func getFirstTrendTitle(page playwright.Page, swiperIndex int) (string, error) {
	titleEl := page.Locator(".u_ni_search_swiper").
		Nth(swiperIndex).
		Locator(".swiper-slide-active h3.u_ni_trend_title").
		First()

	count, err := titleEl.Count()
	if err != nil || count == 0 {
		return "", nil
	}

	text, err := titleEl.InnerText()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(text), nil
}

func waitForTrendTitleChange(page playwright.Page, prevTitle string, swiperIndex int) bool {
	start := time.Now()
	for time.Since(start) < 2*time.Second {
		title, _ := getFirstTrendTitle(page, swiperIndex)
		if title != "" && title != prevTitle {
			return true
		}
		page.WaitForTimeout(120)
	}
	return false
}

func clickTrendButton(page playwright.Page, label string) (bool, error) {
	btn := page.Locator("button.u_ni_tab_btn").Filter(playwright.LocatorFilterOptions{
		HasText: label,
	}).First()

	count, err := btn.Count()
	if err != nil || count == 0 {
		return false, nil
	}

	btn.ScrollIntoViewIfNeeded()
	if err := btn.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)}); err != nil {
		btn.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(1000),
			Force:   playwright.Bool(true),
		})
	}

	page.WaitForTimeout(300)
	return true, nil
}

func resetTrendSwiper(page playwright.Page, swiperIndex int) error {
	_, err := page.Evaluate(`(targetIndex) => {
		const list = document.querySelectorAll(".u_ni_search_swiper");
		const el = list?.[targetIndex];
		const swiper = el && el.swiper;
		if (swiper && typeof swiper.slideTo === "function") {
			swiper.slideTo(0, 0);
		}
	}`, swiperIndex)
	return err
}

func getTrendSwiperLength(page playwright.Page, swiperIndex int) (int, error) {
	result, err := page.Evaluate(`(targetIndex) => {
		const list = document.querySelectorAll(".u_ni_search_swiper");
		const el = list?.[targetIndex];
		const swiper = el && el.swiper;
		if (swiper && Array.isArray(swiper.slides)) return swiper.slides.length;
		return 0;
	}`, swiperIndex)

	if err != nil {
		return 0, err
	}

	if length, ok := result.(float64); ok {
		return int(length), nil
	}

	return 0, nil
}

func moveTrendSwiperTo(page playwright.Page, index int, swiperIndex int) (bool, error) {
	moved, err := page.Evaluate(`({ idx, swiperIndex: targetIndex }) => {
		const list = document.querySelectorAll(".u_ni_search_swiper");
		const el = list?.[targetIndex];
		const swiper = el && el.swiper;
		if (swiper && typeof swiper.slideTo === "function") {
			swiper.slideTo(idx, 0);
			return true;
		}
		const btn = el?.querySelector(".swiper-button-next");
		if (btn && btn instanceof HTMLElement) {
			btn.click();
			return true;
		}
		return false;
	}`, map[string]interface{}{
		"idx":         index,
		"swiperIndex": swiperIndex,
	})

	if err != nil {
		return false, err
	}

	if moved, ok := moved.(bool); ok && moved {
		return true, nil
	}

	if index == 0 {
		return true, nil
	}

	swiper := page.Locator(".u_ni_search_swiper").Nth(swiperIndex)
	box, err := swiper.BoundingBox()
	if err != nil || box == nil {
		return false, nil
	}

	startX := box.X + box.Width*0.75
	endX := box.X + box.Width*0.25
	y := box.Y + box.Height*0.5

	mouse := page.Mouse()
	if err := mouse.Move(startX, y); err != nil {
		return false, err
	}
	if err := mouse.Down(); err != nil {
		return false, err
	}
	if err := mouse.Move(endX, y, playwright.MouseMoveOptions{Steps: playwright.Int(8)}); err != nil {
		return false, err
	}
	if err := mouse.Up(); err != nil {
		return false, err
	}

	return true, nil
}

func waitForTrendItemsReady(page playwright.Page, swiperIndex int, minCount int) bool {
	start := time.Now()
	timeout := 900 * time.Millisecond
	if minCount > 0 {
		timeout = 2500 * time.Millisecond
	}

	lastCount := -1
	stableRounds := 0

	for time.Since(start) < timeout {
		hasItemsLocator := page.Locator(".u_ni_search_swiper").
			Nth(swiperIndex).
			Locator(".swiper-slide-active li.u_ni_trend_item")
		hasItems, _ := hasItemsLocator.Count()

		isLoadingLocator := page.Locator(".u_ni_search_swiper").
			Nth(swiperIndex).
			Locator(".swiper-slide-active .u_ni_nodata_wrap")
		isLoading, _ := isLoadingLocator.Count()
		isEmpty := isLoading > 0

		if !isEmpty && hasItems > 0 {
			if minCount > 0 && hasItems >= minCount {
				return true
			}
			if hasItems == lastCount {
				stableRounds++
				if stableRounds >= 2 {
					return true
				}
			} else {
				stableRounds = 0
				lastCount = hasItems
			}
		}

		page.WaitForTimeout(80)
	}

	return false
}

func expandTrendSlideItems(page playwright.Page, swiperIndex int) error {
	_, err := page.Evaluate(`(targetIndex) => {
		const list = document.querySelectorAll(".u_ni_search_swiper");
		const el = list?.[targetIndex];
		const slide = el?.querySelector(".swiper-slide-active");
		if (!slide) return;
		const listWrap = slide.querySelector(".u_ni_item_list");
		if (listWrap) {
			listWrap.style.maxHeight = "none";
			listWrap.style.height = "auto";
			listWrap.style.overflow = "visible";
			listWrap.scrollTop = listWrap.scrollHeight;
			listWrap.scrollTop = 0;
		}
		const container = slide.querySelector(".u_ni_trend_list");
		if (container) {
			container.style.maxHeight = "none";
			container.style.height = "auto";
			container.style.overflow = "visible";
			container.scrollTop = container.scrollHeight;
			container.scrollTop = 0;
		}
	}`, swiperIndex)
	return err
}

func expandAllTrendSlides(page playwright.Page, swiperIndex int) error {
	_, err := page.Evaluate(`(targetIndex) => {
		const list = document.querySelectorAll(".u_ni_search_swiper");
		const el = list?.[targetIndex];
		if (!el) return;
		const slides = Array.from(el.querySelectorAll(".swiper-slide"));
		slides.forEach((slide) => {
			const listWrap = slide.querySelector(".u_ni_item_list");
			if (listWrap) {
				listWrap.style.maxHeight = "none";
				listWrap.style.height = "auto";
				listWrap.style.overflow = "visible";
				listWrap.scrollTop = listWrap.scrollHeight;
			}
			const container = slide.querySelector(".u_ni_trend_list");
			if (container) {
				container.style.maxHeight = "none";
				container.style.height = "auto";
				container.style.overflow = "visible";
				container.scrollTop = container.scrollHeight;
			}
		});
	}`, swiperIndex)
	return err
}

func readActiveTrendSlide(page playwright.Page, swiperIndex int) (*struct {
	Category string
	Items    []TrendItem
}, error) {
	slide := page.Locator(".u_ni_search_swiper").
		Nth(swiperIndex).
		Locator(".swiper-slide-active").
		First()

	count, err := slide.Count()
	if err != nil || count == 0 {
		log.Printf("[readActiveTrendSlide] swiperIndex=%d: slide not found", swiperIndex)
		return nil, nil
	}

	titleEl := slide.Locator("h3.u_ni_trend_title").First()
	titleCount, _ := titleEl.Count()
	if titleCount == 0 {
		log.Printf("[readActiveTrendSlide] swiperIndex=%d: title element not found", swiperIndex)
		return nil, nil
	}

	category, err := titleEl.InnerText()
	if err != nil {
		log.Printf("[readActiveTrendSlide] swiperIndex=%d: failed to get title text: %v", swiperIndex, err)
		return nil, err
	}
	category = strings.TrimSpace(category)
	if category == "" {
		log.Printf("[readActiveTrendSlide] swiperIndex=%d: empty category", swiperIndex)
		return nil, nil
	}

	log.Printf("[readActiveTrendSlide] swiperIndex=%d: category='%s'", swiperIndex, category)

	items, err := readTrendItemsFromSlide(slide)
	if err != nil {
		log.Printf("[readActiveTrendSlide] swiperIndex=%d: failed to read items: %v", swiperIndex, err)
		return nil, err
	}
	if len(items) == 0 {
		log.Printf("[readActiveTrendSlide] swiperIndex=%d: category='%s', no items found", swiperIndex, category)
		return nil, nil
	}

	log.Printf("[readActiveTrendSlide] swiperIndex=%d: category='%s', items count=%d", swiperIndex, category, len(items))
	for i, item := range items {
		log.Printf("[readActiveTrendSlide] swiperIndex=%d: category='%s', item[%d]: keyword='%s', change='%s', status='%s'", 
			swiperIndex, category, i, item.Keyword, item.Change, item.Status)
	}

	return &struct {
		Category string
		Items    []TrendItem
	}{
		Category: category,
		Items:    items,
	}, nil
}

func readTrendItemsFromSlide(slide playwright.Locator) ([]TrendItem, error) {
	result, err := slide.Evaluate(`(el) => {
		const rows = Array.from(
			el.querySelectorAll(".u_ni_item_list li.u_ni_trend_item, li.u_ni_trend_item")
		);
		return rows
			.map((item) => {
				const keywordEl = item.querySelector(".u_ni_trend_text");
				const keyword = keywordEl?.textContent?.trim() || "";
				if (!keyword) return null;
				const dataEl = item.querySelector(".u_ni_data");
				const changeText = dataEl?.textContent?.trim() || "-";
				const dataClass = dataEl?.classList;
				let status = "none";
				if (dataClass?.contains("up")) status = "up";
				if (dataClass?.contains("down")) status = "down";
				if (dataClass?.contains("new")) status = "new";
				const url = item.querySelector("a.u_ni_trend_link")?.getAttribute("href") || "";
				return { keyword, change: changeText, status, url };
			})
			.filter(Boolean);
	}`, nil)

	if err != nil {
		return nil, err
	}

	items := []TrendItem{}
	if resultArray, ok := result.([]interface{}); ok {
		for _, item := range resultArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				keyword, _ := itemMap["keyword"].(string)
				change, _ := itemMap["change"].(string)
				status, _ := itemMap["status"].(string)
				url, _ := itemMap["url"].(string)

				if keyword != "" {
					items = append(items, TrendItem{
						Keyword: keyword,
						Change:  change,
						Status:  status,
						URL:     url,
					})
				}
			}
		}
	}

	return items, nil
}

func scrollToTrendSection(page playwright.Page, label string) (bool, error) {
	if label == "" {
		return false, nil
	}

	target := page.Locator(fmt.Sprintf(`text=%s`, label)).First()
	count, err := target.Count()
	if err != nil || count == 0 {
		return false, nil
	}

	target.ScrollIntoViewIfNeeded()
	page.WaitForTimeout(180)
	return true, nil
}

func ensureDemographicVisible(page playwright.Page) bool {
	for i := 0; i < 6; i++ {
		found := scrollToDemographicTitle(page)
		if found {
			return true
		}
		page.Evaluate(`() => window.scrollBy(0, window.innerHeight * 0.8)`)
		page.WaitForTimeout(200)
	}
	return false
}

func scrollToDemographicTitle(page playwright.Page) bool {
	re := regexp.MustCompile(`(\d+\s*-\s*\d+\s*세|\d+\s*세)\s*(남자|여자)`)
	title := page.Locator("h3.u_ni_trend_title").Filter(playwright.LocatorFilterOptions{
		HasText: re,
	}).First()

	count, _ := title.Count()
	if count == 0 {
		return false
	}

	title.ScrollIntoViewIfNeeded()
	page.WaitForTimeout(300)
	return true
}

func collectTrendSwiper(page playwright.Page, sourceLabel string, maxCategories int, swiperIndex int, expandAllItems bool) ([]SearchInflowTrend, error) {
	log.Printf("[collectTrendSwiper] START: sourceLabel='%s', maxCategories=%d, swiperIndex=%d, expandAllItems=%v", 
		sourceLabel, maxCategories, swiperIndex, expandAllItems)
	
	page.WaitForSelector(".u_ni_search_swiper", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(10000),
	})

	results := make(map[string][]TrendItem)

	swiperLen, _ := getTrendSwiperLength(page, swiperIndex)
	log.Printf("[collectTrendSwiper] sourceLabel='%s', swiperIndex=%d: swiperLen=%d", sourceLabel, swiperIndex, swiperLen)
	
	targetCount := maxCategories
	if targetCount == 0 {
		targetCount = 0
	}

	totalSteps := targetCount * 2
	if totalSteps == 0 {
		totalSteps = swiperLen
	}
	if totalSteps == 0 {
		totalSteps = 20
	}
	log.Printf("[collectTrendSwiper] sourceLabel='%s', swiperIndex=%d: targetCount=%d, totalSteps=%d", 
		sourceLabel, swiperIndex, targetCount, totalSteps)

	if expandAllItems {
		expandAllTrendSlides(page, swiperIndex)
	}

	lastTitle := ""
	for index := 0; index < totalSteps; index++ {
		log.Printf("[collectTrendSwiper] sourceLabel='%s', swiperIndex=%d: moving to slide index=%d", 
			sourceLabel, swiperIndex, index)
		
		moved, _ := moveTrendSwiperTo(page, index, swiperIndex)
		if !moved {
			log.Printf("[collectTrendSwiper] sourceLabel='%s', swiperIndex=%d: failed to move to slide index=%d", 
				sourceLabel, swiperIndex, index)
			break
		}

		waitForTrendTitleChange(page, lastTitle, swiperIndex)
		minCount := 0
		if expandAllItems {
			minCount = 20
		}
		waitForTrendItemsReady(page, swiperIndex, minCount)
		expandTrendSlideItems(page, swiperIndex)

		slideData, _ := readActiveTrendSlide(page, swiperIndex)
		if slideData != nil && slideData.Category != "" && len(slideData.Items) > 0 {
			if _, exists := results[slideData.Category]; !exists {
				results[slideData.Category] = slideData.Items
				log.Printf("[collectTrendSwiper] sourceLabel='%s', swiperIndex=%d: added category='%s' with %d items", 
					sourceLabel, swiperIndex, slideData.Category, len(slideData.Items))
			} else {
				log.Printf("[collectTrendSwiper] sourceLabel='%s', swiperIndex=%d: category='%s' already exists, skipping", 
					sourceLabel, swiperIndex, slideData.Category)
			}
			lastTitle = slideData.Category
		} else {
			log.Printf("[collectTrendSwiper] sourceLabel='%s', swiperIndex=%d: slide index=%d returned no data", 
				sourceLabel, swiperIndex, index)
		}

		if targetCount > 0 && len(results) >= targetCount {
			log.Printf("[collectTrendSwiper] sourceLabel='%s', swiperIndex=%d: reached target count %d, stopping", 
				sourceLabel, swiperIndex, targetCount)
			break
		}

		page.WaitForTimeout(40)
	}

	log.Printf("[collectTrendSwiper] sourceLabel='%s', swiperIndex=%d: collected %d categories", 
		sourceLabel, swiperIndex, len(results))

	trends := []SearchInflowTrend{}
	for category, items := range results {
		log.Printf("[collectTrendSwiper] sourceLabel='%s': final category='%s' with %d items", 
			sourceLabel, category, len(items))
		trends = append(trends, SearchInflowTrend{
			Category: category,
			Items:    items,
			Source:   sourceLabel,
		})
	}

	log.Printf("[collectTrendSwiper] END: sourceLabel='%s', total trends=%d", sourceLabel, len(trends))
	return trends, nil
}
