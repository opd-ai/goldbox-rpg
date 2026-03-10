#!/bin/bash
# GoldBox RPG Engine - Interactive Playtest Script
# Tests the full gameplay loop via JSON-RPC HTTP POST

set -e

BASE_URL="http://localhost:8080"
COOKIE_JAR="/tmp/goldbox-playtest-cookies.txt"
rm -f "$COOKIE_JAR"

rpc() {
    local method="$1"
    local params="$2"
    local id="$3"
    curl -s -b "$COOKIE_JAR" -c "$COOKIE_JAR" -X POST "$BASE_URL/" \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"$method\",\"params\":$params,\"id\":$id}"
}

echo "======================================"
echo "  GoldBox RPG Engine - Playtest"
echo "======================================"
echo ""

# 1. Health check
echo "--- Health Check ---"
HEALTH=$(curl -s "$BASE_URL/health")
STATUS=$(echo "$HEALTH" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])")
echo "Server status: $STATUS"
if [ "$STATUS" != "healthy" ]; then
    echo "ERROR: Server not healthy!"
    exit 1
fi
echo ""

# 2. Join game
echo "--- 1. Join Game ---"
JOIN_RESULT=$(rpc "joinGame" '{"player_name":"Aldric the Bold"}' 1)
echo "$JOIN_RESULT" | python3 -m json.tool
SID=$(echo "$JOIN_RESULT" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])")
echo "Session ID: $SID"
echo ""

# 3. Create character (Fighter with standard array)
echo "--- 2. Create Character (Fighter) ---"
CREATE_RESULT=$(rpc "createCharacter" "{\"session_id\":\"$SID\",\"name\":\"Aldric\",\"class\":\"fighter\",\"attribute_method\":\"standard\"}" 2)
echo "$CREATE_RESULT" | python3 -c "
import sys, json
r = json.load(sys.stdin)['result']
c = r['character']
print(f\"  Name: {c['Name']}\")
print(f\"  Class: Fighter (ID: {c['Class']})\")
print(f\"  HP: {c['HP']}/{c['MaxHP']}\")
print(f\"  AC: {c['ArmorClass']}, THAC0: {c['THAC0']}\")
print(f\"  STR: {c['Strength']}, DEX: {c['Dexterity']}, CON: {c['Constitution']}\")
print(f\"  INT: {c['Intelligence']}, WIS: {c['Wisdom']}, CHA: {c['Charisma']}\")
print(f\"  AP: {c['ActionPoints']}/{c['MaxActionPoints']}\")
print(f\"  Gold: {c['Gold']}\")
print(f\"  Position: ({c['Position']['X']}, {c['Position']['Y']})\")
"
# Capture the session from createCharacter (may differ)
NEW_SID=$(echo "$CREATE_RESULT" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])")
if [ -n "$NEW_SID" ] && [ "$NEW_SID" != "None" ] && [ "$NEW_SID" != "null" ]; then
    SID="$NEW_SID"
fi
echo ""

# 4. Get game state
echo "--- 3. Get Game State ---"
STATE_RESULT=$(rpc "getGameState" "{\"session_id\":\"$SID\"}" 3)
echo "$STATE_RESULT" | python3 -c "
import sys, json
r = json.load(sys.stdin)['result']
print(f\"  Version: {r.get('version', 'N/A')}\")
print(f\"  Sessions: {len(r.get('sessions', {}))}\")
w = r.get('world', {})
if isinstance(w, dict):
    print(f\"  World keys: {list(w.keys())[:5]}\")
print('  Game state OK')
"
echo ""

# 5. Movement sequence
echo "--- 4. Movement Test ---"
for dir in north east south west east north north east; do
    MOVE_RESULT=$(rpc "move" "{\"session_id\":\"$SID\",\"direction\":\"$dir\"}" 4)
    SUCCESS=$(echo "$MOVE_RESULT" | python3 -c "
import sys, json
r = json.load(sys.stdin)
if 'result' in r:
    res = r['result']
    pos = res.get('position', {})
    print(f'Moved {\"$dir\":8s} -> ({pos.get(\"X\",\"?\")}, {pos.get(\"Y\",\"?\")})')
elif 'error' in r:
    print(f'Move {\"$dir\":8s} FAILED: {r[\"error\"][\"message\"]}')
" 2>/dev/null || echo "FAILED: parse error")
    echo "  $SUCCESS"
done
echo ""

# 6. Get game state after moving
echo "--- 5. Final State ---"
FINAL_STATE=$(rpc "getGameState" "{\"session_id\":\"$SID\"}" 10)
echo "$FINAL_STATE" | python3 -c "
import sys, json
r = json.load(sys.stdin)
if 'result' in r:
    print('  Game state retrieved successfully')
    print(f\"  Version: {r['result'].get('version', 'N/A')}\")
elif 'error' in r:
    print(f\"  Error: {r['error']['message']}\")
"
echo ""

# 7. Get all spells
echo "--- 6. Spell System ---"
SPELLS=$(rpc "getAllSpells" "{}" 11)
echo "$SPELLS" | python3 -c "
import sys, json
r = json.load(sys.stdin)
if 'result' in r:
    res = r['result']
    if isinstance(res, dict) and 'spells' in res:
        spells = res['spells']
        print(f'  Total spells: {len(spells)}')
        for s in spells[:5]:
            if isinstance(s, dict):
                print(f'    - {s.get(\"name\", s.get(\"spell_id\", \"?\"))}')
    elif isinstance(res, list):
        print(f'  Total spells: {len(res)}')
    else:
        print(f'  Spells data: {type(res).__name__}')
elif 'error' in r:
    print(f'  Error: {r[\"error\"][\"message\"]}')
"
echo ""

# 8. End turn
echo "--- 7. End Turn ---"
END_RESULT=$(rpc "endTurn" "{\"session_id\":\"$SID\"}" 12)
echo "$END_RESULT" | python3 -c "
import sys, json
r = json.load(sys.stdin)
if 'result' in r:
    print(f'  End turn: {r[\"result\"]}')
elif 'error' in r:
    print(f'  End turn: {r[\"error\"][\"message\"]}')
"
echo ""

echo "======================================"
echo "  Playtest Complete!"
echo "======================================"
echo ""
echo "WASM UI available at: $BASE_URL/wasm.html"
echo "Classic UI available at: $BASE_URL/"
