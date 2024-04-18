package main

import (
	"testing"
)

// I think theres gonna be some problems on testing the main because it calls the functions
// I'll need to look for workourounds for the different fucntions that were used
// calls to functions that make api calls have to be taken care of
// make multiple servers?
// go back to that one function where I made a change to the parameters to do the env file method?
//

func TestSummarizeRoutes(t *testing.T) {

	test_rides := []Route{
		{LengthInMeters: 4124, TravelTimeInSeconds: 3251, TrafficDelayInSeconds: 32523, DepartureTime: "2018-08-10T10:20:42+02:00", ArrivalTime: "2018-08-10T10:20:42+02:00", Source: Location{Latitude: 30.3525234, Longitude: 92.45132523}, Destination: Location{Latitude: 32.532452, Longitude: 92.35343}},
		{LengthInMeters: 4124, TravelTimeInSeconds: 3251, TrafficDelayInSeconds: 32523, DepartureTime: "2018-08-10T10:20:42+02:00", ArrivalTime: "2018-08-10T10:20:42+02:00", Source: Location{Latitude: 30.3525234, Longitude: 92.45132523}, Destination: Location{Latitude: 32.532452, Longitude: 92.35343}},
		{LengthInMeters: 4124, TravelTimeInSeconds: 3251, TrafficDelayInSeconds: 32523, DepartureTime: "2018-08-10T10:20:42+02:00", ArrivalTime: "2018-08-10T10:20:42+02:00", Source: Location{Latitude: 30.3525234, Longitude: 92.45132523}, Destination: Location{Latitude: 32.532452, Longitude: 92.35343}},
	}

	summaries := SummarizeRoutes(test_rides)
	expected_distance := 0.000621371 * 4124
	if summaries[0].Distance != expected_distance {
		t.Errorf("Fail: response was not what was expected, got: %f. expected: %f", summaries[0].Distance, expected_distance)
	}
}

func TestStreamPickupPoints(t *testing.T) {
	test_center := Location{
		Latitude:  31.5,
		Longitude: 91.5,
	}
	rows := 3
	cols := 3

	// Create a 2D array of Location structs with the specified dimensions
	test_locations := make([][]Location, rows)
	for i := range test_locations {
		test_locations[i] = make([]Location, cols)
	}

	// Populate the array with some Location values
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Assigning sample latitude and longitude values for demonstration
			test_locations[i][j] = Location{Latitude: float64(i + 30), Longitude: float64(j + 90)}
		}
	}
	result := StreamPickupPoints(test_center, test_locations)
	if result == nil {
		t.Errorf("Fail: Got unexpected result, nil")
	}

}
