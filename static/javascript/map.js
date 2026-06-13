document.addEventListener("DOMContentLoaded", () => {
    // 1. DATA EXTRACTION
    // Select all hidden container elements containing raw concert data structured by Go
    const relationItems = document.querySelectorAll(".relation-item");
    const events = [];

    relationItems.forEach(item => {
        // Retrieve the location string (e.g., "north-carolina-usa")
        let locText = item.querySelector(".relation-location").innerText;
        
        // Clean the location: replace underscores/hyphens with spaces for better geocoding compatibility
        let cleanLoc = locText.replace(/[-_]/g, ' ');
        
        // Retrieve and process each concert date listed under this location
        const dateItems = item.querySelectorAll(".relation-dates li");
        dateItems.forEach(dateItem => {
            // Strip any leading asterisks if present in the source dataset
            let dateText = dateItem.innerText.replace('*', '').trim(); 
            
            // Parse date in DD-MM-YYYY format
            let parts = dateText.split('-');
            let dateObj;
            if (parts.length === 3) {
                // Javascript Date constructor uses 0-indexed months (0 = January, 11 = December)
                dateObj = new Date(parts[2], parts[1] - 1, parts[0]);
            } else {
                // Fallback to current date if format is invalid
                dateObj = new Date(); 
            }
            
            // Store the compiled event object
            events.push({
                locationStr: locText,      // Raw location name (used in CSS/HTML queries)
                cleanLocation: cleanLoc,  // Normalized name (used for geocoding queries)
                dateStr: dateText,        // Human-readable date string
                dateObj: dateObj          // Date object (used for sorting)
            });
        });
    });

    // 2. CHRONOLOGICAL SORTING
    // Sort all concert events in ascending order to reveal the band's chronological travel path
    events.sort((a, b) => a.dateObj - b.dateObj);

    // 3. GEOCODING DICTIONARY MAP
    // Manual mapping for tricky or shorthand locations that Nominatim might fail to resolve accurately.
    // Maps informal/implicit location strings to precise, structured address strings.
    const customQueryMap = {
        "penrose new zealand": "Penrose, Auckland, New Zealand",
        "dunedin new zealand": "Dunedin, New Zealand",
        "del mar usa": "Del Mar, CA, USA",
        "canton usa": "Canton, OH, USA",
        "oakland us": "Oakland, CA, USA",
        "los angeles usa": "Los Angeles, CA, USA",
        "inglewood usa": "Inglewood, CA, USA",
        "georgia usa": "Georgia, USA",
        "north carolina usa": "North Carolina, USA",
        "las vegas usa": "Las Vegas, NV, USA",
        "new york usa": "New York City, NY, USA",
        "california usa": "California, USA",
        "nagoya japan": "Nagoya, Japan",
        "osaka japan": "Osaka, Japan",
        "saitama japan": "Saitama, Japan",
        "dusseldorf germany": "Düsseldorf, Germany",
        "aarhus denmark": "Aarhus, Denmark",
        "quebec canada": "Quebec City, Canada",
        "monterrey mexico": "Monterrey, Mexico",
        "mexico city mexico": "Mexico City, Mexico",
        "rio de janeiro brazil": "Rio de Janeiro, Brazil",
        "riyadh saudi arabia": "Riyadh, Saudi Arabia",
        "madrid spain": "Madrid, Spain",
        "berlin germany": "Berlin, Germany",
        "manchester uk": "Manchester, UK"
    };

    // 4. MAP & COMPONENT INITIALIZATION
    const mapContainer = document.getElementById("map");
    if (!mapContainer) return; // Exit if map element is not present on this page
    
    // Initialize the Leaflet map centered at a default wide view coordinates
    const map = L.map('map').setView([20, 0], 2);
    
    // Load and render standard OpenStreetMap tile layers
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap contributors'
    }).addTo(map);

    // Create a dashed red polyline to depict travel paths and add it to the map
    const polyline = L.polyline([], {color: 'red', dashArray: '5, 5', weight: 2}).addTo(map);
    const timelineDiv = document.getElementById("timeline-list");

    // 5. ASYNCHRONOUS GEOLOCALIZER
    // Performs geocoding by querying the Nominatim OpenStreetMap REST API.
    async function geocode(location) {
        // Check if a custom mapped address exists; if not, query the cleaned location string
        let query = customQueryMap[location.toLowerCase()] || location;
        const url = `https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=1`;
        
        try {
            const response = await fetch(url);
            const data = await response.json();
            if (data && data.length > 0) {
                // Return latitude and longitude as floats
                return { lat: parseFloat(data[0].lat), lon: parseFloat(data[0].lon) };
            }
        } catch (e) {
            console.error("Geocoding failed for:", location, e);
        }
        return null; // Return null if API fetch failed or no matches found
    }

    // 6. MAIN CHRONOLOGICAL PROCESSOR LOOP
    // Iterates through events, geocodes locations, renders markers, builds polylines, and updates the timeline.
    async function processEvents() {
        let count = 1;
        const locationCache = {}; // Cache map to store coordinates and prevent duplicate API queries for identical locations
        
        for (const event of events) {
            // Check cache memory first to optimize speed and API usage
            let coords = locationCache[event.cleanLocation];
            
            if (!coords) {
                // Request geolocalization coordinates from the API
                coords = await geocode(event.cleanLocation);
                if (coords) {
                    // Save to local cache for future iterations of identical locations
                    locationCache[event.cleanLocation] = coords;
                }
            }

            if (coords) {
                // A. Render Leaflet Marker
                const marker = L.marker([coords.lat, coords.lon]).addTo(map);
                
                // B. Construct Marker Popup contents
                const popupContent = `<b>#${count} ${event.cleanLocation.toUpperCase()}</b><br>Date: ${event.dateStr}`;
                marker.bindPopup(popupContent);
                
                // C. Append coordinates to the polyline to draw the chronological connector line
                polyline.addLatLng([coords.lat, coords.lon]);

                // D. Update Map Bounds dynamically so the view expands as coordinates are populated
                if (count === 1) {
                    // Zoom into first location initially
                    map.setView([coords.lat, coords.lon], 4);
                } else {
                    // Fit map bounds to contain the entire drawn path with comfortable padding
                    map.fitBounds(polyline.getBounds(), { padding: [50, 50], maxZoom: 5 });
                }

                // E. Append entry dynamically to the HTML timeline container
                if (timelineDiv) {
                    const li = document.createElement("li");
                    li.innerText = `#${count}: ${event.cleanLocation} (${event.dateStr})`;
                    timelineDiv.appendChild(li);
                    // Automatically scroll to the bottom of the timeline feed
                    timelineDiv.scrollTop = timelineDiv.scrollHeight;
                }

                count++;
            }

            // F. 1-SECOND DELAY ENFORCER
            // Mandatory delay to respect OpenStreetMap Nominatim guidelines (max 1 req/sec).
            // Additionally provides a progressive animation effect where elements render step-by-step.
            await new Promise(r => setTimeout(r, 1000));
        }

        // G. FINAL ZOOM ALIGNMENT
        // Re-fit the map view bounds after all locations have been successfully loaded and drawn.
        if (polyline.getLatLngs().length > 1) {
            map.fitBounds(polyline.getBounds(), { padding: [20, 20] });
        }
    }

    // Trigger the main process
    processEvents();
});
