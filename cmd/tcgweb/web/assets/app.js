const state = {
  decks: [],
  currentDeck: null,
};

const deckSelect = document.getElementById("deckSelect");
const deckNotice = document.getElementById("deckNotice");
const deckMeta = document.getElementById("deckMeta");
const deckCards = document.getElementById("deckCards");
const battleHistory = document.getElementById("battleHistory");
const searchResults = document.getElementById("searchResults");
const statsGrid = document.getElementById("statsGrid");
const lossList = document.getElementById("lossList");
const deckStatus = document.getElementById("deckStatus");
const connectionStatus = document.getElementById("connectionStatus");
const themeSelect = document.getElementById("themeSelect");
const backgroundUpload = document.getElementById("backgroundUpload");
const clearBackground = document.getElementById("clearBackground");
const battleChart = document.getElementById("battleChart");
const battleChartNotice = document.getElementById("battleChartNotice");

const THEME_KEY = "tcgcli-theme";
const BG_KEY = "tcgcli-background";

function setStatus(text, variant = "info") {
  connectionStatus.textContent = text;
  connectionStatus.style.background = variant === "error" ? "rgba(248, 113, 113, 0.2)" : "rgba(56, 189, 248, 0.2)";
  connectionStatus.style.color = variant === "error" ? "#fecaca" : "#38bdf8";
}

function applyTheme(theme) {
  if (!theme || theme === "default") {
    document.documentElement.removeAttribute("data-theme");
    return;
  }
  document.documentElement.setAttribute("data-theme", theme);
}

function setCustomBackground(dataUrl) {
  if (!dataUrl) {
    return;
  }
  document.body.style.setProperty("--custom-bg", `url("${dataUrl}")`);
  if (clearBackground) {
    clearBackground.disabled = false;
  }
}

function clearCustomBackground() {
  document.body.style.setProperty("--custom-bg", "none");
  localStorage.removeItem(BG_KEY);
  if (backgroundUpload) {
    backgroundUpload.value = "";
  }
  if (clearBackground) {
    clearBackground.disabled = true;
  }
}

function loadAppearanceSettings() {
  const storedTheme = localStorage.getItem(THEME_KEY) || "default";
  if (themeSelect) {
    themeSelect.value = storedTheme;
    applyTheme(storedTheme);
  }

  const storedBackground = localStorage.getItem(BG_KEY);
  if (storedBackground) {
    setCustomBackground(storedBackground);
  } else {
    if (clearBackground) {
      clearBackground.disabled = true;
    }
  }
}

async function apiFetch(path, options = {}) {
  try {
    const response = await fetch(path, {
      headers: { "Content-Type": "application/json" },
      ...options,
    });
    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || "Request failed");
    }
    setStatus("Connected");
    return data;
  } catch (error) {
    setStatus(error.message, "error");
    throw error;
  }
}

function updateDeckSelect() {
  deckSelect.innerHTML = '<option value="">Select a deck</option>';
  state.decks.forEach((deck) => {
    const option = document.createElement("option");
    option.value = deck;
    option.textContent = deck;
    deckSelect.appendChild(option);
  });
}

