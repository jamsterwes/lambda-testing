package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
)

type Route struct {
	LengthInMeters        int      `json:"lengthInMeters"`
	TravelTimeInSeconds   int      `json:"travelTimeInSeconds"`
	TrafficDelayInSeconds int      `json:"trafficDelayInSeconds"`
	DepartureTime         string   `json:"departureTime"`
	ArrivalTime           string   `json:"arrivalTime"`
	Source                Location `json:"source"`
	Destination           Location `json:"destination"`
}

func locationToJSON(location Location) string {
	return fmt.Sprintf(`{"latitude": %f, "longitude": %f}`, location.Latitude, location.Longitude)
}

func ttCalculateRouteURL(src Location, dst Location) string {
	return fmt.Sprintf(`/calculateRoute/%.6f,%.6f:%.6f,%.6f/json?travelMode=car&routeType=fastest&traffic=true&departAt=now&maxAlternatives=0&routeRepresentation=summaryOnly`,
		src.Latitude,
		src.Longitude,
		dst.Latitude,
		dst.Longitude)
}

// Get a list of routes from TomTom
func getTomTomRoutes(sources []Location, destination Location) []Route {
	// If source empty, return empty
	if len(sources) == 0 {
		return []Route{}
	}

	// Get PromCache
	cache := NewPromCache()

	// Get ttl setting
	ttl, err := strconv.Atoi(os.Getenv("TT_TTL"))
	if err != nil {
		// Default to 5min ttl
		ttl = 60 * 5
	}

	// TODO: Pull from cache

	// Start request body
	requestBody := `{"batchItems":[`

	// Add (src,dst) pairs
	for i := range sources {
		requestBody += fmt.Sprintf(`{"query": "%s"},`, ttCalculateRouteURL(sources[i], destination))
	}

	// Trim trailing comma
	requestBody = requestBody[:len(requestBody)-1]

	// Finish request body
	requestBody += `]}`

	// PRINT REQUEST BODY
	fmt.Println(string(requestBody))

	// Now get the URL
	url := os.Getenv("TOMTOM_API_URL") + os.Getenv("TOMTOM_API_KEY")

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

	// PRINT RESPONSE BODY
	fmt.Println(string(resBody))

	// Decode the response JSON
	var p fastjson.Parser
	var routes []Route
	v, err := p.Parse(string(resBody))
	if err != nil {
		fmt.Printf("Error parsing response JSON: %s", err)
		os.Exit(1)
	}

	// Step 1. Loop through the data array
	for _, route := range v.GetArray("batchItems") {
		// Get the route summary
		routeSummary := route.Get("response").GetArray("routes")[0].Get("summary")

		// Create a new route
		newRoute := Route{
			LengthInMeters:        routeSummary.GetInt("lengthInMeters"),
			TravelTimeInSeconds:   routeSummary.GetInt("travelTimeInSeconds"),
			TrafficDelayInSeconds: routeSummary.GetInt("trafficDelayInSeconds"),
			DepartureTime:         string(routeSummary.GetStringBytes("departureTime")),
			ArrivalTime:           string(routeSummary.GetStringBytes("arrivalTime")),
			Source:                sources[route.GetInt("originIndex")],
			Destination:           destination,
		}

		// Append the new route to the routes slice
		routes = append(routes, newRoute)

		// Store this route in memcache
		cache.StoreRoute("tt", newRoute, int32(ttl))
	}

	return routes
}
