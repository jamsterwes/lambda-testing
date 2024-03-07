package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"

	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/valyala/fastjson"
)

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

// GIS helper function
func milesToDegLatitude(miles float64, latitude float64) float64 {
	// Constants
	const a = 3963.190592
	const e = 0.081819191
	phi := latitude * (math.Pi / 180)

	// M = length of 1 radian of latitude in miles
	M := a * (1 - e*e) / math.Pow((1-math.Pow(e*math.Sin(phi), 2)), 1.5)

	// Convert miles to degrees latitude
	return miles * 180 / (math.Pi * M)
}

// GIS helper function
func milesToDegLongitude(miles float64, latitude float64) float64 {
	// Constants
	const a = 3963.190592
	const e = 0.081819191
	phi := latitude * (math.Pi / 180)

	// N = length of 1 radian of longitude in miles
	N := a * math.Cos(phi) / math.Pow(1-math.Pow(e*math.Sin(phi), 2), 0.5)

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
	left := longitude - degLong/2
	bottom := latitude - degLat/2
	right := longitude + degLong/2
	top := latitude + degLat/2

	return left, bottom, right, top
}

// Get street geometry
func getStreetGeometry(radius float64, latitude float64, longitude float64) [][]Location {
	// Get bounding box
	left, bottom, right, top := getUserBoundingBox(radius, latitude, longitude)

	// Query OSM for streets within the bounding box
	bbox := fmt.Sprintf("%f,%f,%f,%f", bottom, left, top, right)
	query := fmt.Sprintf(`
		[out:json];
		(
			way["highway"="primary"](%s);
			way["highway"="secondary"](%s);
			way["highway"="tertiary"](%s);
			way["highway"="residential"](%s);
			way["highway"="service"](%s);
			way["highway"="unclassified"](%s);
		);
		out geom;`,
		bbox, bbox, bbox, bbox, bbox, bbox)

	// Make the request
	q := make(url.Values)
	q.Set("data", query)

	apiURL := &url.URL{
		Scheme:   "https",
		Host:     "overpass-api.de",
		Path:     "/api/interpreter",
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequest(http.MethodGet, apiURL.String(), nil)
	if err != nil {
		fmt.Printf("Error creating http request: %s", err)
		os.Exit(1)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error making http request: %s", err)
		os.Exit(1)
	}

	resBody, err := io.ReadAll(res.Body)

	// Decode response JSON (elements only)
	var geometries [][]Location
	var p fastjson.Parser

	v, err := p.Parse(string(resBody))
	if err != nil {
		log.Fatal(err)
	}

	// Unpack each street's line segments
	// elements: Street[]
	// Street = []{ lat: number, lon: number }
	for _, street := range v.GetArray("elements") {
		var streetGeometry []Location
		for _, coords := range street.GetArray("geometry") {
			streetGeometry = append(streetGeometry, Location{
				Latitude:  coords.GetFloat64("lat"),
				Longitude: coords.GetFloat64("lon"),
			})
		}
		geometries = append(geometries, streetGeometry)
	}

	return geometries
}

// Returns [][lat, long]
func intersectLineRing(cx float64, cy float64, a float64, b float64, x1 float64, y1 float64, x2 float64, y2 float64) []Location {
	// Step 1. Set origin + scale
	X1 := (x1 - cx) / a
	X2 := (x2 - cx) / a
	Y1 := (y1 - cy) / b
	Y2 := (y2 - cy) / b

	// Step 2. Solve
	// Step 2a. Get quadratic coefficients
	A := math.Pow(X2-X1, 2.0) + math.Pow(Y2-Y1, 2.0)
	B := 2.0 * (X1*(X2-X1) + Y1*(Y2-Y1))
	C := X1*X1 + Y1*Y1 - 1.0

	// Step 2b. Find discriminant
	D := B*B - 4.0*A*C

	// Step 2c. Get # of solutions
	var ts []float64
	if D < 0 {
		return []Location{}
	} else if D > 0 {
		ts = append(ts,
			(-B+math.Sqrt(D))/(2*A),
			(-B-math.Sqrt(D))/(2*A),
		)
	} else {
		ts = append(ts, -B/(2.0*A))
	}

	// Step 3. Convert t values into coords
	var solutions []Location
	for _, t := range ts {
		// Ignore t outside (0, 1)
		if t < 0 || t > 1 {
			continue
		}

		// Get coords
		solutions = append(solutions, Location{
			Longitude: (x2-x1)*t + x1,
			Latitude:  (y2-y1)*t + y1,
		})
	}

	return solutions
}

// Returns [][lat, long]
func intersectWayRing(wayGeom []Location, radius float64, latitude float64, longitude float64) []Location {
	// Step 0. Ignore empty geom
	if len(wayGeom) == 0 {
		return []Location{}
	}

	// Step 1. Get the two radii of the ellipse
	a := milesToDegLongitude(radius, latitude)
	b := milesToDegLatitude(radius, latitude)

	// Step 2. Accumulate points
	var points []Location
	for i := range wayGeom[:len(wayGeom)-1] {
		// Get line segment
		x1 := wayGeom[i].Longitude
		y1 := wayGeom[i].Latitude
		x2 := wayGeom[i+1].Longitude
		y2 := wayGeom[i+1].Latitude

		// Get solutions and concat
		solutions := intersectLineRing(longitude, latitude, a, b, x1, y1, x2, y2)
		points = append(points, solutions...)
	}

	return points
}

type PickupSelectionResponse struct {
	Points     []Location `json:"points"`
	PointCount int        `json:"pointCount"`
}

func HandleRequest(ctx context.Context, event *Location) (*PickupSelectionResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	// Get the street geometry in a 1mi x 1mi box centered at user position
	var points []Location
	streetGeometries := getStreetGeometry(1, event.Latitude, event.Longitude)

	for _, radius := range []float64{0.1, 0.25, 0.5, 0.75} {
		for _, streetGeom := range streetGeometries {
			solutions := intersectWayRing(streetGeom, radius, event.Latitude, event.Longitude)
			points = append(points, solutions...)
		}
	}

	response := &PickupSelectionResponse{
		Points:     points,
		PointCount: len(points),
	}

	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
