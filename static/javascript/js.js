const searchInput = document.getElementById("searchInput");
const searchResults = document.getElementById("searchResults");
const artistsGrid = document.getElementById("artistsGrid");
const filtersForm = document.querySelector(".filters-form");
const searchSuggestions = document.getElementById("searchSuggestions");

if (searchInput && searchResults && artistsGrid) {
  let debounceTimer;

  // Listen to user typing and trigger the client-server event.
  searchInput.addEventListener("input", () => {
    const query = searchInput.value.trim();

    // Check if the search suggestions element exists
    if (searchSuggestions) {

      // Get the current text typed by the user
      let typedText = searchInput.value;

      // Retrieve all available options from the suggestions datalist
      let allOptions = searchSuggestions.querySelectorAll("option");

      // Check options one by one
      for (let option of allOptions) {

        let optionText = option.value;

        // If option text matches user input exactly
        if (optionText === typedText) {

          // Found match! Retrieve the hidden artist ID
          let artistId = option.getAttribute("data-artist-id");

          // Redirect user to artist details page using this ID
          if (artistId) {
            window.location.href = "/artist?id=" + artistId;
          }

          return;
        }
      }
    }

    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      fetchArtists();
      fetchSuggestions(query);
    }, 250);
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

  if (query.length < 3) {
    searchSuggestions.innerHTML = ""; // Empty suggestions if query is less than 3 characters.
    return;
  }

  // STEP 2: Request data from server (inside try-catch to handle network errors gracefully)
  try {
    // Construct URL with URL-encoded query parameter
    let url = "/api/suggestions?q=" + encodeURIComponent(query);

    // Send request to server and await response
    let response = await fetch(url);
    if (!response.ok) {
      throw new Error("Suggestions request failed");
    }

    // Parse response body as JSON
    let suggestionsList = await response.json();

    // STEP 3: Construct the HTML options list
    let newHtml = "";

    // Loop through each suggestion sent by the server
    for (let item of suggestionsList) {
      let text = item.displayText; // E.g., "Phil Collins - member"
      let id = item.artistId;      // E.g., 10

      // Append a formatted option tag
      newHtml = newHtml + `<option value="${escapeHtml(text)}" data-artist-id="${id}"></option>`;
    }

    // STEP 4: Render the suggestions list to page
    searchSuggestions.innerHTML = newHtml;

  } catch (error) {
    console.error("Error loading suggestions:", error);
  }
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}
