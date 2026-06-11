package service

import "testing"

func TestPageNormalizeDefaultsZeroLimit(t *testing.T) {
	p := Page{}.normalize()

	if p.Limit != defaultPageLimit {
		t.Errorf("expected default limit %d, got %d", defaultPageLimit, p.Limit)
	}
}

func TestPageNormalizeCapsOversizedLimit(t *testing.T) {
	p := Page{Limit: 100000}.normalize()

	if p.Limit != maxPageLimit {
		t.Errorf("expected limit capped at %d, got %d", maxPageLimit, p.Limit)
	}
}

func TestPageNormalizeKeepsValidLimit(t *testing.T) {
	p := Page{Limit: 50}.normalize()

	if p.Limit != 50 {
		t.Errorf("expected limit 50 untouched, got %d", p.Limit)
	}
}

func TestPageNormalizeResetsNegativeOffset(t *testing.T) {
	p := Page{Offset: -10}.normalize()

	if p.Offset != 0 {
		t.Errorf("expected negative offset reset to 0, got %d", p.Offset)
	}
}
