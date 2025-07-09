package pcg

import (
	"context"
	"fmt"
	"sync"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// Registry manages all registered PCG generators and provides factory methods
// Thread-safe registry following the established locking patterns
type Registry struct {
	mu         sync.RWMutex
	generators map[ContentType]map[string]Generator
	logger     *logrus.Logger
}

// NewRegistry creates a new generator registry
func NewRegistry(logger *logrus.Logger) *Registry {
	if logger == nil {
		logger = logrus.New()
	}

	return &Registry{
		generators: make(map[ContentType]map[string]Generator),
		logger:     logger,
	}
}

// RegisterGenerator registers a new generator with the registry
// Generator names must be unique within their content type
func (r *Registry) RegisterGenerator(name string, generator Generator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	contentType := generator.GetType()

	// Initialize content type map if needed
	if r.generators[contentType] == nil {
		r.generators[contentType] = make(map[string]Generator)
	}

	// Check for duplicate names
	if _, exists := r.generators[contentType][name]; exists {
		return fmt.Errorf("generator '%s' already registered for content type '%s'", name, contentType)
	}

	r.generators[contentType][name] = generator

	r.logger.WithFields(logrus.Fields{
		"generator":    name,
		"content_type": contentType,
		"version":      generator.GetVersion(),
	}).Info("Registered PCG generator")

	return nil
}

// UnregisterGenerator removes a generator from the registry
func (r *Registry) UnregisterGenerator(contentType ContentType, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.generators[contentType] == nil {
		return fmt.Errorf("no generators registered for content type '%s'", contentType)
	}

	if _, exists := r.generators[contentType][name]; !exists {
		return fmt.Errorf("generator '%s' not found for content type '%s'", name, contentType)
	}

	delete(r.generators[contentType], name)

	r.logger.WithFields(logrus.Fields{
		"generator":    name,
		"content_type": contentType,
	}).Info("Unregistered PCG generator")

	return nil
}

// GetGenerator retrieves a specific generator by content type and name
func (r *Registry) GetGenerator(contentType ContentType, name string) (Generator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.generators[contentType] == nil {
		return nil, fmt.Errorf("no generators registered for content type '%s'", contentType)
	}

	generator, exists := r.generators[contentType][name]
	if !exists {
		return nil, fmt.Errorf("generator '%s' not found for content type '%s'", name, contentType)
	}

	return generator, nil
}

// ListGenerators returns all registered generators for a content type
func (r *Registry) ListGenerators(contentType ContentType) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.generators[contentType] == nil {
		return []string{}
	}

	names := make([]string, 0, len(r.generators[contentType]))
	for name := range r.generators[contentType] {
		names = append(names, name)
	}

	return names
}

// ListAllGenerators returns all registered generators grouped by content type
func (r *Registry) ListAllGenerators() map[ContentType][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[ContentType][]string)
	for contentType, generators := range r.generators {
		names := make([]string, 0, len(generators))
		for name := range generators {
			names = append(names, name)
		}
		result[contentType] = names
	}

	return result
}

// GenerateContent creates content using the specified generator
func (r *Registry) GenerateContent(ctx context.Context, contentType ContentType, generatorName string, params GenerationParams) (interface{}, error) {
	generator, err := r.GetGenerator(contentType, generatorName)
	if err != nil {
		return nil, err
	}

	// Validate parameters before generation
	if err := generator.Validate(params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"generator":    generatorName,
		"content_type": contentType,
		"seed":         params.Seed,
		"difficulty":   params.Difficulty,
	}).Info("Starting content generation")

	// Generate content with timeout handling
	resultChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := generator.Generate(ctx, params)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		r.logger.WithFields(logrus.Fields{
			"generator":    generatorName,
			"content_type": contentType,
		}).Info("Content generation completed successfully")
		return result, nil

	case err := <-errorChan:
		r.logger.WithFields(logrus.Fields{
			"generator":    generatorName,
			"content_type": contentType,
			"error":        err.Error(),
		}).Error("Content generation failed")
		return nil, err

	case <-ctx.Done():
		r.logger.WithFields(logrus.Fields{
			"generator":    generatorName,
			"content_type": contentType,
		}).Warn("Content generation cancelled or timed out")
		return nil, ctx.Err()
	}
}

// Factory provides convenient factory methods for content generation
type Factory struct {
	registry *Registry
	logger   *logrus.Logger
}

// NewFactory creates a new factory instance
func NewFactory(registry *Registry, logger *logrus.Logger) *Factory {
	return &Factory{
		registry: registry,
		logger:   logger,
	}
}

// GenerateTerrain generates terrain using the specified generator
func (f *Factory) GenerateTerrain(ctx context.Context, generatorName string, params TerrainParams) (*game.GameMap, error) {
	result, err := f.registry.GenerateContent(ctx, ContentTypeTerrain, generatorName, params.GenerationParams)
	if err != nil {
		return nil, err
	}

	gameMap, ok := result.(*game.GameMap)
	if !ok {
		return nil, fmt.Errorf("terrain generator returned unexpected type: %T", result)
	}

	return gameMap, nil
}

// GenerateItems generates items using the specified generator
func (f *Factory) GenerateItems(ctx context.Context, generatorName string, params ItemParams) ([]*game.Item, error) {
	result, err := f.registry.GenerateContent(ctx, ContentTypeItems, generatorName, params.GenerationParams)
	if err != nil {
		return nil, err
	}

	// Handle both single items and item arrays
	switch v := result.(type) {
	case *game.Item:
		return []*game.Item{v}, nil
	case []*game.Item:
		return v, nil
	case []game.Item:
		items := make([]*game.Item, len(v))
		for i := range v {
			items[i] = &v[i]
		}
		return items, nil
	default:
		return nil, fmt.Errorf("item generator returned unexpected type: %T", result)
	}
}

// GenerateLevel generates a level using the specified generator
func (f *Factory) GenerateLevel(ctx context.Context, generatorName string, params LevelParams) (*game.Level, error) {
	result, err := f.registry.GenerateContent(ctx, ContentTypeLevels, generatorName, params.GenerationParams)
	if err != nil {
		return nil, err
	}

	level, ok := result.(*game.Level)
	if !ok {
		return nil, fmt.Errorf("level generator returned unexpected type: %T", result)
	}

	return level, nil
}

// GenerateQuest generates a quest using the specified generator
func (f *Factory) GenerateQuest(ctx context.Context, generatorName string, params QuestParams) (*game.Quest, error) {
	result, err := f.registry.GenerateContent(ctx, ContentTypeQuests, generatorName, params.GenerationParams)
	if err != nil {
		return nil, err
	}

	quest, ok := result.(*game.Quest)
	if !ok {
		return nil, fmt.Errorf("quest generator returned unexpected type: %T", result)
	}

	return quest, nil
}

// GetDefaultRegistry returns a registry with default generators registered
func GetDefaultRegistry(logger *logrus.Logger) *Registry {
	registry := NewRegistry(logger)

	// Register default generators (these would be implemented in their respective packages)
	// This demonstrates the registration pattern

	return registry
}
