package game

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiplomacyManager_InitializeRelation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	rel, err := dm.InitializeRelation("faction1", "faction2")
	require.NoError(t, err)
	assert.NotEmpty(t, rel.ID)
	assert.Equal(t, DiplomaticStateNeutral, rel.State)
	assert.Equal(t, 0, rel.Opinion)
	assert.Equal(t, 0, rel.Trust)

	// Get same relation with reversed order
	rel2, err := dm.GetRelation("faction2", "faction1")
	require.NoError(t, err)
	assert.Equal(t, rel.ID, rel2.ID)

	// Cannot create relation with self
	_, err = dm.InitializeRelation("faction1", "faction1")
	assert.ErrorIs(t, err, ErrCannotSelfRelation)
}

func TestDiplomacyManager_DeclareWar(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("kingdom", "empire")

	err := dm.DeclareWar("kingdom", "empire", "territorial dispute")
	require.NoError(t, err)

	rel, _ := dm.GetRelation("kingdom", "empire")
	assert.Equal(t, DiplomaticStateWar, rel.State)
	assert.Equal(t, -100, rel.Opinion)
	assert.Equal(t, -100, rel.Trust)
	assert.False(t, rel.TradeTreaty)
	assert.False(t, rel.DefensivePact)

	// Cannot declare war again
	err = dm.DeclareWar("kingdom", "empire", "another reason")
	assert.ErrorIs(t, err, ErrAlreadyAtWar)

	// Check history
	assert.Len(t, rel.History, 1)
	assert.Equal(t, ActionDeclareWar, rel.History[0].Action)
}

func TestDiplomacyManager_PeaceProcess(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("kingdom", "empire")
	dm.DeclareWar("kingdom", "empire", "war")

	// Offer peace
	err := dm.OfferPeace("empire", "kingdom")
	require.NoError(t, err)

	// Accept peace
	err = dm.AcceptPeace("kingdom", "empire")
	require.NoError(t, err)

	rel, _ := dm.GetRelation("kingdom", "empire")
	assert.Equal(t, DiplomaticStateHostile, rel.State)
	assert.Equal(t, -50, rel.Opinion)
}

func TestDiplomacyManager_Alliance(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("kingdom", "republic")

	// Propose alliance
	err := dm.ProposeAlliance("kingdom", "republic")
	require.NoError(t, err)

	// Accept alliance
	err = dm.AcceptAlliance("republic", "kingdom")
	require.NoError(t, err)

	rel, _ := dm.GetRelation("kingdom", "republic")
	assert.Equal(t, DiplomaticStateAllied, rel.State)
	assert.True(t, rel.DefensivePact)
	assert.True(t, rel.MilitaryAccess)
	assert.Equal(t, 75, rel.Opinion)
	assert.Equal(t, 50, rel.Trust)

	// Check helper function
	assert.True(t, dm.AreAllied("kingdom", "republic"))
	assert.False(t, dm.AreAtWar("kingdom", "republic"))
}

func TestDiplomacyManager_BreakAlliance(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("kingdom", "republic")
	dm.AcceptAlliance("kingdom", "republic")

	err := dm.BreakAlliance("kingdom", "republic")
	require.NoError(t, err)

	rel, _ := dm.GetRelation("kingdom", "republic")
	assert.Equal(t, DiplomaticStateTense, rel.State)
	assert.Equal(t, -25, rel.Opinion)
	assert.Equal(t, -50, rel.Trust)
	assert.False(t, rel.DefensivePact)
	assert.False(t, rel.MilitaryAccess)
}

func TestDiplomacyManager_TradeAgreement(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("kingdom", "merchants")

	err := dm.SignTradeAgreement("kingdom", "merchants")
	require.NoError(t, err)

	rel, _ := dm.GetRelation("kingdom", "merchants")
	assert.True(t, rel.TradeTreaty)
	assert.Equal(t, 10, rel.Opinion)

	// Cannot sign trade during war
	dm.InitializeRelation("kingdom", "empire")
	dm.DeclareWar("kingdom", "empire", "war")

	err = dm.SignTradeAgreement("kingdom", "empire")
	assert.ErrorIs(t, err, ErrInvalidAction)
}

func TestDiplomacyManager_SendGift(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("kingdom", "republic")

	err := dm.SendGift("kingdom", "republic", 500)
	require.NoError(t, err)

	rel, _ := dm.GetRelation("kingdom", "republic")
	assert.Greater(t, rel.Opinion, 0)
	assert.Greater(t, rel.Trust, 0)
}

