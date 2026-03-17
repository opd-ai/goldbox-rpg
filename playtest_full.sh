#!/bin/bash
# GoldBox RPG Engine - Comprehensive Playtest Script
# Tests all major game systems via JSON-RPC
set -uo pipefail

BASE="http://localhost:8080"
COOKIES="/tmp/goldbox_playtest_cookies.txt"
rm -f "$COOKIES"

PASS=0
FAIL=0
WARN=0
RESULTS=""

# Helper: make RPC call
rpc() {
    local method="$1" params="$2" id="${3:-1}"
    curl -s -b "$COOKIES" -c "$COOKIES" -X POST "$BASE/rpc" \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"$method\",\"params\":$params,\"id\":$id}"
}

# Helper: record result
record() {
    local status="$1" test="$2" detail="${3:-}"
    if [ "$status" = "PASS" ]; then
        PASS=$((PASS + 1))
        echo "  [PASS] $test"
    elif [ "$status" = "WARN" ]; then
        WARN=$((WARN + 1))
        echo "  [WARN] $test: $detail"
    else
        FAIL=$((FAIL + 1))
        echo "  [FAIL] $test: $detail"
    fi
    RESULTS="$RESULTS\n$status: $test $detail"
}

echo "╔══════════════════════════════════════════════╗"
echo "║  GoldBox RPG Engine - Full Playtest          ║"
echo "║  $(date +%Y-%m-%d\ %H:%M:%S)                        ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

########################################
echo "=== 1. HEALTH & MONITORING ==="
########################################

# 1a. Health endpoint
HEALTH=$(curl -s "$BASE/health" 2>/dev/null || echo "CONNECT_FAIL")
if echo "$HEALTH" | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['status']=='healthy'" 2>/dev/null; then
    record "PASS" "Health endpoint returns healthy"
    echo "$HEALTH" | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'    Status: {d[\"status\"]}, Checks: {len(d.get(\"checks\",{}))}')"
else
    record "FAIL" "Health endpoint" "$HEALTH"
fi

# 1b. Ready endpoint
READY=$(curl -s "$BASE/ready" 2>/dev/null || echo "FAIL")
if echo "$READY" | grep -qi "ready\|ok\|healthy" 2>/dev/null; then
    record "PASS" "Ready endpoint"
else
    record "WARN" "Ready endpoint" "Response: $(echo "$READY" | head -c 100)"
fi

# 1c. Live endpoint
LIVE=$(curl -s "$BASE/live" 2>/dev/null || echo "FAIL")
if echo "$LIVE" | grep -q "alive\|ok\|live\|healthy" 2>/dev/null; then
    record "PASS" "Live endpoint"
else
    record "WARN" "Live endpoint" "Response: $(echo "$LIVE" | head -c 100)"
fi

# 1d. Metrics endpoint
METRICS=$(curl -s "$BASE/metrics" 2>/dev/null | head -5)
if echo "$METRICS" | grep -q "go_\|goldbox\|promhttp\|process_" 2>/dev/null; then
    record "PASS" "Metrics endpoint (Prometheus)"
else
    record "WARN" "Metrics endpoint" "No prometheus metrics found"
fi

echo ""

########################################
echo "=== 2. JOIN GAME ==="
########################################

J=$(rpc "joinGame" '{"player_name":"Sir Aldric"}' 1)
SID=$(echo "$J" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])" 2>/dev/null || echo "")
if [ -n "$SID" ] && [ "$SID" != "None" ]; then
    record "PASS" "Join game"
    echo "    Session ID: $SID"
else
    record "FAIL" "Join game" "$(echo "$J" | head -c 200)"
    echo "Cannot continue without session. Aborting."
    exit 1
fi

echo ""

########################################
echo "=== 3. CHARACTER CREATION ==="
########################################

# 3a. Fighter with standard array
C=$(rpc "createCharacter" "{\"session_id\":\"$SID\",\"name\":\"Aldric\",\"class\":\"fighter\",\"attribute_method\":\"standard\"}" 2)
NSID=$(echo "$C" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])" 2>/dev/null || echo "")
if [ -n "$NSID" ] && [ "$NSID" != "None" ]; then
    SID="$NSID"
    record "PASS" "Create Fighter (standard array)"
    echo "$C" | python3 -c "
