package match_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/synth"
)

func TestEngine_Determinism(t *testing.T) {
	engine := match.NewEngine(match.DefaultConfig())
	if !harness.CheckDeterminism(engine, synth.AllExtended()) {
		t.Fatal("matching engine output must be deterministic")
	}
}
