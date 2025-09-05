import json
import os
import urllib.error
import urllib.request
from colorama import Fore, init
from datetime import datetime

# Initialize Colorama
init(autoreset=True)


class Deck:
    def __init__(self, name, deck_file):
        self.name = name
        self.deck_file = deck_file
        self.cards = []  # List of card dictionaries
        self.battle_history = []  # List of battle record dictionaries
        self.valid_cards = self.load_valid_cards()
        self.load_deck()

    def load_valid_cards(self):
        """Load valid cards from online database with local fallback.

        Normalizes ``name`` and ``set`` fields to lowercase for searching.
        """
        url = (
            "https://raw.githubusercontent.com/flibustier/"
            "pokemon-tcg-pocket-database/main/dist/cards.json"
        )
        cards = []

        try:
            with urllib.request.urlopen(url) as response:
                cards = json.load(response)
            print(Fore.GREEN + "Loaded latest card data from online database.")
        except (urllib.error.URLError, urllib.error.HTTPError, ValueError) as e:
            print(
                Fore.YELLOW
                + f"Warning: Could not fetch latest card data ({e}). Using local cache."
            )
            try:
                with open("valid_cards.json", "r") as f:
                    cards = json.load(f)
            except FileNotFoundError:
                print(Fore.RED + "Error: valid_cards.json not found.")
                return []

        for card in cards:
            if "name" in card:
                card["name"] = card["name"].lower()
            if "set" in card and isinstance(card["set"], str):
                card["set"] = card["set"].lower()
        return cards

    def load_deck(self):
        """
        Load deck data from file. The file is expected to be a JSON dict
        with keys 'cards' and 'battle_history'.
        """
        if os.path.exists(self.deck_file):
            try:
                with open(self.deck_file, "r") as f:
                    data = json.load(f)
                self.cards = data.get("cards", [])
                self.battle_history = data.get("battle_history", [])
                print(Fore.GREEN + f"Deck '{self.name}' loaded from {self.deck_file}.")
            except json.JSONDecodeError:
                print(
                    Fore.RED
                    + f"Error decoding {self.deck_file}. Starting with an empty deck."
                )
                self.cards, self.battle_history = [], []
        else:
            print(
                Fore.YELLOW
                + f"Deck file '{self.deck_file}' not found. Starting new deck '{self.name}'."
            )
            self.cards, self.battle_history = [], []

    def save_deck(self):
        """Save the current deck data to file."""
        data = {"cards": self.cards, "battle_history": self.battle_history}
        with open(self.deck_file, "w") as f:
            json.dump(data, f, indent=2)
        print(Fore.GREEN + f"Deck '{self.name}' saved successfully!")

    def list_available_cards(self):
        """List all valid cards with name, set, and id."""
        if not self.valid_cards:
            print(Fore.RED + "No valid cards available.")
            return

        print(Fore.CYAN + "\nAvailable Cards:")
        for card in self.valid_cards:
            name = card.get("name", "?").capitalize()
            card_set = card.get("set", "?").title()
            card_id = card.get("id", "?")
            print(f" - {name} (Set: {card_set}, ID: {card_id})")

    def search_cards(self, search_term):
        """
        Search for valid cards by name or set (case-insensitive).
        Returns a list of matching card dictionaries.
        """
        term = search_term.lower()
        return [
            card
            for card in self.valid_cards
            if term in card.get("name", "") or term in card.get("set", "")
        ]

    def total_copies_in_deck(self, card_name):
        """
        Return the total number of copies in deck for the given card name.
        Comparison is case-insensitive.
        """
        total = 0
        for entry in self.cards:
            if entry.get("name", "").lower() == card_name.lower():
                total += entry.get("count", 0)
        return total

    def add_card(self, search_term):
        """
        Search and add a card to the deck.
        Limits overall copies (across sets) to 2.
        If multiple cards match, prompt the user to choose.
        """
        results = self.search_cards(search_term)
        if not results:
            print(Fore.RED + f"No valid card found matching '{search_term}'.")
            return False

        if len(results) > 1:
            print(Fore.CYAN + "\nMultiple matches found:")
            for idx, card in enumerate(results, start=1):
                name = card.get("name", "?").capitalize()
                card_set = card.get("set", "?").title()
                card_id = card.get("id", "?")
                print(f"  {idx}. {name} (Set: {card_set}, ID: {card_id})")
            try:
                choice = int(
                    input(Fore.WHITE + "Enter the number of the card you want to add: ")
                )
                if not (1 <= choice <= len(results)):
                    print(Fore.RED + "Invalid selection.")
                    return False
                selected_card = results[choice - 1]
            except ValueError:
                print(Fore.RED + "Please enter a valid number.")
                return False
        else:
            selected_card = results[0]

        # Capitalize for display/storage consistency
        card_name = selected_card.get("name", "").capitalize()
        card_set = selected_card.get("set", "").title()

        # Limit overall copies across sets to 2
        if self.total_copies_in_deck(card_name) >= 2:
            print(
                Fore.YELLOW
                + f"Warning: Already have 2 copies of {card_name} (across all sets). Cannot add more."
            )
            return False

        # If an entry exists for card with the same name & set, increment count
        for entry in self.cards:
            if entry.get("name") == card_name and entry.get("set") == card_set:
                if entry.get("count", 0) >= 2:
                    print(
                        Fore.YELLOW
                        + f"Warning: Already have 2 copies of {card_name} from {card_set}."
                    )
                    return False
                else:
                    entry["count"] += 1
                    if entry["count"] == 2:
                        print(
                            Fore.GREEN
                            + f"{card_name} from {card_set} added. You now have 2 copies in this set."
                        )
                    else:
                        print(Fore.GREEN + f"{card_name} from {card_set} added.")
                    return True

        # Otherwise, add card as a new entry
        self.cards.append({"name": card_name, "set": card_set, "count": 1})
        print(Fore.GREEN + f"{card_name} from {card_set} added to your deck.")
        return True

    def view_deck(self):
        """Display current deck contents."""
        if not self.cards:
            print(Fore.YELLOW + "Your deck is empty.")
            return

        print(Fore.LIGHTCYAN_EX + f"\nDeck: {self.name}")
        for idx, entry in enumerate(self.cards, start=1):
            name = entry.get("name", "?")
            card_set = entry.get("set", "?")
            count = entry.get("count", 0)
            print(Fore.LIGHTCYAN_EX + f"  {idx}. {name} x {count} from {card_set}")

    def remove_card(self, index):
        """
        Remove a card from the deck by index.
        If multiple copies exist, decrement the count; else remove the card entry.
        """
        if index < 0 or index >= len(self.cards):
            print(Fore.RED + "Invalid index. Nothing was removed.")
            return False

        entry = self.cards[index]
        name = entry.get("name", "?")
        card_set = entry.get("set", "?")
        if entry.get("count", 0) > 1:
            entry["count"] -= 1
            print(
                Fore.GREEN
                + f"One copy of {name} from {card_set} removed. Now you have {entry['count']} copy(ies)."
            )
        else:
            self.cards.pop(index)
            print(Fore.GREEN + f"{name} from {card_set} removed from your deck.")
        return True

    def record_battle(self):
        """
        Record the outcome of a battle.
        User inputs outcome ('W' for win, 'L' for loss) and opponent details.
        """
        outcome = (
            input(Fore.WHITE + "Enter battle outcome (W for win, L for loss): ")
            .strip()
            .upper()
        )
        if outcome not in ("W", "L"):
            print(Fore.RED + "Invalid outcome. Use 'W' or 'L'.")
            return

        opponent_info = input(
            Fore.WHITE + "Enter opponent deck details (or other metadata): "
        ).strip()
        battle_record = {
            "date": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            "result": outcome,
            "opponent": opponent_info,
        }
        self.battle_history.append(battle_record)
        print(Fore.GREEN + f"Battle record added for deck '{self.name}'.")

    def show_statistics(self):
        """
        Display battle statistics and a simple text-based win/loss graph.
        Also lists loss frequency by opponent deck.
        """
        total_battles = len(self.battle_history)
        if total_battles == 0:
            print(Fore.YELLOW + "No battle records to show statistics.")
            return

        wins = sum(1 for battle in self.battle_history if battle.get("result") == "W")
        losses = total_battles - wins
        win_percentage = (wins / total_battles) * 100

        print(Fore.CYAN + f"\nBattle Statistics for '{self.name}':")
        print(Fore.CYAN + f"  Total Battles: {total_battles}")
        print(Fore.CYAN + f"  Wins: {wins}")
        print(Fore.CYAN + f"  Losses: {losses}")
        print(Fore.CYAN + f"  Win Percentage: {win_percentage:.2f}%")

        print(Fore.BLUE + "\nWin/Loss Graph:")
        print(Fore.GREEN + "Wins  : " + "*" * wins)
        print(Fore.RED + "Losses: " + "*" * losses)

        lose_matchups = {}
        for battle in self.battle_history:
            if battle.get("result") == "L":
                opponent = battle.get("opponent", "Unknown")
                lose_matchups[opponent] = lose_matchups.get(opponent, 0) + 1

        if lose_matchups:
            print(Fore.MAGENTA + "\nLoss Frequency by Opponent Deck:")
            for opponent, count in lose_matchups.items():
                print(Fore.MAGENTA + f"  {opponent}: {count} loss(es)")


