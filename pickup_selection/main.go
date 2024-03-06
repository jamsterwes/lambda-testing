package main

import (
	"fmt"
	"math"
	"net/http"
	"context"

	// "github.com/valyala/fastjson"

	"github.com/aws/aws-lambda-go/lambda"
)

func milesToDegLatitude(miles float64, latitude float64) float64 {
	// Constants
	const a = 3963.190592
	const e = 0.081819191
	phi := latitude * (3.141592653589793 / 180)

	// M = length of 1 radian of latitude in miles
	M := a * (1 - e * e) / math.Pow((1 - math.Pow(e * math.Sin(phi), 2)), 1.5)

	// Convert miles to degrees latitude
	return miles * 180 / (math.Pi * M)
}

func milesToDegLongitude(miles float64, latitude float64) float64 {
	// Constants
	const a = 3963.190592
	const e = 0.081819191
	phi := latitude * (3.141592653589793 / 180)

	// N = length of 1 radian of longitude in miles
	N := a * math.Cos(phi) / math.Pow(1 - math.Pow(e * math.Sin(phi), 2), 0.5)

	// Convert miles to degrees longitude
	return miles * 180 / (math.Pi * N)
}

// Size - the size of the bounding box in miles
// Latitude - the latitude of the center of the bounding box
// Longitude - the longitude of the center of the bounding box
// Returns: the bounding box in the form of
// .. (left, bottom, right, top)
func getUserBoundingBox(size float64, latitude float64, longitude float64) (float64, float64, float64, float64) {
	// Convert miles to degrees latitude and longitude
	degLat := milesToDegLatitude(size, latitude)
	degLong := milesToDegLongitude(size, latitude)

	// Calculate the bounding box
	left := longitude - degLong / 2
	bottom := latitude - degLat / 2
	right := longitude + degLong / 2
	top := latitude + degLat / 2

	return left, bottom, right, top
}

// Get street geometry
func getStreetGeometry(radius float64, latitude float64, longitude float64) [](float64, float64) {
	// Get bounding box
	(left, bottom, right, top) = getUserBoundingBox(radius, latitude, longitude)

	// Query OSM for streets within the bounding box
	const apiURL = "https://overpass-api.de/api/interpreter"
	bbox := fmt.Sprintf("%f,%f,%f,%f", bottom, left, top, right)
	query := fmt.Sprintf(`[out:json];(way["highway"="primary"](%s);way["highway"="secondary"](%s);way["highway"="tertiary"](%s);way["highway"="residential"](%s);way["highway"="service"](%s);way["highway"="unclassified"](%s););out geom;`,
		bbox, bbox, bbox, bbox, bbox, bbox)

	// Make the request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte(query)))
	resp, err := http.DefaultClient.Do(req)

	return [(0, 0)]
}


type MyEvent struct {
	Latitude float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

func HandleRequest(ctx context.Context, event *MyEvent) (*string, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	fmt.Printf("Received event: %f, %f\n", event.Latitude, event.Longitude)

	// Get the bounding box (1mi x 1mi) centered at user position
	left, bottom, right, top := getUserBoundingBox(1, event.Latitude, event.Longitude)

	// TODO: Selection

	// For now, return the bounding box as JSON
	response := fmt.Sprintf("{\"left\": %f, \"bottom\": %f, \"right\": %f, \"top\": %f}", left, bottom, right, top)
	return &response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
