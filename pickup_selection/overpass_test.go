package main

import (
	"testing"
)

// Tests all functions found in location.go
// As of 4/1 - all functions are covered

func TestGetStreetGeometry(t *testing.T) {

	test_source := Location{
		Latitude: 30.616016382236353, Longitude: -96.3370441554713,
	}

	geomtries := getStreetGeometry(1, test_source, "nil")

	if geomtries == nil {
		t.Errorf("Error posting request to overpass API. Result was nil")
	}
}
