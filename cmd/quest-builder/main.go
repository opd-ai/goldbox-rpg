// Package main provides the quest-builder CLI tool for creating quest YAML files.
// This tool guides users through the process of defining quests with objectives
// and rewards, then outputs valid YAML configuration files.
package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"goldbox-rpg/pkg/game"

	"gopkg.in/yaml.v3"
	"nhooyr.io/websocket"
)

//go:embed preview.html
var previewHTML embed.FS

// Config holds the command-line configuration for the quest builder.
type Config struct {
	// OutputFile specifies the path to write the generated YAML.
	OutputFile string
	// Interactive enables interactive mode for guided quest creation.
	Interactive bool
	// Template specifies a template type for quick quest scaffolding.
	Template string
	// PreviewPort specifies the port for WebSocket-based live preview (0 = disabled).
	PreviewPort int
}

// previewServer manages WebSocket connections for live quest preview.
type previewServer struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
	port    int
}

// newPreviewServer creates a new preview server on the specified port.
func newPreviewServer(port int) *previewServer {
	return &previewServer{
		clients: make(map[*websocket.Conn]struct{}),
		port:    port,
	}
}

// addClient registers a new WebSocket client.
func (ps *previewServer) addClient(conn *websocket.Conn) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.clients[conn] = struct{}{}
}

// removeClient unregisters a WebSocket client.
func (ps *previewServer) removeClient(conn *websocket.Conn) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.clients, conn)
}

// broadcastQuest sends the current quest state to all connected clients.
func (ps *previewServer) broadcastQuest(quest game.Quest) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	data, err := json.Marshal(quest)
	if err != nil {
		return
	}

	for conn := range ps.clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		conn.Write(ctx, websocket.MessageText, data)
		cancel()
	}
}

// start begins serving the preview HTTP server.
func (ps *previewServer) start() error {
	mux := http.NewServeMux()

	// Serve the embedded preview HTML
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := previewHTML.ReadFile("preview.html")
		if err != nil {
			http.Error(w, "Preview page not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})

	// WebSocket endpoint for live updates
	mux.HandleFunc("/ws", ps.handleWebSocket)

	addr := fmt.Sprintf(":%d", ps.port)
	fmt.Printf("Preview server starting at http://localhost%s\n", addr)
	fmt.Println("Open this URL in a browser to see live quest updates")

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Preview server error: %v\n", err)
		}
	}()

	return nil
}

// handleWebSocket handles WebSocket connection upgrades and message handling.
func (ps *previewServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "connection closed")

	ps.addClient(conn)
	defer ps.removeClient(conn)

	// Keep connection open until client disconnects
	for {
		_, _, err := conn.Read(context.Background())
		if err != nil {
			return
		}
	}
}

// parseFlags parses command-line flags and returns the configuration.
func parseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.OutputFile, "output", "", "output file path for generated YAML (default: stdout)")
	flag.StringVar(&cfg.OutputFile, "o", "", "output file path (shorthand)")
	flag.BoolVar(&cfg.Interactive, "interactive", false, "enable interactive mode for guided quest creation")
	flag.BoolVar(&cfg.Interactive, "i", false, "interactive mode (shorthand)")
	flag.StringVar(&cfg.Template, "template", "", "quest template type: fetch, kill, escort, explore, puzzle")
	flag.StringVar(&cfg.Template, "t", "", "template type (shorthand)")
	flag.IntVar(&cfg.PreviewPort, "preview", 0, "enable live preview server on specified port (e.g., 9001)")
	flag.IntVar(&cfg.PreviewPort, "p", 0, "preview port (shorthand)")
	flag.Usage = printUsage
	flag.Parse()
	return cfg
}

// printUsage prints the usage information for the quest builder.
func printUsage() {
	fmt.Fprintf(os.Stderr, `Quest Builder - Create quest YAML files for GoldBox RPG

Usage:
  quest-builder [options]

Options:
  -o, --output FILE      Write output to FILE instead of stdout
  -i, --interactive      Enable interactive mode for guided quest creation
  -t, --template TYPE    Use a quest template (fetch, kill, escort, explore, puzzle)
  -p, --preview PORT     Enable live preview server on PORT (e.g., 9001)
  -h, --help             Show this help message

Examples:
  # Interactive mode - guided quest creation
  quest-builder -i -o my_quest.yaml

  # Generate from template
  quest-builder -t fetch -o fetch_quest.yaml

  # Combine template with interactive refinement
  quest-builder -t kill -i -o kill_quest.yaml

  # Interactive with live browser preview
  quest-builder -i -o my_quest.yaml --preview 9001
  # Then open http://localhost:9001 in your browser

Templates:
  fetch   - Retrieve an item from a location
  kill    - Defeat a number of enemies
  escort  - Protect an NPC to a destination
  explore - Discover new areas
  puzzle  - Solve a riddle or puzzle

`)
}