class DeckManager:
    """
    Manages creation and loading of decks.
    Decks are stored as separate JSON files within the 'decks' folder.
    """

    def __init__(self):
        self.decks_dir = "decks"
        if not os.path.exists(self.decks_dir):
            os.makedirs(self.decks_dir)
        self.current_deck = None

    def list_existing_decks(self):
        return [f[:-5] for f in os.listdir(self.decks_dir) if f.endswith(".json")]

    def create_new_deck(self):
        deck_name = input(Fore.WHITE + "Enter a name for your new deck: ").strip()
        if not deck_name:
            print(Fore.RED + "Deck name cannot be empty.")
            return

        deck_file = os.path.join(self.decks_dir, deck_name + ".json")
        if os.path.exists(deck_file):
            print(Fore.RED + "A deck with that name already exists.")
            return

        self.current_deck = Deck(deck_name, deck_file)
        print(Fore.GREEN + f"New deck '{deck_name}' created.")

    def load_existing_deck(self):
        decks = self.list_existing_decks()
        if not decks:
            print(Fore.YELLOW + "No saved decks found.")
            return

        print(Fore.CYAN + "\nExisting decks:")
        for idx, deck_name in enumerate(decks, start=1):
            print(Fore.CYAN + f"  {idx}. {deck_name}")
        try:
            choice = int(input(Fore.WHITE + "Enter the number of the deck to load: "))
            if 1 <= choice <= len(decks):
                selected_deck = decks[choice - 1]
                deck_file = os.path.join(self.decks_dir, selected_deck + ".json")
                self.current_deck = Deck(selected_deck, deck_file)
                print(Fore.GREEN + f"Deck '{selected_deck}' loaded.")
            else:
                print(Fore.RED + "Invalid selection.")
        except ValueError:
            print(Fore.RED + "Please enter a valid number.")

    def select_deck(self):
        """Allows user to create a new deck or load an existing one."""
        while True:
            print(Fore.MAGENTA + "\nDeck Manager Options:")
            print("  1: Create a new deck")
            print("  2: Load an existing deck")
            print("  3: Exit")
            choice = input(Fore.WHITE + "Enter your choice (1-3): ").strip()
            if choice == "1":
                self.create_new_deck()
                break
            elif choice == "2":
                self.load_existing_deck()
                if self.current_deck:
                    break
            elif choice == "3":
                exit(Fore.GREEN + "Goodbye!")
            else:
                print(Fore.RED + "Invalid option. Please try again.")


