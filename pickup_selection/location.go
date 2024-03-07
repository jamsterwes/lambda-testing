package main

import "math"

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

// Constants
const GIS_A = 3963.190592
const GIS_E = 0.081819191

// GIS helper function
func milesToDegLatitude(miles float64, latitude float64) float64 {
	phi := latitude * (math.Pi / 180)

	// M = length of 1 radian of latitude in miles
	M := GIS_A * (1 - GIS_E*GIS_E) / math.Pow((1-math.Pow(GIS_E*math.Sin(phi), 2)), 1.5)

	// Convert miles to degrees latitude
	return miles * 180 / (math.Pi * M)
}

// GIS helper function
func milesToDegLongitude(miles float64, latitude float64) float64 {
	phi := latitude * (math.Pi / 180)

	// N = length of 1 radian of longitude in miles
	N := GIS_A * math.Cos(phi) / math.Pow(1-math.Pow(GIS_E*math.Sin(phi), 2), 0.5)

	// Convert miles to degrees longitude
	return miles * 180 / (math.Pi * N)
}

// size - the size of the bounding box in miles
// center - the center of the bounding box
// Returns: the bounding box in the form of
// .. (left, bottom, right, top)
func getUserBoundingBox(size float64, center Location) (float64, float64, float64, float64) {
	// Convert miles to degrees latitude and longitude
	degLat := milesToDegLatitude(size, center.Latitude)
	degLong := milesToDegLongitude(size, center.Latitude)

	// Calculate the bounding box
	left := center.Longitude - degLong/2
	bottom := center.Latitude - degLat/2
	right := center.Longitude + degLong/2
	top := center.Latitude + degLat/2

	return left, bottom, right, top
}
