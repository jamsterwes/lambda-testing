package main

import "math"

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
func intersectWayRing(wayGeom []Location, radius float64, center Location) []Location {
	// Step 0. Ignore empty geom
	if len(wayGeom) == 0 {
		return []Location{}
	}

	// Step 1. Get the two radii of the ellipse
	a := milesToDegLongitude(radius, center.Latitude)
	b := milesToDegLatitude(radius, center.Latitude)

	// Step 2. Accumulate points
	var points []Location
	for i := range wayGeom[:len(wayGeom)-1] {
		// Get line segment
		x1 := wayGeom[i].Longitude
		y1 := wayGeom[i].Latitude
		x2 := wayGeom[i+1].Longitude
		y2 := wayGeom[i+1].Latitude

		// Get solutions and concat
		solutions := intersectLineRing(center.Longitude, center.Latitude, a, b, x1, y1, x2, y2)
		points = append(points, solutions...)
	}

	return points
}

// Cull by angle
// - numberSegments int: number of segments to divide the circle into
// - pointsPerSegment int: number of points to keep per segment
func cullByAngle(points []Location, center Location, numberSegments int, pointsPerSegment int) []Location {
	// Step 0. Ignore empty points
	if len(points) == 0 {
		return []Location{}
	}

	// Step 1. Allocate space for the segment indexes
	segmentIndexes := make([][]int, numberSegments)

	// Step 2. Sort each point into its segment
	for i, point := range points {
		// Get angle of point
		angle := math.Atan2(point.Latitude-center.Latitude, point.Longitude-center.Longitude)

		// Wrap angle to [0, 2pi)
		if angle < 0 {
			angle += 2 * math.Pi
		}

		// Turn angle into segment index
		segmentIndex := int(angle / (2 * math.Pi / float64(numberSegments)))

		// Safety: ensure segmentIndex is within bounds
		segmentIndex = segmentIndex % numberSegments

		// Ignore if segment is full
		if len(segmentIndexes[segmentIndex]) >= pointsPerSegment {
			continue
		}

		// Now append this index into segmentIndexes
		segmentIndexes[segmentIndex] = append(segmentIndexes[segmentIndex], i)
	}

	// Step 3. Extract the points from the segment indexes
	var culledPoints []Location
	for _, segment := range segmentIndexes {
		for _, index := range segment {
			culledPoints = append(culledPoints, points[index])
		}
	}

	return culledPoints
}
