package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/valyala/fastjson"
)

// Get street geometry
func getStreetGeometry(radius float64, center Location) [][]Location {
	// Get bounding box
	left, bottom, right, top := getUserBoundingBox(radius, center)

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
