package main

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                context.Context
	creatorAdvisorJob  *CreatorAdvisorJob
	creatorAdvisorLock sync.Mutex
}

type CreatorAdvisorJob struct {
	ID      string
	Running bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetConfig returns the current configuration
func (a *App) GetConfig() (Config, error) {
	return readConfig()
}

// SaveConfig saves the configuration
func (a *App) SaveConfig(config Config) error {
	// Ensure all nested structs are properly initialized
	if config.Naver.ClientID == "" {
		config.Naver.ClientID = ""
	}
	if config.Naver.ClientSecret == "" {
		config.Naver.ClientSecret = ""
	}
	if config.SearchAd.CustomerID == "" {
		config.SearchAd.CustomerID = ""
	}
	if config.SearchAd.AccessKey == "" {
		config.SearchAd.AccessKey = ""
	}
	if config.SearchAd.SecretKey == "" {
		config.SearchAd.SecretKey = ""
	}
	return writeConfig(config)
}

// GetUsage returns the current usage information
func (a *App) GetUsage() (Usage, error) {
	config, err := readConfig()
	if err != nil {
		return Usage{}, err
	}
	return config.Usage, nil
}

// AnalyzeKeyword analyzes a keyword
func (a *App) AnalyzeKeyword(keyword string) (*KeywordAnalysisResult, error) {
	config, err := readConfig()
	if err != nil {
		return nil, err
	}

	progressChan := make(chan Progress, 10)
	go func() {
		for progress := range progressChan {
			runtime.EventsEmit(a.ctx, "analysis:keyword:progress", progress)
		}
	}()

	result, err := analyzeKeyword(config, keyword, func(progress Progress) {
		select {
		case progressChan <- progress:
		default:
		}
	})
	close(progressChan)

	if err != nil {
		return nil, err
	}

	_, err = incrementUsage(1)
	if err != nil {
		// Log error but don't fail the request
	}

	return result, nil
}

// RunAutoExtract runs automatic keyword extraction
func (a *App) RunAutoExtract(payload map[string]interface{}) ([]AutoExtractResult, error) {
	config, err := readConfig()
	if err != nil {
		return nil, err
	}

	var seeds []string
	if seedsVal, ok := payload["seeds"].([]interface{}); ok {
		for _, s := range seedsVal {
			if str, ok := s.(string); ok && str != "" {
				seeds = append(seeds, str)
			}
		}
	}
	if len(seeds) == 0 {
		seeds = config.Seeds
	}

	progressChan := make(chan Progress, 10)
	go func() {
		for progress := range progressChan {
			runtime.EventsEmit(a.ctx, "analysis:auto:progress", progress)
		}
	}()

	results, err := autoExtract(config, seeds, func(progress Progress) {
		select {
		case progressChan <- progress:
		default:
		}
	})
	close(progressChan)

	if err != nil {
		return nil, err
	}

	_, err = incrementUsage(len(results))
	if err != nil {
		// Log error but don't fail the request
	}

	return results, nil
}

// CollectCreatorAdvisor collects creator advisor data
func (a *App) CollectCreatorAdvisor(payloadJSON string) (*CreatorAdvisorResponse, error) {
	var payload CreatorAdvisorPayload
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return &CreatorAdvisorResponse{
			OK:      false,
			Message: "Invalid payload",
		}, err
	}

	a.creatorAdvisorLock.Lock()
	if a.creatorAdvisorJob != nil && a.creatorAdvisorJob.Running {
		a.creatorAdvisorLock.Unlock()
		return &CreatorAdvisorResponse{
			OK:      false,
			Message: "이미 수집이 진행 중입니다.",
		}, nil
	}

	jobID := generateJobID()
	a.creatorAdvisorJob = &CreatorAdvisorJob{
		ID:      jobID,
		Running: true,
	}
	a.creatorAdvisorLock.Unlock()

	go func() {
		defer func() {
			a.creatorAdvisorLock.Lock()
			a.creatorAdvisorJob = nil
			a.creatorAdvisorLock.Unlock()
		}()

		progressChan := make(chan Progress, 10)
		go func() {
			for progress := range progressChan {
				progress.JobID = jobID
				runtime.EventsEmit(a.ctx, "creator:progress", progress)
			}
		}()

		result, err := collectCreatorAdvisor(payload, func(progress Progress) {
			select {
			case progressChan <- progress:
			default:
			}
		})
		close(progressChan)

		if err != nil {
			runtime.EventsEmit(a.ctx, "creator:error", CreatorAdvisorError{
				JobID:   jobID,
				Message: err.Error(),
			})
			return
		}

		runtime.EventsEmit(a.ctx, "creator:done", CreatorAdvisorDone{
			JobID:  jobID,
			Result: *result,
		})
	}()

	return &CreatorAdvisorResponse{
		OK:    true,
		JobID: jobID,
	}, nil
}

// TestSearchAd tests SearchAd API connection
func (a *App) TestSearchAd() (*SearchAdTestResult, error) {
	config, err := readConfig()
	if err != nil {
		return nil, err
	}

	searchAdResult := testSearchAd(config)
	naverResult := testNaverSearch(config)

	return &SearchAdTestResult{
		OK:       searchAdResult.OK && naverResult.OK,
		SearchAd: *searchAdResult,
		Naver:    *naverResult,
	}, nil
}

// GetSearchAdLog returns SearchAd log tail
func (a *App) GetSearchAdLog() (string, error) {
	return readLogTail("searchad.log", 8000)
}

// SetZoom sets the window zoom factor
func (a *App) SetZoom(delta float64) (float64, error) {
	// Wails v2 doesn't have direct zoom control like Electron
	// We can emit an event for the frontend to handle zoom via CSS transform
	// For now, return the delta as a placeholder
	// Frontend can implement zoom using CSS transform: scale()
	runtime.EventsEmit(a.ctx, "zoom:changed", delta)
	return delta, nil
}

func generateJobID() string {
	id := uuid.New()
	return id.String()
}
