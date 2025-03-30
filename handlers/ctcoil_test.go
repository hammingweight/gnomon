package handlers

import "testing"

func TestUpperTriggerOnSoc(t *testing.T) {
	threshold := 50
	expected := 95
	actual := upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 30
	expected = 85
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 40
	expected = 85
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 55
	expected = 99
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 65
	expected = 99
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 73
	expected = 99
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 74
	expected = 101
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 85
	expected = 101
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 95
	expected = 101
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}
}
