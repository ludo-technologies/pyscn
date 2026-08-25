package app

import (
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestProgressTrackerAdvancesWithElapsedTime(t *testing.T) {
	tracker := newAnalysisProgressTracker(map[domain.AnalysisKind]float64{
		domain.AnalysisKindComplexity: 1.0,
		domain.AnalysisKindClones:     10.0,
	})
	// Simulate 5s elapsed: projection is the largest pending task (10s)
	tracker.start = time.Now().Add(-5 * time.Second)

	p := tracker.Percent()
	// 5/10 of the projection at 95% scale => ~47%
	if p < 40 || p > 55 {
		t.Errorf("expected progress around 47%%, got %d%%", p)
	}
}

func TestProgressTrackerRecalibratesOnTaskCompletion(t *testing.T) {
	tracker := newAnalysisProgressTracker(map[domain.AnalysisKind]float64{
		domain.AnalysisKindComplexity: 1.0,
		domain.AnalysisKindClones:     10.0,
	})
	// The small task (estimated 1s) actually took 3s => machine is 3x slower
	tracker.start = time.Now().Add(-3 * time.Second)
	tracker.TaskCompleted(domain.AnalysisKindComplexity)

	p := tracker.Percent()
	// Projection rescales to 10 * 3 = 30s; at 3s elapsed => ~9.5%
	if p > 15 {
		t.Errorf("expected recalibrated progress below 15%%, got %d%%", p)
	}
}

func TestProgressTrackerApproachesCapOnOverrun(t *testing.T) {
	tracker := newAnalysisProgressTracker(map[domain.AnalysisKind]float64{domain.AnalysisKindComplexity: 1.0})
	// Elapsed time far beyond the 1s estimate
	tracker.start = time.Now().Add(-30 * time.Second)

	p := tracker.Percent()
	if p < 95 || p > 99 {
		t.Errorf("expected progress in [95, 99] on overrun, got %d%%", p)
	}
}

func TestProgressTrackerIsMonotonic(t *testing.T) {
	tracker := newAnalysisProgressTracker(map[domain.AnalysisKind]float64{
		domain.AnalysisKindComplexity: 1.0,
		domain.AnalysisKindClones:     10.0,
	})
	tracker.start = time.Now().Add(-8 * time.Second)
	before := tracker.Percent()

	// Completing the small task triples the projection, which would lower the
	// raw time fraction; the displayed value must not move backwards
	tracker.TaskCompleted(domain.AnalysisKindComplexity)
	after := tracker.Percent()

	if after < before {
		t.Errorf("progress moved backwards: %d%% -> %d%%", before, after)
	}
}

func TestProgressTrackerReports99WhenAllTasksDone(t *testing.T) {
	tracker := newAnalysisProgressTracker(map[domain.AnalysisKind]float64{domain.AnalysisKindComplexity: 1.0})
	tracker.TaskCompleted(domain.AnalysisKindComplexity)

	if p := tracker.Percent(); p != 99 {
		t.Errorf("expected 99%% when all tasks completed, got %d%%", p)
	}
}

func TestProgressTrackerIgnoresUnknownTask(t *testing.T) {
	tracker := newAnalysisProgressTracker(map[domain.AnalysisKind]float64{domain.AnalysisKindComplexity: 1.0})
	tracker.TaskCompleted(domain.AnalysisKind("unknown"))

	if d := tracker.CompletedDurations(); len(d) != 0 {
		t.Errorf("expected no recorded durations, got %v", d)
	}
}

func TestApplyTimingFactors(t *testing.T) {
	estimates := map[domain.AnalysisKind]float64{domain.AnalysisKindComplexity: 2.0, domain.AnalysisKindClones: 4.0}
	factors := map[domain.AnalysisKind]float64{domain.AnalysisKindComplexity: 3.0}

	calibrated := applyTimingFactors(estimates, factors)

	if calibrated[domain.AnalysisKindComplexity] != 6.0 {
		t.Errorf("expected complexity calibrated to 6.0, got %f", calibrated[domain.AnalysisKindComplexity])
	}
	if calibrated[domain.AnalysisKindClones] != 4.0 {
		t.Errorf("expected clones unchanged at 4.0, got %f", calibrated[domain.AnalysisKindClones])
	}
}
