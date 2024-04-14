package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocationToJSON(t *testing.T) {

	test_location := Location{
		Latitude:  30.532431,
		Longitude: 92.352342,
	}

	response := locationToJSON(test_location)
	expected := "{\"latitude\": 30.532431, \"longitude\": 92.352342}"
	if response != expected {
		t.Errorf("Fail: response was not what was expected, got: %s. expected: %s", response, expected)
	}
}

func TestMakeBatchSSMDRoutingRequest(t *testing.T) {

	test_sources := []Location{
		{Latitude: 30.245234235, Longitude: 93.352341235},
		{Latitude: 30.265234235, Longitude: 93.232341235},
		{Latitude: 30.635234235, Longitude: 93.632341235},
	}
	test_destinations := []Location{
		{Latitude: 30.5325234235, Longitude: 93.742341235},
		{Latitude: 30.645234235, Longitude: 93.632341235},
		{Latitude: 30.355234235, Longitude: 93.362341235},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		//w.Write([]byte(`{"durations":[[0,74076.57,1767380.88,7558253],[74076.57,0,1712921.25,7503793.5],[1767380.88,1712921.25,0,5797053.5],[7558253,7503793.5,5797053.5,0]],"destinations":[{"location":[9.700817,48.476406],"snapped_distance":118.9},{"location":[9.207773,49.153882],"snapped_distance":10.54},{"location":[37.572963,55.801279],"snapped_distance":17.44},{"location":[115.665017,38.100717],"snapped_distance":648.79}],"sources":[{"location":[9.700817,48.476406],"snapped_distance":118.9},{"location":[9.207773,49.153882],"snapped_distance":10.54},{"location":[37.572963,55.801279],"snapped_distance":17.44},{"location":[115.665017,38.100717],"snapped_distance":648.79}],"metadata":{"attribution":"openrouteservice.org | OpenStreetMap contributors","service":"matrix","timestamp":1712591855420,"query":{"locations":[[9.70093,48.477473],[9.207916,49.153868],[37.573242,55.801281],[115.663757,38.106467]],"profile":"foot-walking","responseType":"json"},"engine":{"version":"7.1.1","build_date":"2024-01-29T14:41:12Z","graph_date":"2024-03-25T03:50:25Z"}}}`))
		w.Write([]byte(`{"data":[{"originIndex":0,"destinationIndex":0,"routeSummary":{"lengthInMeters":681999,"travelTimeInSeconds":25106,"trafficDelayInSeconds":1769,"departureTime":"2018-08-10T10:20:42+02:00","arrivalTime":"2018-08-10T17:19:07+02:00"}},{"originIndex":0,"destinationIndex":0,"routeSummary":{"lengthInMeters":681999,"travelTimeInSeconds":25106,"trafficDelayInSeconds":1769,"departureTime":"2018-08-10T10:20:42+02:00","arrivalTime":"2018-08-10T17:19:07+02:00"}}]}`))

	}))
	defer ts.Close()
	ThirdPartyURL := ts.URL
	t.Setenv("TOMTOM_API_URL", string(ThirdPartyURL))

	routes := makeBatchSSMDRoutingRequest(test_sources, test_destinations, "car")

	if routes[0].LengthInMeters != 681999 {
		t.Errorf("Fail: Unexpected return from function")
	}

}
