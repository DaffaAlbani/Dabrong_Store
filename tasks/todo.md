# TODO: Streamline Diamond Variants to Popular Tiers

- [ ] 1. Update `sync_excel.py` and `sync_api.py` with streamlined numeric spacing thresholds (~6-10 clean, popular diamond/coin tiers per game).
- [ ] 2. Run `python3 sync_api.py` and `python3 sync_excel.py` to regenerate `database/products.json` and SQLite `products_cache`.
- [ ] 3. Run `go test ./...` and `go build` to verify backend integrity.
- [ ] 4. Commit changes to Git, push to GitHub, and re-deploy to Vercel (`npx vercel --prod`).
- [ ] 5. Document review & final item counts in `tasks/todo.md` and `walkthrough.md`.