function renderDeck(deck) {
  state.currentDeck = deck;
  deckStatus.textContent = deck.load_status || "ready";
  deckStatus.style.background = deck.cards_warning ? "rgba(248, 113, 113, 0.2)" : "rgba(74, 222, 128, 0.2)";
  deckStatus.style.color = deck.cards_warning ? "#fecaca" : "#4ade80";

  const warning = deck.cards_warning ? `<br /><span class="muted">Card data warning: ${deck.cards_warning}</span>` : "";
  deckMeta.innerHTML = `Cards source: ${deck.cards_source || "unknown"}.${warning}`;

  deckCards.innerHTML = "";
  if (deck.cards.length === 0) {
    deckCards.innerHTML = "<li class=\"notice\">No cards in this deck yet.</li>";
  } else {
    deck.cards.forEach((entry, index) => {
      const item = document.createElement("li");
      item.className = "card-item";
      item.innerHTML = `
        <header>
          <strong>${entry.name}</strong>
          <span class="muted">${entry.count}x</span>
        </header>
        <div class="muted">${entry.set}</div>
        <button type="button" class="danger" data-index="${index}">Remove</button>
      `;
      item.querySelector("button").addEventListener("click", () => removeCard(index));
      deckCards.appendChild(item);
    });
  }

  battleHistory.innerHTML = "";
  if (deck.battles.length === 0) {
    battleHistory.innerHTML = "<li class=\"notice\">No battles recorded yet.</li>";
  } else {
    deck.battles.slice().reverse().forEach((battle) => {
      const item = document.createElement("li");
      const isLoss = battle.result.toUpperCase() === "L";
      item.className = `battle-item ${isLoss ? "loss" : ""}`;
      item.innerHTML = `
        <strong>${battle.result === "W" ? "Win" : "Loss"}</strong>
        <span class="muted">${battle.date}</span>
        <div>${battle.opponent}</div>
      `;
      battleHistory.appendChild(item);
    });
  }

  statsGrid.innerHTML = "";
  const stats = deck.stats || {};
  const statItems = [
    { label: "Total Battles", value: stats.totalBattles ?? 0 },
    { label: "Wins", value: stats.wins ?? 0 },
    { label: "Losses", value: stats.losses ?? 0 },
    { label: "Win %", value: stats.winPercentage ? stats.winPercentage.toFixed(2) + "%" : "0%" },
  ];

  statItems.forEach((stat) => {
    const card = document.createElement("div");
    card.className = "stat";
    card.innerHTML = `<h4>${stat.label}</h4><p>${stat.value}</p>`;
    statsGrid.appendChild(card);
  });

  lossList.innerHTML = "";
  if (stats.lossByOpponent && Object.keys(stats.lossByOpponent).length > 0) {
    const list = document.createElement("ul");
    list.className = "card-list";
    Object.entries(stats.lossByOpponent).forEach(([opponent, count]) => {
      const item = document.createElement("li");
      item.className = "result-item";
      item.innerHTML = `<strong>${opponent}</strong><span class="muted">${count} loss(es)</span>`;
      list.appendChild(item);
    });
    lossList.appendChild(list);
  } else {
    lossList.textContent = "No loss breakdown yet.";
  }

  renderBattleChart(deck);
}

function renderBattleChart(deck) {
  if (!battleChart || !battleChartNotice) {
    return;
  }

  const battles = deck?.battles || [];
  battleChart.innerHTML = "";
  if (battles.length === 0) {
    battleChartNotice.textContent = "No battle results to chart yet.";
    return;
  }

  battleChartNotice.textContent = "";
  const width = 600;
  const height = 220;
  const padding = 28;
  let wins = 0;
  const points = battles.map((battle, index) => {
    if (battle.result === "W") {
      wins += 1;
    }
    const winRate = (wins / (index + 1)) * 100;
    return { index, winRate };
  });

  const maxX = Math.max(points.length - 1, 1);
  const xStep = (width - padding * 2) / maxX;
  const yScale = (height - padding * 2) / 100;
  const coords = points.map((point, index) => {
    const x = padding + index * xStep;
    const y = height - padding - point.winRate * yScale;
    return { x, y };
  });

  const axis = `
    <line class="chart-axis" x1="${padding}" y1="${padding}" x2="${padding}" y2="${height - padding}" />
    <line class="chart-axis" x1="${padding}" y1="${height - padding}" x2="${width - padding}" y2="${height - padding}" />
  `;

  const pathData = coords
    .map((point, index) => `${index === 0 ? "M" : "L"}${point.x},${point.y}`)
    .join(" ");

  const circles = coords
    .map((point) => `<circle cx="${point.x}" cy="${point.y}" r="4" />`)
    .join("");

  battleChart.innerHTML = `${axis}<path d="${pathData}" />${circles}`;
}

async function loadDecks() {
  const data = await apiFetch("/api/decks");
  state.decks = data.decks || [];
  updateDeckSelect();
}

