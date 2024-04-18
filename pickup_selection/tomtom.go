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
	LengthInMeters                       int      `json:"lengthInMeters"`
	TravelTimeInSeconds                  int      `json:"travelTimeInSeconds"`
	HistoricalTrafficTravelTimeInSeconds int      `json:"historicTrafficTravelTimeInSeconds"`
	NoTrafficTravelTimeInSeconds         int      `json:"noTrafficTravelTimeInSeconds"`
	TrafficDelayInSeconds                int      `json:"trafficDelayInSeconds"`
	DepartureTime                        string   `json:"departureTime"`
	ArrivalTime                          string   `json:"arrivalTime"`
	Source                               Location `json:"source"`
	Destination                          Location `json:"destination"`
}

func ttCalculateRouteURL(src Location, dst Location) string {
	return fmt.Sprintf(`/calculateRoute/%.6f,%.6f:%.6f,%.6f/json?travelMode=car&routeType=fastest&traffic=true&departAt=now&maxAlternatives=0&computeTravelTimeFor=all&routeRepresentation=summaryOnly`,
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

	// Make response array
	routes := make([]Route, len(sources))

	// Build src lookup, destinations
	lookup := make(map[Location]int)
	destinations := make([]Location, len(sources))
	for i, src := range sources {
		lookup[src] = i
		destinations[i] = destination
	}

	// Get PromCache
	cache := NewPromCache()

	// Get ttl setting
	ttl, err := strconv.Atoi(os.Getenv("TT_TTL"))
	if err != nil {
		// Default to 5min ttl
		ttl = 60 * 5
	}

	// Pull from cache
	cachedRoutes, missedSrcs, _ := cache.GetRoutes("tt", sources, destinations)
	for _, route := range cachedRoutes {
		// Insert into proper index
		i := lookup[route.Source]
		routes[i] = route
	}

	// If there were no missed sources, simply return routes
	if len(missedSrcs) == 0 {
		return routes
	}

	// Start request body
	requestBody := `{"batchItems":[`

	// Add (src,dst) pairs
	for _, source := range missedSrcs {
		fmt.Printf("Cache hit @ (%.6f, %.6f)->(%.6f, %.6f)\n", source.Latitude, source.Longitude, destination.Latitude, destination.Longitude)
		requestBody += fmt.Sprintf(`{"query": "%s"},`, ttCalculateRouteURL(source, destination))
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
	v, err := p.Parse(string(resBody))
	if err != nil {
		fmt.Printf("Error parsing response JSON: %s", err)
		os.Exit(1)
	}

	// Step 1. Loop through the data array
	for i, route := range v.GetArray("batchItems") {
		// Get the route summary
		routeSummary := route.Get("response").GetArray("routes")[0].Get("summary")

		// Create a new route
		newRoute := Route{
			LengthInMeters:                       routeSummary.GetInt("lengthInMeters"),
			TravelTimeInSeconds:                  routeSummary.GetInt("travelTimeInSeconds"),
			HistoricalTrafficTravelTimeInSeconds: routeSummary.GetInt("historicTrafficTravelTimeInSeconds"),
			NoTrafficTravelTimeInSeconds:         routeSummary.GetInt("noTrafficTravelTimeInSeconds"),
			TrafficDelayInSeconds:                routeSummary.GetInt("trafficDelayInSeconds"),
			DepartureTime:                        string(routeSummary.GetStringBytes("departureTime")),
			ArrivalTime:                          string(routeSummary.GetStringBytes("arrivalTime")),
			Source:                               missedSrcs[i],
			Destination:                          destination,
		}

		// Add to routes in proper index
		routes[lookup[newRoute.Source]] = newRoute

		// Store this route in memcache
		cache.StoreRoute("tt", newRoute, int32(ttl))
	}

	return routes
}

// Multiple Source Multiple Destination
func makeRouteRequest(source Location, destination Location) Route {

	// Now get the URL
	locationsStrings := fmt.Sprintf("%f%%2C%f/json?travelMode=carkey=&", source.Latitude, source.Longitude)
	url := os.Getenv("TOMTOM_API_ROUTE_URL") + locationsStrings + os.Getenv("TOMTOM_API_KEY")
	// Make the request
	res, err := http.Get(url)
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
		fmt.Printf("Error parsing response JSON: %s", err)
		os.Exit(1)
	}

	// Step 1. Loop through the data array
	for _, route := range v.GetArray("routes") {
		// Get the route summary
		summary := route.Get("summary")

		// Create a new route
		newRoute := Route{
			LengthInMeters:        summary.GetInt("lengthInMeters"),
			TravelTimeInSeconds:   summary.GetInt("travelTimeInSeconds"),
			TrafficDelayInSeconds: summary.GetInt("trafficDelayInSeconds"),
			DepartureTime:         string(summary.GetStringBytes("departureTime")),
			ArrivalTime:           string(summary.GetStringBytes("arrivalTime")),
			Source:                source,
			Destination:           destination,
		}

		// Append the new route to the routes slice
		routes = append(routes, newRoute)
	}

	return routes[0]
}

// // Multiple Source Multiple Destination
// func makeRouteRequest(source Location, destination Location) Route {

// 	// Now get the URL
// 	locationsStrings := fmt.Sprintf("%f%%2C%f%%3A%f%%2C%f/json?travelMode=carkey=&", source.Latitude, source.Longitude, destination.Latitude, destination.Longitude)
// 	url := os.Getenv("TOMTOM_API_ROUTE_URL") + locationsStrings + os.Getenv("TOMTOM_API_KEY")

// 	// Make the request
// 	res, err := http.Get(url)
// 	if err != nil {
// 		fmt.Printf("Error making http request: %s", err)
// 		os.Exit(1)
// 	}

// 	// Decode the response
// 	resBody, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		fmt.Printf("Error reading response body: %s", err)
// 		os.Exit(1)
// 	}

// 	// Decode the response JSON
// 	var p fastjson.Parser
// 	var routes []Route
// 	v, err := p.Parse(string(resBody))
// 	if err != nil {
// 		fmt.Printf("Error parsing response JSON: %s", err)
// 		os.Exit(1)
// 	}

// 	// Step 1. Loop through the data array
// 	for _, route := range v.GetArray("routes") {
// 		// Get the route summary
// 		summary := route.Get("summary")

// 		// Create a new route
// 		newRoute := Route{
// 			LengthInMeters:        summary.GetInt("lengthInMeters"),
// 			TravelTimeInSeconds:   summary.GetInt("travelTimeInSeconds"),
// 			TrafficDelayInSeconds: summary.GetInt("trafficDelayInSeconds"),
// 			DepartureTime:         string(summary.GetStringBytes("departureTime")),
// 			ArrivalTime:           string(summary.GetStringBytes("arrivalTime")),
// 			Source:                source,
// 			Destination:           destination,
// 		}

// 		// Append the new route to the routes slice
// 		routes = append(routes, newRoute)
// 	}

// 	return routes[0]
// }
