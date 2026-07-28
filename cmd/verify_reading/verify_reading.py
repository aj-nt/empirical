#!/usr/bin/env python3
"""
verify_reading.py — Cross-check factual claims in an astrological reading
against the empirical engine.

Usage:
    # Verify claims from stdin against a chart
    echo "ASPECT: Mars trine Saturn 4.5
PATTERN: Grand Trine Mars Saturn Sun
DISPOSITOR: Pluto -> Mercury
ELEMENT: Aquarius earth
SIGN: Sun Aquarius
HOUSE: Mars 1
DIGNITY: Venus fall" | python3 verify_reading.py AJ 1969 2 15 23 10 -8 47.038 -122.901

    # Verify claims from a file
    python3 verify_reading.py AJ 1969 2 15 23 10 -8 47.038 -122.901 claims.txt

    # Just dump the fact sheet (no claims to verify)
    python3 verify_reading.py AJ 1969 2 15 23 10 -8 47.038 -122.901

Claim format (one per line):
    ASPECT: <planet1> <aspect_type> <planet2> [<orb>]
    PATTERN: <pattern_type> <planet1> <planet2> ...
    DISPOSITOR: <planet> -> <final_dispositor>
    ELEMENT: <sign> <element>
    SIGN: <planet> <sign>
    HOUSE: <planet> <house_number>
    DIGNITY: <planet> <dignity>
    MODALITY: <sign> <modality>
    STELLIUM: <sign> <planet1> <planet2> ...
    CHART_RULER: <planet>
    FINAL_DISPOSITOR: <planet>
    GRAND_TRINE: <planet1> <planet2> <planet3>
    YOD: <planet1> <planet2> <planet3>
    T_SQUARE: <planet1> <planet2> <planet3>
    CRADLE: <planet1> <planet2> <planet3> <planet4>
"""

import json
import subprocess
import sys
import os
from pathlib import Path

RECOVER = os.path.expanduser("~/Documents/repos/empirical/recover")
EMPIRICAL = os.path.expanduser("~/Documents/repos/empirical/empirical")

# ── Element/Modality tables ──────────────────────────────────────────
ELEMENTS = {
    "Aries": "fire", "Leo": "fire", "Sagittarius": "fire",
    "Taurus": "earth", "Virgo": "earth", "Capricorn": "earth",
    "Gemini": "air", "Libra": "air", "Aquarius": "air",
    "Cancer": "water", "Scorpio": "water", "Pisces": "water",
}

MODALITIES = {
    "Aries": "cardinal", "Cancer": "cardinal", "Libra": "cardinal", "Capricorn": "cardinal",
    "Taurus": "fixed", "Leo": "fixed", "Scorpio": "fixed", "Aquarius": "fixed",
    "Gemini": "mutable", "Virgo": "mutable", "Sagittarius": "mutable", "Pisces": "mutable",
}

ASPECT_ANGLES = {
    "conjunction": 0, "conjunct": 0, "conj": 0,
    "opposition": 180, "opposite": 180, "opp": 180,
    "trine": 120, "square": 90, "sextile": 60,
    "semi-sextile": 30, "semi_sextile": 30, "semisextile": 30,
    "quincunx": 150, "inconjunct": 150,
    "semi-square": 45, "semi_square": 45, "semisquare": 45,
    "sesquiquadrate": 135, "sesqui": 135,
    "quintile": 72, "biquintile": 144,
    "septile": 51.4, "biseptile": 102.9, "triseptile": 154.3,
}

# ── Engine queries ────────────────────────────────────────────────────
def run_recover(args):
    """Run the recover binary and return parsed JSON."""
    cmd = [RECOVER] + args
    result = subprocess.run(cmd, capture_output=True, text=True)
    stdout = result.stdout.strip()
    if not stdout:
        return {}
    try:
        return json.loads(stdout)
    except json.JSONDecodeError:
        for line in stdout.split('\n'):
            line = line.strip()
            if line.startswith('{'):
                try:
                    return json.loads(line)
                except:
                    continue
        return {}

def run_empirical(args):
    """Run the empirical binary and return parsed JSON."""
    cmd = [EMPIRICAL] + args
    result = subprocess.run(cmd, capture_output=True, text=True)
    stdout = result.stdout.strip()
    if not stdout:
        return {}
    try:
        return json.loads(stdout)
    except json.JSONDecodeError:
        for line in stdout.split('\n'):
            line = line.strip()
            if line.startswith('{'):
                try:
                    return json.loads(line)
                except:
                    continue
        return {}

