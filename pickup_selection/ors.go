package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/valyala/fastjson"
)

func CeilToInt(x float64) int {
	return int(math.Ceil(x))
}

// "walking"-only
func ORSMatrix(sources []Location, destinations []Location) []Route {
	// If source empty, return empty
	if len(sources) == 0 {
		return []Route{}
	}

	// If destination empty, return empty
	if len(destinations) == 0 {
		return []Route{}
	}

	// Create request body
	requestBody := `{"locations":[`

	// Add the sources first
	for _, source := range sources {
		// Add source
		// Always add commas (there will be at least one destination)
		requestBody += fmt.Sprintf("[%0.6f,%0.6f],", source.Longitude, source.Latitude)
	}

	// Now add the destinations
	for _, destination := range destinations {
		// Add destination
		requestBody += fmt.Sprintf("[%0.6f,%0.6f],", destination.Longitude, destination.Latitude)
	}

	// Remove the last comma
	requestBody = requestBody[:len(requestBody)-1]

	// Now specify the source indices
	requestBody += `],"sources":[`
	for i := range sources {
		requestBody += fmt.Sprintf("%d,", i)
	}

	// Remove the last comma
	requestBody = requestBody[:len(requestBody)-1]

	// Now specify the destination indices
	requestBody += `],"destinations":[`
	for i := range destinations {
		// Destinations are offset by len(sources)
		requestBody += fmt.Sprintf("%d,", len(sources)+i)
	}

	// Remove the last comma
	requestBody = requestBody[:len(requestBody)-1]

	// Now request distance AND duration
	requestBody += `],"metrics":["distance","duration"]`

	// Now request meters and finish the request
	requestBody += `,"units":"m"}`

	// Send the request to the ORS API
	url := "https://api.openrouteservice.org/v2/matrix/foot-walking"

	// Set the HTTP header Authorization to API Key
	req, err := http.NewRequest("POST", url, strings.NewReader(requestBody))
	if err != nil {
		fmt.Printf("Error creating http request: %s", err)
		os.Exit(1)
	}

	// Set the content type
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", os.Getenv("ORS_API_KEY"))

	// Make the request
	res, err := http.DefaultClient.Do(req)
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

	// Print response
	fmt.Printf("ORSMatrix: %s\n", resBody)

	// Unpack JSON
	var p fastjson.Parser
	v, err := p.Parse(string(resBody))

	// Get routes
	var routes []Route
	for i, row := range v.GetArray("durations") {
		for j, cell := range row.GetArray() {
			// Get v["distances"][i][j] as float64
			length := CeilToInt(v.GetArray("distances")[i].GetArray()[j].GetFloat64())
			duration := CeilToInt(cell.GetFloat64())

			routes = append(routes, Route{
				Source:                sources[i],
				Destination:           destinations[j],
				LengthInMeters:        length,
				TravelTimeInSeconds:   duration,
				TrafficDelayInSeconds: 0, // unsupported for walking
			})
		}
	}

	fmt.Printf("ORSMatrix: %d routes\n", len(routes))

	return routes
}