def main_menu(deck):
    while True:
        print(Fore.MAGENTA + "\nMain Menu:")
        print("  0: List all available cards")
        print("  1: Add a card to your deck (search by name or set)")
        print("  2: View your deck")
        print("  3: Remove a card from your deck")
        print("  4: Record a battle outcome")
        print("  5: Show deck battle statistics")
        print("  6: Save and exit")
        choice = input(Fore.WHITE + "Enter your choice (0-6): ").strip()

        if choice == "0":
            deck.list_available_cards()
            cont = (
                input(
                    Fore.MAGENTA
                    + "\nDo you want to add a card or go back to the main menu? (add/main): "
                )
                .strip()
                .lower()
            )
            if cont == "add":
                search_term = input(
                    Fore.MAGENTA + "\nEnter search term (name or set): "
                ).strip()
                deck.add_card(search_term)
            elif cont != "main":
                print(Fore.RED + "Invalid choice. Going back to the main menu.")
        elif choice == "1":
            search_term = input(
                Fore.MAGENTA + "\nEnter search term (name or set): "
            ).strip()
            deck.add_card(search_term)
            cont = (
                input(Fore.MAGENTA + "\nDo you want to add another card? (yes/no): ")
                .strip()
                .lower()
            )
            while cont == "yes":
                search_term = input(
                    Fore.MAGENTA + "\nEnter search term (name or set): "
                ).strip()
                deck.add_card(search_term)
                cont = (
                    input(
                        Fore.MAGENTA + "\nDo you want to add another card? (yes/no): "
                    )
                    .strip()
                    .lower()
                )
            if cont != "no":
                print(Fore.RED + "Invalid choice. Going back to the main menu.")
        elif choice == "2":
            deck.view_deck()
            cont = (
                input(
                    Fore.MAGENTA
                    + "\nDo you want to remove a card or go back to the main menu? (rm/main): "
                )
                .strip()
                .lower()
            )
            if cont == "rm":
                if not deck.cards:
                    print(Fore.RED + "Cannot remove from an empty deck.")
                else:
                    try:
                        index = (
                            int(
                                input(
                                    Fore.MAGENTA
                                    + "\nEnter the position (number) of the card to remove: "
                                )
                            )
                            - 1
                        )
                        deck.remove_card(index)
                    except ValueError:
                        print(Fore.RED + "Please enter a valid number.")
            elif cont != "main":
                print(Fore.RED + "Invalid choice. Going back to the main menu.")
        elif choice == "3":
            if not deck.cards:
                print(Fore.RED + "Cannot remove from an empty deck.")
            else:
                deck.view_deck()
                try:
                    index = (
                        int(
                            input(
                                Fore.MAGENTA
                                + "\nEnter the position (number) of the card to remove: "
                            )
                        )
                        - 1
                    )
                    deck.remove_card(index)
                except ValueError:
                    print(Fore.RED + "Please enter a valid number.")
                cont = (
                    input(
                        Fore.MAGENTA
                        + "\nDo you want to remove another card? (yes/no): "
                    )
                    .strip()
                    .lower()
                )
                while cont == "yes":
                    if not deck.cards:
                        print(Fore.RED + "Cannot remove from an empty deck.")
                        break
                    deck.view_deck()
                    try:
                        index = (
                            int(
                                input(
                                    Fore.MAGENTA
                                    + "\nEnter the position (number) of the card to remove: "
                                )
                            )
                            - 1
                        )
                        deck.remove_card(index)
                    except ValueError:
                        print(Fore.RED + "Please enter a valid number.")
                    cont = (
                        input(
                            Fore.MAGENTA
                            + "\nDo you want to remove another card? (yes/no): "
                        )
                        .strip()
                        .lower()
                    )
                if cont != "no":
                    print(Fore.RED + "Invalid choice. Going back to the main menu.")
        elif choice == "4":
            deck.record_battle()
            cont = (
                input(
                    Fore.MAGENTA
                    + "\nDo you want to record another battle or go back to the main menu? (add/main): "
                )
                .strip()
                .lower()
            )
            while cont == "add":
                deck.record_battle()
                cont = (
                    input(
                        Fore.MAGENTA
                        + "\nDo you want to record another battle or go back to the main menu? (add/main): "
                    )
                    .strip()
                    .lower()
                )
            if cont != "main":
                print(Fore.RED + "Invalid choice. Going back to the main menu.")
        elif choice == "5":
            deck.show_statistics()
        elif choice == "6":
            deck.save_deck()
            print(Fore.GREEN + "\nExiting. Your deck has been saved!")
            break
        else:
            print(Fore.RED + "Invalid choice. Please try again.")


if __name__ == "__main__":
    manager = DeckManager()
    manager.select_deck()
    if manager.current_deck:
        main_menu(manager.current_deck)