def get_ground_truth(name, year, month, day, hour, minute, tz_off, lat, lng):
    """Query the engine for all verifiable facts about a chart."""
    args = [
        "--json", name,
        str(year), str(month), str(day), str(hour), str(minute),
        str(tz_off), str(lat), str(lng),
    ]
    data = run_recover(args)

    # Also get western reading for patterns, dispositors, etc.
    western_args = [
        "western", "--json", "--reading", name,
        str(year), str(month), str(day), str(hour), str(minute),
        str(tz_off), str(lat), str(lng),
    ]
    western = run_empirical(western_args)

    return data, western


def build_fact_sheet(data, western):
    """Extract all verifiable facts from engine output."""
    facts = {}

    # ── Planet signs ──
    phase1 = data.get("phase1_dignity", {})
    planets_data = phase1.get("Planets", [])
    facts["signs"] = {}
    facts["dignities"] = {}
    for p in planets_data:
        name = p.get("Planet", "")
        sign = p.get("TropSign", "")
        western_dig = p.get("Western", "")
        facts["signs"][name] = sign
        facts["dignities"][name] = western_dig

    # ── Houses ──
    phase3 = data.get("phase3_houses", {})
    house_planets = phase3.get("Planets", [])
    facts["houses"] = {}
    for p in house_planets:
        name = p.get("Planet", "")
        placements = p.get("Placements", {})
        # Use whole_sign as canonical
        facts["houses"][name] = placements.get("whole_sign", "?")

    # ── Aspects ──
    aspects_raw = western.get("aspects", [])
    facts["aspects"] = []
    for a in aspects_raw:
        # Parse "Mars trine Saturn (orb 4.5°)" format
        text = a if isinstance(a, str) else str(a)
        facts["aspects"].append(text)

    # ── Patterns ──
    patterns_raw = western.get("patterns", [])
    facts["patterns"] = []
    for p in patterns_raw:
        text = p if isinstance(p, str) else str(p)
        facts["patterns"].append(text)

    # ── Dispositor trees ──
    trees = western.get("dispositor_trees", {})
    facts["dispositors"] = {}
    for planet, chain in trees.items():
        if chain and isinstance(chain, list):
            # Chains may loop — detect the loop and report it
            # Format: "Mars in Scorpio→Pluto in Virgo"
            seen = set()
            final = None
            first_source = None
            for step in chain:
                if "→" in step:
                    source = step.split("→")[0].strip().split()[0]
                    target = step.split("→")[1].strip().split()[0]
                    if first_source is None:
                        first_source = source
                    if target in seen:
                        final = f"LOOP({target})"
                        break
                    seen.add(target)
                    final = target
            # Also check: does the last target loop back to the first source?
            if final and not final.startswith("LOOP") and first_source and final == first_source:
                final = f"LOOP({final})"
            if final:
                facts["dispositors"][planet] = final

    # ── Chart ruler ──
    facts["chart_ruler"] = western.get("chart_ruler", "?")

    # ── Final dispositor (from chart_ruler chain or first planet's chain) ──
    facts["final_dispositor"] = western.get("final_dispositor", "?")

    # ── Rulership chains ──
    chains = western.get("rulership_chains", {})
    facts["rulership_chains"] = chains

    # ── Element/modality balance ──
    facts["element_balance"] = western.get("element_balance", {})
    facts["modality_balance"] = western.get("modality_balance", {})

    # ── Planet signs from western (includes outer planets) ──
    planet_signs = western.get("planet_signs", [])
    for ps in planet_signs:
        if isinstance(ps, str):
            # "Sun in Virgo: ..."
            parts = ps.split(" in ", 1)
            if len(parts) == 2:
                planet = parts[0].strip()
                sign = parts[1].split(":")[0].strip()
                if planet not in facts["signs"]:
                    facts["signs"][planet] = sign

    return facts


# ── Claim verification ────────────────────────────────────────────────
def verify_claim(claim_line, facts):
    """Verify a single claim. Returns (pass/fail, message)."""
    line = claim_line.strip()
    if not line or line.startswith("#"):
        return None, None

    parts = line.split(":", 1)
    if len(parts) != 2:
        return "ERROR", f"Malformed claim: {line}"

    claim_type = parts[0].strip().upper()
    claim_body = parts[1].strip()

    if claim_type == "ASPECT":
        return verify_aspect(claim_body, facts)
    elif claim_type == "PATTERN":
        return verify_pattern(claim_body, facts)
    elif claim_type == "DISPOSITOR":
        return verify_dispositor(claim_body, facts)
    elif claim_type == "ELEMENT":
        return verify_element(claim_body, facts)
    elif claim_type == "SIGN":
        return verify_sign(claim_body, facts)
    elif claim_type == "HOUSE":
        return verify_house(claim_body, facts)
    elif claim_type == "DIGNITY":
        return verify_dignity(claim_body, facts)
    elif claim_type == "MODALITY":
        return verify_modality(claim_body, facts)
    elif claim_type == "STELLIUM":
        return verify_stellium(claim_body, facts)
    elif claim_type == "CHART_RULER":
        return verify_chart_ruler(claim_body, facts)
    elif claim_type == "FINAL_DISPOSITOR":
        return verify_final_dispositor(claim_body, facts)
    elif claim_type in ("GRAND_TRINE", "YOD", "T_SQUARE", "CRADLE", "KITE", "GRAND_CROSS"):
        return verify_pattern(claim_body, facts)
    else:
        return "ERROR", f"Unknown claim type: {claim_type}"


