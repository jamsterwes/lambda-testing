// Find the lat/long bounds of a box centered at (long, lat) with a s = radius
// Returns { left: number, right: number, top: number, bottom: number }
const getBoundingBox = (long, lat, radius) => {
    // See https://gis.stackexchange.com/questions/142326/calculating-longitude-length-in-miles
    const phi = lat * Math.PI / 180.0;
    const a = 3963.190592;
    const e = 0.081819191;
    const M = a * (1 - e * e) / Math.pow(1 - Math.pow(e * Math.sin(phi), 2), 1.5);
    const R = a * Math.cos(phi) / Math.pow(1 - Math.pow(e * Math.sin(phi), 2), 0.5);

    // M = length of 1rad latitude in mi
    // R = length of 1rad longitude in mi

    const width = radius * 180 / (Math.PI * M);
    const height = radius * 180 / (Math.PI * R);

    const left = long - width / 2;
    const right = long + width / 2;
    const top = lat + height / 2;
    const bottom = lat - height / 2;

    return { left, right, top, bottom };
};

// Returns XML text response
const getMapBox = async (long, lat, radius) => {
    // Get bounding box
    const { left, right, top, bottom } = getBoundingBox(long, lat, radius);

    // Now go fetch the XML
    const api_url = `https://overpass-api.de/api/interpreter`;
    const bbox = `${bottom},${left},${top},${right}`;
    const osm = await fetch(api_url, {
        method: 'POST',
        body: 'data=' + encodeURIComponent(`[out:json];(way["highway"="primary"](${bbox});way["highway"="secondary"](${bbox});way["highway"="tertiary"](${bbox});way["highway"="residential"](${bbox}););out geom;`)
    });
    return osm.json();
};

export const handler = async (event) => {
    // TODO implement
    const response = {
        statusCode: 200,
        body: await getMapBox(event['long'], event['lat'], 1.0)
    };
    return response;
};
