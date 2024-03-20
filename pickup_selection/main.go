package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-lambda-go/lambda"
)

type PickupSelectionRequest = struct {
	Source      Location `json:"source"`
	Destination Location `json:"destination"`
	MaxPoints   int      `json:"maxPoints"`
}

type PickupSelectionResponse struct {
	Rides []Ride `json:"rides"`
}

var RING_RADII []float64 = []float64{0.1, 0.25, 0.5, 0.75}
var CULL_SEGMENTS []int = []int{4, 4, 3, 3}
var CULL_AMOUNTS []int = []int{1, 1, 1, 1}

type RouteSummary struct {
	Source      Location `json:"source"`
	Destination Location `json:"destination"`
	Time        float64  `json:"time"`
	Distance    float64  `json:"distance"` // in mi
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
	var culledPoints []Location
	streetGeometries := getStreetGeometry(1, event.Source)

	// Loop through 4 preset radii to find the intersecting points
	for ringID, radius := range RING_RADII {
		// Store the points for this ring
		var points []Location

		// For this specific radius, find the intersecting points
		// and append them to the points slice
		for _, streetGeom := range streetGeometries {
			solutions := intersectWayRing(streetGeom, radius, event.Source)
			points = append(points, solutions...)
		}

		// Now cull the points
		points = cullByAngle(points, event.Source, CULL_SEGMENTS[ringID], CULL_AMOUNTS[ringID])
		culledPoints = append(culledPoints, points...)
	}

	// Now get inbound summaries
	inboundRoutes := ORSMatrix(culledPoints, []Location{event.Source})
	inboundSummaries := SummarizeRoutes(inboundRoutes)
	fmt.Printf("Inbound Summaries: %+v\n", inboundSummaries)

	// Now get outbound summaries
	outboundRoutes := makeBatchSSMDRoutingRequest(
		culledPoints,
		[]Location{event.Destination},
		"car",
	)
	outboundSummaries := SummarizeRoutes(outboundRoutes)
	fmt.Printf("Outbound Summaries: %+v\n", outboundSummaries)

	// Build rides
	rides := BuildRides(inboundSummaries, outboundSummaries)

	// Price rides
	rides = PriceRides(rides)
	fmt.Println(rides)

	// TODO: do something with ride prices, etc
	// sort rides by price
	sort.Slice(rides, func(i, j int) bool {
		return rides[i].Price < rides[j].Price
	})

	// Return the response
	response := &PickupSelectionResponse{
		Rides: rides,
	}

	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
