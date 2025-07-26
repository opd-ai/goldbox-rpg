package pcg

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// WorldGenerator creates overworld maps with regions, settlements, and travel networks
// Uses spatial indexing for efficient placement and pathfinding
type WorldGenerator struct {
	version string
	logger  *logrus.Logger
	rng     *rand.Rand
}

// GeneratedWorld represents a complete overworld campaign setting
type GeneratedWorld struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Width       int                    `json:"width"`
	Height      int                    `json:"height"`
	Regions     []*Region              `json:"regions"`
	Settlements []*Settlement          `json:"settlements"`
	TravelPaths []*TravelPath          `json:"travel_paths"`
	Landmarks   []*Landmark            `json:"landmarks"`
	Climate     ClimateType            `json:"climate"`
	Metadata    map[string]interface{} `json:"metadata"`
	Generated   time.Time              `json:"generated"`
}

// Region represents a geographical area with specific characteristics
type Region struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Bounds     Rectangle              `json:"bounds"`
	Biome      BiomeType              `json:"biome"`
	Difficulty int                    `json:"difficulty"`
	Resources  []ResourceType         `json:"resources"`
	Climate    ClimateType            `json:"climate"`
	Population int                    `json:"population"`
	Features   []RegionFeature        `json:"features"`
	Properties map[string]interface{} `json:"properties"`
}

// Settlement represents a town, city, or other inhabited location
type Settlement struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Position    game.Position          `json:"position"`
	Type        SettlementType         `json:"type"`
	Population  int                    `json:"population"`
	Government  GovernmentType         `json:"government"`
	Economy     EconomyType            `json:"economy"`
	Defenses    DefenseLevel           `json:"defenses"`
	Services    []ServiceType          `json:"services"`
	TradeRoutes []string               `json:"trade_routes"`
	Connections []string               `json:"connections"`
	RegionID    string                 `json:"region_id"`
	Properties  map[string]interface{} `json:"properties"`
}

// TravelPath represents roads, rivers, and other travel routes
type TravelPath struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       PathType               `json:"type"`
	Points     []game.Position        `json:"points"`
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Difficulty int                    `json:"difficulty"`
	TravelTime int                    `json:"travel_time"`
	Hazards    []HazardType           `json:"hazards"`
	Properties map[string]interface{} `json:"properties"`
}

// Landmark represents significant geographical or man-made features
type Landmark struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Position    game.Position          `json:"position"`
	Type        LandmarkType           `json:"type"`
	Importance  int                    `json:"importance"`
	Description string                 `json:"description"`
	Properties  map[string]interface{} `json:"properties"`
}

// Enums for world generation types

type ClimateType string

const (
	ClimateTemperate ClimateType = "temperate"
	ClimateArctic    ClimateType = "arctic"
	ClimateTropical  ClimateType = "tropical"
	ClimateArid      ClimateType = "arid"
	ClimateMountain  ClimateType = "mountain"
)

type SettlementType string

const (
	SettlementHamlet    SettlementType = "hamlet"
	SettlementVillage   SettlementType = "village"
	SettlementTown      SettlementType = "town"
	SettlementCity      SettlementType = "city"
	SettlementCapital   SettlementType = "capital"
	SettlementFortress  SettlementType = "fortress"
	SettlementMonastery SettlementType = "monastery"
	SettlementOutpost   SettlementType = "outpost"
)

type GovernmentType string

const (
	GovernmentMonarchy  GovernmentType = "monarchy"
	GovernmentRepublic  GovernmentType = "republic"
	GovernmentTheocracy GovernmentType = "theocracy"
	GovernmentMilitary  GovernmentType = "military"
	GovernmentTribal    GovernmentType = "tribal"
	GovernmentAnarchy   GovernmentType = "anarchy"
)

type EconomyType string

const (
	EconomyAgriculture EconomyType = "agriculture"
	EconomyMining      EconomyType = "mining"
	EconomyTrading     EconomyType = "trading"
	EconomyFishing     EconomyType = "fishing"
	EconomyCrafting    EconomyType = "crafting"
	EconomyMagical     EconomyType = "magical"
)

type DefenseLevel string

