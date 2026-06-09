package dignity

import (
	"encoding/json"
	"fmt"
)

// ── Phase 5: Node Axis Convergence ─────────────────────────────────────────
//
// The lunar nodes are orbital points — the 180-degree axis is invariant by
// definition. The question is whether the sign-level interpretation survives
// the tropical → sidereal coordinate shift.
//
// Western: North Node (evolutionary direction), South Node (past mastery)
// Vedic:   Rahu (future desire), Ketu (past wisdom)
//
// MEANING CONVERGENCE: Western and Vedic agree on the fundamental architecture.

// NodeConvergence captures the node axis comparison across tropical and
// sidereal traditions.
type NodeConvergence struct {
	Name       string
	NNTropSign string
	NNSidSign  string
	SNTropSign string
	SNSidSign  string
	SignMatch  bool
	AxisAngle  float64
}

// AxisPreserved reports whether the 180-degree node opposition is intact.
func (n *NodeConvergence) AxisPreserved() bool {
	diff := n.AxisAngle - 180.0
	if diff < 0 {
		diff = -diff
	}
	return diff < 1.0
}

// MeaningConverges is always true — every surviving tradition agrees on the
// node axis as a soul-trajectory encoding.
func (n *NodeConvergence) MeaningConverges() bool { return true }

// ComputeNodeConvergence computes node axis convergence. nnLong is the
// tropical longitude of the North (Mean) Node. ayan is the Lahiri ayanamsa.
func ComputeNodeConvergence(nnLong, ayan float64, name string) *NodeConvergence {
	nnTrop := normalizeLon(nnLong)
	snTrop := normalizeLon(nnLong + 180.0)
	nnSid := normalizeLon(nnLong - ayan)
	snSid := normalizeLon(nnLong + 180.0 - ayan)

	return &NodeConvergence{
		Name:       name,
		NNTropSign: SignForLongitude(nnTrop),
		NNSidSign:  SignForLongitude(nnSid),
		SNTropSign: SignForLongitude(snTrop),
		SNSidSign:  SignForLongitude(snSid),
		SignMatch:  SignForLongitude(nnTrop) == SignForLongitude(nnSid),
		AxisAngle:  angleDist(nnTrop, snTrop),
	}
}

// FormatNodeConvergence formats a human-readable node convergence report.
func FormatNodeConvergence(n *NodeConvergence) string {
	status := "DEVIATION"
	if n.AxisPreserved() {
		status = "180.00 — preserved"
	}
	sig := "SIGNAL"
	if !n.SignMatch {
		sig = "NOISE (ayanamsa shift)"
	}

	return fmt.Sprintf(`Node Axis Convergence Report — %s

  North Node (Rahu):  trop %-12s sid %s
  South Node (Ketu):  trop %-12s sid %s
  Axis angle:         %.2f deg (%s)
  Sign agreement:     %s

`, n.Name, n.NNTropSign, n.NNSidSign, n.SNTropSign, n.SNSidSign,
		n.AxisAngle, status, sig) +
		axisInterpretation(n)
}

func axisInterpretation(n *NodeConvergence) string {
	pres := "AXIS BROKEN: Node opposition deviates from 180 degrees.\n\n"
	if n.AxisPreserved() {
		pres = "AXIS PRESERVED: The 180-degree opposition is invariant. " +
			"The node axis survives as a structural fact regardless " +
			"of coordinate system.\n\n"
	}
	return pres +
		"MEANING CONVERGENCE: Western (NN=evolution, SN=past) and " +
		"Vedic (Rahu=future desire, Ketu=past mastery) agree on " +
		"the fundamental architecture — a polar pair encoding " +
		"soul trajectory.\n\n" +
		"RECOVERY IMPLICATION: The node axis is almost certainly " +
		"original infrastructure. The ayanamsa question (which sign) " +
		"is the Phase 6 coordinate problem — but the axis itself is " +
		"signal, not noise.\n"
}

// NodeConvergenceJSON serializes the node report for the API.
func (n *NodeConvergence) NodeConvergenceJSON() ([]byte, error) {
	return json.MarshalIndent(map[string]interface{}{
		"name":              n.Name,
		"nn_trop_sign":      n.NNTropSign,
		"nn_sid_sign":       n.NNSidSign,
		"sn_trop_sign":      n.SNTropSign,
		"sn_sid_sign":       n.SNSidSign,
		"sign_match":        n.SignMatch,
		"axis_preserved":    n.AxisPreserved(),
		"meaning_converges": n.MeaningConverges(),
		"axis_angle":        n.AxisAngle,
	}, "", "  ")
}
