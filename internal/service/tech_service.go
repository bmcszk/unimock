package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// techService implements the TechService interface
type techService struct {
	startTime      time.Time
	requestCounter atomic.Int64
	endpointStats  map[string]*atomic.Int64
	statsMutex     sync.RWMutex
}

// NewTechService creates a new instance of TechService
func NewTechService(startTime time.Time) TechService {
	return &techService{
		startTime:     startTime,
		endpointStats: make(map[string]*atomic.Int64),
	}
}

// GetHealthStatus returns the health status of the service
func (s *techService) GetHealthStatus(ctx context.Context) map[string]interface{} {
	uptime := time.Since(s.startTime).String()

	return map[string]interface{}{
		"status": "ok",
		"uptime": uptime,
	}
}

// GetMetrics returns metrics about the service
func (s *techService) GetMetrics(ctx context.Context) map[string]interface{} {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	// Create API endpoint stats
	apiEndpoints := make(map[string]int64)
	for path, counter := range s.endpointStats {
		apiEndpoints[path] = counter.Load()
	}

	return map[string]interface{}{
		"request_count": s.requestCounter.Load(),
		"api_endpoints": apiEndpoints,
	}
}

// IncrementRequestCount increments the request counter
func (s *techService) IncrementRequestCount(ctx context.Context, path string) {
	s.requestCounter.Add(1)

	// Track endpoint stats
	s.statsMutex.RLock()
	counter, exists := s.endpointStats[path]
	s.statsMutex.RUnlock()

	if exists {
		counter.Add(1)
	} else {
		s.statsMutex.Lock()
		// Re-check after acquiring write lock, in case another goroutine added it
		counter, exists = s.endpointStats[path]
		if exists {
			s.statsMutex.Unlock()
			counter.Add(1)
		} else {
			var newCounter atomic.Int64
			newCounter.Add(1)
			s.endpointStats[path] = &newCounter
			s.statsMutex.Unlock()
		}
	}
}
