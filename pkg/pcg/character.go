package pcg

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// NPCGenerator creates NPCs with procedural personalities and motivations
// Generates cohesive character profiles that enhance narrative depth and world immersion
type NPCGenerator struct {
	version string
	logger  *logrus.Logger
	rng     *rand.Rand
}

// NewNPCGenerator creates a new character generator instance
func NewNPCGenerator(logger *logrus.Logger) *NPCGenerator {
	if logger == nil {
		logger = logrus.New()
	}

	return &NPCGenerator{
		version: "1.0.0",
		logger:  logger,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Generate creates characters based on the provided parameters
// Returns generated NPCs with complete personality profiles
func (cg *NPCGenerator) Generate(ctx context.Context, params GenerationParams) (interface{}, error) {
	if err := cg.Validate(params); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Use seed for deterministic generation
	rng := rand.New(rand.NewSource(params.Seed))
	cg.rng = rng

	characterParams, ok := params.Constraints["character_params"].(CharacterParams)
	if !ok {
		// Use default parameters
		characterParams = CharacterParams{
			GenerationParams: params,
			CharacterType:    CharacterTypeGeneric,
			PersonalityDepth: 3,
			MotivationCount:  rng.Intn(3) + 1, // 1-3 motivations
			BackgroundType:   BackgroundUrban,
			SocialClass:      SocialClassPeasant,
			AgeRange:         AgeRangeAdult,
			UniqueTraits:     rng.Intn(3) + 2, // 2-4 traits
		}
	} else {
		// Apply defaults for unset values
		if len(characterParams.Alignment) == 0 {
			characterParams.Alignment = cg.generateAlignment(rng)
		}
		if characterParams.PersonalityDepth == 0 {
			characterParams.PersonalityDepth = 3
		}
		if characterParams.MotivationCount == 0 {
			characterParams.MotivationCount = rng.Intn(3) + 1
		}
		if characterParams.UniqueTraits == 0 {
			characterParams.UniqueTraits = rng.Intn(3) + 2
		}
	}

	cg.logger.WithFields(logrus.Fields{
		"seed":              params.Seed,
		"character_type":    characterParams.CharacterType,
		"personality_depth": characterParams.PersonalityDepth,
		"background":        characterParams.BackgroundType,
	}).Info("generating character")

	start := time.Now()

	// Generate the character
	npc, err := cg.GenerateNPC(ctx, characterParams.CharacterType, characterParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate character: %w", err)
	}

	duration := time.Since(start)
	cg.logger.WithFields(logrus.Fields{
		"duration":  duration,
		"character": npc.Character.Name,
		"generated": "success",
	}).Info("character generation completed")

	return npc, nil
}

// GenerateNPC creates a single NPC with personality and motivations
func (cg *NPCGenerator) GenerateNPC(ctx context.Context, characterType CharacterType, params CharacterParams) (*game.NPC, error) {
	// Use seed for deterministic generation
	rng := rand.New(rand.NewSource(params.Seed))
	cg.rng = rng

	// Generate base character attributes
	baseChar, err := cg.generateBaseCharacter(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base character: %w", err)
	}

	// Generate personality profile
	personality, err := cg.GeneratePersonality(ctx, baseChar, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate personality: %w", err)
	}

	// Create NPC with behavior and faction
	// Note: We'll store personality in Dialog metadata for now until we extend Character
	npc := &game.NPC{
		Character: *baseChar.Clone(), // Use Clone to avoid mutex copy issues
		Behavior:  cg.generateBehavior(characterType, params),
		Faction:   params.Faction,
		Dialog:    cg.generateDialog(personality, params),
		LootTable: cg.generateLootTable(characterType, params),
	}

	return npc, nil
}

// GenerateNPCGroup creates a collection of related NPCs
func (cg *NPCGenerator) GenerateNPCGroup(ctx context.Context, groupType NPCGroupType, params CharacterParams) ([]*game.NPC, error) {
	var npcs []*game.NPC
	var groupSize int

	// Determine group size based on type
	switch groupType {
	case NPCGroupFamily:
		groupSize = cg.rng.Intn(5) + 2 // 2-6 family members
	case NPCGroupGuards:
		groupSize = cg.rng.Intn(6) + 3 // 3-8 guards
	case NPCGroupMerchants:
		groupSize = cg.rng.Intn(4) + 2 // 2-5 merchants
	case NPCGroupCultists:
		groupSize = cg.rng.Intn(8) + 4 // 4-11 cultists
	case NPCGroupBandits:
		groupSize = cg.rng.Intn(7) + 3 // 3-9 bandits
	case NPCGroupScholars:
		groupSize = cg.rng.Intn(4) + 2 // 2-5 scholars
	case NPCGroupCrafters:
		groupSize = cg.rng.Intn(5) + 3 // 3-7 crafters
	default:
		groupSize = cg.rng.Intn(4) + 2 // 2-5 default
	}
	// Generate related characters
	for i := 0; i < groupSize; i++ {
		// Adjust character type based on group and position
		charType := cg.selectCharacterTypeForGroup(groupType, i, groupSize)

		// Create modified parameters for group coherence
		groupParams := params
		groupParams.CharacterType = charType
		groupParams.Seed = params.Seed + int64(i*1000) // Ensure unique seed for each group member
		if i == 0 {
			// Leader gets higher social class and more complex personality
			groupParams.SocialClass = cg.elevatedSocialClass(params.SocialClass)
			groupParams.PersonalityDepth = params.PersonalityDepth + 1
		}

		npc, err := cg.GenerateNPC(ctx, charType, groupParams)
		if err != nil {
			return nil, fmt.Errorf("failed to generate group member %d: %w", i, err)
		}

		npcs = append(npcs, npc)
	}

	// Add group relationships and connections
	cg.establishGroupRelationships(npcs, groupType)

	return npcs, nil
}

// GeneratePersonality creates personality traits and motivations
func (cg *NPCGenerator) GeneratePersonality(ctx context.Context, character *game.Character, params CharacterParams) (*PersonalityProfile, error) {
	profile := &PersonalityProfile{
		Alignment:   params.Alignment,
		Temperament: cg.generateTemperament(),
		Values:      cg.generateValues(params),
		Fears:       cg.generateFears(params),
		Speech:      cg.generateSpeechPattern(params),
	}

	// Generate personality traits
	traits, err := cg.generatePersonalityTraits(params.UniqueTraits, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate personality traits: %w", err)
	}
	profile.Traits = traits

	// Generate motivations
	motivations, err := cg.generateMotivations(params.MotivationCount, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate motivations: %w", err)
	}
	profile.Motivations = motivations

	return profile, nil
}

// GetType returns the content type for character generation
func (cg *NPCGenerator) GetType() ContentType {
	return ContentTypeCharacters
}

// GetVersion returns the generator version
func (cg *NPCGenerator) GetVersion() string {
	return cg.version
}

// Validate checks if the provided parameters are valid
func (cg *NPCGenerator) Validate(params GenerationParams) error {
	if params.Seed == 0 {
		return fmt.Errorf("seed cannot be zero")
	}

	if params.Difficulty < 1 || params.Difficulty > 20 {
		return fmt.Errorf("difficulty must be between 1 and 20, got %d", params.Difficulty)
	}

	if params.PlayerLevel < 1 || params.PlayerLevel > 20 {
		return fmt.Errorf("player level must be between 1 and 20, got %d", params.PlayerLevel)
	}

	// Validate character-specific constraints if present
	if characterParams, ok := params.Constraints["character_params"].(CharacterParams); ok {
		if characterParams.PersonalityDepth < 1 || characterParams.PersonalityDepth > 5 {
			return fmt.Errorf("personality depth must be between 1 and 5, got %d", characterParams.PersonalityDepth)
		}

		if characterParams.MotivationCount < 0 || characterParams.MotivationCount > 10 {
			return fmt.Errorf("motivation count must be between 0 and 10, got %d", characterParams.MotivationCount)
		}

		if characterParams.UniqueTraits < 1 || characterParams.UniqueTraits > 10 {
			return fmt.Errorf("unique traits must be between 1 and 10, got %d", characterParams.UniqueTraits)
		}
	}

	return nil
}

// generateBaseCharacter creates the fundamental character attributes
func (cg *NPCGenerator) generateBaseCharacter(params CharacterParams) (*game.Character, error) {
	// Generate basic attributes based on character type and social class
	stats := cg.generateAttributesByType(params.CharacterType, params.SocialClass)

	// Generate name based on background and gender
	name := cg.generateName(params.BackgroundType, params.Gender)

	// Generate description
	description := cg.generateDescription(params)

	// Create character with generated attributes
	character := &game.Character{
		ID:           fmt.Sprintf("npc_%d", cg.rng.Int63()),
		Name:         name,
		Description:  description,
		Class:        cg.selectCharacterClass(params.CharacterType),
		Strength:     stats.Strength,
		Dexterity:    stats.Dexterity,
		Constitution: stats.Constitution,
		Intelligence: stats.Intelligence,
		Wisdom:       stats.Wisdom,
		Charisma:     stats.Charisma,
		Level:        cg.generateLevel(params),
		Gold:         cg.generateStartingGold(params.SocialClass),
		Equipment:    make(map[game.EquipmentSlot]game.Item),
		Inventory:    []game.Item{},
	}

	// Calculate derived stats
	character.MaxHP = cg.calculateMaxHP(character)
	character.HP = character.MaxHP
	character.ArmorClass = cg.calculateArmorClass(character)
	character.THAC0 = cg.calculateTHAC0(character)
	character.MaxActionPoints = cg.calculateActionPoints(character)
	character.ActionPoints = character.MaxActionPoints

	return character, nil
}

// Helper functions for character generation
func (cg *NPCGenerator) generateAlignment(rng *rand.Rand) string {
	alignments := []string{
		"Lawful Good", "Neutral Good", "Chaotic Good",
		"Lawful Neutral", "True Neutral", "Chaotic Neutral",
		"Lawful Evil", "Neutral Evil", "Chaotic Evil",
	}
	return alignments[rng.Intn(len(alignments))]
}

func (cg *NPCGenerator) generateTemperament() string {
	temperaments := []string{
		"sanguine", "choleric", "melancholic", "phlegmatic",
		"optimistic", "pessimistic", "stoic", "passionate",
		"cautious", "bold", "gentle", "fierce",
	}
	return temperaments[cg.rng.Intn(len(temperaments))]
}

func (cg *NPCGenerator) generateValues(params CharacterParams) []string {
	allValues := []string{
		"honor", "wealth", "power", "knowledge", "family",
		"freedom", "justice", "beauty", "tradition", "progress",
		"loyalty", "independence", "compassion", "strength", "wisdom",
	}

	// Select 2-4 values based on personality depth
	numValues := params.PersonalityDepth + cg.rng.Intn(2)
	if numValues > len(allValues) {
		numValues = len(allValues)
	}

	values := make([]string, 0, numValues)
	used := make(map[int]bool)

	for len(values) < numValues {
		idx := cg.rng.Intn(len(allValues))
		if !used[idx] {
			values = append(values, allValues[idx])
			used[idx] = true
		}
	}

	return values
}

func (cg *NPCGenerator) generateFears(params CharacterParams) []string {
	allFears := []string{
		"death", "failure", "betrayal", "abandonment", "powerlessness",
		"poverty", "ignorance", "chaos", "authority", "magic",
		"undead", "heights", "water", "fire", "darkness",
	}

	// Select 1-3 fears
	numFears := cg.rng.Intn(3) + 1
	fears := make([]string, 0, numFears)
	used := make(map[int]bool)

	for len(fears) < numFears {
		idx := cg.rng.Intn(len(allFears))
		if !used[idx] {
			fears = append(fears, allFears[idx])
			used[idx] = true
		}
	}

	return fears
}

func (cg *NPCGenerator) generateSpeechPattern(params CharacterParams) SpeechPattern {
	formalities := []string{"formal", "casual", "crude", "archaic", "pompous"}
	vocabularies := []string{"simple", "moderate", "complex", "technical", "poetic"}
	accents := []string{"none", "regional", "foreign", "aristocratic", "rural"}

	pattern := SpeechPattern{
		Formality:  formalities[cg.rng.Intn(len(formalities))],
		Vocabulary: vocabularies[cg.rng.Intn(len(vocabularies))],
		Accent:     accents[cg.rng.Intn(len(accents))],
		Mannerisms: cg.generateSpeechMannerisms(),
	}

	// Sometimes add a catchphrase
	if cg.rng.Float64() < 0.3 {
		pattern.Catchphrase = cg.generateCatchphrase(params)
	}

	return pattern
}

func (cg *NPCGenerator) generateSpeechMannerisms() []string {
	allMannerisms := []string{
		"repeats key words", "speaks quickly", "speaks slowly",
		"uses elaborate gestures", "avoids eye contact", "speaks loudly",
		"whispers often", "clears throat frequently", "pauses dramatically",
		"uses archaic terms", "mixes languages", "speaks in rhyme occasionally",
	}

	numMannerisms := cg.rng.Intn(3) + 1 // 1-3 mannerisms
	mannerisms := make([]string, 0, numMannerisms)
	used := make(map[int]bool)

	for len(mannerisms) < numMannerisms {
		idx := cg.rng.Intn(len(allMannerisms))
		if !used[idx] {
			mannerisms = append(mannerisms, allMannerisms[idx])
			used[idx] = true
		}
	}

	return mannerisms
}

func (cg *NPCGenerator) generateCatchphrase(params CharacterParams) string {
	catchphrases := []string{
		"By my honor!", "Mark my words!", "As sure as sunrise!",
		"Trust me on this!", "You can count on it!", "Without a doubt!",
		"I swear by the gods!", "As I live and breathe!", "Upon my soul!",
	}
	return catchphrases[cg.rng.Intn(len(catchphrases))]
}

func (cg *NPCGenerator) generatePersonalityTraits(count int, params CharacterParams) ([]PersonalityTrait, error) {
	allTraits := []string{
		"brave", "cowardly", "honest", "deceitful", "generous", "greedy",
		"patient", "impatient", "wise", "foolish", "kind", "cruel",
		"ambitious", "lazy", "loyal", "treacherous", "humble", "arrogant",
		"creative", "mundane", "curious", "incurious", "optimistic", "pessimistic",
	}

	traits := make([]PersonalityTrait, 0, count)
	used := make(map[int]bool)

	for len(traits) < count && len(traits) < len(allTraits) {
		idx := cg.rng.Intn(len(allTraits))
		if !used[idx] {
			trait := PersonalityTrait{
				Name:        allTraits[idx],
				Intensity:   cg.rng.Float64()*0.7 + 0.3, // 0.3-1.0
				Description: fmt.Sprintf("Character displays %s behavior", allTraits[idx]),
			}
			traits = append(traits, trait)
			used[idx] = true
		}
	}

	return traits, nil
}

func (cg *NPCGenerator) generateMotivations(count int, params CharacterParams) ([]Motivation, error) {
	motivationTypes := []string{
		"power", "wealth", "knowledge", "love", "revenge", "redemption",
		"survival", "family", "honor", "freedom", "justice", "fame",
	}

	motivations := make([]Motivation, 0, count)
	used := make(map[int]bool)

	for len(motivations) < count && len(motivations) < len(motivationTypes) {
		idx := cg.rng.Intn(len(motivationTypes))
		if !used[idx] {
			mType := motivationTypes[idx]
			motivation := Motivation{
				Type:        mType,
				Target:      cg.generateMotivationTarget(mType),
				Intensity:   cg.rng.Float64()*0.6 + 0.4, // 0.4-1.0
				Description: fmt.Sprintf("Driven by desire for %s", mType),
			}
			motivations = append(motivations, motivation)
			used[idx] = true
		}
	}

	return motivations, nil
}

func (cg *NPCGenerator) generateMotivationTarget(motivationType string) string {
	targets := map[string][]string{
		"power":      {"political control", "magical ability", "influence over others"},
		"wealth":     {"gold and treasure", "valuable items", "profitable business"},
		"knowledge":  {"ancient secrets", "magical lore", "historical truth"},
		"love":       {"romantic partner", "family member", "lost friend"},
		"revenge":    {"past enemy", "corrupt official", "betrayer"},
		"redemption": {"past mistakes", "family honor", "personal guilt"},
		"survival":   {"personal safety", "family protection", "clan preservation"},
		"family":     {"children's future", "family legacy", "ancestral home"},
		"honor":      {"reputation", "code of conduct", "sworn oath"},
		"freedom":    {"personal liberty", "oppressed people", "enslaved kin"},
		"justice":    {"wronged innocent", "corrupt system", "fair treatment"},
		"fame":       {"legendary status", "heroic recognition", "artistic acclaim"},
	}

	if typeTargets, exists := targets[motivationType]; exists {
		return typeTargets[cg.rng.Intn(len(typeTargets))]
	}
	return "unknown goal"
}

// Placeholder implementations for remaining methods
func (cg *NPCGenerator) generateBehavior(characterType CharacterType, params CharacterParams) string {
	behaviors := map[CharacterType][]string{
		CharacterTypeGuard:    {"patrol", "guard_post", "challenge_strangers"},
		CharacterTypeMerchant: {"haggle", "advertise_wares", "count_coins"},
		CharacterTypeNoble:    {"command", "judge", "social_gathering"},
		CharacterTypePeasant:  {"work", "gossip", "simple_tasks"},
		CharacterTypeCrafter:  {"craft", "teach", "perfectionist"},
	}

	if behaviorList, exists := behaviors[characterType]; exists {
		return behaviorList[cg.rng.Intn(len(behaviorList))]
	}
	return "generic_npc"
}

func (cg *NPCGenerator) generateDialog(personality *PersonalityProfile, params CharacterParams) []game.DialogEntry {
	// Generate basic dialog entries based on personality
	// This would be expanded with more sophisticated dialog generation
	return []game.DialogEntry{
		{
			ID:   "greeting",
			Text: cg.generateGreeting(personality),
			Responses: []game.DialogResponse{
				{Text: "Hello", NextDialog: "conversation", Action: ""},
			},
		},
	}
}

func (cg *NPCGenerator) generateGreeting(personality *PersonalityProfile) string {
	greetings := []string{
		"Well met, traveler!", "Good day to you!", "What brings you here?",
		"Greetings, stranger.", "Welcome!", "State your business.",
	}
	return greetings[cg.rng.Intn(len(greetings))]
}

func (cg *NPCGenerator) generateLootTable(characterType CharacterType, params CharacterParams) []game.LootEntry {
	// Generate appropriate loot based on character type
	return []game.LootEntry{} // Placeholder
}

// Additional helper methods for character generation
func (cg *NPCGenerator) generateAttributesByType(charType CharacterType, socialClass SocialClass) CharacterAttributes {
	// Base attributes
	base := CharacterAttributes{
		Strength:     10,
		Dexterity:    10,
		Constitution: 10,
		Intelligence: 10,
		Wisdom:       10,
		Charisma:     10,
	}

	// Modify based on character type
	switch charType {
	case CharacterTypeGuard:
		base.Strength += cg.rng.Intn(4) + 2     // +2 to +5
		base.Constitution += cg.rng.Intn(3) + 1 // +1 to +3
	case CharacterTypeMerchant:
		base.Charisma += cg.rng.Intn(4) + 2     // +2 to +5
		base.Intelligence += cg.rng.Intn(3) + 1 // +1 to +3
	case CharacterTypeMage:
		base.Intelligence += cg.rng.Intn(6) + 3 // +3 to +8
		base.Wisdom += cg.rng.Intn(3) + 1       // +1 to +3
	case CharacterTypeNoble:
		base.Charisma += cg.rng.Intn(4) + 3     // +3 to +6
		base.Intelligence += cg.rng.Intn(3) + 2 // +2 to +4
	}

	// Modify based on social class
	switch socialClass {
	case SocialClassNoble, SocialClassRoyalty:
		base.Charisma += cg.rng.Intn(3) + 1
		base.Intelligence += cg.rng.Intn(2) + 1
	case SocialClassSlave, SocialClassSerf:
		base.Constitution += cg.rng.Intn(2) + 1 // Hardy from hard work
	}

	return base
}

// CharacterAttributes helper struct
type CharacterAttributes struct {
	Strength     int
	Dexterity    int
	Constitution int
	Intelligence int
	Wisdom       int
	Charisma     int
}

// More helper methods
func (cg *NPCGenerator) generateName(background BackgroundType, gender string) string {
	// Simple name generation - could be expanded with more sophisticated systems
	firstNames := []string{"Aiden", "Bella", "Connor", "Diana", "Ethan", "Fiona", "Gareth", "Helen"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Garcia"}

	first := firstNames[cg.rng.Intn(len(firstNames))]
	last := lastNames[cg.rng.Intn(len(lastNames))]

	return fmt.Sprintf("%s %s", first, last)
}

func (cg *NPCGenerator) generateDescription(params CharacterParams) string {
	age := cg.ageRangeToDescription(params.AgeRange)
	background := string(params.BackgroundType)
	socialClass := string(params.SocialClass)

	return fmt.Sprintf("A %s %s from a %s background", age, socialClass, background)
}

func (cg *NPCGenerator) ageRangeToDescription(ageRange AgeRange) string {
	switch ageRange {
	case AgeRangeChild:
		return "young"
	case AgeRangeAdolescent:
		return "teenage"
	case AgeRangeYoungAdult:
		return "young adult"
	case AgeRangeAdult:
		return "middle-aged"
	case AgeRangeMiddleAged:
		return "mature"
	case AgeRangeElderly:
		return "elderly"
	case AgeRangeAncient:
		return "ancient"
	default:
		return "adult"
	}
}

func (cg *NPCGenerator) selectCharacterClass(charType CharacterType) game.CharacterClass {
	switch charType {
	case CharacterTypeGuard:
		return game.ClassFighter
	case CharacterTypeMage:
		return game.ClassMage
	case CharacterTypeCleric:
		return game.ClassCleric
	case CharacterTypeRogue:
		return game.ClassThief
	case CharacterTypeMerchant:
		return game.ClassFighter // Merchants often need to defend themselves
	case CharacterTypeNoble:
		return game.ClassFighter // Nobles often have military training
	case CharacterTypeBard:
		return game.ClassThief // Bards use thief-like skills
	case CharacterTypeCrafter:
		return game.ClassFighter // Crafters need physical strength
	default:
		// For generic characters, use a deterministic selection based on RNG
		classes := []game.CharacterClass{
			game.ClassFighter, game.ClassMage, game.ClassCleric, game.ClassThief,
		}
		return classes[cg.rng.Intn(len(classes))]
	}
}

// Additional methods for group generation and relationship management
func (cg *NPCGenerator) selectCharacterTypeForGroup(groupType NPCGroupType, position, groupSize int) CharacterType {
	switch groupType {
	case NPCGroupGuards:
		if position == 0 {
			return CharacterTypeGuard // Captain is also a guard, not noble
		}
		return CharacterTypeGuard
	case NPCGroupMerchants:
		return CharacterTypeMerchant
	case NPCGroupScholars:
		if position == 0 {
			return CharacterTypeMage // Lead scholar
		}
		return CharacterTypeGeneric
	case NPCGroupCrafters:
		return CharacterTypeCrafter
	case NPCGroupFamily:
		// Mix of character types for family diversity
		if position == 0 {
			return CharacterTypeGeneric // Family head
		}
		return CharacterTypeGeneric
	default:
		return CharacterTypeGeneric
	}
}

func (cg *NPCGenerator) elevatedSocialClass(current SocialClass) SocialClass {
	switch current {
	case SocialClassSlave:
		return SocialClassSerf
	case SocialClassSerf:
		return SocialClassPeasant
	case SocialClassPeasant:
		return SocialClassCrafter
	case SocialClassCrafter:
		return SocialClassMerchant
	case SocialClassMerchant:
		return SocialClassGentry
	case SocialClassGentry:
		return SocialClassNoble
	case SocialClassNoble:
		return SocialClassRoyalty
	default:
		return current
	}
}

func (cg *NPCGenerator) establishGroupRelationships(npcs []*game.NPC, groupType NPCGroupType) {
	// Add logic to establish relationships between group members
	// This could include setting up dialog references, shared motivations, etc.
	// Implementation would depend on the specific relationship system
}

// Remaining calculation methods
func (cg *NPCGenerator) generateLevel(params CharacterParams) int {
	// Base level on difficulty and social class
	baseLevel := params.Difficulty / 4 // 1-5 base level
	if baseLevel < 1 {
		baseLevel = 1
	}

	// Add variation
	return baseLevel + cg.rng.Intn(3) // +0 to +2
}

func (cg *NPCGenerator) generateStartingGold(socialClass SocialClass) int {
	base := map[SocialClass]int{
		SocialClassSlave:    10,
		SocialClassSerf:     25,
		SocialClassPeasant:  50,
		SocialClassCrafter:  100,
		SocialClassMerchant: 200,
		SocialClassGentry:   500,
		SocialClassNoble:    1000,
		SocialClassRoyalty:  2000,
	}

	if amount, exists := base[socialClass]; exists {
		// Add random variation ±50%
		variation := int(float64(amount) * 0.5)
		return amount + cg.rng.Intn(variation*2) - variation
	}
	return 50 // Default
}

func (cg *NPCGenerator) calculateMaxHP(char *game.Character) int {
	// Simple HP calculation based on level and constitution
	baseHP := 4 + cg.getModifier(char.Constitution) // Base 4 HP
	levelHP := char.Level * (2 + cg.getModifier(char.Constitution))
	return baseHP + levelHP
}

func (cg *NPCGenerator) calculateArmorClass(char *game.Character) int {
	// Base AC 10, modified by dexterity
	return 10 - cg.getModifier(char.Dexterity)
}

func (cg *NPCGenerator) calculateTHAC0(char *game.Character) int {
	// Classic D&D THAC0 calculation
	baseTHAC0 := 20 - char.Level
	strMod := cg.getModifier(char.Strength)
	return baseTHAC0 - strMod
}

func (cg *NPCGenerator) calculateActionPoints(char *game.Character) int {
	// Base action points modified by dexterity
	base := 3 + cg.getModifier(char.Dexterity)
	if base < 1 {
		base = 1
	}
	return base
}

func (cg *NPCGenerator) getModifier(attribute int) int {
	// Standard D&D attribute modifier calculation
	return (attribute - 10) / 2
}
