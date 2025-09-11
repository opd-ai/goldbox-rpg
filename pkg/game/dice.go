package game

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DiceRoll represents the result of rolling dice
type DiceRoll struct {
	Rolls    []int // Individual die results
	Total    int   // Sum of all rolls
	Modifier int   // Modifier applied to the total
	Final    int   // Final result (Total + Modifier)
}

// DiceRoller handles rolling dice with various expressions
type DiceRoller struct {
	rng *rand.Rand
}

// NewDiceRoller creates a new dice roller with a random seed
func NewDiceRoller() *DiceRoller {
	return &DiceRoller{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewDiceRollerWithSeed creates a new dice roller with a specific seed (for testing)
func NewDiceRollerWithSeed(seed int64) *DiceRoller {
	return &DiceRoller{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Roll parses and rolls a dice expression like "3d6+2", "1d20", "2d4-1"
func (dr *DiceRoller) Roll(expression string) (*DiceRoll, error) {
	logrus.WithFields(logrus.Fields{
		"function":   "Roll",
		"package":    "game",
		"expression": expression,
	}).Debug("entering Roll")

	if expression == "" {
		logrus.WithFields(logrus.Fields{
			"function":   "Roll",
			"package":    "game",
			"expression": expression,
		}).Debug("empty expression, returning empty roll")
		return &DiceRoll{}, nil
	}

	// Clean up the expression
	expression = strings.ReplaceAll(expression, " ", "")
	expression = strings.ToLower(expression)

	// Regex to parse dice expressions like "3d6+2", "1d20-1", "2d4"
	re := regexp.MustCompile(`^(\d+)d(\d+)([+-]\d+)?$`)
	matches := re.FindStringSubmatch(expression)

	if len(matches) < 3 {
		logrus.WithFields(logrus.Fields{
			"function":   "Roll",
			"package":    "game",
			"expression": expression,
			"matches":    len(matches),
		}).Error("invalid dice expression format")
		return nil, fmt.Errorf("invalid dice expression: %s", expression)
	}

	// Parse number of dice
	numDice, err := strconv.Atoi(matches[1])
	if err != nil || numDice <= 0 {
		logrus.WithFields(logrus.Fields{
			"function":   "Roll",
			"package":    "game",
			"expression": expression,
			"num_dice":   matches[1],
			"error":      err,
		}).Error("invalid number of dice")
		return nil, fmt.Errorf("invalid number of dice: %s", matches[1])
	}

	// Parse die size
	dieSize, err := strconv.Atoi(matches[2])
	if err != nil || dieSize <= 0 {
		logrus.WithFields(logrus.Fields{
			"function":   "Roll",
			"package":    "game",
			"expression": expression,
			"die_size":   matches[2],
			"error":      err,
		}).Error("invalid die size")
		return nil, fmt.Errorf("invalid die size: %s", matches[2])
	}

	// Parse modifier (optional)
	var modifier int
	if len(matches) >= 4 && matches[3] != "" {
		modifier, err = strconv.Atoi(matches[3])
		if err != nil {
			return nil, fmt.Errorf("invalid modifier: %s", matches[3])
		}
	}

	// Roll the dice
	rolls := make([]int, numDice)
	total := 0

	for i := 0; i < numDice; i++ {
		roll := dr.rng.Intn(dieSize) + 1
		rolls[i] = roll
		total += roll
	}

	final := total + modifier

	return &DiceRoll{
		Rolls:    rolls,
		Total:    total,
		Modifier: modifier,
		Final:    final,
	}, nil
}

// RollMultiple rolls multiple dice expressions and returns the sum
func (dr *DiceRoller) RollMultiple(expressions []string) (*DiceRoll, error) {
	var allRolls []int
	var totalSum int
	var totalModifier int

	for _, expr := range expressions {
		if expr == "" {
			continue
		}

		roll, err := dr.Roll(expr)
		if err != nil {
			return nil, fmt.Errorf("failed to roll %s: %w", expr, err)
		}

		allRolls = append(allRolls, roll.Rolls...)
		totalSum += roll.Total
		totalModifier += roll.Modifier
	}

	final := totalSum + totalModifier

	return &DiceRoll{
		Rolls:    allRolls,
		Total:    totalSum,
		Modifier: totalModifier,
		Final:    final,
	}, nil
}

// String returns a string representation of the dice roll
func (dr *DiceRoll) String() string {
	if len(dr.Rolls) == 0 {
		return "0"
	}

	rollStr := fmt.Sprintf("%v", dr.Rolls)
	if dr.Modifier == 0 {
		return fmt.Sprintf("%s = %d", rollStr, dr.Total)
	} else if dr.Modifier > 0 {
		return fmt.Sprintf("%s + %d = %d", rollStr, dr.Modifier, dr.Final)
	} else {
		return fmt.Sprintf("%s - %d = %d", rollStr, -dr.Modifier, dr.Final)
	}
}

// CalculateDiceAverage calculates the average result for a dice expression without rolling
func CalculateDiceAverage(expression string) (float64, error) {
	if expression == "" {
		return 0, nil
	}

	// Clean up the expression
	expression = strings.ReplaceAll(expression, " ", "")
	expression = strings.ToLower(expression)

	// Regex to parse dice expressions
	re := regexp.MustCompile(`^(\d+)d(\d+)([+-]\d+)?$`)
	matches := re.FindStringSubmatch(expression)

	if len(matches) < 3 {
		return 0, fmt.Errorf("invalid dice expression: %s", expression)
	}

	// Parse values
	numDice, err := strconv.Atoi(matches[1])
	if err != nil || numDice <= 0 {
		return 0, fmt.Errorf("invalid number of dice: %s", matches[1])
	}

	dieSize, err := strconv.Atoi(matches[2])
	if err != nil || dieSize <= 0 {
		return 0, fmt.Errorf("invalid die size: %s", matches[2])
	}

	var modifier int
	if len(matches) >= 4 && matches[3] != "" {
		modifier, err = strconv.Atoi(matches[3])
		if err != nil {
			return 0, fmt.Errorf("invalid modifier: %s", matches[3])
		}
	}

	// Calculate average: (dieSize + 1) / 2 is the average of a single die
	avgPerDie := float64(dieSize+1) / 2.0
	totalAvg := float64(numDice)*avgPerDie + float64(modifier)

	return totalAvg, nil
}

// GlobalDiceRoller is the global dice roller instance used throughout the game
var GlobalDiceRoller = NewDiceRoller()
