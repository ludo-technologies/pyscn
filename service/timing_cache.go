package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const analysisTimingsVersion = 1

// analysisTimings stores per-task calibration factors (actual seconds divided by
// estimated seconds) observed in previous runs for a given project directory.
type analysisTimings struct {
	Version int                `json:"version"`
	Factors map[string]float64 `json:"factors"`
}

// analysisTimingsPath returns the per-project cache file path, keyed by the
// current working directory so repeated runs on the same project share timings.
func analysisTimingsPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(cwd))
	name := hex.EncodeToString(sum[:8]) + ".json"
	return filepath.Join(cacheDir, "pyscn", "timings", name), nil
}

// LoadAnalysisTimingFactors returns calibration factors recorded by previous
// runs on this project. Returns an empty map when no usable cache exists.
// Disabled under `go test` so test runs neither read nor pollute the cache.
func LoadAnalysisTimingFactors() map[string]float64 {
	if testing.Testing() {
		return map[string]float64{}
	}
	path, err := analysisTimingsPath()
	if err != nil {
		return map[string]float64{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]float64{}
	}
	var timings analysisTimings
	if err := json.Unmarshal(data, &timings); err != nil || timings.Version != analysisTimingsVersion || timings.Factors == nil {
		return map[string]float64{}
	}
	return timings.Factors
}

// UpdateAnalysisTimingFactors records observed task durations against their
// estimates, smoothing with previously stored factors. Failures are ignored:
// the cache only improves progress estimation and must never affect analysis.
func UpdateAnalysisTimingFactors(estimatedSeconds, actualSeconds map[string]float64) {
	if testing.Testing() {
		return
	}
	factors := LoadAnalysisTimingFactors()
	for name, est := range estimatedSeconds {
		actual, ok := actualSeconds[name]
		if !ok || est <= 0 || actual <= 0 {
			continue
		}
		factor := clampTimingFactor(actual / est)
		if prev, ok := factors[name]; ok {
			// Exponential moving average to absorb run-to-run noise
			factor = 0.5*prev + 0.5*factor
		}
		factors[name] = factor
	}
	if len(factors) == 0 {
		return
	}

	path, err := analysisTimingsPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	data, err := json.Marshal(analysisTimings{Version: analysisTimingsVersion, Factors: factors})
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

// clampTimingFactor bounds a calibration factor so a single pathological run
// (sleeping laptop, antivirus scan) cannot poison future estimates.
func clampTimingFactor(f float64) float64 {
	if f < 0.05 {
		return 0.05
	}
	if f > 50 {
		return 50
	}
	return f
}
