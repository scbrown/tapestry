package dolt

import "testing"

func TestEpicProgress_Zero(t *testing.T) {
	p := EpicProgress{}
	if p.Total != 0 || p.Closed != 0 {
		t.Errorf("zero EpicProgress = %+v, want zero values", p)
	}
}

func TestEpicProgress_Partial(t *testing.T) {
	p := EpicProgress{Total: 5, Closed: 3}
	wantPct := 60
	gotPct := 0
	if p.Total > 0 {
		gotPct = p.Closed * 100 / p.Total
	}
	if gotPct != wantPct {
		t.Errorf("progress = %d%%, want %d%%", gotPct, wantPct)
	}
}

func TestEpicProgress_Complete(t *testing.T) {
	p := EpicProgress{Total: 4, Closed: 4}
	wantPct := 100
	gotPct := p.Closed * 100 / p.Total
	if gotPct != wantPct {
		t.Errorf("progress = %d%%, want %d%%", gotPct, wantPct)
	}
}
