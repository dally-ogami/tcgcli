import Foundation
import tcgcli

final class TCGClient: ObservableObject {
    private let manager: TcgmobileManager

    @Published var decks: [String] = []
    @Published var cards: [CardEntry] = []
    @Published var stats: StatsPayload?
    @Published var errorMessage: String?

    init(decksDir: String) throws {
        self.manager = try TcgmobileManager(decksDir)
    }

    func loadDecks() {
        do {
            let json = try manager.listDecksJSON()
            decks = try decode(json)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func loadDeck(named name: String) {
        do {
            try manager.loadDeck(name)
            let json = try manager.deckCardsJSON()
            cards = try decode(json)
            stats = try decodeStats(try manager.statsJSON())
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func addCard(byID cardID: String) {
        do {
            _ = try manager.addCardByIDJSON(cardID)
            let json = try manager.deckCardsJSON()
            cards = try decode(json)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func recordBattle(result: String, opponent: String) {
        do {
            try manager.recordBattle(result, opponent)
            stats = try decodeStats(try manager.statsJSON())
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func decode<T: Decodable>(_ json: String) throws -> T {
        guard let data = json.data(using: .utf8) else {
            throw NSError(domain: "TCGClient", code: 1, userInfo: [NSLocalizedDescriptionKey: "Invalid JSON string"]) 
        }
        return try JSONDecoder().decode(T.self, from: data)
    }

    private func decodeStats(_ json: String) throws -> StatsPayload {
        return try decode(json)
    }
}

struct CardEntry: Decodable, Identifiable {
    let id = UUID()
    let name: String
    let set: String
    let count: Int
}

struct StatsPayload: Decodable {
    let totalBattles: Int
    let wins: Int
    let losses: Int
    let winPercentage: Double
    let lossByOpponent: [String: Int]
}