const (
	DefenseNone     DefenseLevel = "none"
	DefensePalisade DefenseLevel = "palisade"
	DefenseWalls    DefenseLevel = "walls"
	DefenseFortress DefenseLevel = "fortress"
	DefenseCastle   DefenseLevel = "castle"
	DefenseMagical  DefenseLevel = "magical"
)

type ServiceType string

const (
	ServiceInn        ServiceType = "inn"
	ServiceTavern     ServiceType = "tavern"
	ServiceShop       ServiceType = "shop"
	ServiceBlacksmith ServiceType = "blacksmith"
	ServiceTemple     ServiceType = "temple"
	ServiceMage       ServiceType = "mage"
	ServiceHealer     ServiceType = "healer"
	ServiceStables    ServiceType = "stables"
	ServiceBank       ServiceType = "bank"
	ServiceLibrary    ServiceType = "library"
)

type PathType string

const (
	PathRoad   PathType = "road"
	PathTrail  PathType = "trail"
	PathRiver  PathType = "river"
	PathSea    PathType = "sea"
	PathBridge PathType = "bridge"
	PathTunnel PathType = "tunnel"
)

type HazardType string

const (
	HazardBandits   HazardType = "bandits"
	HazardMonsters  HazardType = "monsters"
	HazardWeather   HazardType = "weather"
	HazardTerrain   HazardType = "terrain"
	HazardMagical   HazardType = "magical"
	HazardPolitical HazardType = "political"
)

type LandmarkType string

const (
	LandmarkMountain LandmarkType = "mountain"
	LandmarkRuins    LandmarkType = "ruins"
	LandmarkForest   LandmarkType = "forest"
	LandmarkLake     LandmarkType = "lake"
	LandmarkDesert   LandmarkType = "desert"
	LandmarkVolcano  LandmarkType = "volcano"
	LandmarkTower    LandmarkType = "tower"
	LandmarkBridge   LandmarkType = "bridge"
	LandmarkShrine   LandmarkType = "shrine"
)

type ResourceType string

const (
	ResourceIron     ResourceType = "iron"
	ResourceGold     ResourceType = "gold"
	ResourceGems     ResourceType = "gems"
	ResourceWood     ResourceType = "wood"
	ResourceStone    ResourceType = "stone"
	ResourceMagicite ResourceType = "magicite"
	ResourceFood     ResourceType = "food"
	ResourceWater    ResourceType = "water"
)

type RegionFeature string

const (
	FeatureRiver       RegionFeature = "river"
	FeatureMountains   RegionFeature = "mountains"
	FeatureForestArea  RegionFeature = "forest"
	FeatureDesertArea  RegionFeature = "desert"
	FeatureSwampArea   RegionFeature = "swamp"
	FeatureMagicalZone RegionFeature = "magical_zone"
	FeatureBattlefield RegionFeature = "battlefield"
)

// WorldParams provides world-specific generation parameters
type WorldParams struct {
	GenerationParams  `yaml:",inline"`
	WorldWidth        int               `yaml:"world_width"`        // Width of the world map
	WorldHeight       int               `yaml:"world_height"`       // Height of the world map
	RegionCount       int               `yaml:"region_count"`       // Number of regions to generate
	SettlementCount   int               `yaml:"settlement_count"`   // Target number of settlements
	LandmarkCount     int               `yaml:"landmark_count"`     // Number of major landmarks
	Climate           ClimateType       `yaml:"climate"`            // Overall world climate
	Connectivity      ConnectivityLevel `yaml:"connectivity"`       // Travel route density
	PopulationDensity float64           `yaml:"population_density"` // Overall population density
	MagicLevel        int               `yaml:"magic_level"`        // Prevalence of magic (1-10)
	DangerLevel       int               `yaml:"danger_level"`       // Overall danger level
}

