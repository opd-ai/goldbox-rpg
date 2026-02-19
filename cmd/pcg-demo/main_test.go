package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Seed != 12345 {
		t.Errorf("expected default seed 12345, got %d", cfg.Seed)
	}
	if cfg.Logger == nil {
		t.Error("expected non-nil logger")
	}
	if cfg.Output == nil {
		t.Error("expected non-nil output")
	}
	if cfg.NumTerrains != 3 {
		t.Errorf("expected 3 terrains, got %d", cfg.NumTerrains)
	}
	if cfg.NumItems != 5 {
		t.Errorf("expected 5 items, got %d", cfg.NumItems)
	}
}

func TestRunDemo(t *testing.T) {
	tests := []struct {
		name        string
		seed        int64
		numTerrains int
		numItems    int
		wantErr     bool
		wantOutput  []string
	}{
		{
			name:        "zero iterations - no generators needed",
			seed:        1,
			numTerrains: 0,
			numItems:    0,
			wantErr:     false,
			wantOutput: []string{
				"=== PCG Performance Metrics Demo ===",
				"=== Initial Metrics ===",
				"=== Performing Content Generation ===",
				"=== Final Metrics ===",
				"=== Manager Statistics ===",
				"=== Resetting Metrics ===",
				"=== Demo Complete ===",
			},
		},
		{
			name:        "different seed no generations",
			seed:        99999,
			numTerrains: 0,
			numItems:    0,
			wantErr:     false,
			wantOutput: []string{
				"seed 99999",
			},
		},
		{
			name:        "terrain generation fails without generators",
			seed:        42,
			numTerrains: 1,
			numItems:    0,
			wantErr:     true,
			wantOutput:  []string{},
		},
		{
			name:        "item generation fails without generators",
			seed:        42,
			numTerrains: 0,
			numItems:    1,
			wantErr:     true,
			wantOutput:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)
			logger.SetOutput(&buf)

			cfg := Config{
				Seed:        tt.seed,
				Logger:      logger,
				Output:      &buf,
				NumTerrains: tt.numTerrains,
				NumItems:    tt.numItems,
			}

			err := RunDemo(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunDemo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected string %q", want)
				}
			}
		})
	}
}

func TestRunDemoMetricsContent(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	logger.SetOutput(&buf)

	cfg := Config{
		Seed:        12345,
		Logger:      logger,
		Output:      &buf,
		NumTerrains: 0, // No generation without registered generators
		NumItems:    0,
	}

	err := RunDemo(cfg)
	if err != nil {
		t.Fatalf("RunDemo() error = %v", err)
	}

	output := buf.String()

	// Verify metrics are shown
	if !strings.Contains(output, "Total Generations") {
		t.Error("output should show Total Generations metric")
	}

	if !strings.Contains(output, "Cache Hit Ratio") {
		t.Error("output should show Cache Hit Ratio")
	}

	// Verify manager statistics section is shown
	if !strings.Contains(output, "Manager Statistics") {
		t.Error("output should show Manager Statistics section")
	}
}

func TestRunDemoNilOutput(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := Config{
		Seed:        1,
		Logger:      logger,
		Output:      nil, // Will default to stdout
		NumTerrains: 0,
		NumItems:    0,
	}

	// This should not panic with nil output
	err := RunDemo(cfg)
	if err != nil {
		t.Errorf("RunDemo() with nil output error = %v", err)
	}
}

func TestPrettyPrint(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name:    "simple map",
			data:    map[string]int{"count": 5},
			wantErr: false,
		},
		{
			name:    "struct",
			data:    struct{ Name string }{"test"},
			wantErr: false,
		},
		{
			name:    "nil",
			data:    nil,
			wantErr: false,
		},
		{
			name:    "channel cannot marshal",
			data:    make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := prettyPrint(&buf, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("prettyPrint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunDemoOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	logger.SetOutput(&buf)

	cfg := Config{
		Seed:        54321,
		Logger:      logger,
		Output:      &buf,
		NumTerrains: 0, // No generation without registered generators
		NumItems:    0,
	}

	err := RunDemo(cfg)
	if err != nil {
		t.Fatalf("RunDemo() error = %v", err)
	}

	output := buf.String()

	// Check for proper section headers
	sections := []string{
		"=== PCG Performance Metrics Demo ===",
		"=== Initial Metrics ===",
		"=== Performing Content Generation ===",
		"=== Final Metrics ===",
		"=== Manager Statistics ===",
		"=== Resetting Metrics ===",
		"=== Demo Complete ===",
	}

	for _, section := range sections {
		if !strings.Contains(output, section) {
			t.Errorf("missing section header: %s", section)
		}
	}
}
