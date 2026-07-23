# TODO: Ultra-Clean Catalog Thinning & Re-deploy

- [x] 1. Update `sync_excel.py` and `sync_api.py` with ultra-tight item thinning thresholds (reducing catalog options to ~10-25 items per game, eliminating 2x/3x/4x pass duplicates).
- [x] 2. Run `python3 sync_api.py` and `python3 sync_excel.py` to regenerate `database/products.json` and SQLite `products_cache`.
- [x] 3. Run `go test ./...` and `go build` to verify backend integrity.
- [x] 4. Commit changes to Git, push to GitHub, and re-deploy to Vercel (`npx vercel --prod`).
- [x] 5. Document review & final item counts in `tasks/todo.md`.

## Final Review & Results
- Total catalog items reduced from **784** to **232 total ultra-clean items** across all 15 game catalogs.
- **Mobile Legends (AGML)**: 274 -> 55 -> **25 items**
- **Free Fire (AGFF)**: 148 -> 57 -> **36 items**
- **PUBG Mobile (AGPUBG)**: 109 -> 45 -> **23 items**
- **CODM (AGCODM)**: 51 -> 33 -> **22 items**
- **Valorant (AGVALO)**: 35 -> 23 -> **11 items**
- **Hago (AGHAGO)**: 58 -> 39 -> **24 items**
- **Honor of Kings (AGHOK)**: 15 -> **9 items**
- Re-deployed to Vercel Production (`https://www.dabrongstore.my.id`).