// NewWorldGenerator creates a new world generator
func NewWorldGenerator(logger *logrus.Logger) *WorldGenerator {
	if logger == nil {
		logger = logrus.New()
	}

	return &WorldGenerator{
		version: "1.0.0",
		logger:  logger,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Generate creates a complete overworld campaign setting
// Implements Generator interface for PCG system integration
func (wg *WorldGenerator) Generate(ctx context.Context, params GenerationParams) (interface{}, error) {
	worldParams, ok := params.Constraints["world_params"].(WorldParams)
	if !ok {
		return nil, fmt.Errorf("invalid parameters for world generation: expected world_params in constraints")
	}

	// Validate parameters before generation
	if err := wg.Validate(params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Initialize RNG with provided seed for deterministic generation
	wg.rng = rand.New(rand.NewSource(params.Seed))

	wg.logger.WithFields(logrus.Fields{
		"world_width":  worldParams.WorldWidth,
		"world_height": worldParams.WorldHeight,
		"regions":      worldParams.RegionCount,
		"settlements":  worldParams.SettlementCount,
	}).Info("generating overworld campaign setting")

	world, err := wg.generateWorld(ctx, params, worldParams)
	if err != nil {
		return nil, fmt.Errorf("world generation failed: %w", err)
	}

	wg.logger.WithField("world_id", world.ID).Info("world generation completed")
	return world, nil
}

// generateWorld creates the complete world structure
func (wg *WorldGenerator) generateWorld(ctx context.Context, params GenerationParams, worldParams WorldParams) (*GeneratedWorld, error) {
	world := &GeneratedWorld{
		ID:          fmt.Sprintf("world_%d", wg.rng.Int63()),
		Name:        wg.generateWorldName(worldParams.Climate),
		Width:       worldParams.WorldWidth,
		Height:      worldParams.WorldHeight,
		Regions:     make([]*Region, 0),
		Settlements: make([]*Settlement, 0),
		TravelPaths: make([]*TravelPath, 0),
		Landmarks:   make([]*Landmark, 0),
		Climate:     worldParams.Climate,
		Metadata:    make(map[string]interface{}),
		Generated:   time.Now(),
	}

	// Step 1: Generate regions using spatial partitioning
	if err := wg.generateRegions(world, worldParams); err != nil {
		return nil, fmt.Errorf("region generation failed: %w", err)
	}

	// Step 2: Place landmarks strategically
	if err := wg.generateLandmarks(world, worldParams); err != nil {
		return nil, fmt.Errorf("landmark generation failed: %w", err)
	}

	// Step 3: Generate settlements with spatial distribution
	if err := wg.generateSettlements(world, worldParams); err != nil {
		return nil, fmt.Errorf("settlement generation failed: %w", err)
	}

	// Step 4: Create travel network connecting settlements
	if err := wg.generateTravelNetwork(world, worldParams); err != nil {
		return nil, fmt.Errorf("travel network generation failed: %w", err)
	}

	// Step 5: Add metadata for debugging and validation
	world.Metadata["total_population"] = wg.calculateTotalPopulation(world)
	world.Metadata["trade_route_count"] = len(world.TravelPaths)
	world.Metadata["generation_seed"] = params.Seed

	return world, nil
}

// generateRegions creates geographical regions using spatial partitioning
func (wg *WorldGenerator) generateRegions(world *GeneratedWorld, params WorldParams) error {
	// Use simple grid-based partitioning for initial implementation
	regionWidth := world.Width / int(math.Sqrt(float64(params.RegionCount)))
	regionHeight := world.Height / int(math.Sqrt(float64(params.RegionCount)))

	regionID := 0
	for y := 0; y < world.Height; y += regionHeight {
		for x := 0; x < world.Width; x += regionWidth {
			if regionID >= params.RegionCount {
				break
			}

			// Adjust last region to cover remaining area
			width := regionWidth
			height := regionHeight
			if x+width > world.Width {
				width = world.Width - x
			}
			if y+height > world.Height {
				height = world.Height - y
			}

			region := &Region{
				ID:         fmt.Sprintf("region_%d", regionID),
				Name:       wg.generateRegionName(regionID),
				Bounds:     Rectangle{X: x, Y: y, Width: width, Height: height},
				Biome:      wg.chooseBiome(params.Climate),
				Difficulty: 1 + wg.rng.Intn(params.DangerLevel),
				Resources:  wg.generateResources(),
				Climate:    params.Climate,
				Population: wg.rng.Intn(10000) + 1000,
				Features:   wg.generateRegionFeatures(),
				Properties: make(map[string]interface{}),
			}

			world.Regions = append(world.Regions, region)
			regionID++
		}
		if regionID >= params.RegionCount {
			break
		}
	}

	return nil
}

// generateLandmarks places significant features across the world
func (wg *WorldGenerator) generateLandmarks(world *GeneratedWorld, params WorldParams) error {
	for i := 0; i < params.LandmarkCount; i++ {
		landmark := &Landmark{
			ID:          fmt.Sprintf("landmark_%d", i),
			Name:        wg.generateLandmarkName(),
			Position:    game.Position{X: wg.rng.Intn(world.Width), Y: wg.rng.Intn(world.Height)},
			Type:        wg.chooseLandmarkType(),
			Importance:  1 + wg.rng.Intn(5),
			Description: wg.generateLandmarkDescription(),
			Properties:  make(map[string]interface{}),
		}

		world.Landmarks = append(world.Landmarks, landmark)
	}

	return nil
}

// generateSettlements places settlements with proper spacing and context
func (wg *WorldGenerator) generateSettlements(world *GeneratedWorld, params WorldParams) error {
	minDistance := 10 // Minimum distance between settlements

	for i := 0; i < params.SettlementCount; i++ {
		var position game.Position
		attempts := 0
		maxAttempts := 100

		// Find position with proper spacing
		for attempts < maxAttempts {
			position = game.Position{
				X: wg.rng.Intn(world.Width),
				Y: wg.rng.Intn(world.Height),
			}

			// Check distance to existing settlements
			validPosition := true
			for _, existing := range world.Settlements {
				distance := wg.calculateDistance(position, existing.Position)
				if distance < minDistance {
					validPosition = false
					break
				}
			}

			if validPosition {
				break
			}
			attempts++
		}

		if attempts >= maxAttempts {
			wg.logger.Warn("could not find valid position for settlement after max attempts")
			continue
		}

		// Find which region contains this settlement
		regionID := wg.findRegionForPosition(world, position)

		settlement := &Settlement{
			ID:          fmt.Sprintf("settlement_%d", i),
			Name:        wg.generateSettlementName(),
			Position:    position,
			Type:        wg.chooseSettlementType(params.PopulationDensity),
			Population:  wg.calculateSettlementPopulation(params.PopulationDensity),
			Government:  wg.chooseGovernmentType(),
			Economy:     wg.chooseEconomyType(),
			Defenses:    wg.chooseDefenseLevel(params.DangerLevel),
			Services:    wg.generateServices(),
			TradeRoutes: make([]string, 0),
			Connections: make([]string, 0),
			RegionID:    regionID,
			Properties:  make(map[string]interface{}),
		}

		world.Settlements = append(world.Settlements, settlement)
	}

	return nil
}

// generateTravelNetwork creates roads and paths connecting settlements
func (wg *WorldGenerator) generateTravelNetwork(world *GeneratedWorld, params WorldParams) error {
	// Connect each settlement to its nearest neighbors
	for _, settlement := range world.Settlements {
		nearestNeighbors := wg.findNearestSettlements(world, settlement, 3)

		for _, neighbor := range nearestNeighbors {
			// Check if connection already exists
			connectionExists := false
			for _, existing := range world.TravelPaths {
				if (existing.From == settlement.ID && existing.To == neighbor.ID) ||
					(existing.From == neighbor.ID && existing.To == settlement.ID) {
					connectionExists = true
					break
				}
			}

			if !connectionExists {
				path := wg.createTravelPath(settlement, neighbor)
				world.TravelPaths = append(world.TravelPaths, path)

				// Update settlement connections
				settlement.Connections = append(settlement.Connections, neighbor.ID)
				neighbor.Connections = append(neighbor.Connections, settlement.ID)
			}
		}
	}

	return nil
}

// Helper methods

func (wg *WorldGenerator) generateWorldName(climate ClimateType) string {
	prefixes := map[ClimateType][]string{
		ClimateTemperate: {"Green", "Fair", "Golden", "Pleasant"},
		ClimateArctic:    {"Frozen", "White", "Frost", "Ice"},
		ClimateTropical:  {"Verdant", "Lush", "Steam", "Jungle"},
		ClimateArid:      {"Sand", "Burning", "Dry", "Sun"},
		ClimateMountain:  {"High", "Stone", "Peak", "Ridge"},
	}

	suffixes := []string{"Realm", "Lands", "Domain", "Kingdom", "Territory", "Expanse"}

	prefix := prefixes[climate][wg.rng.Intn(len(prefixes[climate]))]
	suffix := suffixes[wg.rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s", prefix, suffix)
}

func (wg *WorldGenerator) generateRegionName(id int) string {
	prefixes := []string{"North", "South", "East", "West", "Central", "Upper", "Lower", "Inner", "Outer"}
	suffixes := []string{"Reach", "March", "Vale", "Moor", "Waste", "Wood", "Hold", "Shire", "Gate"}

	prefix := prefixes[wg.rng.Intn(len(prefixes))]
	suffix := suffixes[wg.rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s", prefix, suffix)
}

func (wg *WorldGenerator) generateSettlementName() string {
	prefixes := []string{"Stone", "Iron", "Gold", "Silver", "Wood", "River", "Hill", "Vale", "Red", "White"}
	suffixes := []string{"ford", "burg", "ton", "ham", "stead", "haven", "bridge", "gate", "port", "mill"}

	prefix := prefixes[wg.rng.Intn(len(prefixes))]
	suffix := suffixes[wg.rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s%s", prefix, suffix)
}

func (wg *WorldGenerator) generateLandmarkName() string {
	adjectives := []string{"Ancient", "Forgotten", "Lost", "Sacred", "Cursed", "Hidden", "Great", "Old"}
	nouns := []string{"Tower", "Ruins", "Mountain", "Forest", "Lake", "Bridge", "Shrine", "Monument"}

	adj := adjectives[wg.rng.Intn(len(adjectives))]
	noun := nouns[wg.rng.Intn(len(nouns))]

	return fmt.Sprintf("%s %s", adj, noun)
}

func (wg *WorldGenerator) generateLandmarkDescription() string {
	descriptions := []string{
		"A mysterious structure of unknown origin",
		"Ancient ruins from a forgotten civilization",
		"A place of natural beauty and wonder",
		"Remnants of a great battle fought long ago",
		"A sacred site revered by local inhabitants",
		"A geographical feature visible from great distances",
	}

	return descriptions[wg.rng.Intn(len(descriptions))]
}

func (wg *WorldGenerator) chooseBiome(climate ClimateType) BiomeType {
	biomeWeights := map[ClimateType]map[BiomeType]int{
		ClimateTemperate: {BiomeForest: 40, BiomeMountain: 20, BiomeCoastal: 20, BiomeUrban: 20},
		ClimateArctic:    {BiomeMountain: 60, BiomeForest: 20, BiomeWasteland: 20},
		ClimateTropical:  {BiomeForest: 50, BiomeSwamp: 30, BiomeCoastal: 20},
		ClimateArid:      {BiomeDesert: 70, BiomeMountain: 20, BiomeWasteland: 10},
		ClimateMountain:  {BiomeMountain: 80, BiomeForest: 20},
	}

	return wg.weightedRandomBiome(biomeWeights[climate])
}

func (wg *WorldGenerator) weightedRandomBiome(weights map[BiomeType]int) BiomeType {
	totalWeight := 0
	for _, weight := range weights {
		totalWeight += weight
	}

	randomValue := wg.rng.Intn(totalWeight)
	currentWeight := 0

	for biome, weight := range weights {
		currentWeight += weight
		if randomValue < currentWeight {
			return biome
		}
	}

	return BiomeForest // fallback
}

func (wg *WorldGenerator) generateResources() []ResourceType {
	allResources := []ResourceType{ResourceIron, ResourceGold, ResourceGems, ResourceWood, ResourceStone, ResourceMagicite, ResourceFood, ResourceWater}
	resourceCount := 1 + wg.rng.Intn(4)

	resources := make([]ResourceType, 0, resourceCount)
	for i := 0; i < resourceCount; i++ {
		resource := allResources[wg.rng.Intn(len(allResources))]
		// Avoid duplicates
		found := false
		for _, existing := range resources {
			if existing == resource {
				found = true
				break
			}
		}
		if !found {
			resources = append(resources, resource)
		}
	}

	return resources
}

func (wg *WorldGenerator) generateRegionFeatures() []RegionFeature {
	allFeatures := []RegionFeature{FeatureRiver, FeatureMountains, FeatureForestArea, FeatureDesertArea, FeatureSwampArea, FeatureMagicalZone, FeatureBattlefield}
	featureCount := wg.rng.Intn(3) + 1

	features := make([]RegionFeature, 0, featureCount)
	for i := 0; i < featureCount; i++ {
		feature := allFeatures[wg.rng.Intn(len(allFeatures))]
		features = append(features, feature)
	}

	return features
}

func (wg *WorldGenerator) chooseLandmarkType() LandmarkType {
	types := []LandmarkType{LandmarkMountain, LandmarkRuins, LandmarkForest, LandmarkLake, LandmarkDesert, LandmarkVolcano, LandmarkTower, LandmarkBridge, LandmarkShrine}
	return types[wg.rng.Intn(len(types))]
}

func (wg *WorldGenerator) chooseSettlementType(populationDensity float64) SettlementType {
	// Higher population density = more likely to have larger settlements
	roll := wg.rng.Float64() * populationDensity

	if roll > 0.8 {
		return SettlementCity
	} else if roll > 0.6 {
		return SettlementTown
	} else if roll > 0.4 {
		return SettlementVillage
	} else {
		return SettlementHamlet
	}
}

func (wg *WorldGenerator) calculateSettlementPopulation(populationDensity float64) int {
	basePopulation := int(100 * populationDensity)
	variation := wg.rng.Intn(basePopulation)
	return basePopulation + variation
}

func (wg *WorldGenerator) chooseGovernmentType() GovernmentType {
	types := []GovernmentType{GovernmentMonarchy, GovernmentRepublic, GovernmentTheocracy, GovernmentMilitary, GovernmentTribal, GovernmentAnarchy}
	return types[wg.rng.Intn(len(types))]
}

func (wg *WorldGenerator) chooseEconomyType() EconomyType {
	types := []EconomyType{EconomyAgriculture, EconomyMining, EconomyTrading, EconomyFishing, EconomyCrafting, EconomyMagical}
	return types[wg.rng.Intn(len(types))]
}

func (wg *WorldGenerator) chooseDefenseLevel(dangerLevel int) DefenseLevel {
	// Higher danger level = better defenses
	roll := wg.rng.Intn(10) + dangerLevel

	if roll > 15 {
		return DefenseCastle
	} else if roll > 12 {
		return DefenseFortress
	} else if roll > 8 {
		return DefenseWalls
	} else if roll > 5 {
		return DefensePalisade
	} else {
		return DefenseNone
	}
}

func (wg *WorldGenerator) generateServices() []ServiceType {
	allServices := []ServiceType{ServiceInn, ServiceTavern, ServiceShop, ServiceBlacksmith, ServiceTemple, ServiceMage, ServiceHealer, ServiceStables, ServiceBank, ServiceLibrary}
	serviceCount := 2 + wg.rng.Intn(4)

	services := make([]ServiceType, 0, serviceCount)
	for i := 0; i < serviceCount; i++ {
		service := allServices[wg.rng.Intn(len(allServices))]
		services = append(services, service)
	}

	return services
}

func (wg *WorldGenerator) calculateDistance(pos1, pos2 game.Position) int {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return int(math.Sqrt(float64(dx*dx + dy*dy)))
}

func (wg *WorldGenerator) findRegionForPosition(world *GeneratedWorld, position game.Position) string {
	for _, region := range world.Regions {
		if region.Bounds.Contains(position.X, position.Y) {
			return region.ID
		}
	}
	return "unknown"
}

func (wg *WorldGenerator) findNearestSettlements(world *GeneratedWorld, settlement *Settlement, count int) []*Settlement {
	type settlementDistance struct {
		settlement *Settlement
		distance   int
	}

	distances := make([]settlementDistance, 0, len(world.Settlements))

	for _, other := range world.Settlements {
		if other.ID != settlement.ID {
			distance := wg.calculateDistance(settlement.Position, other.Position)
			distances = append(distances, settlementDistance{other, distance})
		}
	}

	// Sort by distance
	for i := 0; i < len(distances)-1; i++ {
		for j := i + 1; j < len(distances); j++ {
			if distances[i].distance > distances[j].distance {
				distances[i], distances[j] = distances[j], distances[i]
			}
		}
	}

	// Return up to 'count' nearest settlements
	result := make([]*Settlement, 0, count)
	for i := 0; i < count && i < len(distances); i++ {
		result = append(result, distances[i].settlement)
	}

	return result
}

func (wg *WorldGenerator) createTravelPath(from, to *Settlement) *TravelPath {
	distance := wg.calculateDistance(from.Position, to.Position)

	return &TravelPath{
		ID:         fmt.Sprintf("path_%s_%s", from.ID, to.ID),
		Name:       fmt.Sprintf("Road from %s to %s", from.Name, to.Name),
		Type:       PathRoad,
		Points:     []game.Position{from.Position, to.Position}, // Simplified straight line
		From:       from.ID,
		To:         to.ID,
		Difficulty: 1 + wg.rng.Intn(3),
		TravelTime: distance / 10, // Simplified travel time calculation
		Hazards:    wg.generatePathHazards(),
		Properties: make(map[string]interface{}),
	}
}

func (wg *WorldGenerator) generatePathHazards() []HazardType {
	allHazards := []HazardType{HazardBandits, HazardMonsters, HazardWeather, HazardTerrain, HazardMagical, HazardPolitical}
	hazardCount := wg.rng.Intn(3) // 0-2 hazards per path

	hazards := make([]HazardType, 0, hazardCount)
	for i := 0; i < hazardCount; i++ {
		hazard := allHazards[wg.rng.Intn(len(allHazards))]
		hazards = append(hazards, hazard)
	}

	return hazards
}

func (wg *WorldGenerator) calculateTotalPopulation(world *GeneratedWorld) int {
	total := 0
	for _, settlement := range world.Settlements {
		total += settlement.Population
	}
	return total
}

// Interface compliance methods

// GetType returns the content type this generator produces
func (wg *WorldGenerator) GetType() ContentType {
	return ContentTypeTerrain // Using existing content type for world generation
}

// GetVersion returns the generator version for compatibility
func (wg *WorldGenerator) GetVersion() string {
	return wg.version
}

// Validate checks if the provided parameters are valid
func (wg *WorldGenerator) Validate(params GenerationParams) error {
	worldParams, ok := params.Constraints["world_params"].(WorldParams)
	if !ok {
		return fmt.Errorf("invalid parameters: expected world_params in constraints")
	}

	if worldParams.WorldWidth < 50 || worldParams.WorldWidth > 1000 {
		return fmt.Errorf("world width must be between 50 and 1000, got %d", worldParams.WorldWidth)
	}

	if worldParams.WorldHeight < 50 || worldParams.WorldHeight > 1000 {
		return fmt.Errorf("world height must be between 50 and 1000, got %d", worldParams.WorldHeight)
	}

	if worldParams.RegionCount < 1 || worldParams.RegionCount > 100 {
		return fmt.Errorf("region count must be between 1 and 100, got %d", worldParams.RegionCount)
	}

	if worldParams.SettlementCount < 1 || worldParams.SettlementCount > 500 {
		return fmt.Errorf("settlement count must be between 1 and 500, got %d", worldParams.SettlementCount)
	}

	if worldParams.PopulationDensity < 0.1 || worldParams.PopulationDensity > 10.0 {
		return fmt.Errorf("population density must be between 0.1 and 10.0, got %f", worldParams.PopulationDensity)
	}

	if worldParams.MagicLevel < 1 || worldParams.MagicLevel > 10 {
		return fmt.Errorf("magic level must be between 1 and 10, got %d", worldParams.MagicLevel)
	}

	if worldParams.DangerLevel < 1 || worldParams.DangerLevel > 20 {
		return fmt.Errorf("danger level must be between 1 and 20, got %d", worldParams.DangerLevel)
	}

	return nil
}
