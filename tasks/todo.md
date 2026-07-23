# TODO: Catalog Item Selection & Deduplication

- [x] 1. Implement dynamic catalog thinning algorithm in `sync_excel.py` and `sync_api.py` (filter out near-duplicate quantities with threshold spacing for small/mid/large tiers).
- [x] 2. Run python sync scripts (`sync_api.py` and `sync_excel.py`) to regenerate `database/products.json` and update SQLite `products_cache`.
- [x] 3. Verify filtered item counts per catalog, run `go test ./...` and `go build`.
- [x] 4. Document results & review in `tasks/todo.md`.

## Review & Results
- Total catalog items reduced from **784** to **358** clean, high-value items across 15 game catalogs.
- **Mobile Legends (AGML)**: 274 -> 55 items
- **Free Fire (AGFF)**: 148 -> 57 items
- **PUBG Mobile (AGPUBG)**: 109 -> 45 items
- **CODM (AGCODM)**: 51 -> 33 items
- **Valorant (AGVALO)**: 35 -> 23 items
- **Hago (AGHAGO)**: 58 -> 39 items
- Special passes, memberships, and unique packages (*Weekly Diamond Pass*, *Twilight Pass*, *Welkin Moon*, *Starlight*, *Level Up Pass*) are **100% preserved**.
- All tests (`go test ./...`) pass and `go build` compiles cleanly.
