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

function formatBattleTimestamp(value) {
  if (!value) {
    return "";
  }
  const [datePart, timePart] = value.split(" ");
  if (!datePart) {
    return value;
  }
  const [year, month, day] = datePart.split("-");
  if (!month || !day) {
    return value;
  }
  const formattedDate = `${Number(month)}/${Number(day)}`;
  if (!timePart) {
    return formattedDate;
  }
  const [hour, minute] = timePart.split(":");
  if (!hour || !minute) {
    return formattedDate;
  }
  return `${formattedDate} ${hour}:${minute}`;
}

function splitOpponent(opponent) {
  if (!opponent) {
    return { name: "Unknown", details: "" };
  }
  const separators = [" — ", " - ", " | ", ": "];
  for (const separator of separators) {
    if (opponent.includes(separator)) {
      const [name, ...rest] = opponent.split(separator);
      return {
        name: name.trim() || "Unknown",
        details: rest.join(separator).trim(),
      };
    }
  }
  return { name: opponent.trim(), details: "" };
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
      const opponent = splitOpponent(battle.opponent);
      const opponentDetails = opponent.details ? `<span class="muted">${opponent.details}</span>` : "";
      item.className = `battle-item ${isLoss ? "loss" : ""}`;
      item.innerHTML = `
        <strong>${battle.result === "W" ? "Win" : "Loss"}</strong>
        <span class="muted">${formatBattleTimestamp(battle.date)}</span>
        <div class="battle-opponent">
          <span>${opponent.name}</span>
          ${opponentDetails}
        </div>
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
      const opponentInfo = splitOpponent(opponent);
      const details = opponentInfo.details || "No details";
      item.innerHTML = `
        <strong>${opponentInfo.name}</strong>
        <div class="result-meta">
          <span>${details}</span>
          <span>${count} loss(es)</span>
        </div>
      `;
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
  const height = 260;
  const paddingX = 40;
  const paddingY = 32;
  const widthPerPoint = 64;
  const minWidth = 640;
  let wins = 0;
  let losses = 0;

  const points = battles.map((battle, index) => {
    if (battle.result === "W") {
      wins += 1;
    } else if (battle.result === "L") {
      losses += 1;
    }
    return {
      index,
      wins,
      losses,
      date: formatBattleTimestamp(battle.date),
    };
  });

  const maxX = Math.max(points.length - 1, 1);
  const width = Math.max(minWidth, paddingX * 2 + maxX * widthPerPoint);
  const maxY = Math.max(1, ...points.map((point) => Math.max(point.wins, point.losses)));
  const yScale = (height - paddingY * 2) / maxY;

  const xCoord = (index) => paddingX + index * widthPerPoint;
  const yCoord = (value) => height - paddingY - value * yScale;

  const winsCoords = points.map((point) => ({
    x: xCoord(point.index),
    y: yCoord(point.wins),
  }));
  const lossCoords = points.map((point) => ({
    x: xCoord(point.index),
    y: yCoord(point.losses),
  }));

  const axis = `
    <line class="chart-axis" x1="${paddingX}" y1="${paddingY}" x2="${paddingX}" y2="${height - paddingY}" />
    <line class="chart-axis" x1="${paddingX}" y1="${height - paddingY}" x2="${width - paddingX}" y2="${height - paddingY}" />
  `;

  const gridLines = Array.from({ length: 3 }, (_, index) => {
    const y = paddingY + ((height - paddingY * 2) / 3) * (index + 1);
    return `<line class="chart-grid" x1="${paddingX}" y1="${y}" x2="${width - paddingX}" y2="${y}" />`;
  }).join("");

  const winsPath = winsCoords
    .map((point, index) => `${index === 0 ? "M" : "L"}${point.x},${point.y}`)
    .join(" ");
  const lossesPath = lossCoords
    .map((point, index) => `${index === 0 ? "M" : "L"}${point.x},${point.y}`)
    .join(" ");

  const winsCircles = winsCoords
    .map((point) => `<circle class="chart-point wins" cx="${point.x}" cy="${point.y}" r="4" />`)
    .join("");
  const lossesCircles = lossCoords
    .map((point) => `<circle class="chart-point losses" cx="${point.x}" cy="${point.y}" r="4" />`)
    .join("");

  const labelInterval = Math.max(1, Math.ceil(points.length / 6));
  const labels = points
    .filter((point) => point.index % labelInterval === 0 || point.index === points.length - 1)
    .map((point) => {
      const x = xCoord(point.index);
      const y = height - paddingY + 18;
      return `<text class="chart-label" x="${x}" y="${y}" text-anchor="middle">${point.date}</text>`;
    })
    .join("");

  battleChart.setAttribute("viewBox", `0 0 ${width} ${height}`);
  battleChart.setAttribute("width", width);
  battleChart.setAttribute("height", height);
  battleChart.innerHTML = `${gridLines}${axis}<path class="chart-path wins" d="${winsPath}" /><path class="chart-path losses" d="${lossesPath}" />${winsCircles}${lossesCircles}${labels}`;
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
  const opponentName = document.getElementById("opponentDeck").value.trim();
  const opponentDetails = document.getElementById("opponentDetails").value.trim();
  const opponentBase = opponentName || "Unknown";
  const opponent = opponentDetails ? `${opponentBase} — ${opponentDetails}` : opponentBase;
  const deck = await apiFetch(`/api/decks/${encodeURIComponent(state.currentDeck.name)}/battles`, {
    method: "POST",
    body: JSON.stringify({ result, opponent }),
  });
  document.getElementById("opponentDeck").value = "";
  document.getElementById("opponentDetails").value = "";
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
