// Package physics provides the core astrological physics types.
// During the Phase B refactor, these are type aliases to dignity types.
// Once implementations move, these become real types.
package physics

import "github.com/aj-nt/empirical/internal/dignity"

// ── Core types ───────────────────────────────────────────────────────────

// BaseChart holds all computed astrological positions for a single chart.
type BaseChart = dignity.BaseChart

// Position holds a planet's longitude, latitude, daily speed, and distance.
type Position = dignity.Position

// BirthData holds the input data for computing a natal chart.
type BirthData = dignity.BirthData

// PlanetID pairs a planet name with its Swiss Ephemeris ID.
type PlanetID = dignity.PlanetID

// AspectDef defines an aspect angle and its name.
type AspectDef = dignity.AspectDef

// ── System and Frame ─────────────────────────────────────────────────────

// System identifies an astrological tradition or interpretive system.
type System = dignity.System

// Frame identifies a coordinate reference frame for positions.
type Frame = dignity.Frame

// ── Planet sets ──────────────────────────────────────────────────────────

// AllPlanets is the full planet list (28 bodies).
var AllPlanets = dignity.AllPlanets

// ClassicalPlanets is the 7 classical planets.
var ClassicalPlanets = dignity.ClassicalPlanets

// ── Functions ────────────────────────────────────────────────────────────

// ComputeBaseChart computes all astrological positions for a birth chart.
var ComputeBaseChart = dignity.ComputeBaseChart

// TropicalToLonMap extracts a longitude-only map from tropical positions.
var TropicalToLonMap = dignity.TropicalToLonMap

// DefaultAspects returns the default Ptolemaic aspect set.
var DefaultAspects = dignity.DefaultAspects

// FindNatalAspects finds all aspects between planets.
var FindNatalAspects = dignity.FindNatalAspects

// FindStarConjunctions finds all star-planet conjunctions.
var FindStarConjunctions = dignity.FindStarConjunctions

// FindStarAspects finds all aspects between a star and planets.
var FindStarAspects = dignity.FindStarAspects

// DetectPatterns detects geometric patterns in a planet set.
var DetectPatterns = dignity.DetectPatterns

// ComputeParts computes all Arabic Parts.
var ComputeParts = dignity.ComputeParts

// ComputeGMST computes Greenwich Mean Sidereal Time.
var ComputeGMST = dignity.ComputeGMST

// NormalizeLon normalizes a longitude to [0, 360).
var NormalizeLon = dignity.NormalizeLon
