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

func StreamPickupPoints(center Location, streetGeometries [][]Location) []Location {
	pointsChannel := make(chan []Location)

	// Loop through 4 preset radii to find the intersecting points
	for ringID, radius := range RING_RADII {
		go func() {
			// Store the points for this ring
			var points []Location

			// For this specific radius, find the intersecting points
			// and append them to the points slice
			for _, streetGeom := range streetGeometries {
				solutions := intersectWayRing(streetGeom, radius, center)
				points = append(points, solutions...)
			}

			// Now cull the points
			pointsChannel <- cullByAngle(points, center, CULL_SEGMENTS[ringID], CULL_AMOUNTS[ringID])
		}()
	}

	// Receive from channels
	var culledPoints []Location
	for range RING_RADII {
		culledPoints = append(culledPoints, <-pointsChannel...)
	}

	// Return response
	return culledPoints
}

func StreamBuildRides(source Location, destination Location, pickups []Location) []Ride {
	// Make a channel to receive inboundSummaries
	inboundSummariesChannel := make(chan []RouteSummary)

	// Make a channel to receive outboundSummaries
	outboundSummariesChannel := make(chan []RouteSummary)

	// Goroutine to retrieve inbound summaries
	go func(c chan []RouteSummary) {
		// Go get inbound summaries
		urlTest := "nil"
		inboundRoutes := ORSMatrix(pickups, []Location{source}, urlTest)
		inboundSummaries := SummarizeRoutes(inboundRoutes)
		c <- inboundSummaries
	}(inboundSummariesChannel)

	// Goroutine to retrieve outbound summaries
	go func(c chan []RouteSummary) {
		// Go get inbound summaries
		outboundRoutes := makeBatchSSMDRoutingRequest(
			pickups,
			[]Location{destination},
			"car",
		)
		outboundSummaries := SummarizeRoutes(outboundRoutes)
		c <- outboundSummaries
	}(outboundSummariesChannel)

	// Now build rides
	inboundSummaries := <-inboundSummariesChannel
	outboundSummaries := <-outboundSummariesChannel
	return BuildRides(inboundSummaries, outboundSummaries)
}

func HandleRequest(ctx context.Context, event *PickupSelectionRequest) (*PickupSelectionResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	// Get the street geometry in a 1mi x 1mi box centered at user position
	streetGeometries := getStreetGeometry(1, event.Source, "nil")
	culledPoints := StreamPickupPoints(event.Source, streetGeometries)

	// Build rides in parallel
	rides := StreamBuildRides(event.Source, event.Destination, culledPoints)

	// Price rides
	rides = PriceRides(rides, []MLPricingData{})

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
