package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/aws/aws-lambda-go/lambda"
)

type PickupSelectionRequest = Location

type PickupSelectionResponse struct {
	Routes []Route `json:"routes"`
}

var RING_RADII []float64 = []float64{0.1, 0.25, 0.5, 0.75}

// TODO: write this to be more spatially-aware
func CullPoints(points []Location, maxPoints int) []Location {
	// Step 1. Randomly shuffle points
	rand.Shuffle(len(points), func(i, j int) { points[i], points[j] = points[j], points[i] })

	// Step 2. Return the first 10 points
	return points[:min(len(points), maxPoints)]
}

func HandleRequest(ctx context.Context, event *PickupSelectionRequest) (*PickupSelectionResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	// Get the street geometry in a 1mi x 1mi box centered at user position
	var points []Location
	streetGeometries := getStreetGeometry(1, *event)

	// Loop through 4 preset radii to find the intersecting points
	for _, radius := range RING_RADII {
		// For this specific radius, find the intersecting points
		// and append them to the points slice
		for _, streetGeom := range streetGeometries {
			solutions := intersectWayRing(streetGeom, radius, *event)
			points = append(points, solutions...)
		}
	}

	// TODO: properly cull points
	culledPoints := CullPoints(points, 10)

	// Now call TomTom API
	routes := makeBatchSSMDRoutingRequest(
		[]Location{*event},
		culledPoints,
		"pedestrian",
	)

	// TODO: use routes to select the best pickup points

	// Return the response
	response := &PickupSelectionResponse{
		Routes: routes,
	}

	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