def verify_aspect(body, facts):
    """Verify an aspect claim like 'Mars trine Saturn 4.5'"""
    tokens = body.split()
    if len(tokens) < 3:
        return "ERROR", f"Malformed aspect claim: {body}"

    p1 = tokens[0]
    aspect_type = tokens[1].lower()
    p2 = tokens[2]
    claimed_orb = float(tokens[3]) if len(tokens) > 3 else None

    # Normalize aspect type
    expected_angle = ASPECT_ANGLES.get(aspect_type)
    if expected_angle is None:
        return "ERROR", f"Unknown aspect type: {aspect_type}"

    # Search aspect list
    for a_text in facts["aspects"]:
        # "Mars trine Saturn (orb 4.5°)"
        a_lower = a_text.lower()
        if p1.lower() in a_lower and p2.lower() in a_lower and aspect_type in a_lower:
            # Extract orb
            if "(orb " in a_lower:
                orb_str = a_lower.split("(orb ")[1].split("°")[0].split(")")[0]
                try:
                    actual_orb = float(orb_str)
                except:
                    actual_orb = None

                if claimed_orb is not None and actual_orb is not None:
                    if abs(claimed_orb - actual_orb) < 0.2:
                        return "PASS", f"{body} — orb matches ({actual_orb}°)"
                    else:
                        return "FAIL", f"{body} — orb mismatch: claimed {claimed_orb}°, actual {actual_orb}°"
                return "PASS", f"{body} — aspect found (orb {actual_orb}°)"

    return "FAIL", f"{body} — aspect not found in chart"


def verify_pattern(body, facts):
    """Verify a pattern claim like 'Grand Trine Mars Saturn Sun'"""
    body_lower = body.lower()
    tokens = body_lower.split()

    # Try to match against pattern list
    for p_text in facts["patterns"]:
        p_lower = p_text.lower()
        # Check if all named planets appear in the pattern
        all_found = True
        for token in tokens:
            if token in ("grand", "trine", "yod", "t-square", "t_square", "cradle",
                         "kite", "cross", "finger", "god", "of", "involving", "stellium"):
                continue
            if token not in p_lower:
                all_found = False
                break
        if all_found:
            return "PASS", f"{body} — pattern found: {p_text[:80]}"

    return "FAIL", f"{body} — pattern not found in chart"


def verify_dispositor(body, facts):
    """Verify a dispositor claim like 'Pluto -> Mercury'"""
    parts = body.split("->")
    if len(parts) != 2:
        return "ERROR", f"Malformed dispositor claim: {body}"

    planet = parts[0].strip()
    claimed_final = parts[1].strip()

    actual_final = facts["dispositors"].get(planet)
    if actual_final is None:
        return "FAIL", f"{body} — planet {planet} not in dispositor tree"

    if actual_final.lower() == claimed_final.lower():
        return "PASS", f"{body} — correct"
    else:
        return "FAIL", f"{body} — actual final dispositor for {planet} is {actual_final}"


def verify_element(body, facts):
    """Verify an element claim like 'Aquarius earth'"""
    parts = body.split()
    if len(parts) != 2:
        return "ERROR", f"Malformed element claim: {body}"

    sign = parts[0]
    claimed_element = parts[1].lower()

    actual_element = ELEMENTS.get(sign, "").lower()
    if not actual_element:
        return "ERROR", f"Unknown sign: {sign}"

    if actual_element == claimed_element:
        return "PASS", f"{body} — correct"
    else:
        return "FAIL", f"{body} — {sign} is {actual_element}, not {claimed_element}"


def verify_sign(body, facts):
    """Verify a sign claim like 'Sun Aquarius'"""
    parts = body.split()
    if len(parts) < 2:
        return "ERROR", f"Malformed sign claim: {body}"

    planet = parts[0]
    claimed_sign = " ".join(parts[1:])

    actual_sign = facts["signs"].get(planet)
    if actual_sign is None:
        return "FAIL", f"{body} — planet {planet} not found in chart"

    if actual_sign.lower() == claimed_sign.lower():
        return "PASS", f"{body} — correct"
    else:
        return "FAIL", f"{body} — {planet} is in {actual_sign}, not {claimed_sign}"


