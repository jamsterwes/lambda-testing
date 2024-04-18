package main

import (
	"fmt"
	"os"

	"github.com/memcachier/mc/v3"
	"github.com/valyala/fastjson"
)

type PromCache struct {
	client *mc.Client
}

// Function to get a new PromCache instance
func NewPromCache() *PromCache {
	// Make cache
	cache := PromCache{
		client: mc.NewMC(os.Getenv("CACHE_URL"), os.Getenv("MEMCACHED_USERNAME"), os.Getenv("MEMCACHED_PASSWORD")),
	}

	// Return pointer to cache
	return &cache
}

// Function to store Route in cache
func (cache *PromCache) StoreRoute(prefix string, route Route, ttl int32) {
	// Get the key for this route
	key := fmt.Sprintf("%s_%.6f_%.6f_%.6f_%.6f",
		prefix,
		route.Source.Latitude,
		route.Source.Longitude,
		route.Destination.Latitude,
		route.Destination.Longitude)

	// Store route (length in meter, travel time in sec, traffic delay in sec)
	routeJSON := fmt.Sprintf(`{"lengthInMeters": %d, "travelTimeInSeconds": %d, "trafficDelayInSeconds": %d, "historical": %d, "noTraffic": %d}`,
		route.LengthInMeters,
		route.TravelTimeInSeconds,
		route.TrafficDelayInSeconds,
		route.HistoricalTrafficTravelTimeInSeconds,
		route.NoTrafficTravelTimeInSeconds)

	// Debug key
	fmt.Printf("stored key: %s\n", key)

	// Store in cache
	_, err := cache.client.Set(key, routeJSON, 0, uint32(ttl), 0)
	if err != nil {
		fmt.Printf("error: %+v\n", err)
	}
}

// Helper function to get route from JSON
func ParseRouteJSON(routeJSON string, source Location, destination Location) *Route {
	// Decode JSON
	var p fastjson.Parser
	v, err := p.Parse(routeJSON)
	if err != nil {
		return nil
	}

	return &Route{
		TravelTimeInSeconds:                  v.GetInt("travelTimeInSeconds"),
		LengthInMeters:                       v.GetInt("lengthInMeters"),
		TrafficDelayInSeconds:                v.GetInt("trafficDelayInSeconds"),
		HistoricalTrafficTravelTimeInSeconds: v.GetInt("historical"),
		NoTrafficTravelTimeInSeconds:         v.GetInt("noTraffic"),
		ArrivalTime:                          "",
		DepartureTime:                        "",
		Source:                               source,
		Destination:                          destination,
	}
}

// Function to retrieve Route from cache (or nil if not found)
func (cache *PromCache) GetRoute(prefix string, source Location, destination Location) *Route {
	// Get the key for this route
	key := fmt.Sprintf("%s_%.6f_%.6f_%.6f_%.6f",
		prefix,
		source.Latitude,
		source.Longitude,
		destination.Latitude,
		destination.Longitude)

	// Get from cache
	data, _, _, err := cache.client.Get(key)
	if err != nil {
		if err != mc.ErrNotFound {
			fmt.Printf("error: %+v\n", err)
		}
		return nil
	}

	// Decode JSON + return route
	return ParseRouteJSON(data, source, destination)
}

// Function to retrieve multiple Routes from cache (or nil if not found)
// Returns:
// 1. Routes from cache
// 2. Sources not found
// 3. Destinations not found
func (cache *PromCache) GetRoutes(prefix string, sources []Location, destinations []Location) ([]Route, []Location, []Location) {
	// Store output arrays
	var routes []Route
	var missedSrcs []Location
	var missedDsts []Location

	// For each (src,dst) query the cache
	for i := range sources {
		route := cache.GetRoute(prefix, sources[i], destinations[i])
		if route != nil {
			routes = append(routes, *route)
		} else {
			missedSrcs = append(missedSrcs, sources[i])
			missedDsts = append(missedDsts, destinations[i])
		}
	}

	return routes, missedSrcs, missedDsts
}
