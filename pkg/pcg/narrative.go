package pcg

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// NarrativeGenerator creates overarching storylines and campaign narratives
// Builds on existing quest narratives to create cohesive campaign stories
type NarrativeGenerator struct {
	version             string
	logger              *logrus.Logger
	rng                 *rand.Rand
	storyArchetypes     map[string]*StoryArchetype
	narrativeThemes     map[string]*NarrativeTheme
	characterArchetypes map[string]*CharacterArchetype
}

// CampaignNarrative represents a complete campaign storyline
type CampaignNarrative struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Theme        string                 `json:"theme"`
	Setting      string                 `json:"setting"`
	MainPlotline *Plotline              `json:"main_plotline"`
	Subplots     []*Plotline            `json:"subplots"`
	NPCs         []*NarrativeCharacter  `json:"npcs"`
	KeyLocations []*NarrativeLocation   `json:"key_locations"`
	Timeline     []*StoryEvent          `json:"timeline"`
	Metadata     map[string]interface{} `json:"metadata"`
	Generated    time.Time              `json:"generated"`
}

// Plotline represents a story arc with beginning, middle, and end
type Plotline struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Type       PlotType               `json:"type"`
	Acts       []*StoryAct            `json:"acts"`
	Characters []string               `json:"characters"`
	Locations  []string               `json:"locations"`
	Hooks      []string               `json:"hooks"`
	Climax     string                 `json:"climax"`
	Resolution string                 `json:"resolution"`
	Properties map[string]interface{} `json:"properties"`
}

// StoryAct represents a major section of a plotline
type StoryAct struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Events      []*StoryEvent          `json:"events"`
	Objectives  []string               `json:"objectives"`
	Properties  map[string]interface{} `json:"properties"`
}

// StoryEvent represents a significant narrative moment
type StoryEvent struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Type         EventType              `json:"type"`
	Participants []string               `json:"participants"`
	Location     string                 `json:"location"`
	Trigger      string                 `json:"trigger"`
	Consequences []string               `json:"consequences"`
	Properties   map[string]interface{} `json:"properties"`
}

// NarrativeCharacter represents a key story character
type NarrativeCharacter struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Archetype     string                 `json:"archetype"`
	Role          CharacterRole          `json:"role"`
	Motivation    string                 `json:"motivation"`
	Background    string                 `json:"background"`
	Personality   []string               `json:"personality"`
	Relationships map[string]string      `json:"relationships"`
	Arc           *CharacterArc          `json:"arc"`
	Properties    map[string]interface{} `json:"properties"`
}

// NarrativeLocation represents a key story location
type NarrativeLocation struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         LocationType           `json:"type"`
	Description  string                 `json:"description"`
	Significance string                 `json:"significance"`
	History      string                 `json:"history"`
	Properties   map[string]interface{} `json:"properties"`
}

// CharacterArc represents a character's journey through the story
type CharacterArc struct {
	StartState     string   `json:"start_state"`
	Developments   []string `json:"developments"`
	Transformation string   `json:"transformation"`
	EndState       string   `json:"end_state"`
}

// Story template types for generation
type StoryArchetype struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Structure   []string `yaml:"structure"`
	Themes      []string `yaml:"themes"`
	Conflicts   []string `yaml:"conflicts"`
}

type NarrativeTheme struct {
	Name     string   `yaml:"name"`
	Tone     string   `yaml:"tone"`
	Motifs   []string `yaml:"motifs"`
	Symbols  []string `yaml:"symbols"`
	Messages []string `yaml:"messages"`
}

type CharacterArchetype struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Motivations []string `yaml:"motivations"`
	Traits      []string `yaml:"traits"`
	SpeechStyle []string `yaml:"speech_style"`
}

// Enums for story structure
type PlotType string

