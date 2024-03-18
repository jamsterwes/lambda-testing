package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/aws/aws-lambda-go/lambda"
)

type PickupSelectionRequest = struct {
	Source      Location `json:"source"`
	Destination Location `json:"destination"`
	MaxPoints   int      `json:"maxPoints"`
}

type PickupSelectionResponse struct {
	Points []RouteSummary `json:"points"`
}

var RING_RADII []float64 = []float64{0.1, 0.25, 0.5, 0.75}

// TODO: write this to be more spatially-aware
func CullPoints(points []Location, maxPoints int) []Location {
	// Step 1. Randomly shuffle points
	rand.Shuffle(len(points), func(i, j int) { points[i], points[j] = points[j], points[i] })

	// Step 2. Return the first 10 points
	return points[:min(len(points), maxPoints)]
}

type RouteSummary struct {
	Source      Location `json:"source"`
	Destination Location `json:"destination"`
	Time        float64  `json:"time"`
	Distance    float64  `json:"distance"`
}

const MetersToMiles float64 = 0.000621371

func SummarizeRoutes(routes []Route) []RouteSummary {
	var summaries []RouteSummary
	for _, route := range routes {
		summaries = append(summaries, RouteSummary{
			Source:      route.Source,
			Destination: route.Destination,
			Time:        float64(route.TravelTimeInSeconds),
			Distance:    float64(route.LengthInMeters) * MetersToMiles,
		})
	}
	return summaries
}

func RankPickupPoints(inboundSummaries []RouteSummary, outboundSummaries []RouteSummary, maxPoints int) []RouteSummary {
	return []RouteSummary{}
}

func HandleRequest(ctx context.Context, event *PickupSelectionRequest) (*PickupSelectionResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	// Get the street geometry in a 1mi x 1mi box centered at user position
	var points []Location
	streetGeometries := getStreetGeometry(1, event.Source)

	// Loop through 4 preset radii to find the intersecting points
	for _, radius := range RING_RADII {
		// For this specific radius, find the intersecting points
		// and append them to the points slice
		for _, streetGeom := range streetGeometries {
			solutions := intersectWayRing(streetGeom, radius, event.Source)
			points = append(points, solutions...)
		}
	}

	// Cull the potential pickup points down to some predetermined threshold/density
	// TODO: properly cull points
	culledPoints := CullPoints(points, 10)

	// TODO: Now get inbound summaries
	var inboundSummaries []RouteSummary
	fmt.Print(inboundSummaries)

	// Now get outbound summaries
	outboundRoutes := makeBatchSSMDRoutingRequest(
		culledPoints,
		[]Location{event.Destination},
		"car",
	)

	outboundSummaries := SummarizeRoutes(outboundRoutes)
	fmt.Print(outboundSummaries)

	// TODO: use summaries to rank points

	// Return the response
	response := &PickupSelectionResponse{
		Points: outboundSummaries,
	}

	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
