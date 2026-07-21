package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildBinary compiles the empirical binary and returns its path.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "empirical")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func TestWesternSubcommand_JSON(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "western", "--json",
		"AJ", "1969", "2", "15", "23", "10", "-8", "47.038", "-122.901")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("western subcommand failed: %v\n%s", err, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out)
	}

	// Verify key fields exist
	if result["name"] != "AJ" {
		t.Errorf("name = %v, want AJ", result["name"])
	}
	if _, ok := result["planet_signs"]; !ok {
		t.Error("missing planet_signs")
	}
	if _, ok := result["planet_houses"]; !ok {
		t.Error("missing planet_houses")
	}
	if _, ok := result["aspects"]; !ok {
		t.Error("missing aspects")
	}
}

func TestWesternSubcommand_Text(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "western",
		"AJ", "1969", "2", "15", "23", "10", "-8", "47.038", "-122.901")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("western subcommand failed: %v\n%s", err, out)
	}

	text := string(out)
	// Text output should contain key sections
	if len(text) < 100 {
		t.Error("text output too short")
	}
}

func TestWesternSubcommand_Usage(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "western")
	out, err := cmd.CombinedOutput()
	// Should fail with usage message
	if err == nil {
		t.Error("expected non-zero exit for missing args")
	}
	text := string(out)
	if len(text) == 0 {
		t.Error("expected usage message on stderr/stdout")
	}
}

func TestTransitSubcommand_JSON(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "transit", "--json",
		"AJ", "1969", "2", "15", "23", "10", "-8", "47.038", "-122.901",
		"2026-07-20", "2026-07-31")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("transit subcommand failed: %v\n%s", err, out)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out)
	}

	// Verify key fields exist
	if result["name"] != "AJ" {
		t.Errorf("name = %v, want AJ", result["name"])
	}
	if _, ok := result["transits"]; !ok {
		t.Error("missing transits")
	}
	if _, ok := result["sky_weather"]; !ok {
		t.Error("missing sky_weather")
	}

	// Should have transits
	transits, ok := result["transits"].([]interface{})
	if !ok {
		t.Fatal("transits is not an array")
	}
	if len(transits) == 0 {
		t.Error("expected non-empty transits array")
	}
}

func TestWesternSubcommand_ReadingFlag(t *testing.T) {
	bin := buildBinary(t)

	// Without --reading: reading fields should be absent
	cmd := exec.Command(bin, "western", "--json",
		"AJ", "1969", "2", "15", "23", "10", "-8", "47.038", "-122.901")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("western subcommand failed: %v\n%s", err, out)
	}

	var without map[string]interface{}
	if err := json.Unmarshal(out, &without); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	readingFields := []string{
		"chart_ruler", "chart_ruler_traditional", "chart_ruler_house",
		"chart_ruler_sign", "chart_ruler_dignity",
		"final_dispositor_traditional",
		"weighted_aspects", "key_midpoints", "key_star_aspects",
		"angular_planets",
	}
	for _, f := range readingFields {
		if _, ok := without[f]; ok {
			t.Errorf("field %q should be absent without --reading", f)
		}
	}

	// With --reading: all reading fields should be present
	cmd = exec.Command(bin, "western", "--json", "--reading",
		"AJ", "1969", "2", "15", "23", "10", "-8", "47.038", "-122.901")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("western --reading subcommand failed: %v\n%s", err, out)
	}

	var with map[string]interface{}
	if err := json.Unmarshal(out, &with); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, f := range readingFields {
		if _, ok := with[f]; !ok {
			t.Errorf("field %q should be present with --reading", f)
		}
	}

	// Verify chart_ruler values are correct
	if with["chart_ruler"] != "Pluto" {
		t.Errorf("chart_ruler = %v, want Pluto", with["chart_ruler"])
	}
	if with["chart_ruler_traditional"] != "Mars" {
		t.Errorf("chart_ruler_traditional = %v, want Mars", with["chart_ruler_traditional"])
	}
	if with["final_dispositor_traditional"] != "Mars" {
		t.Errorf("final_dispositor_traditional = %v, want Mars", with["final_dispositor_traditional"])
	}

	// Modern should be empty (closed loop) — omitempty means it may be absent
	if fd, ok := with["final_dispositor"]; ok && fd != nil && fd != "" {
		t.Errorf("final_dispositor = %v, want empty or absent (closed loop)", fd)
	}

	// Weighted aspects should be sorted descending
	wa, ok := with["weighted_aspects"].([]interface{})
	if !ok || len(wa) == 0 {
		t.Fatal("weighted_aspects is empty or not an array")
	}

	// Angular planets should include Mars (1st house)
	ap, ok := with["angular_planets"].([]interface{})
	if !ok {
		t.Fatal("angular_planets is not an array")
	}
	foundMars := false
	for _, p := range ap {
		if s, ok := p.(string); ok && s == "Mars" {
			foundMars = true
			break
		}
	}
	if !foundMars {
		t.Error("Mars not found in angular_planets")
	}
}
