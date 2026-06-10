const searchInput = document.getElementById("searchInput");
const searchResults = document.getElementById("searchResults");
const artistsGrid = document.getElementById("artistsGrid");
const filtersForm = document.querySelector(".filters-form");

if (searchInput && searchResults && artistsGrid) {
  let debounceTimer;

  // Task 08: listen to user typing and trigger the client-server event.
  searchInput.addEventListener("input", () => {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      fetchArtists();
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
  }
}

function renderArtists(artists) {
  if (!artists.length) {
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

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}
