#!/bin/bash
# Comprehensive playtest of GoldBox RPG engine
set -e

BASE="http://localhost:8080"
COOKIES="/tmp/pt_cookies.txt"
rm -f "$COOKIES"

rpc() {
    local id=$1 method=$2 params=$3
    curl -s -b "$COOKIES" -c "$COOKIES" "$BASE/rpc" \
        -d "{\"jsonrpc\":\"2.0\",\"id\":$id,\"method\":\"$method\",\"params\":$params}"
}

echo "============================="
echo "  GoldBox RPG Playtest"
echo "============================="

# 1. Health check
echo -e "\n--- Health Check ---"
curl -s "$BASE/health" | python3 -c "import sys,json; d=json.load(sys.stdin); print('Status:', d['status'], '| Checks:', len(d['checks']))"

# 2. Join game
echo -e "\n--- Join Game ---"
J=$(rpc 1 joinGame '{"player_name":"Hero"}')
SID=$(echo "$J" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])")
echo "Session: $SID"

# 3. Create character
echo -e "\n--- Create Fighter ---"
C=$(rpc 2 createCharacter "{\"session_id\":\"$SID\",\"name\":\"Conan\",\"class\":\"fighter\",\"attribute_method\":\"standard\"}")
NSID=$(echo "$C" | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['session_id'])")
PID=$(echo "$C" | python3 -c "import sys,json; c=json.load(sys.stdin)['result']['character']; print(c['ID'])")
echo "$C" | python3 -c "
import sys,json
c=json.load(sys.stdin)['result']['character']
print(f'  Name: {c[\"Name\"]} | Class: Fighter | HP: {c[\"HP\"]}/{c[\"MaxHP\"]}')
print(f'  STR:{c[\"Strength\"]} DEX:{c[\"Dexterity\"]} CON:{c[\"Constitution\"]} INT:{c[\"Intelligence\"]} WIS:{c[\"Wisdom\"]} CHA:{c[\"Charisma\"]}')
print(f'  Position: ({c[\"Position\"][\"X\"]}, {c[\"Position\"][\"Y\"]})')
print(f'  Gold: {c[\"Gold\"]} | AC: {c[\"ArmorClass\"]} | THAC0: {c[\"THAC0\"]}')
"
echo "  Session: $NSID"
echo "  Player ID: $PID"

# 4. Movement test
echo -e "\n--- Movement Test ---"
echo -n "  Start: (5,5)"

for DIR in north east south west north north east east south south west west; do
    M=$(rpc 3 move "{\"session_id\":\"$NSID\",\"direction\":\"$DIR\"}")
    POS=$(echo "$M" | python3 -c "import sys,json; p=json.load(sys.stdin)['result']['position']; print(f'({p[\"X\"]},{p[\"Y\"]})')" 2>/dev/null || echo "FAIL")
    echo -n " -> $DIR: $POS"
done
echo ""

# 5. Get spells
echo -e "\n--- Spell List ---"
S=$(rpc 4 getAllSpells '{}')
echo "$S" | python3 -c "
import sys,json
spells=json.load(sys.stdin)['result']['spells']
for s in spells:
    dmg=f' ({s[\"DamageDice\"]} {s[\"DamageType\"]})' if s['DamageDice'] else ''
    heal=f' (heal {s[\"HealingDice\"]})' if s['HealingDice'] else ''
    print(f'  Lvl {s[\"Level\"]}: {s[\"Name\"]}{dmg}{heal} - Range:{s[\"Range\"]}')
"

# 6. Start combat
echo -e "\n--- Combat ---"
CB=$(rpc 5 startCombat "{\"session_id\":\"$NSID\",\"participant_ids\":[\"$PID\"]}")
echo "$CB" | python3 -c "
import sys,json
d=json.load(sys.stdin)
if 'result' in d:
    print(f'  Combat started! First turn: {d[\"result\"].get(\"first_turn\",\"?\")}')
else:
    print(f'  Error: {d[\"error\"][\"message\"]}')
"

# 7. Movement during combat (should use action points)
echo -e "\n--- Combat Movement ---"
M=$(rpc 6 move "{\"session_id\":\"$NSID\",\"direction\":\"north\"}")
echo "$M" | python3 -c "
import sys,json
d=json.load(sys.stdin)
if 'result' in d:
    p=d['result']['position']
    print(f'  Moved to ({p[\"X\"]},{p[\"Y\"]})')
else:
    print(f'  Error: {d[\"error\"][\"message\"]}')
"

# 8. End turn
echo -e "\n--- End Turn ---"
ET=$(rpc 7 endTurn "{\"session_id\":\"$NSID\"}")
echo "$ET" | python3 -c "
import sys,json
d=json.load(sys.stdin)
if 'result' in d:
    r=d['result']
    print(f'  Next turn: {r.get(\"next_turn\",\"?\")} | Round: {r.get(\"current_round\",\"?\")}')
else:
    print(f'  Error: {d[\"error\"][\"message\"]}')
"

# 9. Get equipment
echo -e "\n--- Equipment ---"
EQ=$(rpc 8 getEquipment "{\"session_id\":\"$NSID\"}")
echo "$EQ" | python3 -c "
import sys,json
d=json.load(sys.stdin)
if 'result' in d:
    eq=d['result'].get('equipment',{})
    inv=d['result'].get('inventory',[])
    print(f'  Equipped: {len(eq)} items | Inventory: {len(inv)} items')
else:
    print(f'  Error: {d[\"error\"][\"message\"]}')
"

echo -e "\n============================="
echo "  Playtest Complete!"
echo "============================="