const (
	PlotTypeMain      PlotType = "main"
	PlotTypeSubplot   PlotType = "subplot"
	PlotTypeSideQuest PlotType = "side_quest"
	PlotTypePersonal  PlotType = "personal"
)

type EventType string

const (
	EventTypeEncounter  EventType = "encounter"
	EventTypeDiscovery  EventType = "discovery"
	EventTypeBetrayal   EventType = "betrayal"
	EventTypeRevelation EventType = "revelation"
	EventTypeConflict   EventType = "conflict"
	EventTypeResolution EventType = "resolution"
	EventTypeSacrifice  EventType = "sacrifice"
	EventTypeTransition EventType = "transition"
)

type CharacterRole string

const (
	RoleProtagonist  CharacterRole = "protagonist"
	RoleAntagonist   CharacterRole = "antagonist"
	RoleAlly         CharacterRole = "ally"
	RoleMentor       CharacterRole = "mentor"
	RoleGuardian     CharacterRole = "guardian"
	RoleHerald       CharacterRole = "herald"
	RoleTrickster    CharacterRole = "trickster"
	RoleShapeshifter CharacterRole = "shapeshifter"
)

type LocationType string

const (
	LocationHometown   LocationType = "hometown"
	LocationDungeon    LocationType = "dungeon"
	LocationCastle     LocationType = "castle"
	LocationWilderness LocationType = "wilderness"
	LocationCity       LocationType = "city"
	LocationShrine     LocationType = "shrine"
	LocationRuins      LocationType = "ruins"
	LocationPortal     LocationType = "portal"
)

// NewNarrativeGenerator creates a new narrative generator
func NewNarrativeGenerator(logger *logrus.Logger) *NarrativeGenerator {
	if logger == nil {
		logger = logrus.New()
	}

	ng := &NarrativeGenerator{
		version:             "1.0.0",
		logger:              logger,
		rng:                 rand.New(rand.NewSource(time.Now().UnixNano())),
		storyArchetypes:     make(map[string]*StoryArchetype),
		narrativeThemes:     make(map[string]*NarrativeTheme),
		characterArchetypes: make(map[string]*CharacterArchetype),
	}

	ng.initializeTemplates()
	return ng
}

