package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDegradationManager(t *testing.T) {
	dm := NewDegradationManager()
	require.NotNil(t, dm)
	assert.Equal(t, LevelFull, dm.GetDegradationLevel())
	assert.Empty(t, dm.GetAllStatuses())
}

func TestRegisterSubsystem(t *testing.T) {
	dm := NewDegradationManager()

	dm.RegisterSubsystem("metrics", false)
	dm.RegisterSubsystem("database", true)

	statuses := dm.GetAllStatuses()
	assert.Len(t, statuses, 2)

	metricsStatus, exists := dm.GetSubsystemStatus("metrics")
	require.True(t, exists)
	assert.Equal(t, "metrics", metricsStatus.Name)
	assert.True(t, metricsStatus.Healthy)
	assert.False(t, metricsStatus.Critical)

	dbStatus, exists := dm.GetSubsystemStatus("database")
	require.True(t, exists)
	assert.Equal(t, "database", dbStatus.Name)
	assert.True(t, dbStatus.Healthy)
	assert.True(t, dbStatus.Critical)
}

func TestUpdateSubsystemStatus(t *testing.T) {
	dm := NewDegradationManager()
	dm.RegisterSubsystem("cache", false)

	// Mark as unhealthy
	testErr := errors.New("connection timeout")
	dm.UpdateSubsystemStatus("cache", false, testErr)

	status, exists := dm.GetSubsystemStatus("cache")
	require.True(t, exists)
	assert.False(t, status.Healthy)
	assert.Equal(t, testErr, status.Error)
	assert.WithinDuration(t, time.Now(), status.LastCheck, time.Second)

	// Mark as healthy again
	dm.UpdateSubsystemStatus("cache", true, nil)

	status, exists = dm.GetSubsystemStatus("cache")
	require.True(t, exists)
	assert.True(t, status.Healthy)
	assert.Nil(t, status.Error)
}

func TestDegradationLevelCalculation(t *testing.T) {
	tests := []struct {
		name          string
		subsystems    map[string]bool // name -> critical
		failures      []string        // subsystems to mark as failed
		expectedLevel DegradationLevel
	}{
		{
			name: "all healthy",
			subsystems: map[string]bool{
				"metrics": false,
				"pcg":     false,
				"db":      true,
			},
			failures:      []string{},
			expectedLevel: LevelFull,
		},
		{
			name: "non-critical failure",
			subsystems: map[string]bool{
				"metrics": false,
				"pcg":     false,
				"db":      true,
			},
			failures:      []string{"metrics"},
			expectedLevel: LevelDegraded,
		},
		{
			name: "critical failure",
			subsystems: map[string]bool{
				"metrics": false,
				"pcg":     false,
				"db":      true,
			},
			failures:      []string{"db"},
			expectedLevel: LevelMinimal,
		},
		{
			name: "multiple non-critical failures",
			subsystems: map[string]bool{
				"metrics": false,
				"pcg":     false,
				"cache":   false,
			},
			failures:      []string{"metrics", "pcg"},
			expectedLevel: LevelDegraded,
		},
		{
			name: "mixed failures",
			subsystems: map[string]bool{
				"metrics": false,
				"db":      true,
			},
			failures:      []string{"metrics", "db"},
			expectedLevel: LevelMinimal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := NewDegradationManager()

			// Register all subsystems
			for name, critical := range tt.subsystems {
				dm.RegisterSubsystem(name, critical)
			}

			// Mark specified subsystems as failed
			for _, name := range tt.failures {
				dm.UpdateSubsystemStatus(name, false, errors.New("test failure"))
			}

			assert.Equal(t, tt.expectedLevel, dm.GetDegradationLevel())
		})
	}
}

func TestIsSubsystemHealthy(t *testing.T) {
	dm := NewDegradationManager()
	dm.RegisterSubsystem("service1", false)
	dm.RegisterSubsystem("service2", false)

	assert.True(t, dm.IsSubsystemHealthy("service1"))
	assert.False(t, dm.IsSubsystemHealthy("nonexistent"))

	dm.UpdateSubsystemStatus("service1", false, errors.New("test error"))
	assert.False(t, dm.IsSubsystemHealthy("service1"))
	assert.True(t, dm.IsSubsystemHealthy("service2"))
}