// main is the entry point for the quest builder application.
func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the quest builder with the given configuration.
func run(cfg *Config) error {
	var quest game.Quest

	// Start with a template if specified
	if cfg.Template != "" {
		var err error
		quest, err = createFromTemplate(cfg.Template)
		if err != nil {
			return fmt.Errorf("invalid template: %w", err)
		}
	}

	// Start preview server if requested
	var preview *previewServer
	if cfg.PreviewPort > 0 {
		preview = newPreviewServer(cfg.PreviewPort)
		if err := preview.start(); err != nil {
			return fmt.Errorf("failed to start preview server: %w", err)
		}
		// Give server a moment to start
		time.Sleep(100 * time.Millisecond)
		// Send initial quest state
		preview.broadcastQuest(quest)
	}

	// Interactive mode for guided creation or refinement
	if cfg.Interactive {
		var err error
		quest, err = interactiveCreate(quest, preview)
		if err != nil {
			return fmt.Errorf("interactive creation failed: %w", err)
		}
	} else if cfg.Template == "" {
		// No template and not interactive - show help
		printUsage()
		return nil
	}

	// Validate the quest
	if err := validateQuest(quest); err != nil {
		return fmt.Errorf("quest validation failed: %w", err)
	}

	// Output the quest YAML
	return outputQuest(quest, cfg.OutputFile)
}

// createFromTemplate creates a quest based on a template type.
func createFromTemplate(templateType string) (game.Quest, error) {
	templates := map[string]game.Quest{
		"fetch": {
			ID:          "quest_fetch_001",
			Title:       "The Lost Artifact",
			Description: "Retrieve the ancient artifact from the dungeon depths.",
			Status:      game.QuestNotStarted,
			Objectives: []game.QuestObjective{
				{Description: "Find the artifact location", Progress: 0, Required: 1, Completed: false},
				{Description: "Retrieve the artifact", Progress: 0, Required: 1, Completed: false},
				{Description: "Return the artifact", Progress: 0, Required: 1, Completed: false},
			},
			Rewards: []game.QuestReward{
				{Type: "gold", Value: 100},
				{Type: "exp", Value: 50},
			},
		},
		"kill": {
			ID:          "quest_kill_001",
			Title:       "Goblin Threat",
			Description: "Eliminate the goblins threatening the village.",
			Status:      game.QuestNotStarted,
			Objectives: []game.QuestObjective{
				{Description: "Defeat goblins", Progress: 0, Required: 10, Completed: false},
				{Description: "Defeat the goblin leader", Progress: 0, Required: 1, Completed: false},
			},
			Rewards: []game.QuestReward{
				{Type: "gold", Value: 200},
				{Type: "exp", Value: 100},
			},
		},
		"escort": {
			ID:          "quest_escort_001",
			Title:       "Safe Passage",
			Description: "Escort the merchant safely to the next town.",
			Status:      game.QuestNotStarted,
			Objectives: []game.QuestObjective{
				{Description: "Meet the merchant", Progress: 0, Required: 1, Completed: false},
				{Description: "Protect the merchant on the journey", Progress: 0, Required: 1, Completed: false},
				{Description: "Arrive at destination", Progress: 0, Required: 1, Completed: false},
			},
			Rewards: []game.QuestReward{
				{Type: "gold", Value: 150},
				{Type: "exp", Value: 75},
				{Type: "item", Value: 1, ItemID: "item_potion_health"},
			},
		},
		"explore": {
			ID:          "quest_explore_001",
			Title:       "Uncharted Territory",
			Description: "Map the unexplored regions of the wilderness.",
			Status:      game.QuestNotStarted,
			Objectives: []game.QuestObjective{
				{Description: "Discover new locations", Progress: 0, Required: 5, Completed: false},
				{Description: "Find hidden landmarks", Progress: 0, Required: 3, Completed: false},
			},
			Rewards: []game.QuestReward{
				{Type: "exp", Value: 150},
				{Type: "item", Value: 1, ItemID: "item_map_enchanted"},
			},
		},
		"puzzle": {
			ID:          "quest_puzzle_001",
			Title:       "The Ancient Riddle",
			Description: "Solve the riddle to unlock the sealed door.",
			Status:      game.QuestNotStarted,
			Objectives: []game.QuestObjective{
				{Description: "Find the riddle clues", Progress: 0, Required: 4, Completed: false},
				{Description: "Solve the riddle", Progress: 0, Required: 1, Completed: false},
			},
			Rewards: []game.QuestReward{
				{Type: "exp", Value: 200},
				{Type: "item", Value: 1, ItemID: "item_key_ancient"},
			},
		},
	}

	quest, ok := templates[strings.ToLower(templateType)]
	if !ok {
		return game.Quest{}, fmt.Errorf("unknown template: %s (valid: fetch, kill, escort, explore, puzzle)", templateType)
	}
	return quest, nil
}