// Generate creates a complete campaign narrative
func (ng *NarrativeGenerator) Generate(ctx context.Context, params GenerationParams) (interface{}, error) {
	narrativeParams, ok := params.Constraints["narrative_params"].(NarrativeParams)
	if !ok {
		return nil, fmt.Errorf("invalid parameters for narrative generation: expected narrative_params in constraints")
	}

	if err := ng.Validate(params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Initialize RNG with seed for deterministic generation
	ng.rng = rand.New(rand.NewSource(params.Seed))

	ng.logger.WithFields(logrus.Fields{
		"narrative_type": narrativeParams.NarrativeType,
		"theme":          narrativeParams.Theme,
		"length":         narrativeParams.CampaignLength,
	}).Info("generating campaign narrative")

	narrative, err := ng.generateCampaignNarrative(ctx, params, narrativeParams)
	if err != nil {
		return nil, fmt.Errorf("narrative generation failed: %w", err)
	}

	ng.logger.WithField("narrative_id", narrative.ID).Info("campaign narrative generation completed")
	return narrative, nil
}

// generateCampaignNarrative creates the complete campaign story
func (ng *NarrativeGenerator) generateCampaignNarrative(ctx context.Context, params GenerationParams, narrativeParams NarrativeParams) (*CampaignNarrative, error) {
	// Select story archetype based on parameters
	archetype := ng.selectStoryArchetype(narrativeParams.Theme, narrativeParams.NarrativeType)
	theme := ng.selectNarrativeTheme(narrativeParams.Theme)

	narrative := &CampaignNarrative{
		ID:           fmt.Sprintf("narrative_%d", ng.rng.Int63()),
		Title:        ng.generateTitle(archetype, theme),
		Theme:        narrativeParams.Theme,
		Setting:      ng.generateSetting(narrativeParams),
		Subplots:     make([]*Plotline, 0),
		NPCs:         make([]*NarrativeCharacter, 0),
		KeyLocations: make([]*NarrativeLocation, 0),
		Timeline:     make([]*StoryEvent, 0),
		Metadata:     make(map[string]interface{}),
		Generated:    time.Now(),
	}

	// Generate main plotline
	mainPlot, err := ng.generateMainPlotline(archetype, theme, narrativeParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate main plotline: %w", err)
	}
	narrative.MainPlotline = mainPlot

	// Generate key characters
	narrative.NPCs = ng.generateKeyCharacters(mainPlot, narrativeParams)

	// Generate key locations
	narrative.KeyLocations = ng.generateKeyLocations(mainPlot, narrativeParams)

	// Generate subplots
	subplotCount := ng.calculateSubplotCount(narrativeParams.CampaignLength)
	for i := 0; i < subplotCount; i++ {
		subplot := ng.generateSubplot(narrative.NPCs, narrative.KeyLocations, theme, narrativeParams)
		narrative.Subplots = append(narrative.Subplots, subplot)
	}

	// Create timeline from all plot events
	narrative.Timeline = ng.createTimeline(narrative.MainPlotline, narrative.Subplots)

	// Add metadata
	narrative.Metadata["character_count"] = len(narrative.NPCs)
	narrative.Metadata["location_count"] = len(narrative.KeyLocations)
	narrative.Metadata["subplot_count"] = len(narrative.Subplots)
	narrative.Metadata["total_events"] = len(narrative.Timeline)

	return narrative, nil
}

// generateMainPlotline creates the central story arc
func (ng *NarrativeGenerator) generateMainPlotline(archetype *StoryArchetype, theme *NarrativeTheme, params NarrativeParams) (*Plotline, error) {
	plotline := &Plotline{
		ID:         "main_plot",
		Title:      ng.generatePlotTitle(archetype),
		Type:       PlotTypeMain,
		Acts:       make([]*StoryAct, 0),
		Characters: make([]string, 0),
		Locations:  make([]string, 0),
		Hooks:      ng.generatePlotHooks(archetype, theme),
		Properties: make(map[string]interface{}),
	}

	// Generate story acts based on campaign length
	actCount := ng.calculateActCount(params.CampaignLength)
	for i := 0; i < actCount; i++ {
		act := ng.generateStoryAct(i, actCount, archetype, theme, params)
		plotline.Acts = append(plotline.Acts, act)
	}

	// Generate climax and resolution
	plotline.Climax = ng.generateClimax(archetype, theme)
	plotline.Resolution = ng.generateResolution(archetype, theme)

	return plotline, nil
}

// generateKeyCharacters creates important NPCs for the narrative
func (ng *NarrativeGenerator) generateKeyCharacters(mainPlot *Plotline, params NarrativeParams) []*NarrativeCharacter {
	characters := make([]*NarrativeCharacter, 0)

	// Generate protagonist (if NPC-focused campaign)
	if params.NarrativeType == NarrativeLinear || params.NarrativeType == NarrativeEpisodic {
		protagonist := ng.generateCharacter(RoleProtagonist, params)
		characters = append(characters, protagonist)
	}

	// Generate antagonist
	antagonist := ng.generateCharacter(RoleAntagonist, params)
	characters = append(characters, antagonist)

	// Generate supporting characters based on campaign length
	supportingCount := ng.calculateSupportingCharacterCount(params.CampaignLength)
	supportingRoles := []CharacterRole{RoleAlly, RoleMentor, RoleGuardian, RoleHerald, RoleTrickster}

	for i := 0; i < supportingCount; i++ {
		role := supportingRoles[i%len(supportingRoles)]
		character := ng.generateCharacter(role, params)
		characters = append(characters, character)
	}

	return characters
}

// generateCharacter creates a single narrative character
func (ng *NarrativeGenerator) generateCharacter(role CharacterRole, params NarrativeParams) *NarrativeCharacter {
	archetype := ng.selectCharacterArchetype(role)

	character := &NarrativeCharacter{
		ID:            fmt.Sprintf("char_%d", ng.rng.Int63()),
		Name:          ng.generateCharacterName(archetype),
		Archetype:     archetype.Name,
		Role:          role,
		Motivation:    ng.selectRandom(archetype.Motivations),
		Background:    ng.generateCharacterBackground(archetype, params),
		Personality:   ng.selectMultipleRandom(archetype.Traits, 2, 4),
		Relationships: make(map[string]string),
		Properties:    make(map[string]interface{}),
	}

	// Generate character arc
	character.Arc = ng.generateCharacterArc(character, params)

	return character
}

// Helper methods for generation

// initializeTemplates sets up the default story templates
func (ng *NarrativeGenerator) initializeTemplates() {
	// Story archetypes
	ng.storyArchetypes["hero_journey"] = &StoryArchetype{
		Name:        "Hero's Journey",
		Description: "Classic hero's journey with call to adventure",
		Structure:   []string{"ordinary_world", "call_to_adventure", "refusal", "mentor", "threshold", "tests", "ordeal", "reward", "road_back", "resurrection", "return"},
		Themes:      []string{"growth", "sacrifice", "destiny"},
		Conflicts:   []string{"good_vs_evil", "order_vs_chaos", "individual_vs_society"},
	}

	ng.storyArchetypes["tragedy"] = &StoryArchetype{
		Name:        "Tragedy",
		Description: "Downfall of protagonist due to fatal flaw",
		Structure:   []string{"exposition", "rising_action", "climax", "falling_action", "catastrophe"},
		Themes:      []string{"hubris", "fate", "justice"},
		Conflicts:   []string{"man_vs_self", "man_vs_fate", "corruption"},
	}

	ng.storyArchetypes["mystery"] = &StoryArchetype{
		Name:        "Mystery",
		Description: "Investigation and revelation of hidden truth",
		Structure:   []string{"incident", "investigation", "clues", "red_herrings", "revelation", "resolution"},
		Themes:      []string{"truth", "justice", "hidden_knowledge"},
		Conflicts:   []string{"truth_vs_deception", "order_vs_chaos"},
	}

	// Narrative themes
	ng.narrativeThemes["classic"] = &NarrativeTheme{
		Name:     "Classic Fantasy",
		Tone:     "heroic",
		Motifs:   []string{"ancient_prophecy", "magical_artifact", "dark_lord", "chosen_one"},
		Symbols:  []string{"sword", "crown", "tower", "dragon"},
		Messages: []string{"good_triumphs", "friendship_matters", "courage_conquers_fear"},
	}

	ng.narrativeThemes["grimdark"] = &NarrativeTheme{
		Name:     "Grimdark",
		Tone:     "dark",
		Motifs:   []string{"corruption", "moral_ambiguity", "pyrrhic_victory", "survival"},
		Symbols:  []string{"broken_crown", "blood", "ravens", "ruins"},
		Messages: []string{"power_corrupts", "survival_at_any_cost", "hope_is_fleeting"},
	}

	// Character archetypes
	ng.characterArchetypes["noble_hero"] = &CharacterArchetype{
		Name:        "Noble Hero",
		Description: "Honorable champion of justice",
		Motivations: []string{"protect_innocent", "uphold_justice", "fulfill_destiny"},
		Traits:      []string{"brave", "honorable", "selfless", "determined"},
		SpeechStyle: []string{"formal", "inspiring", "direct"},
	}

	ng.characterArchetypes["dark_lord"] = &CharacterArchetype{
		Name:        "Dark Lord",
		Description: "Powerful evil ruler seeking dominion",
		Motivations: []string{"conquer_world", "gain_power", "spread_darkness"},
		Traits:      []string{"ruthless", "intelligent", "charismatic", "cruel"},
		SpeechStyle: []string{"commanding", "menacing", "eloquent"},
	}
}

// selectStoryArchetype chooses appropriate story structure
func (ng *NarrativeGenerator) selectStoryArchetype(theme string, narrativeType NarrativeType) *StoryArchetype {
	// Simple selection based on theme - in a full implementation this would be more sophisticated
	switch theme {
	case "classic":
		return ng.storyArchetypes["hero_journey"]
	case "grimdark":
		return ng.storyArchetypes["tragedy"]
	case "mystery":
		return ng.storyArchetypes["mystery"]
	default:
		return ng.storyArchetypes["hero_journey"]
	}
}

// selectNarrativeTheme chooses thematic elements
func (ng *NarrativeGenerator) selectNarrativeTheme(theme string) *NarrativeTheme {
	if narrativeTheme, exists := ng.narrativeThemes[theme]; exists {
		return narrativeTheme
	}
	return ng.narrativeThemes["classic"]
}

// selectCharacterArchetype chooses character template based on role
func (ng *NarrativeGenerator) selectCharacterArchetype(role CharacterRole) *CharacterArchetype {
	switch role {
	case RoleProtagonist, RoleAlly:
		return ng.characterArchetypes["noble_hero"]
	case RoleAntagonist:
		return ng.characterArchetypes["dark_lord"]
	default:
		return ng.characterArchetypes["noble_hero"]
	}
}

// Utility methods for random selection and generation
func (ng *NarrativeGenerator) selectRandom(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[ng.rng.Intn(len(items))]
}

func (ng *NarrativeGenerator) selectMultipleRandom(items []string, min, max int) []string {
	if len(items) == 0 {
		return []string{}
	}

	count := min + ng.rng.Intn(max-min+1)
	if count > len(items) {
		count = len(items)
	}

	// Shuffle and take first count items
	shuffled := make([]string, len(items))
	copy(shuffled, items)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := ng.rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:count]
}

