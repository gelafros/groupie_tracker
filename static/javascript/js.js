const searchInput = document.getElementById("searchInput");
const searchResults = document.getElementById("searchResults");
const artistsGrid = document.getElementById("artistsGrid");
const filtersForm = document.querySelector(".filters-form");
const searchSuggestions = document.getElementById("searchSuggestions");

let activeScope = "all";

// Setup scope button listeners
const scopeButtons = document.querySelectorAll(".scope-btn");
scopeButtons.forEach(btn => {
  btn.addEventListener("click", () => {
    scopeButtons.forEach(b => b.classList.remove("active"));
    btn.classList.add("active");
    activeScope = btn.getAttribute("data-scope");
    
    // Immediately refresh suggestions and filtered artists on scope change
    if (searchInput) {
      const query = searchInput.value.trim();
      fetchArtists();
      fetchSuggestions(query);
    }
  });
});

if (searchInput && searchSuggestions && searchResults && artistsGrid) {
  let debounceTimer;

  // Listen to user typing and trigger the client-server event.
  searchInput.addEventListener("input", () => {
    const query = searchInput.value.trim();

    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      fetchArtists();
      fetchSuggestions(query);
    }, 250);
  });

  // Hide suggestions dropdown when clicking outside
  document.addEventListener("click", (e) => {
    if (!searchInput.contains(e.target) && !searchSuggestions.contains(e.target)) {
      searchSuggestions.classList.remove("show");
    }
  });

  // Show suggestions dropdown on focus if input has value
  searchInput.addEventListener("focus", () => {
    if (searchInput.value.trim().length > 0 && searchSuggestions.children.length > 0) {
      searchSuggestions.classList.add("show");
    }
  });
}

if (filtersForm) {
  // Intercept the filter form submission to stop the page reload
  filtersForm.addEventListener("submit", (e) => {
    e.preventDefault();
    fetchArtists();
  });

  let filterDebounceTimer;
  // Listen to any input changes inside the form (checkboxes, number inputs)
  filtersForm.addEventListener("input", () => {
    clearTimeout(filterDebounceTimer);
    filterDebounceTimer = setTimeout(() => {
      fetchArtists();
    }, 250);
  });

  // Intercept the reset button to clear everything asynchronously
  const resetBtn = filtersForm.querySelector(".btn-reset");
  if (resetBtn) {
    resetBtn.addEventListener("click", (e) => {
      e.preventDefault(); // Prevent standard link navigation
      filtersForm.reset();
      if (searchInput) searchInput.value = "";
      
      // Reset scope back to all
      activeScope = "all";
      scopeButtons.forEach(b => b.classList.remove("active"));
      const allBtn = document.querySelector('.scope-btn[data-scope="all"]');
      if (allBtn) allBtn.classList.add("active");

      // Hide suggestions
      if (searchSuggestions) {
        searchSuggestions.innerHTML = "";
        searchSuggestions.classList.remove("show");
      }

      fetchArtists();
    });
  }
}

async function fetchArtists() {
  // Grab all the checked boxes and typed years directly from the form
  const params = filtersForm ? new URLSearchParams(new FormData(filtersForm)) : new URLSearchParams();

  // Add the search query to the params if it exists
  const query = searchInput ? searchInput.value.trim() : "";
  if (query) {
    params.set("q", query);
  }

  // Add the search scope if not "all"
  if (activeScope && activeScope !== "all") {
    params.set("scope", activeScope);
  }

  try {
    searchResults.textContent = "Loading...";

    // Task 08: call the backend search endpoint without reloading the page.
    const response = await fetch(`/api/search?${params.toString()}`);
    if (!response.ok) {
      throw new Error("Search request failed");
    }

    const artists = await response.json();
    renderArtists(artists);
    renderSearchMessage(query, artists.length);
  } catch (error) {
    searchResults.textContent = "Search is temporarily unavailable.";
    if (artistsGrid) artistsGrid.innerHTML = '<p class="error-state">Failed to load artists. Please try again later.</p>';
  }
}

