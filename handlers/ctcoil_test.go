package handlers

import "testing"

func TestUpperTriggerOnSoc(t *testing.T) {
	threshold := 50
	expected := 90
	actual := upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 30
	expected = 80
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 40
	expected = 80
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 60
	expected = 99
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 70
	expected = 99
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 78
	expected = 99
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 79
	expected = 101
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 80
	expected = 101
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}

	threshold = 90
	expected = 101
	actual = upperTriggerOnSoc(threshold)
	if actual != expected {
		t.Errorf("expected %d, got %d", expected, actual)
	}
}
