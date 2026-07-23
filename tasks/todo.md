# TODO: Consolidate Weekly Diamond Pass to 1 Single Item at Rp 28.000

- [ ] 1. Update `sync_excel.py` and `sync_api.py` to consolidate all Weekly Diamond Pass / Weekly Card variations into 1 single canonical `Weekly Diamond Pass` item priced at Rp 28.000.
- [ ] 2. Run `python3 sync_api.py` and `python3 sync_excel.py` to regenerate `database/products.json` and SQLite `products_cache`.
- [ ] 3. Run `go test ./...` and `go build` to verify backend integrity.
- [ ] 4. Commit changes to Git, push to GitHub, and re-deploy to Vercel (`npx vercel --prod`).
- [ ] 5. Document review & final item counts in `tasks/todo.md`.
