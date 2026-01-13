import SwiftUI

struct ContentView: View {
    @StateObject private var client: TCGClient
    @State private var selectedDeck: String?
    @State private var cardID: String = ""
    @State private var battleResult: String = "W"
    @State private var opponent: String = ""

    init() {
        _client = StateObject(wrappedValue: try! TCGClient(decksDir: "decks"))
    }

    var body: some View {
        NavigationView {
            VStack(spacing: 16) {
                if let error = client.errorMessage {
                    Text(error)
                        .foregroundColor(.red)
                }

                List(client.decks, id: \.self, selection: $selectedDeck) { deck in
                    Text(deck)
                        .onTapGesture {
                            selectedDeck = deck
                            client.loadDeck(named: deck)
                        }
                }
                .frame(height: 200)

                HStack {
                    TextField("Card ID", text: $cardID)
                        .textFieldStyle(RoundedBorderTextFieldStyle())
                    Button("Add") {
                        client.addCard(byID: cardID)
                    }
                }

                List(client.cards) { entry in
                    VStack(alignment: .leading) {
                        Text(entry.name)
                        Text("\(entry.set) x\(entry.count)")
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }
                }

                VStack(alignment: .leading, spacing: 8) {
                    Text("Record Battle")
                        .font(.headline)
                    HStack {
                        Picker("Result", selection: $battleResult) {
                            Text("W").tag("W")
                            Text("L").tag("L")
                        }
                        .pickerStyle(SegmentedPickerStyle())
                        TextField("Opponent", text: $opponent)
                            .textFieldStyle(RoundedBorderTextFieldStyle())
                        Button("Save") {
                            client.recordBattle(result: battleResult, opponent: opponent)
                        }
                    }
                }

                if let stats = client.stats {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Stats")
                            .font(.headline)
                        Text("Battles: \(stats.totalBattles)")
                        Text("Wins: \(stats.wins)  Losses: \(stats.losses)")
                        Text(String(format: "Win %%: %.2f", stats.winPercentage))
                    }
                }

                Spacer()
            }
            .padding()
            .navigationTitle("TCGCLI")
            .onAppear {
                client.loadDecks()
            }
        }
    }
}