// Generation helper methods (simplified implementations)
func (ng *NarrativeGenerator) generateTitle(archetype *StoryArchetype, theme *NarrativeTheme) string {
	prefixes := []string{"The", "Rise of", "Fall of", "Chronicles of", "Legend of"}
	subjects := theme.Symbols

	prefix := ng.selectRandom(prefixes)
	subject := ng.selectRandom(subjects)

	return fmt.Sprintf("%s %s", prefix, strings.Title(subject))
}

func (ng *NarrativeGenerator) generateSetting(params NarrativeParams) string {
	settings := []string{
		"A realm torn by ancient conflicts",
		"A world where magic and technology clash",
		"A dying empire on the edge of collapse",
		"A frontier land full of mysteries",
		"A kingdom threatened by dark forces",
	}
	return ng.selectRandom(settings)
}

func (ng *NarrativeGenerator) generatePlotTitle(archetype *StoryArchetype) string {
	return fmt.Sprintf("The %s", archetype.Name)
}

func (ng *NarrativeGenerator) generatePlotHooks(archetype *StoryArchetype, theme *NarrativeTheme) []string {
	hooks := []string{
		fmt.Sprintf("Ancient %s surfaces with dark purpose", ng.selectRandom(theme.Symbols)),
		fmt.Sprintf("Mysterious events plague the land"),
		fmt.Sprintf("A %s emerges to challenge the status quo", ng.selectRandom(theme.Motifs)),
	}
	return hooks
}