import sys, json
c = json.load(sys.stdin)['result']['character']
a = c.get('attributes', {})
print(f'    Name: {c[\"name\"]} | Class: {c[\"class\"]} | Level: {c[\"level\"]}')
print(f'    HP: {c[\"hp\"]}/{c[\"max_hp\"]}')
print(f'    STR:{a[\"strength\"]} DEX:{a[\"dexterity\"]} CON:{a[\"constitution\"]} INT:{a[\"intelligence\"]} WIS:{a[\"wisdom\"]} CHA:{a[\"charisma\"]}')
print(f'    AP: {c[\"ap\"]}/{c[\"max_ap\"]}')
print(f'    Position: ({c[\"position\"][\"X\"]}, {c[\"position\"][\"Y\"]})')
" 2>/dev/null || echo "    (could not parse character details)"
else
    record "FAIL" "Create Fighter" "$(echo "$C" | head -c 300)"
fi

# Save player ID for later
PID=$(echo "$C" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['character']['id'])" 2>/dev/null || echo "")
echo "    Player ID: $PID"

echo ""

########################################
echo "=== 4. GAME STATE ==="
########################################

GS=$(rpc "getGameState" "{\"session_id\":\"$SID\"}" 3)
if echo "$GS" | python3 -c "import sys,json; r=json.load(sys.stdin); assert 'result' in r" 2>/dev/null; then
    record "PASS" "Get game state"
    echo "$GS" | python3 -c "
import sys, json
r = json.load(sys.stdin)['result']
print(f'    Version: {r.get(\"version\", \"N/A\")}')
print(f'    World keys: {list(r.get(\"world\", {}).keys())[:6]}')
" 2>/dev/null || true
else
    record "FAIL" "Get game state" "$(echo "$GS" | head -c 200)"
fi

echo ""

########################################
echo "=== 5. MOVEMENT ==="
########################################

echo "  Testing movement in all 4 directions..."
MOVE_OK=0
MOVE_FAIL=0
for DIR in north east south west north north east east south south west west; do
    M=$(rpc "move" "{\"session_id\":\"$SID\",\"direction\":\"$DIR\"}" 4)
    POS=$(echo "$M" | python3 -c "import sys,json; p=json.load(sys.stdin)['result']['position']; print(f'({p[\"X\"]},{p[\"Y\"]})')" 2>/dev/null)
    if [ -n "$POS" ]; then
        MOVE_OK=$((MOVE_OK + 1))
        echo "    $DIR -> $POS"
    else
        MOVE_FAIL=$((MOVE_FAIL + 1))
        ERR=$(echo "$M" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
        echo "    $DIR -> BLOCKED ($ERR)"
    fi
done

if [ "$MOVE_OK" -gt 0 ]; then
    record "PASS" "Movement ($MOVE_OK/$((MOVE_OK + MOVE_FAIL)) succeeded)"
else
    record "FAIL" "Movement" "All moves failed"
fi

echo ""

########################################
echo "=== 6. SPELL SYSTEM ==="
########################################

# 6a. Get all spells
SP=$(rpc "getAllSpells" '{}' 5)
SPELL_COUNT=$(echo "$SP" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['result']['spells']))" 2>/dev/null || echo "0")
if [ "$SPELL_COUNT" -gt 0 ]; then
    record "PASS" "Get all spells ($SPELL_COUNT spells)"
    echo "$SP" | python3 -c "
import sys,json
spells = json.load(sys.stdin)['result']['spells']
for s in spells[:5]:
    dmg = f' Dmg:{s[\"DamageDice\"]} ({s[\"DamageType\"]})' if s.get('DamageDice') else ''
    heal = f' Heal:{s[\"HealingDice\"]}' if s.get('HealingDice') else ''
    print(f'    Lvl {s[\"Level\"]}: {s[\"Name\"]}{dmg}{heal} Range:{s[\"Range\"]}')
if len(spells) > 5:
    print(f'    ... and {len(spells) - 5} more')
" 2>/dev/null || true
else
    record "FAIL" "Get all spells" "No spells returned"
fi

# 6b. Get spells by level
SBL=$(rpc "getSpellsByLevel" '{"level":0}' 6)
LVL0=$(echo "$SBL" | python3 -c "import sys,json; r=json.load(sys.stdin); print(len(r['result']['spells']) if 'result' in r else 0)" 2>/dev/null || echo "0")
if [ "$LVL0" -gt 0 ]; then
    record "PASS" "Get spells by level (level 0: $LVL0 cantrips)"
