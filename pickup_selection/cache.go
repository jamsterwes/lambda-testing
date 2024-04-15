package main

import (
	"fmt"
	"os"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/valyala/fastjson"
)

type PromCache struct {
	client *memcache.Client
}

// Function to get a new PromCache instance
func NewPromCache() *PromCache {
	// Make cache
	cache := PromCache{
		client: memcache.New(os.Getenv("CACHE_URL")),
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
	routeJSON := fmt.Sprintf(`{"lengthInMeters": %d, "travelTimeInSeconds": %d, "trafficDelayInSeconds": %d}`,
		route.LengthInMeters,
		route.TravelTimeInSeconds,
		route.TrafficDelayInSeconds)

	// Debug key
	fmt.Printf("stored key: %s\n", key)

	// Store in cache
	err := cache.client.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(routeJSON),
		Flags:      0,
		Expiration: ttl,
	})
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
		TravelTimeInSeconds:   v.GetInt("travelTimeInSeconds"),
		LengthInMeters:        v.GetInt("lengthInMeters"),
		TrafficDelayInSeconds: v.GetInt("trafficDelayInSeconds"),
		ArrivalTime:           "",
		DepartureTime:         "",
		Source:                source,
		Destination:           destination,
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
	data, err := cache.client.Get(key)
	if err != nil {
		if err != memcache.ErrCacheMiss {
			fmt.Printf("error: %+v\n", err)
		}
		return nil
	}

	// Decode JSON + return route
	return ParseRouteJSON(string(data.Value), source, destination)
}

// Function to retrieve multiple Routes from cache (or nil if not found)
// Returns:
// 1. Routes from cache
// 2. Sources not found
// 3. Destinations not found
func (cache *PromCache) GetRoutes(prefix string, sources []Location, destinations []Location) ([]Route, []Location, []Location) {
	// Store the (src,dest) per key for quick translation
	srcs := make(map[string]Location)
	dsts := make(map[string]Location)

	// Get the keys for these routes
	var keys []string
	for i := range sources {
		key := fmt.Sprintf("%s_%.6f_%.6f_%.6f_%.6f",
			prefix,
			sources[i].Latitude,
			sources[i].Longitude,
			destinations[i].Latitude,
			destinations[i].Longitude)
		srcs[key] = sources[i]
		dsts[key] = destinations[i]
		keys = append(keys, key)
	}

	// Get from cache
	dataMap, err := cache.client.GetMulti(keys)
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		return nil, sources, destinations
	}

	// Store routes
	var routes []Route
	for k, v := range dataMap {
		// Get src,dest for this key
		src := srcs[k]
		dest := dsts[k]

		// Remove this key from srcs+dsts
		delete(srcs, k)
		delete(dsts, k)

		// Append to routes
		routes = append(routes, *ParseRouteJSON(string(v.Value), src, dest))
	}

	var missedSrcs []Location
	var missedDsts []Location
	for k := range srcs {
		missedSrcs = append(missedSrcs, srcs[k])
		missedDsts = append(missedDsts, dsts[k])
	}

	return routes, missedSrcs, missedDsts
}
