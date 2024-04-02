package main

import (
	"strconv"
	"testing"
)

// Tests all functions found in location.go
// As of 4/1 - all functions are covered
func TestMilesToDegLatitude(t *testing.T) {

	response := strconv.FormatFloat(milesToDegLatitude(10.5, 30.62), 'f', -6, 64)
	expected := strconv.FormatFloat(0.1524234288179642, 'f', -6, 64)
	if response != expected {
		t.Errorf("Result was incorrect, got: %s, want: %s", response, expected)

	}
}

func TestMilesToDegLongitude(t *testing.T) {

	response := strconv.FormatFloat(milesToDegLongitude(10.5, 30.62), 'f', -6, 64)
	expected := strconv.FormatFloat(0.17624069777397156, 'f', -6, 64)
	if response != expected {
		t.Errorf("Result was incorrect, got: %s, want: %s", response, expected)

	}
}

func TestGetUserBoundingBox(t *testing.T) {
	loc := Location{
		Latitude:  37.7749,
		Longitude: -122.4194,
	}
	responseLeft, responseBottom, responseRight, responseTop := getUserBoundingBox(10.5, loc)
	responseArray := [4]string{strconv.FormatFloat(responseLeft, 'f', -5, 32), strconv.FormatFloat(responseBottom, 'f', -6, 32), strconv.FormatFloat(responseRight, 'f', -6, 32), strconv.FormatFloat(responseTop, 'f', -6, 32)}

	expectedArray := [4]string{"-122.515305", "37.698776", "-122.323494", "37.851025"}
	if responseArray != expectedArray {
		t.Errorf("Result was incorrect, got: %s, want: %s", responseArray, expectedArray)
	}
}