// interactiveCreate guides the user through quest creation interactively.
func interactiveCreate(base game.Quest, preview *previewServer) (game.Quest, error) {
	reader := bufio.NewReader(os.Stdin)
	quest := base

	fmt.Println("\n=== Quest Builder - Interactive Mode ===")
	if preview != nil {
		fmt.Printf("Live preview available at http://localhost:%d\n", preview.port)
	}
	fmt.Println()

	// Quest ID
	quest.ID = promptString(reader, "Quest ID", quest.ID, "quest_")
	if preview != nil {
		preview.broadcastQuest(quest)
	}

	// Quest Title
	quest.Title = promptString(reader, "Quest Title", quest.Title, "")
	if preview != nil {
		preview.broadcastQuest(quest)
	}

	// Quest Description
	quest.Description = promptString(reader, "Quest Description", quest.Description, "")
	if preview != nil {
		preview.broadcastQuest(quest)
	}

	// Objectives
	fmt.Println("\n--- Quest Objectives ---")
	quest.Objectives = promptObjectivesWithPreview(reader, quest.Objectives, &quest, preview)

	// Rewards
	fmt.Println("\n--- Quest Rewards ---")
	quest.Rewards = promptRewardsWithPreview(reader, quest.Rewards, &quest, preview)

	return quest, nil
}

// promptObjectivesWithPreview prompts for quest objectives and broadcasts updates.
func promptObjectivesWithPreview(reader *bufio.Reader, existing []game.QuestObjective, quest *game.Quest, preview *previewServer) []game.QuestObjective {
	objectives := promptObjectives(reader, existing)
	quest.Objectives = objectives
	if preview != nil {
		preview.broadcastQuest(*quest)
	}
	return objectives
}

// promptRewardsWithPreview prompts for quest rewards and broadcasts updates.
func promptRewardsWithPreview(reader *bufio.Reader, existing []game.QuestReward, quest *game.Quest, preview *previewServer) []game.QuestReward {
	rewards := promptRewards(reader, existing)
	quest.Rewards = rewards
	if preview != nil {
		preview.broadcastQuest(*quest)
	}
	return rewards
}

// promptString prompts for a string value with a default.
func promptString(reader *bufio.Reader, prompt, defaultVal, prefix string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else if prefix != "" {
		fmt.Printf("%s [%s...]: ", prompt, prefix)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		if defaultVal != "" {
			return defaultVal
		}
		if prefix != "" {
			return prefix + "new"
		}
	}
	return input
}

// promptInt prompts for an integer value with a default.
func promptInt(reader *bufio.Reader, prompt string, defaultVal int) int {
	fmt.Printf("%s [%d]: ", prompt, defaultVal)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("Invalid number, using default: %d\n", defaultVal)
		return defaultVal
	}
	return val
}

// promptObjectives prompts for quest objectives.
func promptObjectives(reader *bufio.Reader, existing []game.QuestObjective) []game.QuestObjective {
	objectives := existing

	// Show existing objectives
	if len(objectives) > 0 {
		fmt.Println("Current objectives:")
		for i, obj := range objectives {
			fmt.Printf("  %d. %s (0/%d)\n", i+1, obj.Description, obj.Required)
		}
		fmt.Println()
	}

	// Ask if user wants to modify
	fmt.Print("Modify objectives? (y/n) [n]: ")
	input, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(input)) != "y" {
		if len(objectives) > 0 {
			return objectives
		}
		// No objectives yet, must add some
	} else {
		objectives = nil // Clear and rebuild
	}

	// Add objectives
	fmt.Println("Enter objectives (empty description to finish):")
	for i := 1; ; i++ {
		fmt.Printf("\nObjective %d:\n", i)
		desc := promptString(reader, "  Description", "", "")
		if desc == "" {
			if len(objectives) == 0 {
				fmt.Println("  At least one objective is required!")
				i--
				continue
			}
			break
		}
		required := promptInt(reader, "  Required count", 1)
		objectives = append(objectives, game.QuestObjective{
			Description: desc,
			Progress:    0,
			Required:    required,
			Completed:   false,
		})
	}

	return objectives
}