func TestGetAllStatuses(t *testing.T) {
	dm := NewDegradationManager()
	dm.RegisterSubsystem("metrics", false)
	dm.RegisterSubsystem("pcg", false)
	dm.RegisterSubsystem("db", true)

	dm.UpdateSubsystemStatus("metrics", false, errors.New("metrics down"))

	statuses := dm.GetAllStatuses()
	assert.Len(t, statuses, 3)

	metricsStatus := statuses["metrics"]
	assert.False(t, metricsStatus.Healthy)
	assert.NotNil(t, metricsStatus.Error)

	pcgStatus := statuses["pcg"]
	assert.True(t, pcgStatus.Healthy)
	assert.Nil(t, pcgStatus.Error)
}

func TestDegradationLevelString(t *testing.T) {
	tests := []struct {
		level    DegradationLevel
		expected string
	}{
		{LevelFull, "Full"},
		{LevelDegraded, "Degraded"},
		{LevelMinimal, "Minimal"},
		{DegradationLevel(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestExecuteWithFallback_Success(t *testing.T) {
	dm := NewDegradationManager()
	dm.RegisterSubsystem("test_service", false)

	ctx := context.Background()
	expectedResult := "primary result"

	primary := func(ctx context.Context) (interface{}, error) {
		return expectedResult, nil
	}

	fallback := func(ctx context.Context) (interface{}, error) {
		t.Fatal("fallback should not be called on success")
		return nil, nil
	}

	result, err := dm.ExecuteWithFallback(ctx, "test_service", primary, fallback)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.True(t, dm.IsSubsystemHealthy("test_service"))
}

func TestExecuteWithFallback_PrimaryFailsWithFallback(t *testing.T) {
	dm := NewDegradationManager()
	dm.RegisterSubsystem("test_service", false)

	ctx := context.Background()
	primaryErr := errors.New("primary failure")
	fallbackResult := "fallback result"

	primary := func(ctx context.Context) (interface{}, error) {
		return nil, primaryErr
	}

	fallback := func(ctx context.Context) (interface{}, error) {
		return fallbackResult, nil
	}

	result, err := dm.ExecuteWithFallback(ctx, "test_service", primary, fallback)

	assert.ErrorIs(t, err, ErrFallbackUsed)
	assert.Equal(t, fallbackResult, result)
	assert.False(t, dm.IsSubsystemHealthy("test_service"))
}

func TestExecuteWithFallback_BothFail(t *testing.T) {
	dm := NewDegradationManager()
	dm.RegisterSubsystem("test_service", false)

	ctx := context.Background()
	primaryErr := errors.New("primary failure")
	fallbackErr := errors.New("fallback failure")

	primary := func(ctx context.Context) (interface{}, error) {
		return nil, primaryErr
	}

	fallback := func(ctx context.Context) (interface{}, error) {
		return nil, fallbackErr
	}

	result, err := dm.ExecuteWithFallback(ctx, "test_service", primary, fallback)

	assert.Error(t, err)
	assert.Equal(t, fallbackErr, err)
	assert.Nil(t, result)
	assert.False(t, dm.IsSubsystemHealthy("test_service"))
}

func TestExecuteWithFallback_NoFallback(t *testing.T) {
	dm := NewDegradationManager()
	dm.RegisterSubsystem("test_service", false)

	ctx := context.Background()
	primaryErr := errors.New("primary failure")

	primary := func(ctx context.Context) (interface{}, error) {
		return nil, primaryErr
	}

	result, err := dm.ExecuteWithFallback(ctx, "test_service", primary, nil)

	assert.Error(t, err)
	assert.Equal(t, primaryErr, err)
	assert.Nil(t, result)
	assert.False(t, dm.IsSubsystemHealthy("test_service"))
}

func TestGetGlobalDegradationManager(t *testing.T) {
	dm1 := GetGlobalDegradationManager()
	dm2 := GetGlobalDegradationManager()

	assert.NotNil(t, dm1)
	assert.NotNil(t, dm2)
	assert.Same(t, dm1, dm2, "should return same instance")
}

func TestConcurrentUpdates(t *testing.T) {
	dm := NewDegradationManager()
	dm.RegisterSubsystem("concurrent_test", false)

	const goroutines = 10
	const iterations = 100

	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				healthy := (j % 2) == 0
				dm.UpdateSubsystemStatus("concurrent_test", healthy, nil)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Should not panic and subsystem should exist
	status, exists := dm.GetSubsystemStatus("concurrent_test")
	assert.True(t, exists)
	assert.NotNil(t, status)
}

func TestUpdateUnknownSubsystem(t *testing.T) {
	dm := NewDegradationManager()

	// Should not panic
	dm.UpdateSubsystemStatus("unknown", false, errors.New("test"))

	_, exists := dm.GetSubsystemStatus("unknown")
	assert.False(t, exists)
}
