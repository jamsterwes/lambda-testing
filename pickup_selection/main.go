package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

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

func StreamBuildRides(source Location, destination Location, pickups []Location) ([]Ride, []MLPricingData) {
	// Make a channel to receive inboundSummaries
	inboundSummariesChannel := make(chan []RouteSummary)

	// Make a channel to receive outboundSummaries
	outboundSummariesChannel := make(chan []RouteSummary)

	// Make a channel to receive pricingData
	pricingDataChannel := make(chan []MLPricingData)

	// Goroutine to retrieve inbound summaries
	go func(c chan []RouteSummary) {
		// Go get inbound summaries
		urlTest := "nil"
		inboundRoutes := ORSMatrix(pickups, []Location{source}, urlTest)
		inboundSummaries := SummarizeRoutes(inboundRoutes)
		c <- inboundSummaries
	}(inboundSummariesChannel)

	// Goroutine to retrieve outbound summaries
	go func(c chan []RouteSummary, m chan []MLPricingData) {
		// Go get inbound summaries
		outboundRoutes := getTomTomRoutes(
			pickups,
			destination,
		)

		// Now summarize routes
		outboundSummaries := SummarizeRoutes(outboundRoutes)
		c <- outboundSummaries

		// Get the day-of-week and time-of-day
		// TODO: in future we would want the user's time zone...
		// assume CDT for now
		loc := time.FixedZone("CDT", -5*60*60)
		now := time.Now().In(loc)
		day_of_week := float64(now.Weekday()) / 7
		time_of_day := (float64(now.Hour()) + (float64(now.Minute()) / 60)) / 24

		// Now build pricing data
		pricingData := make([]MLPricingData, len(outboundRoutes))
		for i, route := range outboundRoutes {
			data := MLPricingData{
				TimeInSeconds:        float64(route.TravelTimeInSeconds),
				DistanceInMeters:     float64(route.LengthInMeters),
				TimeToHistoricRatio:  float64(route.TravelTimeInSeconds) / float64(route.HistoricalTrafficTravelTimeInSeconds),
				TimeToNoTrafficRatio: float64(route.TravelTimeInSeconds) / float64(route.NoTrafficTravelTimeInSeconds),
				DayOfWeekSin:         math.Sin(2 * math.Pi * day_of_week),
				DayOfWeekCos:         math.Cos(2 * math.Pi * day_of_week),
				TimeOfDaySin:         math.Sin(2 * math.Pi * time_of_day),
				TimeOfDayCos:         math.Cos(2 * math.Pi * time_of_day),
			}

			pricingData[i] = data
		}
		m <- pricingData
	}(outboundSummariesChannel, pricingDataChannel)

	// Now build rides
	inboundSummaries := <-inboundSummariesChannel
	outboundSummaries := <-outboundSummariesChannel
	return BuildRides(inboundSummaries, outboundSummaries), <-pricingDataChannel
}

func HandleRequest(ctx context.Context, event *PickupSelectionRequest) (*PickupSelectionResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	// Get the street geometry in a 1mi x 1mi box centered at user position
	streetGeometries := getStreetGeometry(1, event.Source, "nil")
	culledPoints := StreamPickupPoints(event.Source, streetGeometries)

	// Add the source to the end of culled points for savings calculations
	// This gets us the pricing data of the no-walking ride for free
	culledPoints = append(culledPoints, event.Source)

	// Build rides in parallel
	rides, pricingData := StreamBuildRides(event.Source, event.Destination, culledPoints)

	// Price rides
	rides = PriceRides(rides, pricingData)

	// Remember to take the no-walking ride out of the slice
	rides = rides[:len(rides)-1]

	// TODO: do something with ride prices, etc
	// sort rides by price lowest -> highest
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