function renderArtists(artists) {
  if (!Array.isArray(artists) || artists.length === 0) {
    artistsGrid.innerHTML = '<p class="empty-state">No matching artists found.</p>';
    return;
  }

  // Task 08: redraw only the artist grid with the returned JSON results.
  artistsGrid.innerHTML = artists
    .map(
      (artist) => `
        <a class="artist-card" href="/artist?id=${artist.id}">
          <img class="card-image" src="${escapeHtml(artist.image)}" alt="${escapeHtml(artist.name)}" loading="lazy">
          <div class="card-body">
            <h2 class="card-name">${escapeHtml(artist.name)}</h2>
            <p class="card-date">Since ${artist.creationDate}</p>
          </div>
        </a>
      `
    )
    .join("");
}

function renderSearchMessage(query, resultCount) {
  if (!query) {
    searchResults.textContent = `Showing all ${resultCount} artists.`;
    return;
  }

  if (resultCount === 0) {
    searchResults.textContent = `No results for "${query}".`;
    return;
  }

  searchResults.textContent = `${resultCount} result(s) for "${query}".`;
}

// Async function to fetch search suggestions from the server
async function fetchSuggestions(query) {
  // STEP 1: Initial checks
  if (!searchSuggestions) {
    return; // If suggestions box does not exist on the page, exit.
  }

  if (query.length === 0) {
    searchSuggestions.innerHTML = "";
    searchSuggestions.classList.remove("show");
    return;
  }

  // STEP 2: Request data from server (inside try-catch to handle network errors gracefully)
  try {
    // Construct URL with URL-encoded query parameter
    let url = "/api/suggestions?q=" + encodeURIComponent(query);
    if (activeScope && activeScope !== "all") {
      url += "&scope=" + encodeURIComponent(activeScope);
    }

    // Send request to server and await response
    let response = await fetch(url);
    if (!response.ok) {
      throw new Error("Suggestions request failed");
    }

    // Parse response body as JSON
    let suggestionsList = await response.json();

    if (suggestionsList.length === 0) {
      searchSuggestions.innerHTML = "";
      searchSuggestions.classList.remove("show");
      return;
    }

    // STEP 3: Construct the HTML suggestions list
    let newHtml = "";

    // Loop through each suggestion sent by the server
    for (let item of suggestionsList) {
      let val = item.value;      // E.g., "Phil Collins"
      let type = item.type;      // E.g., "member"
      let info = item.info;      // E.g., "Genesis"
      let id = item.artistId;    // E.g., 10

      // Highlight the matching part in the value (case-insensitive)
      let highlightedVal = escapeHtml(val);
      if (query) {
        const escapedQuery = escapeRegExp(query);
        const regex = new RegExp(`(${escapedQuery})`, "gi");
        highlightedVal = escapeHtml(val).replace(regex, '<span class="match-highlight">$1</span>');
      }

      let infoSpan = info ? ` <span class="match-info">(${escapeHtml(info)})</span>` : "";

      newHtml += `
        <div class="suggestion-item" data-artist-id="${id}">
          <div class="suggestion-left">
            <span class="match-value">${highlightedVal}</span>${infoSpan}
          </div>
          <span class="match-type">${escapeHtml(type)}</span>
        </div>
      `;
    }

    // STEP 4: Render the suggestions list to page
    searchSuggestions.innerHTML = newHtml;
    searchSuggestions.classList.add("show");

    // Add click listeners to items to navigate directly to the artist detail
    const items = searchSuggestions.querySelectorAll(".suggestion-item");
    items.forEach(item => {
      item.addEventListener("click", () => {
        const artistId = item.getAttribute("data-artist-id");
        if (artistId) {
          window.location.href = "/artist?id=" + artistId;
        }
      });
    });

  } catch (error) {
    console.error("Error loading suggestions:", error);
  }
}

function escapeRegExp(string) {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}
