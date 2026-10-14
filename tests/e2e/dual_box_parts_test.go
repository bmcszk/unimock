package e2e_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Returns fresh parts so caller's own server/baseURL stays untouched.
func (p *parts) startSecondServer(configFile string) *parts {
	p.Helper()

	receiver := &parts{
		T:       p.T,
		require: p.require,
	}
	receiver.setupScenarioTestServer(configFile)
	return receiver
}

// Second instance shares the parts lifecycle so Cleanup closes both boxes.
func (p *parts) a_second_unimock_instance_is_started_with(scenarioYAML string) *parts {
	p.Helper()

	p.receiver = p.startSecondServer(p.writeTempConfig(scenarioYAML))
	return p
}

func (p *parts) writeTempConfig(yamlContent string) string {
	p.Helper()

	tempDir := p.TempDir()
	configFile := filepath.Join(tempDir, "dual-box-receiver-config.yaml")
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	p.require.NoError(err, "Failed to create receiver config file")
	return configFile
}

// Swaps baseURL to box B for the call's duration so shared stream state feeds the_stream_* assertions.
func (p *parts) a_streaming_request_is_made_to_the_second_instance(url string) *parts {
	p.Helper()

	savedBase := p.baseURL
	p.baseURL = p.receiver.baseURL
	defer func() { p.baseURL = savedBase }()
	return p.a_streaming_request_is_made_to(url)
}

func fetchDeliveries(baseURL string) ([]map[string]any, error) {
	resp, err := http.Get(baseURL + "/_uni/webhooks/deliveries")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deliveries endpoint returned %d", resp.StatusCode)
	}
	var records []map[string]any
	if err := json.Unmarshal(body, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// Cross-box delivery is async; poll instead of sleeping.
func (p *parts) the_second_instance_received_the_webhook_request(path string) *parts {
	p.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if p.receiver.metricsHitCountFor(path) >= 1 {
			return p
		}
		time.Sleep(10 * time.Millisecond)
	}
	p.Fatalf("timed out waiting for box B to record an inbound request on %s", path)
	return p
}

func (p *parts) the_second_instance_responded_200_to(path string) *parts {
	p.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if p.receiver.metricsStatusCountFor(path, http.StatusOK) >= 1 {
			return p
		}
		time.Sleep(10 * time.Millisecond)
	}
	p.Fatalf("timed out waiting for box B to record a 200 response on %s", path)
	return p
}

func (p *parts) the_delivery_to_is_recorded_as_successful(targetURL string) *parts {
	p.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		records, err := fetchDeliveries(p.baseURL)
		p.require.NoError(err, "GET /_uni/webhooks/deliveries on broadcaster")
		for _, rec := range records {
			if fmt.Sprintf("%v", rec["url"]) == targetURL &&
				statusIs2xx(rec["statusCode"]) &&
				deliveryHasNoError(rec) {
				return p
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	p.Fatalf("timed out waiting for a successful delivery record targeting %s", targetURL)
	return p
}

// B only received; it never broadcast.
func (p *parts) the_second_instance_records_no_outbound_deliveries() *parts {
	p.Helper()

	records, err := fetchDeliveries(p.receiver.baseURL)
	p.require.NoError(err, "GET /_uni/webhooks/deliveries on receiver")
	p.require.Empty(records, "receiver box should have no outbound deliveries")
	return p
}

func (p *parts) metricsHitCountFor(path string) int {
	p.Helper()

	metrics := p.fetchMetrics()
	endpoints := metricsMap(metrics, "api_endpoints")
	return int(metricsNumber(endpoints, path))
}

func (p *parts) metricsStatusCountFor(path string, statusCode int) int {
	p.Helper()

	metrics := p.fetchMetrics()
	stats := metricsMap(metrics, "status_code_stats")
	pathStats := metricsMap(stats, path)
	return int(metricsNumber(pathStats, fmt.Sprintf("%d", statusCode)))
}

func metricsMap(source map[string]any, key string) map[string]any {
	if value, ok := source[key].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}

func metricsNumber(source map[string]any, key string) float64 {
	if value, ok := source[key].(float64); ok {
		return value
	}
	return 0
}

func (p *parts) fetchMetrics() map[string]any {
	p.Helper()

	resp, err := http.Get(p.baseURL + "/_uni/metrics")
	p.require.NoError(err, "GET /_uni/metrics")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	p.require.NoError(err, "read metrics body")
	p.require.Equal(http.StatusOK, resp.StatusCode, "metrics endpoint status")
	var metrics map[string]any
	p.require.NoError(json.Unmarshal(body, &metrics), "decode metrics JSON")
	return metrics
}

// deliveries API omits the error key entirely on success.
func deliveryHasNoError(record map[string]any) bool {
	errValue, present := record["error"]
	return !present || errValue == nil || errValue == ""
}

func statusIs2xx(statusCode any) bool {
	code, ok := statusCode.(float64)
	return ok && code >= 200 && code < 300
}