func (ng *NarrativeGenerator) generateClimax(archetype *StoryArchetype, theme *NarrativeTheme) string {
	climaxTemplates := []string{
		"Final confrontation with the %s",
		"Ultimate test of %s",
		"Revelation that changes everything",
	}
	template := ng.selectRandom(climaxTemplates)
	return fmt.Sprintf(template, ng.selectRandom(theme.Motifs))
}

func (ng *NarrativeGenerator) generateResolution(archetype *StoryArchetype, theme *NarrativeTheme) string {
	return fmt.Sprintf("The realm finds %s", ng.selectRandom([]string{"peace", "balance", "new hope", "redemption"}))
}

func (ng *NarrativeGenerator) generateCharacterName(archetype *CharacterArchetype) string {
	// Simple name generation - could be expanded with proper name libraries
	prefixes := []string{"Arath", "Drak", "Eld", "Grim", "Kael", "Mor", "Rath", "Thane", "Vel", "Zar"}
	suffixes := []string{"an", "as", "eth", "ion", "or", "us", "wyn", "dar", "rim", "ok"}

	return ng.selectRandom(prefixes) + ng.selectRandom(suffixes)
}

func (ng *NarrativeGenerator) generateCharacterBackground(archetype *CharacterArchetype, params NarrativeParams) string {
	backgrounds := []string{
		fmt.Sprintf("A %s who seeks to %s", strings.ToLower(archetype.Name), ng.selectRandom(archetype.Motivations)),
		fmt.Sprintf("Born into %s, shaped by %s", "humble origins", "great trials"),
		fmt.Sprintf("Once %s, now %s", "a different person", "transformed by events"),
	}
	return ng.selectRandom(backgrounds)
}

