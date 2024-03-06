package main

import (
	"fmt"
	"math"
	"context"
	"io/ioutil"
	"os"
	"log"


	"net/http"
	"net/url"

	"github.com/valyala/fastjson"
	"github.com/aws/aws-lambda-go/lambda"
)

// GIS helper function
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

// GIS helper function
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
func getStreetGeometry(radius float64, latitude float64, longitude float64) [][][2]float64 {
	// Get bounding box
	left, bottom, right, top := getUserBoundingBox(radius, latitude, longitude)

	// Query OSM for streets within the bounding box
	bbox := fmt.Sprintf("%f,%f,%f,%f", bottom, left, top, right)
	query := fmt.Sprintf(`[out:json];(way["highway"="primary"](%s);way["highway"="secondary"](%s);way["highway"="tertiary"](%s);way["highway"="residential"](%s);way["highway"="service"](%s);way["highway"="unclassified"](%s););out geom;`,
		bbox, bbox, bbox, bbox, bbox, bbox)

	// Make the request
	q := make(url.Values)
	q.Set("data", query)

	apiURL := &url.URL{
		Scheme: "https",
		Host: "overpass-api.de",
		Path: "/api/interpreter",
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

	resBody, err := ioutil.ReadAll(res.Body)

	// Decode response JSON (elements only)
	var geometries [][][2]float64
	var p fastjson.Parser

	v, err := p.Parse(string(resBody))
	if err != nil {
		log.Fatal(err)
	}

	// Unpack each street's line segments
	// elements: Street[]
	// Street = []{ lat: number, lon: number }
	for _, street := range v.GetArray("elements") {
		var streetGeometry [][2]float64
		for _, coords := range street.GetArray("geometry") {
			streetGeometry = append(streetGeometry, [2]float64{
				coords.GetFloat64("lat"),
				coords.GetFloat64("lon"),
			})
		}
		geometries = append(geometries, streetGeometry)
	}

	return geometries
}

// Returns [][lat, long]
func intersectLineRing(cx float64, cy float64, a float64, b float64, x1 float64, y1 float64, x2 float64, y2 float64) [][2]float64 {
	// Step 1. Set origin + scale
	X1 := (x1 - cx) / a
	X2 := (x2 - cx) / a
	Y1 := (y1 - cy) / b
	Y2 := (y2 - cy) / b

	// Step 2. Solve
	// Step 2a. Get quadratic coefficients
	A := math.Pow(X2 - X1, 2.0) + math.Pow(Y2 - Y1, 2.0)
	B := 2.0 * ( X1 * (X2 - X1) + Y1 * (Y2 - Y1) )
	C := X1 * X1 + Y1 * Y1 - 1.0

	// Step 2b. Find discriminant
	D := B * B - 4.0 * A * C

	// Step 2c. Get # of solutions
	var ts []float64
	if (D < 0) {
		return [][2]float64{}
	} else if (D > 0) {
		ts = append(ts,
			(-B + math.Sqrt(D)) / (2 * A),
			(-B - math.Sqrt(D)) / (2 * A),
		)
	} else {
		ts = append(ts, -B / (2.0 * A))
	}

	// Step 3. Convert t values into coords
	var solutions [][2]float64
	for _, t := range ts {
		// Ignore t outside (0, 1)
		if (t < 0 || t > 1) {
			continue
		}

		// Get coords
		solutions = append(solutions, [2]float64{
			(x2 - x1) * t + x1,
			(y2 - y1) * t + y1,
		})
	}

	return solutions
}

// Returns [][lat, long]
func intersectWayRing(wayGeom [][2]float64, radius float64, latitude float64, longitude float64) [][2]float64 {
	// Step 0. Ignore empty geom
	if len(wayGeom) == 0 {
		return [][2]float64{}
	}

	// Step 1. Get the two radii of the ellipse
	a := milesToDegLongitude(radius, latitude)
	b := milesToDegLatitude(radius, latitude)

	// Step 2. Accumulate points
	var points [][2]float64
	for i, _ := range wayGeom[:len(wayGeom) - 1] {
		// Get line segment
		y1 := wayGeom[i][0]
		x1 := wayGeom[i][1]
		y2 := wayGeom[i+1][0]
		x2 := wayGeom[i+1][1]

		// Get solutions and concat
		solutions := intersectLineRing(longitude, latitude, a, b, x1, y1, x2, y2)
		points = append(points, solutions...)
	}

	return points
}

type MyEvent struct {
	Latitude float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

func HandleRequest(ctx context.Context, event *MyEvent) (*string, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	// Get the street geometry in a 1mi x 1mi box centered at user position
	var points [][2]float64
	streetGeometries := getStreetGeometry(1, event.Latitude, event.Longitude)

	for _, radius := range []float64{0.1, 0.25, 0.5, 0.75} {
		for _, streetGeom := range streetGeometries {
			solutions := intersectWayRing(streetGeom, radius, event.Latitude, event.Longitude)
			points = append(points, solutions...)
		}
	}

	// Now convert points to string
	var response string = `{"points": [`
	for _, point := range points {
		response += fmt.Sprintf(`{"lat": %f, "long": %f}`, point[0], point[1])
	}
	response += `],"pointCount": ` + fmt.Sprintf(`%d`, len(points)) + `}`

	return &response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
