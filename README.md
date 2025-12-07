# TCGCLI

A really crude tool for building Pokémon TCG Pocket decks, tracking battle outcomes, and generating battle statistics so I can improve my game. The CLI helps you manage your deck by letting you add cards from a list of valid TCG Pocket cards, manage copy limits, and record battle outcomes along with statistics. Pretty useful for my needs, but hey take what's useful and discard the rest <3

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration Files](#configuration-files)
- [Sample Output](#sample-output)
- [Thanks](#thanks)
- [License](#license)

## Features

**Deck Management**  
  Easily create or load decks saved as JSON files within the `decks` directory.
  
**Card Management**  
  List available cards retrieved from an up-to-date online database (with a local fallback `valid_cards.json`), search by card name or set, and add cards to your deck (with a limit of 2 copies per card across all sets).

**Battle Records**  
  Record battle outcomes (win or loss), along with opponent details and a timestamp.

**Statistics & Graphs**  
  Calculate and display overall battle statistics, and generate am ASCII win/loss graph.

**User-Friendly Interface**  
ANSI-colored command-line output (auto-reset).

## Prerequisites

- Go 1.22+

## Build & Run

```bash
# Clone the repository
git clone https://github.com/dally-ogami/tcgcli.git
cd tcgcli

# Build the CLI
go build -o tcgcli ./cmd/tcgcli

# Run it (or use go run ./cmd/tcgcli)
./tcgcli
```

## Sample Output
```bash
./tcgcli

Deck Manager Options:
  1: Create a new deck
  2: Load an existing deck
  3: Exit
Enter your choice (1-3): 1
Enter a name for your new deck: Fighting Aggro
Deck file 'decks/Fighting Aggro.json' not found. Starting new deck 'Fighting Aggro'.
New deck 'Fighting Aggro' created.

Main Menu:
  0: List all available cards
  1: Add a card to your deck (search by name or set)
  2: View your deck
  3: Remove a card from your deck
  4: Record a battle outcome
  5: Show deck battle statistics
  6: Save and exit
Enter your choice (0-6): 4
Enter battle outcome (W for win, L for loss): W
Enter opponent deck details (or other metadata): Mewtwo jumped off the porch and slapped me
Battle record added for deck 'Fighting Aggro'.

Main Menu:
  0: List all available cards
  1: Add a card to your deck (search by name or set)
  2: View your deck
  3: Remove a card from your deck
  4: Record a battle outcome
  5: Show deck battle statistics
  6: Save and exit
Enter your choice (0-6): 4
Enter battle outcome (W for win, L for loss): W      
Enter opponent deck details (or other metadata): Arceus got punked
Battle record added for deck 'Fighting Aggro'.

Main Menu:
  0: List all available cards
  1: Add a card to your deck (search by name or set)
  2: View your deck
  3: Remove a card from your deck
  4: Record a battle outcome
  5: Show deck battle statistics
  6: Save and exit
Enter your choice (0-6): 4
Enter battle outcome (W for win, L for loss): L
Enter opponent deck details (or other metadata): Giratina rocked my dome
Battle record added for deck 'Fighting Aggro'.

Main Menu:
  0: List all available cards
  1: Add a card to your deck (search by name or set)
  2: View your deck
  3: Remove a card from your deck
  4: Record a battle outcome
  5: Show deck battle statistics
  6: Save and exit
Enter your choice (0-6): 5

Battle Statistics for 'Fighting Aggro':
  Total Battles: 3
  Wins: 2
  Losses: 1
  Win Percentage: 66.67%

Win/Loss Graph:
Wins  : **
Losses: *

Loss Frequency by Opponent Deck:
  Giratina rocked my dome: 1 loss(es)

Main Menu:
  0: List all available cards
  1: Add a card to your deck (search by name or set)
  2: View your deck
  3: Remove a card from your deck
  4: Record a battle outcome
  5: Show deck battle statistics
  6: Save and exit
```
###
Note: The actual output may differ based on your interactions with the CLI and the contents of the card database.

## Thanks
Thanks to **@flibustier** for maintaining the [pokemon-tcg-pocket-database](https://github.com/flibustier/pokemon-tcg-pocket-database).

## License
This project is licensed under the MIT License. You can do what you want with this script, the author provides no guarantees with it. 