func (ng *NarrativeGenerator) generateCharacterArc(character *NarrativeCharacter, params NarrativeParams) *CharacterArc {
	return &CharacterArc{
		StartState:     fmt.Sprintf("Begins as %s", strings.ToLower(character.Archetype)),
		Developments:   []string{"faces challenges", "learns hard truths", "makes difficult choices"},
		Transformation: "undergoes fundamental change",
		EndState:       "emerges transformed",
	}
}

// Calculation methods
func (ng *NarrativeGenerator) calculateSubplotCount(length string) int {
	switch length {
	case "short":
		return 1 + ng.rng.Intn(2) // 1-2 subplots
	case "medium":
		return 2 + ng.rng.Intn(3) // 2-4 subplots
	case "long":
		return 3 + ng.rng.Intn(4) // 3-6 subplots
	default:
		return 2
	}
}

func (ng *NarrativeGenerator) calculateActCount(length string) int {
	switch length {
	case "short":
		return 3 // Classic 3-act structure
	case "medium":
		return 5 // Extended structure
	case "long":
		return 7 // Epic structure
	default:
		return 3
	}
}

func (ng *NarrativeGenerator) calculateSupportingCharacterCount(length string) int {
	switch length {
	case "short":
		return 2 + ng.rng.Intn(2) // 2-3 characters
	case "medium":
		return 3 + ng.rng.Intn(3) // 3-5 characters
	case "long":
		return 4 + ng.rng.Intn(4) // 4-7 characters
	default:
		return 3
	}
}

// Stub methods that would be fully implemented
func (ng *NarrativeGenerator) generateStoryAct(index, total int, archetype *StoryArchetype, theme *NarrativeTheme, params NarrativeParams) *StoryAct {
	return &StoryAct{
		ID:          fmt.Sprintf("act_%d", index+1),
		Title:       fmt.Sprintf("Act %d", index+1),
		Description: fmt.Sprintf("Story development phase %d", index+1),
		Events:      make([]*StoryEvent, 0),
		Objectives:  []string{fmt.Sprintf("Complete act %d objectives", index+1)},
		Properties:  make(map[string]interface{}),
	}
}

func (ng *NarrativeGenerator) generateKeyLocations(mainPlot *Plotline, params NarrativeParams) []*NarrativeLocation {
	locations := make([]*NarrativeLocation, 0)
	locationTypes := []LocationType{LocationHometown, LocationDungeon, LocationCastle, LocationWilderness}

	for i, locType := range locationTypes {
		location := &NarrativeLocation{
			ID:           fmt.Sprintf("loc_%d", i),
			Name:         ng.generateLocationName(locType),
			Type:         locType,
			Description:  fmt.Sprintf("A significant %s in the story", string(locType)),
			Significance: "Important to the main plot",
			History:      "Has a rich and troubled past",
			Properties:   make(map[string]interface{}),
		}
		locations = append(locations, location)
	}

	return locations
}

