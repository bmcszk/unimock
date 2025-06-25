package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// incrementValue represents the value to add when incrementing counters
	incrementValue = 1
)

// techService implements the TechService interface
type techService struct {
	startTime      time.Time
	requestCounter atomic.Int64
	endpointStats  map[string]*atomic.Int64
	statusStats    map[string]map[int]*atomic.Int64 // path -> status_code -> counter
	statsMutex     sync.RWMutex
}

// NewTechService creates a new instance of TechService
func NewTechService(startTime time.Time) TechService {
	return &techService{
		startTime:     startTime,
		endpointStats: make(map[string]*atomic.Int64),
		statusStats:   make(map[string]map[int]*atomic.Int64),
	}
}

// GetHealthStatus returns the health status of the service
func (s *techService) GetHealthStatus(_ context.Context) map[string]any {
	uptime := time.Since(s.startTime).String()

	return map[string]any{
		"status": "ok",
		"uptime": uptime,
	}
}

// GetMetrics returns metrics about the service
func (s *techService) GetMetrics(_ context.Context) map[string]any {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	// Create API endpoint stats
	apiEndpoints := make(map[string]int64)
	for path, counter := range s.endpointStats {
		apiEndpoints[path] = counter.Load()
	}

	// Create status code stats grouped by path
	statusCodeStats := make(map[string]map[string]int64)
	for path, statusMap := range s.statusStats {
		pathStats := make(map[string]int64)
		for statusCode, counter := range statusMap {
			pathStats[fmt.Sprintf("%d", statusCode)] = counter.Load()
		}
		if len(pathStats) > 0 {
			statusCodeStats[path] = pathStats
		}
	}

	return map[string]any{
		"request_count":     s.requestCounter.Load(),
		"api_endpoints":     apiEndpoints,
		"status_code_stats": statusCodeStats,
	}
}

// IncrementRequestCount increments the request counter
func (s *techService) IncrementRequestCount(_ context.Context, path string) {
	s.requestCounter.Add(incrementValue)

	// Track endpoint stats
	s.statsMutex.RLock()
	counter, exists := s.endpointStats[path]
	s.statsMutex.RUnlock()

	if exists {
		counter.Add(incrementValue)
	} else {
		s.statsMutex.Lock()
		// Re-check after acquiring write lock, in case another goroutine added it
		counter, exists = s.endpointStats[path]
		if exists {
			s.statsMutex.Unlock()
			counter.Add(incrementValue)
		} else {
			var newCounter atomic.Int64
			newCounter.Add(incrementValue)
			s.endpointStats[path] = &newCounter
			s.statsMutex.Unlock()
		}
	}
}

// TrackResponse tracks a response by path and status code
func (s *techService) TrackResponse(_ context.Context, path string, statusCode int) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	// Get or create path stats
	pathStats, pathExists := s.statusStats[path]
	if !pathExists {
		pathStats = make(map[int]*atomic.Int64)
		s.statusStats[path] = pathStats
	}

	// Get or create status code counter
	counter, statusExists := pathStats[statusCode]
	if !statusExists {
		counter = &atomic.Int64{}
		pathStats[statusCode] = counter
	}

	counter.Add(incrementValue)
}
