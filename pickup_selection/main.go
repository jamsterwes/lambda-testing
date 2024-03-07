package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type PickupSelectionResponse struct {
	Points     []Location `json:"points"`
	PointCount int        `json:"pointCount"`
}

var RING_RADII []float64 = []float64{0.1, 0.25, 0.5, 0.75}

func HandleRequest(ctx context.Context, event *Location) (*PickupSelectionResponse, error) {
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

	// Return the response
	response := &PickupSelectionResponse{
		Points:     points,
		PointCount: len(points),
	}

	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