func (ng *NarrativeGenerator) generateLocationName(locType LocationType) string {
	prefixes := map[LocationType][]string{
		LocationHometown:   {"Green", "Old", "Fair", "New"},
		LocationDungeon:    {"Dark", "Forgotten", "Ancient", "Lost"},
		LocationCastle:     {"High", "Storm", "Iron", "Golden"},
		LocationWilderness: {"Whispering", "Shadow", "Endless", "Wild"},
	}

	suffixes := map[LocationType][]string{
		LocationHometown:   {"haven", "bridge", "ford", "vale"},
		LocationDungeon:    {"depths", "halls", "caverns", "maze"},
		LocationCastle:     {"keep", "tower", "citadel", "fortress"},
		LocationWilderness: {"woods", "plains", "marsh", "peaks"},
	}

	prefix := ng.selectRandom(prefixes[locType])
	suffix := ng.selectRandom(suffixes[locType])

	return fmt.Sprintf("%s %s", prefix, strings.Title(suffix))
}

func (ng *NarrativeGenerator) generateSubplot(characters []*NarrativeCharacter, locations []*NarrativeLocation, theme *NarrativeTheme, params NarrativeParams) *Plotline {
	return &Plotline{
		ID:         fmt.Sprintf("subplot_%d", ng.rng.Int63()),
		Title:      "Character Subplot",
		Type:       PlotTypeSubplot,
		Acts:       make([]*StoryAct, 0),
		Characters: []string{ng.selectRandom([]string{"character1", "character2"})},
		Locations:  []string{ng.selectRandom([]string{"location1", "location2"})},
		Hooks:      []string{"Personal conflict emerges"},
		Climax:     "Personal resolution",
		Resolution: "Character growth achieved",
		Properties: make(map[string]interface{}),
	}
}

func (ng *NarrativeGenerator) createTimeline(mainPlot *Plotline, subplots []*Plotline) []*StoryEvent {
	events := make([]*StoryEvent, 0)

	// Add main plot events
	for i, act := range mainPlot.Acts {
		event := &StoryEvent{
			ID:           fmt.Sprintf("main_event_%d", i),
			Title:        act.Title,
			Description:  act.Description,
			Type:         EventTypeTransition,
			Participants: []string{"protagonist"},
			Location:     "various",
			Trigger:      "story progression",
			Consequences: []string{"plot advancement"},
			Properties:   make(map[string]interface{}),
		}
		events = append(events, event)
	}

	return events
}

// Interface compliance methods

// GetType returns the content type this generator produces
func (ng *NarrativeGenerator) GetType() ContentType {
	return ContentTypeNarrative
}

// GetVersion returns the generator version
func (ng *NarrativeGenerator) GetVersion() string {
	return ng.version
}

// Validate checks if the provided parameters are valid
func (ng *NarrativeGenerator) Validate(params GenerationParams) error {
	narrativeParams, ok := params.Constraints["narrative_params"].(NarrativeParams)
	if !ok {
		return fmt.Errorf("invalid parameters: expected narrative_params in constraints")
	}

	validLengths := []string{"short", "medium", "long"}
	validLength := false
	for _, length := range validLengths {
		if narrativeParams.CampaignLength == length {
			validLength = true
			break
		}
	}
	if !validLength {
		return fmt.Errorf("campaign length must be one of: %v, got %s", validLengths, narrativeParams.CampaignLength)
	}

	if narrativeParams.Theme == "" {
		return fmt.Errorf("theme cannot be empty")
	}

	return nil
}