func TestDiplomacyManager_ModifyOpinion(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("faction1", "faction2")

	// Increase opinion
	err := dm.ModifyOpinion("faction1", "faction2", 50)
	require.NoError(t, err)

	rel, _ := dm.GetRelation("faction1", "faction2")
	assert.Equal(t, 50, rel.Opinion)
	assert.Equal(t, DiplomaticStateFriendly, rel.State)

	// Decrease opinion
	err = dm.ModifyOpinion("faction1", "faction2", -100)
	require.NoError(t, err)

	assert.Equal(t, -50, rel.Opinion)
	assert.Equal(t, DiplomaticStateTense, rel.State)

	// Opinion capped at +/-100
	dm.ModifyOpinion("faction1", "faction2", -200)
	assert.Equal(t, -100, rel.Opinion)
}

func TestDiplomacyManager_GetFactionRelations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("kingdom", "empire")
	dm.InitializeRelation("kingdom", "republic")
	dm.InitializeRelation("empire", "republic")

	kingdomRels := dm.GetFactionRelations("kingdom")
	assert.Len(t, kingdomRels, 2)

	empireRels := dm.GetFactionRelations("empire")
	assert.Len(t, empireRels, 2)
}

func TestDiplomacyManager_GetAllRelations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("f1", "f2")
	dm.InitializeRelation("f1", "f3")
	dm.InitializeRelation("f2", "f3")

	all := dm.GetAllRelations()
	assert.Len(t, all, 3)
}

func TestDiplomacyManager_DecayRelations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("f1", "f2")
	dm.ModifyOpinion("f1", "f2", 50)
	dm.ModifyTrust("f1", "f2", 40)

	rel, _ := dm.GetRelation("f1", "f2")
	initialOpinion := rel.Opinion
	initialTrust := rel.Trust

	// Apply decay
	dm.DecayRelations(0.1)

	assert.Less(t, rel.Opinion, initialOpinion, "opinion should decay")
	assert.Less(t, rel.Trust, initialTrust, "trust should decay")
}

func TestDiplomacyManager_GetOpinion(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("f1", "f2")
	dm.ModifyOpinion("f1", "f2", 30)

	opinion, err := dm.GetOpinion("f1", "f2")
	require.NoError(t, err)
	assert.Equal(t, 30, opinion)

	// Reversed order should work
	opinion2, err := dm.GetOpinion("f2", "f1")
	require.NoError(t, err)
	assert.Equal(t, opinion, opinion2)

	// Non-existent relation
	_, err = dm.GetOpinion("f1", "nonexistent")
	assert.ErrorIs(t, err, ErrRelationNotFound)
}

func TestDiplomacyManager_StateTransitions(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("f1", "f2")

	tests := []struct {
		opinionChange int
		expectedState DiplomaticState
	}{
		{80, DiplomaticStateFriendly},
		{-30, DiplomaticStateFriendly}, // 50, still friendly
		{-40, DiplomaticStateNeutral},  // 10, neutral
		{-40, DiplomaticStateTense},    // -30, tense (below -25)
		{-50, DiplomaticStateHostile},  // -80, hostile
	}

	for _, tc := range tests {
		dm.ModifyOpinion("f1", "f2", tc.opinionChange)
		rel, _ := dm.GetRelation("f1", "f2")
		assert.Equal(t, tc.expectedState, rel.State, "expected state %s for opinion change %d", tc.expectedState, tc.opinionChange)
	}
}

func TestDiplomacyManager_AllianceErrors(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dm := NewDiplomacyManager(logger)

	dm.InitializeRelation("f1", "f2")

	// Cannot propose alliance while at war
	dm.DeclareWar("f1", "f2", "war")
	err := dm.ProposeAlliance("f1", "f2")
	assert.ErrorIs(t, err, ErrInvalidAction)

	// Reset
	dm.AcceptPeace("f1", "f2")

	// Cannot break alliance when not allied
	err = dm.BreakAlliance("f1", "f2")
	assert.ErrorIs(t, err, ErrNotAllied)

	// Accept alliance and try again
	dm.AcceptAlliance("f1", "f2")
	err = dm.AcceptAlliance("f1", "f2")
	assert.ErrorIs(t, err, ErrAlreadyAllied)
}
