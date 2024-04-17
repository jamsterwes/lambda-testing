package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/valyala/fastjson"
)

type Ride struct {
	Source        Location `json:"source"`
	PickupPoint   Location `json:"pickupPoint"`
	Destination   Location `json:"destination"`
	WalkTime      float64  `json:"walkTime"`
	WalkDistance  float64  `json:"walkDistance"`
	DriveTime     float64  `json:"driveTime"`
	DriveDistance float64  `json:"driveDistance"`
	TotalTime     float64  `json:"totalTime"`
	TotalDistance float64  `json:"totalDistance"`
	Price         float64  `json:"price"`
}

// This stores all the data needed to price a ride
type MLPricingData struct {
	TimeInSeconds        float64
	DistanceInMeters     float64
	TimeToHistoricRatio  float64
	TimeToNoTrafficRatio float64
	DayOfWeekSin         float64
	DayOfWeekCos         float64
	TimeOfDaySin         float64
	TimeOfDayCos         float64
}

func BuildRide(inbound RouteSummary, outbound RouteSummary) Ride {
	return Ride{
		Source:        inbound.Source,
		PickupPoint:   inbound.Destination,
		Destination:   outbound.Destination,
		WalkTime:      inbound.Time,
		WalkDistance:  inbound.Distance,
		DriveTime:     outbound.Time,
		DriveDistance: outbound.Distance,
		TotalTime:     inbound.Time + outbound.Time,
		TotalDistance: inbound.Distance + outbound.Distance,
		Price:         0.0,
	}
}

func BuildRides(inbounds []RouteSummary, outbounds []RouteSummary) []Ride {
	var rides []Ride
	for i := range inbounds {
		rides = append(rides, BuildRide(inbounds[i], outbounds[i]))
	}
	return rides
}

func PriceRides(rides []Ride, pricingData []MLPricingData) []Ride {
	// Get miles
	var miles []float64
	for _, ride := range rides {
		miles = append(miles, ride.DriveDistance)
	}

	// Get minutes
	var minutes []float64
	for _, ride := range rides {
		minutes = append(minutes, ride.DriveTime/60)
	}

	// Build request object
	requestBody := `{"miles": [`

	// Add miles
	for _, mile := range miles {
		requestBody += fmt.Sprintf("%f,", mile)
	}

	// Remove trailing comma
	requestBody = requestBody[:len(requestBody)-1]

	// Add minutes
	requestBody += `],"minutes": [`

	for _, minute := range minutes {
		requestBody += fmt.Sprintf("%f,", minute)
	}

	// Remove trailing comma
	requestBody = requestBody[:len(requestBody)-1]

	// Finish request body
	requestBody += `]}`

	// Make HTTP request to pricing service
	url := os.Getenv("PRICING_API_URL")

	fmt.Printf("Making request to %s\n", url)
	fmt.Printf("Request body: %s\n", requestBody)
	req, err := http.Post(url, "application/json", strings.NewReader(requestBody))
	if err != nil {
		fmt.Printf("Error making http request: %s", err)
		os.Exit(1)
	}

	// Decode the response
	resBody, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Response body: %s\n", string(resBody))

	// Parse the response
	var p fastjson.Parser
	v, err := p.Parse(string(resBody))
	if err != nil {
		fmt.Printf("Error parsing response body: %s", err)
		os.Exit(1)
	}

	// Get the prices
	prices := v.GetArray("prices")
	for i, price := range prices {
		rides[i].Price = price.GetFloat64()
	}

	return rides
}