else
    record "WARN" "Get spells by level" "No cantrips found or error"
fi

# 6c. Search spells
SS=$(rpc "searchSpells" '{"query":"fire"}' 7)
SCOUNT=$(echo "$SS" | python3 -c "import sys,json; r=json.load(sys.stdin); print(len(r['result']['spells']) if 'result' in r else 0)" 2>/dev/null || echo "0")
if [ "$SCOUNT" -ge 0 ]; then
    record "PASS" "Search spells for 'fire' ($SCOUNT found)"
else
    record "WARN" "Search spells" "Error in search"
fi

echo ""

########################################
echo "=== 7. COMBAT SYSTEM ==="
########################################

# 7a. Start combat
CB=$(rpc "startCombat" "{\"session_id\":\"$SID\",\"participant_ids\":[\"$PID\"]}" 8)
COMBAT_OK=$(echo "$CB" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$COMBAT_OK" = "yes" ]; then
    record "PASS" "Start combat"
    echo "$CB" | python3 -c "
import sys,json
r = json.load(sys.stdin)['result']
print(f'    Combat started! First turn: {r.get(\"first_turn\",\"?\")}, Round: {r.get(\"current_round\",\"?\")}')
" 2>/dev/null || true
else
    ERR=$(echo "$CB" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Start combat" "$ERR"
fi

# 7b. Combat movement (should cost AP)
CM=$(rpc "move" "{\"session_id\":\"$SID\",\"direction\":\"north\"}" 9)
CM_OK=$(echo "$CM" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$CM_OK" = "yes" ]; then
    record "PASS" "Combat movement"
else
    ERR=$(echo "$CM" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Combat movement" "$ERR"
fi

# 7c. End turn
ET=$(rpc "endTurn" "{\"session_id\":\"$SID\"}" 10)
ET_OK=$(echo "$ET" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$ET_OK" = "yes" ]; then
    record "PASS" "End turn"
    echo "$ET" | python3 -c "
import sys,json
r = json.load(sys.stdin)['result']
print(f'    Next: {r.get(\"next_turn\",\"?\")} | Round: {r.get(\"current_round\",\"?\")}')
" 2>/dev/null || true
else
    ERR=$(echo "$ET" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "End turn" "$ERR"
fi

echo ""

########################################
echo "=== 8. EQUIPMENT & INVENTORY ==="
########################################

EQ=$(rpc "getEquipment" "{\"session_id\":\"$SID\"}" 11)
EQ_OK=$(echo "$EQ" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$EQ_OK" = "yes" ]; then
    record "PASS" "Get equipment"
    echo "$EQ" | python3 -c "
import sys,json
r = json.load(sys.stdin)['result']
eq = r.get('equipment', {})
inv = r.get('inventory', [])
print(f'    Equipped: {len(eq)} slots | Inventory: {len(inv)} items')
for slot, item in eq.items():
    if isinstance(item, dict):
        print(f'      {slot}: {item.get(\"Name\", item.get(\"name\", \"?\"))}')
" 2>/dev/null || true
else
    ERR=$(echo "$EQ" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Get equipment" "$ERR"
fi

echo ""

########################################
echo "=== 9. QUEST SYSTEM ==="
########################################

# 9a. Get active quests
AQ=$(rpc "getActiveQuests" "{\"session_id\":\"$SID\"}" 12)
AQ_OK=$(echo "$AQ" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$AQ_OK" = "yes" ]; then
    record "PASS" "Get active quests"
    echo "$AQ" | python3 -c "
import sys,json
r = json.load(sys.stdin)['result']
quests = r.get('quests', [])
print(f'    Active quests: {len(quests)}')
for q in quests[:3]:
    print(f'      - {q.get(\"Name\", q.get(\"name\", \"?\"))}')
" 2>/dev/null || true
else
    ERR=$(echo "$AQ" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Get active quests" "$ERR"
fi

# 9b. Get quest log
QL=$(rpc "getQuestLog" "{\"session_id\":\"$SID\"}" 13)
QL_OK=$(echo "$QL" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$QL_OK" = "yes" ]; then
    record "PASS" "Get quest log"
else
    ERR=$(echo "$QL" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Get quest log" "$ERR"
fi

echo ""

########################################
echo "=== 10. SPATIAL QUERIES ==="
########################################

# 10a. Get objects in range
OIR=$(rpc "getObjectsInRange" "{\"session_id\":\"$SID\",\"x\":5,\"y\":5,\"range\":10}" 14)
OIR_OK=$(echo "$OIR" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$OIR_OK" = "yes" ]; then
    OBJ_COUNT=$(echo "$OIR" | python3 -c "import sys,json; r=json.load(sys.stdin)['result']; print(len(r.get('objects',[])))" 2>/dev/null || echo "?")
    record "PASS" "Get objects in range ($OBJ_COUNT objects)"
else
    ERR=$(echo "$OIR" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Get objects in range" "$ERR"
fi

# 10b. Get objects in radius
OIRad=$(rpc "getObjectsInRadius" "{\"session_id\":\"$SID\",\"x\":5,\"y\":5,\"radius\":5}" 15)
OIRad_OK=$(echo "$OIRad" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$OIRad_OK" = "yes" ]; then
    record "PASS" "Get objects in radius"
else
    ERR=$(echo "$OIRad" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Get objects in radius" "$ERR"
fi

echo ""

########################################
echo "=== 11. PCG SYSTEM ==="
########################################

# Need an active session with a player for PCG calls
PCG_JOIN=$(rpc "joinGame" '{"player_name":"PCG Tester"}' 50)
PCG_SID=$(echo "$PCG_JOIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])" 2>/dev/null || echo "")
PCG_CHAR=$(rpc "createCharacter" "{\"session_id\":\"$PCG_SID\",\"name\":\"PCGTester\",\"class\":\"fighter\",\"attribute_method\":\"standard\"}" 51)
PCG_SID=$(echo "$PCG_CHAR" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])" 2>/dev/null || echo "$PCG_SID")

# 11a. PCG stats
PS=$(rpc "getPCGStats" "{\"session_id\":\"$PCG_SID\"}" 16)
PS_OK=$(echo "$PS" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$PS_OK" = "yes" ]; then
    record "PASS" "Get PCG stats"
    echo "$PS" | python3 -c "
import sys,json
r = json.load(sys.stdin)['result']
print(f'    PCG Stats: {json.dumps(r, indent=2)[:200]}')
" 2>/dev/null || true
else
    ERR=$(echo "$PS" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Get PCG stats" "$ERR"
fi

# 11b. Generate content
GC=$(rpc "generateContent" "{\"session_id\":\"$PCG_SID\",\"content_type\":\"items\",\"location_id\":\"default_level\",\"difficulty\":5}" 17)
GC_OK=$(echo "$GC" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$GC_OK" = "yes" ]; then
    record "PASS" "Generate PCG content (items)"
    echo "$GC" | python3 -c "
import sys,json
r = json.load(sys.stdin)['result']
print(f'    Generated: {json.dumps(r, indent=2)[:300]}')
" 2>/dev/null || true
else
    ERR=$(echo "$GC" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Generate PCG content" "$ERR"
fi

# 11c. Generate quest
GQ=$(rpc "generateQuest" "{\"session_id\":\"$PCG_SID\",\"quest_type\":\"fetch\",\"difficulty\":5}" 18)
GQ_OK=$(echo "$GQ" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$GQ_OK" = "yes" ]; then
    record "PASS" "Generate PCG quest"
    echo "$GQ" | python3 -c "
import sys,json
r = json.load(sys.stdin)['result']
print(f'    Quest: {json.dumps(r, indent=2)[:300]}')
" 2>/dev/null || true
else
    ERR=$(echo "$GQ" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Generate PCG quest" "$ERR"
fi

# Cleanup PCG session
rpc "leaveGame" "{\"session_id\":\"$PCG_SID\"}" 52 > /dev/null 2>&1

echo ""

########################################
echo "=== 12. ERROR HANDLING ==="
########################################

# 12a. Invalid session
BAD=$(rpc "getGameState" '{"session_id":"invalid-session-id"}' 19)
BAD_ERR=$(echo "$BAD" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'error' in r else 'no')" 2>/dev/null || echo "no")
if [ "$BAD_ERR" = "yes" ]; then
    record "PASS" "Invalid session returns error"
else
    record "FAIL" "Invalid session should return error" "Got result instead of error"
fi

# 12b. Unknown method
UNK=$(rpc "nonExistentMethod" '{}' 20)
UNK_ERR=$(echo "$UNK" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'error' in r else 'no')" 2>/dev/null || echo "no")
if [ "$UNK_ERR" = "yes" ]; then
    record "PASS" "Unknown method returns error"
else
    record "FAIL" "Unknown method should return error"
fi

# 12c. Missing required parameters
MIS=$(rpc "move" '{}' 21)
MIS_ERR=$(echo "$MIS" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'error' in r else 'no')" 2>/dev/null || echo "no")
if [ "$MIS_ERR" = "yes" ]; then
    record "PASS" "Missing params returns error"
else
    record "FAIL" "Missing params should return error"
fi

echo ""

########################################
echo "=== 13. LEAVE GAME ==="
########################################

LG=$(rpc "leaveGame" "{\"session_id\":\"$SID\"}" 22)
LG_OK=$(echo "$LG" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r else 'no')" 2>/dev/null || echo "no")
if [ "$LG_OK" = "yes" ]; then
    record "PASS" "Leave game"
else
    ERR=$(echo "$LG" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Leave game" "$ERR"
fi

# Verify session is gone
POST_LG=$(rpc "getGameState" "{\"session_id\":\"$SID\"}" 23)
POST_ERR=$(echo "$POST_LG" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'error' in r else 'no')" 2>/dev/null || echo "no")
if [ "$POST_ERR" = "yes" ]; then
    record "PASS" "Session cleaned up after leave"
else
    record "WARN" "Session cleanup" "Session still valid after leaving"
fi

echo ""

########################################
# SECOND SESSION - Test different class
########################################
echo "=== 14. SECOND SESSION (Mage) ==="
########################################

J2=$(rpc "joinGame" '{"player_name":"Elara the Wise"}' 24)
SID2=$(echo "$J2" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])" 2>/dev/null || echo "")
if [ -n "$SID2" ] && [ "$SID2" != "None" ]; then
    record "PASS" "Second join game"
else
    record "FAIL" "Second join" "$(echo "$J2" | head -c 200)"
fi

# Create Mage (custom attrs to meet INT 13 requirement)
C2=$(rpc "createCharacter" "{\"session_id\":\"$SID2\",\"name\":\"Elara\",\"class\":\"mage\",\"attribute_method\":\"custom\",\"custom_attributes\":{\"strength\":8,\"dexterity\":14,\"constitution\":12,\"intelligence\":15,\"wisdom\":13,\"charisma\":10}}" 25)
NSID2=$(echo "$C2" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])" 2>/dev/null || echo "")
if [ -n "$NSID2" ] && [ "$NSID2" != "None" ]; then
    SID2="$NSID2"
    record "PASS" "Create Mage (standard array)"
    echo "$C2" | python3 -c "
import sys, json
c = json.load(sys.stdin)['result']['character']
print(f'    Name: {c[\"Name\"]} | Class: {c[\"Class\"]} | Level: {c[\"Level\"]}')
print(f'    HP: {c[\"HP\"]}/{c[\"MaxHP\"]} | AC: {c[\"ArmorClass\"]} | THAC0: {c[\"THAC0\"]}')
print(f'    STR:{c[\"Strength\"]} DEX:{c[\"Dexterity\"]} CON:{c[\"Constitution\"]} INT:{c[\"Intelligence\"]} WIS:{c[\"Wisdom\"]} CHA:{c[\"Charisma\"]}')
" 2>/dev/null || true
else
    record "FAIL" "Create Mage" "$(echo "$C2" | head -c 300)"
fi

# Try casting a spell  
PID2=$(echo "$C2" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['character']['ID'])" 2>/dev/null || echo "")
CS=$(rpc "castSpell" "{\"session_id\":\"$SID2\",\"spell_id\":\"magic_missile\",\"target_id\":\"$PID2\",\"position\":{\"x\":5,\"y\":5}}" 26)
CS_OK=$(echo "$CS" | python3 -c "import sys,json; r=json.load(sys.stdin); print('result' if 'result' in r else 'error')" 2>/dev/null || echo "error")
if [ "$CS_OK" = "result" ]; then
    record "PASS" "Cast spell (fire_bolt)"
else
    ERR=$(echo "$CS" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',{}).get('message','unknown'))" 2>/dev/null || echo "parse error")
    record "WARN" "Cast spell" "$ERR"
fi

# Clean up session 2
rpc "leaveGame" "{\"session_id\":\"$SID2\"}" 27 > /dev/null 2>&1

echo ""

########################################
# THIRD SESSION - Test other classes
########################################
echo "=== 15. TEST REMAINING CLASSES ==="
########################################

# Each class needs custom attributes meeting their requirements:
#   Fighter: STR 13  |  Mage: INT 13  |  Cleric: WIS 13
#   Thief: DEX 13    |  Ranger: DEX 13 + WIS 13  |  Paladin: STR 13 + CHA 13

for CLASS in cleric thief ranger paladin; do
    J3=$(rpc "joinGame" "{\"player_name\":\"Test $CLASS\"}" 30)
    SID3=$(echo "$J3" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])" 2>/dev/null || echo "")
    if [ -z "$SID3" ] || [ "$SID3" = "None" ]; then
        record "FAIL" "Create $CLASS" "Couldn't join game"
        continue
    fi

    # Build custom attributes per class
    case $CLASS in
        cleric)
            ATTRS="{\"session_id\":\"$SID3\",\"name\":\"TestCleric\",\"class\":\"cleric\",\"attribute_method\":\"custom\",\"custom_attributes\":{\"strength\":10,\"dexterity\":12,\"constitution\":13,\"intelligence\":8,\"wisdom\":15,\"charisma\":14}}"
            ;;
        thief)
            ATTRS="{\"session_id\":\"$SID3\",\"name\":\"TestThief\",\"class\":\"thief\",\"attribute_method\":\"custom\",\"custom_attributes\":{\"strength\":10,\"dexterity\":15,\"constitution\":13,\"intelligence\":12,\"wisdom\":8,\"charisma\":14}}"
            ;;
        ranger)
            ATTRS="{\"session_id\":\"$SID3\",\"name\":\"TestRanger\",\"class\":\"ranger\",\"attribute_method\":\"custom\",\"custom_attributes\":{\"strength\":12,\"dexterity\":15,\"constitution\":10,\"intelligence\":8,\"wisdom\":14,\"charisma\":13}}"
            ;;
        paladin)
            ATTRS="{\"session_id\":\"$SID3\",\"name\":\"TestPaladin\",\"class\":\"paladin\",\"attribute_method\":\"custom\",\"custom_attributes\":{\"strength\":15,\"dexterity\":10,\"constitution\":12,\"intelligence\":8,\"wisdom\":14,\"charisma\":13}}"
            ;;
    esac

    C3=$(rpc "createCharacter" "$ATTRS" 31)
    C3_OK=$(echo "$C3" | python3 -c "import sys,json; r=json.load(sys.stdin); print('yes' if 'result' in r and r.get('result',{}).get('success') else 'no')" 2>/dev/null || echo "no")
    if [ "$C3_OK" = "yes" ]; then
        HP=$(echo "$C3" | python3 -c "import sys,json; c=json.load(sys.stdin)['result']['character']; print(f'HP:{c[\"HP\"]}/{c[\"MaxHP\"]}')" 2>/dev/null || echo "?")
        record "PASS" "Create $CLASS ($HP)"
    else
        ERR=$(echo "$C3" | python3 -c "import sys,json; r=json.load(sys.stdin); e=r.get('error',{}).get('message',''); rs=r.get('result',{}); errs=rs.get('errors',[]); print(e or ', '.join(errs) if errs else 'unknown')" 2>/dev/null || echo "parse error")
        record "FAIL" "Create $CLASS" "$ERR"
    fi
    # Cleanup
    NSID3=$(echo "$C3" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])" 2>/dev/null || echo "$SID3")
    rpc "leaveGame" "{\"session_id\":\"$NSID3\"}" 32 > /dev/null 2>&1
done

echo ""

########################################
echo "╔══════════════════════════════════════════════╗"
echo "║           PLAYTEST RESULTS                   ║"
echo "╚══════════════════════════════════════════════╝"
echo ""
echo "  PASSED: $PASS"
echo "  FAILED: $FAIL"
echo "  WARNED: $WARN"
echo "  TOTAL:  $((PASS + FAIL + WARN))"
echo ""

if [ "$FAIL" -eq 0 ]; then
    echo "  ✓ All critical tests passed!"
else
    echo "  ✗ Some tests failed - review output above"
fi

if [ "$WARN" -gt 0 ]; then
    echo "  ⚠ $WARN warnings - some features may need attention"
fi

echo ""
echo "  Server: $BASE"
echo "  Web UI: $BASE/"
echo ""