// promptRewards prompts for quest rewards.
func promptRewards(reader *bufio.Reader, existing []game.QuestReward) []game.QuestReward {
	rewards := existing

	// Show existing rewards
	if len(rewards) > 0 {
		fmt.Println("Current rewards:")
		for i, r := range rewards {
			if r.Type == "item" {
				fmt.Printf("  %d. %s: %d x %s\n", i+1, r.Type, r.Value, r.ItemID)
			} else {
				fmt.Printf("  %d. %s: %d\n", i+1, r.Type, r.Value)
			}
		}
		fmt.Println()
	}

	// Ask if user wants to modify
	fmt.Print("Modify rewards? (y/n) [n]: ")
	input, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(input)) != "y" {
		if len(rewards) > 0 {
			return rewards
		}
	} else {
		rewards = nil // Clear and rebuild
	}

	// Add rewards
	fmt.Println("Enter rewards (empty type to finish):")
	fmt.Println("  Valid types: gold, exp, item")
	for i := 1; ; i++ {
		fmt.Printf("\nReward %d:\n", i)
		rType := promptString(reader, "  Type (gold/exp/item)", "", "")
		if rType == "" {
			break
		}

		rType = strings.ToLower(rType)
		if rType != "gold" && rType != "exp" && rType != "item" {
			fmt.Println("  Invalid type! Use: gold, exp, or item")
			i--
			continue
		}

		value := promptInt(reader, "  Value/Amount", 100)
		reward := game.QuestReward{Type: rType, Value: value}

		if rType == "item" {
			reward.ItemID = promptString(reader, "  Item ID", "", "item_")
		}

		rewards = append(rewards, reward)
	}

	return rewards
}

// validateQuest performs basic validation on the quest.
func validateQuest(quest game.Quest) error {
	if err := validateQuestBasicFields(quest); err != nil {
		return err
	}
	if err := validateQuestObjectivesList(quest.Objectives); err != nil {
		return err
	}
	return validateQuestRewardsList(quest.Rewards)
}

// validateQuestBasicFields validates the quest ID and title.
func validateQuestBasicFields(quest game.Quest) error {
	if quest.ID == "" {
		return fmt.Errorf("quest ID is required")
	}
	if quest.Title == "" {
		return fmt.Errorf("quest title is required")
	}
	return nil
}

// validateQuestObjectivesList validates the quest objectives list.
func validateQuestObjectivesList(objectives []game.QuestObjective) error {
	if len(objectives) == 0 {
		return fmt.Errorf("at least one objective is required")
	}
	for i, obj := range objectives {
		if obj.Description == "" {
			return fmt.Errorf("objective %d has no description", i+1)
		}
		if obj.Required <= 0 {
			return fmt.Errorf("objective %d has invalid required count", i+1)
		}
	}
	return nil
}

// validateQuestRewardsList validates the quest rewards list.
func validateQuestRewardsList(rewards []game.QuestReward) error {
	for i, r := range rewards {
		if r.Type != "gold" && r.Type != "exp" && r.Type != "item" {
			return fmt.Errorf("reward %d has invalid type: %s", i+1, r.Type)
		}
		if r.Value <= 0 {
			return fmt.Errorf("reward %d has invalid value", i+1)
		}
		if r.Type == "item" && r.ItemID == "" {
			return fmt.Errorf("reward %d (item) requires an item ID", i+1)
		}
	}
	return nil
}

// outputQuest writes the quest to the specified output file or stdout.
func outputQuest(quest game.Quest, outputFile string) error {
	data, err := yaml.Marshal(quest)
	if err != nil {
		return fmt.Errorf("failed to marshal quest: %w", err)
	}

	// Add a header comment
	output := fmt.Sprintf("# Quest configuration generated by quest-builder\n# ID: %s\n# Title: %s\n\n%s",
		quest.ID, quest.Title, string(data))

	if outputFile == "" {
		fmt.Print(output)
		return nil
	}

	if err := os.WriteFile(outputFile, []byte(output), 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Quest written to: %s\n", outputFile)
	return nil
}