def verify_house(body, facts):
    """Verify a house claim like 'Mars 1'"""
    parts = body.split()
    if len(parts) != 2:
        return "ERROR", f"Malformed house claim: {body}"

    planet = parts[0]
    try:
        claimed_house = int(parts[1])
    except ValueError:
        return "ERROR", f"Invalid house number: {parts[1]}"

    actual_house = facts["houses"].get(planet)
    if actual_house is None or actual_house == "?":
        return "FAIL", f"{body} — planet {planet} house unknown"

    if actual_house == claimed_house:
        return "PASS", f"{body} — correct"
    else:
        return "FAIL", f"{body} — {planet} is in house {actual_house}, not {claimed_house}"


def verify_dignity(body, facts):
    """Verify a dignity claim like 'Venus fall'"""
    parts = body.split()
    if len(parts) != 2:
        return "ERROR", f"Malformed dignity claim: {body}"

    planet = parts[0]
    claimed_dignity = parts[1].lower()

    actual_dignity = facts["dignities"].get(planet, "").lower()
    if not actual_dignity:
        return "FAIL", f"{body} — planet {planet} not found"

    if actual_dignity == claimed_dignity:
        return "PASS", f"{body} — correct"
    else:
        return "FAIL", f"{body} — {planet} is {actual_dignity}, not {claimed_dignity}"


def verify_modality(body, facts):
    """Verify a modality claim like 'Aquarius fixed'"""
    parts = body.split()
    if len(parts) != 2:
        return "ERROR", f"Malformed modality claim: {body}"

    sign = parts[0]
    claimed_modality = parts[1].lower()

    actual_modality = MODALITIES.get(sign, "").lower()
    if not actual_modality:
        return "ERROR", f"Unknown sign: {sign}"

    if actual_modality == claimed_modality:
        return "PASS", f"{body} — correct"
    else:
        return "FAIL", f"{body} — {sign} is {actual_modality}, not {claimed_modality}"


def verify_stellium(body, facts):
    """Verify a stellium claim like 'Aries Venus Saturn Node Chiron Eris'"""
    return verify_pattern(f"Stellium {body}", facts)


def verify_chart_ruler(body, facts):
    """Verify a chart ruler claim like 'Neptune'"""
    claimed = body.strip()
    actual = facts.get("chart_ruler", "?")
    if actual.lower() == claimed.lower():
        return "PASS", f"Chart ruler: {claimed} — correct"
    else:
        return "FAIL", f"Chart ruler: claimed {claimed}, actual {actual}"


def verify_final_dispositor(body, facts):
    """Verify a final dispositor claim like 'Mercury'"""
    claimed = body.strip()
    actual = facts.get("final_dispositor", "?")
    if actual.lower() == claimed.lower():
        return "PASS", f"Final dispositor: {claimed} — correct"
    else:
        return "FAIL", f"Final dispositor: claimed {claimed}, actual {actual}"


# ── Main ──────────────────────────────────────────────────────────────
def main():
    if len(sys.argv) < 9:
        print(__doc__)
        sys.exit(1)

    name = sys.argv[1]
    year, month, day = int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4])
    hour, minute = int(sys.argv[5]), int(sys.argv[6])
    tz_off = float(sys.argv[7])
    lat, lng = float(sys.argv[8]), float(sys.argv[9])

    claims_file = sys.argv[10] if len(sys.argv) > 10 else None

    # Query engine
    print(f"Querying engine for {name}...", file=sys.stderr)
    data, western = get_ground_truth(name, year, month, day, hour, minute, tz_off, lat, lng)
    facts = build_fact_sheet(data, western)

    # Read claims
    if claims_file:
        with open(claims_file) as f:
            claims = f.readlines()
    else:
        claims = sys.stdin.readlines()

    if not claims or all(not c.strip() or c.strip().startswith("#") for c in claims):
        # No claims — just dump fact sheet
        print(json.dumps(facts, indent=2))
        return

    # Verify
    results = []
    for claim in claims:
        result = verify_claim(claim, facts)
        if result[0] is not None:
            results.append(result)

    # Report
    passes = sum(1 for r in results if r[0] == "PASS")
    fails = sum(1 for r in results if r[0] == "FAIL")
    errors = sum(1 for r in results if r[0] == "ERROR")

    print()
    print("=" * 60)
    print(f"VERIFICATION: {passes} PASS, {fails} FAIL, {errors} ERROR")
    print("=" * 60)

    for status, msg in results:
        symbol = {"PASS": "✓", "FAIL": "✗", "ERROR": "!"}[status]
        print(f"  {symbol} [{status}] {msg}")

    if fails > 0:
        print(f"\n⚠️  {fails} factual errors found. Fix before presenting.")
        sys.exit(1)
    else:
        print(f"\n✓  All claims verified.")


if __name__ == "__main__":
    main()
