package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/valyala/fastjson"
)

type Route struct {
	LengthInMeters        int    `json:"lengthInMeters"`
	TravelTimeInSeconds   int    `json:"travelTimeInSeconds"`
	TrafficDelayInSeconds int    `json:"trafficDelayInSeconds"`
	DepartureTime         string `json:"departureTime"`
	ArrivalTime           string `json:"arrivalTime"`
}

func locationToJSON(location Location) string {
	return fmt.Sprintf(`{"latitude": %f, "longitude": %f}`, location.Latitude, location.Longitude)
}

// Multiple Source Multiple Destination
func makeBatchSSMDRoutingRequest(sources []Location, destinations []Location, travelMode string) []Route {
	// Create the request body
	var requestBody string = `{`

	// Add the origins to the body
	requestBody += `"origins": [`
	for i, source := range sources {
		requestBody += `{
			"point": ` + locationToJSON(source) + `
		}`
		if i < len(sources)-1 {
			requestBody += `,`
		}
	}
	requestBody += `],`

	// Add the destinations to the body
	requestBody += `"destinations": [`
	for i, destination := range destinations {
		requestBody += `{
			"point": ` + locationToJSON(destination) + `
		}`
		if i < len(destinations)-1 {
			requestBody += `,`
		}
	}
	requestBody += `],`

	// Add the rest of the body
	requestBody += `"options": {
		"traffic": "live",
		"departAt": "now",
		"travelMode": "` + travelMode + `"
	}}`

	fmt.Printf("Request body: %s\n", requestBody)

	// Now get the URL
	url := "https://api.tomtom.com/routing/matrix/2?key=" + os.Getenv("TOMTOM_API_KEY")
	fmt.Printf("URL: %s\n", url)

	// Make the request
	res, err := http.Post(url, "application/json", strings.NewReader(requestBody))
	if err != nil {
		fmt.Printf("Error making http request: %s", err)
		os.Exit(1)
	}

	// Decode the response
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s", err)
		os.Exit(1)
	}

	// Decode the response JSON
	var p fastjson.Parser
	var routes []Route
	v, err := p.Parse(string(resBody))
	if err != nil {
		fmt.Printf("Error parsing response body: %s", err)
		os.Exit(1)
	}

	// Step 1. Loop through the data array
	for _, route := range v.GetArray("data") {
		// Get the route summary
		routeSummary := route.Get("routeSummary")

		// Create a new route
		newRoute := Route{
			LengthInMeters:        routeSummary.GetInt("lengthInMeters"),
			TravelTimeInSeconds:   routeSummary.GetInt("travelTimeInSeconds"),
			TrafficDelayInSeconds: routeSummary.GetInt("trafficDelayInSeconds"),
			DepartureTime:         string(routeSummary.GetStringBytes("departureTime")),
			ArrivalTime:           string(routeSummary.GetStringBytes("arrivalTime")),
		}

		// Append the new route to the routes slice
		routes = append(routes, newRoute)
	}

	return routes
}
