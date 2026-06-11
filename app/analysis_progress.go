package app

import (
	"maps"
	"math"
	"sync"
	"time"
)

// analysisProgressTracker estimates overall progress for concurrently running
// analysis tasks. Tasks all start together, so the projected wall time is the
// completion time of the largest still-pending task. Each task completion is
// both a hard checkpoint and a calibration sample: the observed ratio of
// actual to estimated time rescales the projection for the remaining tasks.
type analysisProgressTracker struct {
	mu          sync.Mutex
	start       time.Time
	pending     map[string]float64 // task name -> calibrated estimated seconds
	doneWall    float64            // sum of completion wall times of finished tasks
	doneEst     float64            // sum of estimated seconds of finished tasks
	durations   map[string]float64 // task name -> completion wall time in seconds
	lastPercent int
}

// minTimingSignal is the minimum accumulated estimated-seconds of completed
// tasks before observed timings override the prior estimates; ratios derived
// from near-instant tasks are too noisy to extrapolate from.
const minTimingSignal = 0.3

func newAnalysisProgressTracker(estimatedSeconds map[string]float64) *analysisProgressTracker {
	pending := make(map[string]float64, len(estimatedSeconds))
	for name, est := range estimatedSeconds {
		if est < 0.05 {
			est = 0.05
		}
		pending[name] = est
	}
	return &analysisProgressTracker{
		start:     time.Now(),
		pending:   pending,
		durations: make(map[string]float64, len(estimatedSeconds)),
	}
}

// TaskCompleted records that the named task finished, capturing its wall-clock
// completion time as a calibration sample for the remaining tasks.
func (t *analysisProgressTracker) TaskCompleted(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	est, ok := t.pending[name]
	if !ok {
		return
	}
	delete(t.pending, name)

	wall := time.Since(t.start).Seconds()
	t.durations[name] = wall
	t.doneWall += wall
	t.doneEst += est
}

// CompletedDurations returns the observed wall-clock completion time per task.
func (t *analysisProgressTracker) CompletedDurations() map[string]float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	out := make(map[string]float64, len(t.durations))
	maps.Copy(out, t.durations)
	return out
}

// Percent returns the current progress in [0, 99]. The value is monotonic; it
// never reaches 100 here because completion is signalled explicitly by the
// caller once all tasks have actually finished.
func (t *analysisProgressTracker) Percent() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.pending) == 0 {
		t.lastPercent = 99
		return 99
	}

	// Rescale remaining estimates by the observed actual/estimated ratio once
	// enough completed work has accumulated to be a meaningful sample.
	ratio := 1.0
	if t.doneEst >= minTimingSignal {
		ratio = t.doneWall / t.doneEst
		if ratio < 0.1 {
			ratio = 0.1
		} else if ratio > 20 {
			ratio = 20
		}
	}

	maxPending := 0.0
	for _, est := range t.pending {
		if est > maxPending {
			maxPending = est
		}
	}
	projected := maxPending * ratio
	if projected < 0.1 {
		projected = 0.1
	}

	elapsed := time.Since(t.start).Seconds()
	frac := elapsed / projected
	var percent float64
	if frac <= 1.0 {
		percent = frac * 95.0
	} else {
		// Past the projection: approach 99% asymptotically instead of
		// freezing, so an underestimate still shows visible movement.
		percent = 95.0 + 4.0*(1.0-math.Exp(-(frac-1.0)/0.5))
	}

	p := min(max(int(percent), t.lastPercent), 99)
	t.lastPercent = p
	return p
}

// applyTimingFactors scales estimated task durations by per-task calibration
// factors observed in previous runs. Tasks without a recorded factor keep
// their formula-based estimate.
func applyTimingFactors(estimatedSeconds, factors map[string]float64) map[string]float64 {
	calibrated := make(map[string]float64, len(estimatedSeconds))
	for name, est := range estimatedSeconds {
		if f, ok := factors[name]; ok && f > 0 {
			est *= f
		}
		calibrated[name] = est
	}
	return calibrated
}
