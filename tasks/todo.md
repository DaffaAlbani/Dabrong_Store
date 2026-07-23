# TODO: Unified Admin & Member Login System

- [ ] 1. Update `middleware/middleware.go` to accept tokens with `role == "admin"` from `admin_token` or `member_token` cookies.
- [ ] 2. Update `handlers/member.go` `MemberLogin` to dynamically detect admin role (via env or `users` DB `role == 'admin'`), set cookies, and return `role` & `admin_path`.
- [ ] 3. Update `appsetup/public/script.js` & `vercel_script.js` to handle admin redirect on login.
- [ ] 4. Run `go test ./...` and `go build -o server main.go`.
- [ ] 5. Commit changes to Git, push to GitHub, and re-deploy to Vercel (`npx vercel --prod`).
- [ ] 6. Document results in `tasks/todo.md` and `walkthrough.md`.
