package main

// Config represents the application configuration
type Config struct {
	Naver          NaverConfig          `json:"naver"`
	CreatorAdvisor CreatorAdvisorConfig `json:"creatorAdvisor"`
	SearchAd       SearchAdConfig        `json:"searchad"`
	Scoring        Scoring               `json:"scoring"`
	Seeds          []string              `json:"seeds"`
	Usage          Usage                 `json:"usage"`
}

// NaverConfig represents Naver API configuration
type NaverConfig struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// CreatorAdvisorConfig represents Creator Advisor configuration
type CreatorAdvisorConfig struct {
	LoginID  string `json:"loginId"`
	Password string `json:"password"`
}

// SearchAdConfig represents SearchAd API configuration
type SearchAdConfig struct {
	BaseURL    string `json:"baseUrl"`
	CustomerID string `json:"customerId"`
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
}

// Scoring represents scoring parameters
type Scoring struct {
	WPC   float64 `json:"w_pc"`
	WMob  float64 `json:"w_mob"`
	K     float64 `json:"k"`
	Alpha float64 `json:"alpha"`
	T     float64 `json:"t"`
	R     float64 `json:"r"`
}

// Usage represents usage information
type Usage struct {
	Date string `json:"date"`
	Used int    `json:"used"`
}

// ScoreResult represents the scoring calculation result
type ScoreResult struct {
	Score       float64 `json:"score"`
	ScoreRaw    float64 `json:"scoreRaw"`
	Score4      float64 `json:"score4"`
	Score5      float64 `json:"score5"`
	TrendNorm   float64 `json:"trendNorm"`
	RecentRatio float64 `json:"recentRatio"`
}

// KeywordAnalysisResult represents keyword analysis result
type KeywordAnalysisResult struct {
	Keyword         string     `json:"keyword"`
	SearchPc        int        `json:"searchPc"`
	SearchMobile    int        `json:"searchMobile"`
	TotalSearch     int        `json:"totalSearch"`
	MonthlySearch   *int       `json:"monthlySearch"`
	MonthlyAvgClicks *float64  `json:"monthlyAvgClicks"`
	MonthlyAvgCtr   *float64   `json:"monthlyAvgCtr"`
	Competition     interface{} `json:"competition"`
	DocCount        int        `json:"docCount"`
	Group           string     `json:"group"`
	RelatedDetails  []RelatedDetail `json:"relatedDetails"`
	Score           float64    `json:"score"`
}

// RelatedDetail represents related keyword detail
type RelatedDetail struct {
	Keyword      string  `json:"keyword"`
	SearchPc     int     `json:"searchPc"`
	SearchMobile int     `json:"searchMobile"`
	TotalSearch  int     `json:"totalSearch"`
	DocCount     int     `json:"docCount"`
}

// AutoExtractResult represents auto extraction result
type AutoExtractResult struct {
	Keyword         string   `json:"keyword"`
	SearchPc        int      `json:"searchPc"`
	SearchMobile    int      `json:"searchMobile"`
	TotalSearch     int      `json:"totalSearch"`
	MonthlySearch   *int     `json:"monthlySearch"`
	MonthlyAvgClicks *float64 `json:"monthlyAvgClicks"`
	MonthlyAvgCtr   *float64 `json:"monthlyAvgCtr"`
	Competition     interface{} `json:"competition"`
	DocCount        int      `json:"docCount"`
	Group           string   `json:"group"`
	Score           float64  `json:"score"`
	Cpc             *float64 `json:"cpc"`
	CpcPc           *float64 `json:"cpcPc"`
	CpcMobile       *float64 `json:"cpcMobile"`
	PcClicks        float64  `json:"pcClicks"`
	PcCost          float64  `json:"pcCost"`
	MobileClicks   float64  `json:"mobileClicks"`
	MobileCost      float64  `json:"mobileCost"`
	TotalClicks     float64  `json:"totalClicks"`
	TotalCost       float64  `json:"totalCost"`
}

// Progress represents progress information
type Progress struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Keyword string `json:"keyword,omitempty"`
	Stage   string `json:"stage,omitempty"`
	Step    string `json:"step,omitempty"`
	JobID   string `json:"jobId,omitempty"`
}

// CreatorAdvisorResult represents creator advisor collection result
type CreatorAdvisorResult struct {
	URL                string                   `json:"url"`
	SearchInflowTrends []SearchInflowTrend      `json:"searchInflowTrends"`
	SearchInflowMessage string                  `json:"searchInflowMessage"`
	MainInflowContents []MainInflowContent     `json:"mainInflowContents"`
}

// SearchInflowTrend represents search inflow trend data
type SearchInflowTrend struct {
	Category string      `json:"category"`
	Items    []TrendItem `json:"items"`
	Source   string      `json:"source"`
}

// TrendItem represents a trend item
type TrendItem struct {
	Keyword string `json:"keyword"`
	Change  string `json:"change"`
	Status  string `json:"status"`
	URL     string `json:"url"`
}

// MainInflowContent represents main inflow content
type MainInflowContent struct {
	Rank  int    `json:"rank"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// TestResult represents API test result
type TestResult struct {
	OK      bool   `json:"ok"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// SearchAdTestResult represents SearchAd API test result
type SearchAdTestResult struct {
	OK       bool       `json:"ok"`
	SearchAd TestResult `json:"searchad"`
	Naver    TestResult `json:"naver"`
}

// CreatorAdvisorPayload represents creator advisor collection payload
type CreatorAdvisorPayload struct {
	LoginID     string `json:"loginId"`
	Password    string `json:"password"`
	ShowBrowser bool   `json:"showBrowser"`
}

// CreatorAdvisorResponse represents creator advisor collection response
type CreatorAdvisorResponse struct {
	OK    bool   `json:"ok"`
	JobID string `json:"jobId,omitempty"`
	Message string `json:"message,omitempty"`
}

// CreatorAdvisorDone represents creator advisor done event
type CreatorAdvisorDone struct {
	JobID  string                `json:"jobId"`
	Result CreatorAdvisorResult  `json:"result"`
}

// CreatorAdvisorError represents creator advisor error event
type CreatorAdvisorError struct {
	JobID   string `json:"jobId"`
	Message string `json:"message"`
}
