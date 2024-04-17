package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildRide(t *testing.T) {

	test_inbound := RouteSummary{
		Source:      Location{Latitude: 30.5324314241, Longitude: 92.3523423345},
		Destination: Location{Latitude: 30.542312532, Longitude: 92.33425235},
		Time:        9.324,
		Distance:    2.52,
	}
	test_outbound := RouteSummary{
		Source:      Location{Latitude: 30.542312532, Longitude: 92.33425235},
		Destination: Location{Latitude: 30.535234, Longitude: 92.3523424},
		Time:        7.324,
		Distance:    1.52,
	}

	ride := BuildRide(test_inbound, test_outbound)
	if ride.Source != test_inbound.Source {
		t.Errorf("Fail: Ride was not created correct")
	}
}

func TestBuildRides(t *testing.T) {

	test_inbounds := []RouteSummary{
		{Source: Location{Latitude: 30.5324314241, Longitude: 92.3523423345}, Destination: Location{Latitude: 30.542312532, Longitude: 92.33425235}, Time: 9.324, Distance: 2.52},
		{Source: Location{Latitude: 30.5324314241, Longitude: 92.3523423345}, Destination: Location{Latitude: 30.542312532, Longitude: 92.33425235}, Time: 9.324, Distance: 2.52},
		{Source: Location{Latitude: 30.5324314241, Longitude: 92.3523423345}, Destination: Location{Latitude: 30.542312532, Longitude: 92.33425235}, Time: 9.324, Distance: 2.52},
	}
	test_outbounds := []RouteSummary{
		{Source: Location{Latitude: 30.5324314241, Longitude: 92.3523423345}, Destination: Location{Latitude: 30.542312532, Longitude: 92.33425235}, Time: 9.324, Distance: 2.52},
		{Source: Location{Latitude: 30.5324314241, Longitude: 92.3523423345}, Destination: Location{Latitude: 30.542312532, Longitude: 92.33425235}, Time: 9.324, Distance: 2.52},
		{Source: Location{Latitude: 30.5324314241, Longitude: 92.3523423345}, Destination: Location{Latitude: 30.542312532, Longitude: 92.33425235}, Time: 9.324, Distance: 2.52},
	}

	ride := BuildRides(test_inbounds, test_outbounds)
	if ride[0].Source != test_inbounds[0].Source {
		t.Errorf("Fail: Ride was not created correct")
	}
}

func TestPriceRides(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		//w.Write([]byte(`{"durations":[[0,74076.57,1767380.88,7558253],[74076.57,0,1712921.25,7503793.5],[1767380.88,1712921.25,0,5797053.5],[7558253,7503793.5,5797053.5,0]],"destinations":[{"location":[9.700817,48.476406],"snapped_distance":118.9},{"location":[9.207773,49.153882],"snapped_distance":10.54},{"location":[37.572963,55.801279],"snapped_distance":17.44},{"location":[115.665017,38.100717],"snapped_distance":648.79}],"sources":[{"location":[9.700817,48.476406],"snapped_distance":118.9},{"location":[9.207773,49.153882],"snapped_distance":10.54},{"location":[37.572963,55.801279],"snapped_distance":17.44},{"location":[115.665017,38.100717],"snapped_distance":648.79}],"metadata":{"attribution":"openrouteservice.org | OpenStreetMap contributors","service":"matrix","timestamp":1712591855420,"query":{"locations":[[9.70093,48.477473],[9.207916,49.153868],[37.573242,55.801281],[115.663757,38.106467]],"profile":"foot-walking","responseType":"json"},"engine":{"version":"7.1.1","build_date":"2024-01-29T14:41:12Z","graph_date":"2024-03-25T03:50:25Z"}}}`))
		w.Write([]byte(`{"prices": [9.482263565063477,10.340690612792969,13.169022560119629]}`))

	}))
	defer ts.Close()
	ThirdPartyURL := ts.URL
	t.Setenv("PRICING_API_URL", string(ThirdPartyURL))
	test_rides := []Ride{
		{Source: Location{Latitude: 30.5324314241, Longitude: 92.3523423345}, PickupPoint: Location{Latitude: 30.3324314241, Longitude: 92.5523423345}, Destination: Location{Latitude: 30.3324314241, Longitude: 92.5523423345}, WalkTime: 21.41, WalkDistance: 4.23, DriveTime: 12.43, DriveDistance: 2.43, TotalTime: 35.32, TotalDistance: 6.325, Price: 0.0},
		{Source: Location{Latitude: 30.5324314241, Longitude: 92.3523423345}, PickupPoint: Location{Latitude: 30.5324314241, Longitude: 92.6523423345}, Destination: Location{Latitude: 30.3324314241, Longitude: 92.5523423345}, WalkTime: 21.41, WalkDistance: 5.23, DriveTime: 11.43, DriveDistance: 1.43, TotalTime: 36.32, TotalDistance: 7.325, Price: 0.0},
		{Source: Location{Latitude: 30.5324314241, Longitude: 92.3523423345}, PickupPoint: Location{Latitude: 30.6324314241, Longitude: 92.2523423345}, Destination: Location{Latitude: 30.3324314241, Longitude: 92.5523423345}, WalkTime: 21.41, WalkDistance: 6.23, DriveTime: 15.43, DriveDistance: 6.43, TotalTime: 32.32, TotalDistance: 5.325, Price: 0.0},
	}

	ride := PriceRides(test_rides, []MLPricingData{})
	if ride[0].Price == 0 {
		t.Errorf("Fail: Price model did not return a value")
	}
}
