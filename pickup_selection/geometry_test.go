package main

import "testing"

func TestIntersectLineRing(t *testing.T) {
	response := intersectLineRing(35.23, 53, 532, 34.532, 67.46, 65.74, 27.646, 73.43)

	if response != nil {
		t.Errorf("Result was incorrect, got: nil")
	}
}

func TestIntersectWayRing(t *testing.T) {
	response := intersectWayRing(
		[]Location{
			{Latitude: 37.7749, Longitude: -122.4194},
			{Latitude: 40.7128, Longitude: -74.0060},
			{Latitude: 51.5074, Longitude: -0.1278},
		},
		10.24,
		Location{Latitude: 32.4525, Longitude: -124.423})

	if response != nil {
		t.Errorf("Result was incorrect, got: nil")
	}
}

func TestCullByAngle(t *testing.T) {
	response := cullByAngle(
		[]Location{
			{Latitude: 37.7749, Longitude: -122.4194},
			{Latitude: 40.7128, Longitude: -74.0060},
			{Latitude: 51.5074, Longitude: -0.1278},
		},
		Location{Latitude: 44.4525, Longitude: -93.423},
		4,
		1)
	if response == nil {
		t.Errorf("Result was incorrect, got: nil")
	}
}