async function loadDeck(name) {
  if (!name) {
    deckNotice.textContent = "Select a deck to load.";
    return;
  }
  const deck = await apiFetch(`/api/decks/${encodeURIComponent(name)}`);
  deckNotice.textContent = `Loaded ${deck.name}.`;
  renderDeck(deck);
}

async function createDeck(event) {
  event.preventDefault();
  const input = document.getElementById("deckName");
  const name = input.value.trim();
  if (!name) {
    deckNotice.textContent = "Deck name cannot be empty.";
    return;
  }
  const deck = await apiFetch("/api/decks", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
  deckNotice.textContent = `Created ${deck.name}.`;
  input.value = "";
  await loadDecks();
  deckSelect.value = deck.name;
  renderDeck(deck);
}

async function searchCards(event) {
  event.preventDefault();
  const term = document.getElementById("searchTerm").value.trim();
  searchResults.innerHTML = "<li class=\"notice\">Searching...</li>";
  const data = await apiFetch(`/api/cards?search=${encodeURIComponent(term)}`);
  const cards = data.cards || [];
  searchResults.innerHTML = "";
  if (cards.length === 0) {
    searchResults.innerHTML = "<li class=\"notice\">No cards found.</li>";
    return;
  }

  cards.slice(0, 50).forEach((card) => {
    const item = document.createElement("li");
    item.className = "result-item";
    item.innerHTML = `
      <header>
        <strong>${card.name}</strong>
        <button type="button">Add</button>
      </header>
      <div class="muted">${card.set}</div>
      <div class="muted">ID: ${card.id}</div>
    `;
    item.querySelector("button").addEventListener("click", () => addCard(card.id));
    searchResults.appendChild(item);
  });
}

async function addCard(cardID) {
  if (!state.currentDeck) {
    deckNotice.textContent = "Load a deck before adding cards.";
    return;
  }
  const deck = await apiFetch(`/api/decks/${encodeURIComponent(state.currentDeck.name)}/cards`, {
    method: "POST",
    body: JSON.stringify({ card_id: cardID }),
  });
  renderDeck(deck);
}

async function removeCard(index) {
  if (!state.currentDeck) {
    return;
  }
  const deck = await apiFetch(`/api/decks/${encodeURIComponent(state.currentDeck.name)}/cards/${index}`, {
    method: "DELETE",
  });
  renderDeck(deck);
}

async function recordBattle(event) {
  event.preventDefault();
  if (!state.currentDeck) {
    deckNotice.textContent = "Load a deck before recording battles.";
    return;
  }
  const result = document.getElementById("battleResult").value;
  const opponent = document.getElementById("opponent").value.trim();
  const deck = await apiFetch(`/api/decks/${encodeURIComponent(state.currentDeck.name)}/battles`, {
    method: "POST",
    body: JSON.stringify({ result, opponent }),
  });
  document.getElementById("opponent").value = "";
  renderDeck(deck);
}

async function init() {
  setStatus("Connecting...");
  loadAppearanceSettings();
  await loadDecks();
  setStatus("Connected");
}

document.getElementById("createDeckForm").addEventListener("submit", createDeck);
document.getElementById("searchForm").addEventListener("submit", searchCards);
document.getElementById("battleForm").addEventListener("submit", recordBattle);
deckSelect.addEventListener("change", (event) => loadDeck(event.target.value));

if (themeSelect) {
  themeSelect.addEventListener("change", (event) => {
    const nextTheme = event.target.value;
    localStorage.setItem(THEME_KEY, nextTheme);
    applyTheme(nextTheme);
  });
}

if (backgroundUpload) {
  backgroundUpload.addEventListener("change", (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = reader.result;
      if (typeof dataUrl === "string") {
        localStorage.setItem(BG_KEY, dataUrl);
        setCustomBackground(dataUrl);
      }
    };
    reader.readAsDataURL(file);
  });
}

if (clearBackground) {
  clearBackground.addEventListener("click", () => {
    clearCustomBackground();
  });
}

init();
