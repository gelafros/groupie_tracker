document.addEventListener("DOMContentLoaded", () => {
    const relationItems = document.querySelectorAll(".relation-item");
    const events = [];

    relationItems.forEach(item => {
        let locText = item.querySelector(".relation-location").innerText;
        // Convert something like "north-carolina-usa" to "north carolina usa"
        let cleanLoc = locText.replace(/[-_]/g, ' ');
        
        const dateItems = item.querySelectorAll(".relation-dates li");
        dateItems.forEach(dateItem => {
            // Remove asterisk from dates if present
            let dateText = dateItem.innerText.replace('*', '').trim(); 
            let parts = dateText.split('-');
            let dateObj;
            if (parts.length === 3) {
                // DD-MM-YYYY format
                dateObj = new Date(parts[2], parts[1] - 1, parts[0]);
            } else {
                dateObj = new Date(); // fallback
            }
            events.push({
                locationStr: locText,
                cleanLocation: cleanLoc,
                dateStr: dateText,
                dateObj: dateObj
            });
        });
    });

    // Sort chronologically
    events.sort((a, b) => a.dateObj - b.dateObj);

    // Custom mappings to improve Nominatim geocoding accuracy
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

    // Initialize map
    const mapContainer = document.getElementById("map");
    if (!mapContainer) return;
    
    const map = L.map('map').setView([20, 0], 2);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap contributors'
    }).addTo(map);

    const polyline = L.polyline([], {color: 'red', dashArray: '5, 5', weight: 2}).addTo(map);
    const timelineDiv = document.getElementById("timeline-list");

    async function geocode(location) {
        let query = customQueryMap[location.toLowerCase()] || location;
        const url = `https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=1`;
        try {
            const response = await fetch(url);
            const data = await response.json();
            if (data && data.length > 0) {
                return { lat: parseFloat(data[0].lat), lon: parseFloat(data[0].lon) };
            }
        } catch (e) {
            console.error("Geocoding failed for", location, e);
        }
        return null;
    }

    async function processEvents() {
        let count = 1;
        const locationCache = {};
        
        for (const event of events) {
            let coords = locationCache[event.cleanLocation];
            
            if (!coords) {
                coords = await geocode(event.cleanLocation);
                if (coords) {
                    locationCache[event.cleanLocation] = coords;
                }
            }

            if (coords) {
                // Add marker
                const marker = L.marker([coords.lat, coords.lon]).addTo(map);
                
                // Tooltip and Popup
                const popupContent = `<b>#${count} ${event.cleanLocation.toUpperCase()}</b><br>Date: ${event.dateStr}`;
                marker.bindPopup(popupContent);
                
                // Add point to polyline dynamically
                polyline.addLatLng([coords.lat, coords.lon]);

                // Fit bounds dynamically so the user can see the path growing
                if (count === 1) {
                    map.setView([coords.lat, coords.lon], 4);
                } else {
                    map.fitBounds(polyline.getBounds(), { padding: [50, 50], maxZoom: 5 });
                }

                // Append to timeline
                if (timelineDiv) {
                    const li = document.createElement("li");
                    li.innerText = `#${count}: ${event.cleanLocation} (${event.dateStr})`;
                    timelineDiv.appendChild(li);
                    // auto-scroll timeline
                    timelineDiv.scrollTop = timelineDiv.scrollHeight;
                }

                count++;
            }

            // Always wait 1 second between processing events to respect Nominatim API rate limits
            // and to animate the drawing progressively, even if cached.
            await new Promise(r => setTimeout(r, 1000));
        }

        // Final fit bounds to ensure all are perfectly visible
        if (polyline.getLatLngs().length > 1) {
            map.fitBounds(polyline.getBounds(), { padding: [20, 20] });
        }
    }

    processEvents();
});
