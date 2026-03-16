#!/usr/bin/env bash
# create-legacy-symlinks.sh — Create backward-compatible portrait symlinks
# Maps old naming convention to new expression+tone naming:
#   portrait_{class}_{race}_male.png   → portrait_{class}_{race}_a_medium.png
#   portrait_{class}_{race}_female.png → portrait_{class}_{race}_b_medium.png
set -euo pipefail

PORTRAIT_DIR="${1:-./web/static/assets/sprites/characters}"

CLASSES=("fighter" "mage" "cleric" "thief" "ranger" "paladin")
RACES=("human" "elf" "dwarf" "halfling")

created=0

for class in "${CLASSES[@]}"; do
  for race in "${RACES[@]}"; do
    male_new="${PORTRAIT_DIR}/portrait_${class}_${race}_a_medium.png"
    female_new="${PORTRAIT_DIR}/portrait_${class}_${race}_b_medium.png"
    male_legacy="${PORTRAIT_DIR}/portrait_${class}_${race}_male.png"
    female_legacy="${PORTRAIT_DIR}/portrait_${class}_${race}_female.png"

    if [[ -f "$male_new" ]] && [[ ! -e "$male_legacy" ]]; then
      ln -s "$(basename "$male_new")" "$male_legacy"
      ((created++))
    fi
    if [[ -f "$female_new" ]] && [[ ! -e "$female_legacy" ]]; then
      ln -s "$(basename "$female_new")" "$female_legacy"
      ((created++))
    fi
  done
done

echo "Created ${created} legacy symlinks in ${PORTRAIT_DIR}"
