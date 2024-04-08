package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Tests all functions found in location.go
// As of 4/1 - all functions are covered

func TestCeilToInt(t *testing.T) {
	input := 47.24134
	result := CeilToInt(input)
	expected := 48
	if result != expected {
		t.Errorf("Result was incorrect, got: %d, want: %d", result, expected)
	}

}

func TestORSMatrix(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		//w.Write([]byte(`{"durations":[[0,74076.57,1767380.88,7558253],[74076.57,0,1712921.25,7503793.5],[1767380.88,1712921.25,0,5797053.5],[7558253,7503793.5,5797053.5,0]],"destinations":[{"location":[9.700817,48.476406],"snapped_distance":118.9},{"location":[9.207773,49.153882],"snapped_distance":10.54},{"location":[37.572963,55.801279],"snapped_distance":17.44},{"location":[115.665017,38.100717],"snapped_distance":648.79}],"sources":[{"location":[9.700817,48.476406],"snapped_distance":118.9},{"location":[9.207773,49.153882],"snapped_distance":10.54},{"location":[37.572963,55.801279],"snapped_distance":17.44},{"location":[115.665017,38.100717],"snapped_distance":648.79}],"metadata":{"attribution":"openrouteservice.org | OpenStreetMap contributors","service":"matrix","timestamp":1712591855420,"query":{"locations":[[9.70093,48.477473],[9.207916,49.153868],[37.573242,55.801281],[115.663757,38.106467]],"profile":"foot-walking","responseType":"json"},"engine":{"version":"7.1.1","build_date":"2024-01-29T14:41:12Z","graph_date":"2024-03-25T03:50:25Z"}}}`))
		w.Write([]byte(`{"durations":[[0,74076.57,1767380.88,7558253],[74076.57,0,1712921.25,7503793.5],[1767380.88,1712921.25,0,5797053.5],[7558253,7503793.5,5797053.5,0]],"distances":[[0,102884.97,3152835.5,11191947],[102884.97,0,3077345.5,11116457],[3152835.5,3077345.5,0,8047698],[11191947,11116457,8047698,0]],"destinations":[{"location":[9.700817,48.476406],"snapped_distance":118.9},{"location":[9.207773,49.153882],"snapped_distance":10.54},{"location":[37.572963,55.801279],"snapped_distance":17.44},{"location":[115.665017,38.100717],"snapped_distance":648.79}],"sources":[{"location":[9.700817,48.476406],"snapped_distance":118.9},{"location":[9.207773,49.153882],"snapped_distance":10.54},{"location":[37.572963,55.801279],"snapped_distance":17.44},{"location":[115.665017,38.100717],"snapped_distance":648.79}],"metadata":{"attribution":"openrouteservice.org | OpenStreetMap contributors","service":"matrix","timestamp":1712592739413,"query":{"locations":[[9.70093,48.477473],[9.207916,49.153868],[37.573242,55.801281],[115.663757,38.106467]],"profile":"foot-walking","responseType":"json","metrics":["distance","duration"]},"engine":{"version":"7.1.1","build_date":"2024-01-29T14:41:12Z","graph_date":"2024-03-25T03:50:25Z"}}}`))

	}))
	defer ts.Close()
	ThirdPartyURL := ts.URL

	test_sources := []Location{
		{Latitude: 30.616016382236353, Longitude: -96.3370441554713},
		{Latitude: 30.716016382236353, Longitude: -96.2370441554713},
		{Latitude: 30.816016382236353, Longitude: -96.5370441554713},
		{Latitude: 30.216016382236353, Longitude: -96.4370441554713},
	}
	test_destinations := []Location{
		{Latitude: 30.618016874387585, Longitude: -96.34653115137277},
		{Latitude: 30.513416382236353, Longitude: -96.5260441554713},
		{Latitude: 30.625016382236353, Longitude: -96.4260441554713},
		{Latitude: 30.516016382236353, Longitude: -96.3370441554713},
	}
	routes := ORSMatrix(test_sources, test_destinations, ThirdPartyURL)

	if routes == nil {
		t.Errorf("Error Posting request to ORS API. Result was nil")
	}
}
