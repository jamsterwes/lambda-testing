const milesToLatitude = (mi, latitude) => {
    // See https://gis.stackexchange.com/questions/142326/calculating-longitude-length-in-miles
    const phi = latitude * Math.PI / 180.0;
    const a = 3963.190592;
    const e = 0.081819191;

    // M = length of 1rad latitude in mi
    const M = a * (1 - e * e) / Math.pow(1 - Math.pow(e * Math.sin(phi), 2), 1.5);

    // Convert mi miles to deg latitude
    const lat_deg = mi * 180 / (Math.PI * M);
    return lat_deg;
}

const milesToLongitude = (mi, latitude) => {
    // See https://gis.stackexchange.com/questions/142326/calculating-longitude-length-in-miles
    const phi = latitude * Math.PI / 180.0;
    const a = 3963.190592;
    const e = 0.081819191;

    // R = length of 1rad longitude in mi
    const R = a * Math.cos(phi) / Math.pow(1 - Math.pow(e * Math.sin(phi), 2), 0.5);

    // Convert mi miles to deg latitude
    const long_deg = mi * 180 / (Math.PI * R);
    return long_deg;
}

// Find the lat/long bounds of a box centered at (long, lat) with sideLength
// Returns { left: number, right: number, top: number, bottom: number }
const getBoundingBox = (long, lat, sideLength) => {
    // Convert miles to deg lat/long
    const width = milesToLongitude(sideLength, lat);
    const height = milesToLatitude(sideLength, lat);

    const left = long - width / 2;
    const right = long + width / 2;
    const top = lat + height / 2;
    const bottom = lat - height / 2;

    return { left, right, top, bottom };
};

// Returns [ [ {lat: number, lon: number} ] ]
const getStreetGeom = async (long, lat, radius) => {
    // Get bounding box
    const { left, right, top, bottom } = getBoundingBox(long, lat, radius);

    // Now go fetch the JSON
    const api_url = `https://overpass-api.de/api/interpreter`;
    const bbox = `${bottom},${left},${top},${right}`;

    // Query is URL-encoded, explanation of query:
    // .. [out:json]; -- so that we get JSON not XML
    // .. (  -- start a group of selections
    // ..   way["highway"="primary"](${{bbox}}); -- get all "primary" roads in bbox (1mi x 1mi square around user)
    // ..   way["highway"="secondary"](${{bbox}}); -- get all "secondary" roads in bbox (1mi x 1mi square around user)
    // ..   way["highway"="tertiary"](${{bbox}}); -- get all "tertiary" roads in bbox (1mi x 1mi square around user)
    // ..   way["highway"="residential"](${{bbox}}); -- get all "residential" roads in bbox (1mi x 1mi square around user)
    // ..   way["highway"="service"](${{bbox}}); -- get all "service" roads (parking lots) in bbox (1mi x 1mi square around user)
    // ..   way["highway"="unclassified"](${{bbox}}); -- get all "unclassified" roads in bbox (1mi x 1mi square around user)
    // .. ); -- close the group
    // .. out geom;  -- return only the geometry of the roads, not the buildings connected to them
    const query = `[out:json];(way["highway"="primary"](${bbox});way["highway"="secondary"](${bbox});way["highway"="tertiary"](${bbox});way["highway"="residential"](${bbox});way["highway"="service"](${bbox});way["highway"="unclassified"](${bbox}););out geom;`;
    const osm = await fetch(api_url, {
        method: 'POST',
        body: 'data=' + encodeURIComponent(query)
    });
    const ways = await osm.json();
    // Only really need the geometry
    return ways['elements'].map(way => way.geometry);
};

// Intersect line segment with ring
// All coordinates in degrees lat/long
// cx, cy - center (long = x, lat = y)
// a, b - ellipse parameters
// x1, y1, x2, y2 - line segment
const intersectLineRing = (cx, cy, a, b, x1, y1, x2, y2) => {
    // Step 1. Set origin + scale
    const X1 = (x1 - cx) / a;
    const X2 = (x2 - cx) / a;
    const Y1 = (y1 - cy) / b;
    const Y2 = (y2 - cy) / b;

    // Step 2. Solve

    // Step 2a. Get quadratic coefficients
    const A = Math.pow(X2 - X1, 2) + Math.pow(Y2 - Y1, 2);
    const B = 2 * ( X1 * (X2 - X1) + Y1 * (Y2 - Y1) );
    const C = X1 * X1 + Y1 * Y1 - 1;

    // Step 2b. Find discriminant
    const D = B * B - 4 * A * C;

    // Step 2c. Get # of solutions
    let ts = [];
    if (D < 0) return [];
    if (D > 0)
    {
        // console.log((-B + Math.sqrt(D)) / (2.0 * A));
        ts.push((-B + Math.sqrt(D)) / (2.0 * A));
        ts.push((-B - Math.sqrt(D)) / (2.0 * A));
    }
    else 
    {
        ts.push(-B / (2.0 * A));
    }

    // Step 3. Convert t values into coordinates
    let solutions = [];
    for (let i = 0; i < ts.length; i++)
    {
        // Ignore t outside (0,1)
        if (ts[i] < 0 || ts[i] > 1) continue;

        // Get coordinates
        solutions.push({
            'lon': (x2 - x1) * ts[i] + x1,
            'lat': (y2 - y1) * ts[i] + y1
        });
    }
    return solutions;
}

// Intersect way with ring
// (radius in miles)
const intersectWayRing = (wayGeom, long, lat, radius) => {
    // Step 1: Get the two radii of the ellipse
    const x_radii = milesToLongitude(radius, lat);
    const y_radii = milesToLatitude(radius, lat);

    // Step 2: Accumulate points
    let points = [];
    for (let i = 0; i < wayGeom.length - 1; i++)
    {
        // Get line segment
        const x1 = wayGeom[i]['lon'];
        const y1 = wayGeom[i]['lat'];
        const x2 = wayGeom[i+1]['lon'];
        const y2 = wayGeom[i+1]['lat'];

        // Get solutions
        const solutions = intersectLineRing(long, lat, x_radii, y_radii, x1, y1, x2, y2);

        // Add them to points
        points = points.concat(solutions);
    }

    // Return points
    return points;
}

// Only take ten points from each ring
const cullRing = (points, maxCount) => {
    // If we are under budget, keep them
    if (points.length < maxCount) return points;

    // Randomize points (in-place)
    points.sort((a, b) => (Math.random() > 0.5) ? 1 : -1);

    // Return the top maxCount points
    return points.slice(0, maxCount);
}

// Set ring radii
const RING_RADII = [0.1, 0.25, 0.5, 0.75];
const RING_SIZE = [10, 20, 30, 40];

export const handler = async (event) => {
    // Step 0: unpack request
    const long = event['long'];
    const lat = event['lat'];

    // Step 1: Get geometry
    const wayGeoms = await getStreetGeom(long, lat, 2.0);

    // Step 2: Loop through ways
    let points = [];
    for (let r = 0; r < RING_RADII.length; r++)
    {
    // let ring = [];
        for (let i = 0; i < wayGeoms.length; i++)
        {
            const wayPoints = intersectWayRing(wayGeoms[i], long, lat, RING_RADII[r]);
            points = points.concat(wayPoints);
        }
        // points = points.concat(cullRing(ring, RING_SIZE[r]));
    }
    
    // Step 3: Convert points to array of arrays
    const pointsArray = points.map(point => [point.lon, point.lat]);

    const response = {
        statusCode: 200,
        body: {
            pointCount: points.length,
            points: pointsArray
        }
    };
    return response;
};
