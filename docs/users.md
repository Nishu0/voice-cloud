# User accounts

One shared installation uses admin/member roles. On first startup, open the web
console and enter your name, email, and password. This creates the initial admin
and signs you in. The initial signup endpoint closes afterward; admins add members
from Account settings. Email addresses are trimmed and lowercased. Passwords must
contain 12–1024 bytes; the browser requires at least 12 characters.

Passwords use PBKDF2-HMAC-SHA256 with 600,000 iterations and a random 16-byte salt.
Only encoded hashes are stored. No default password or user is seeded. Signup and
session creation use one transaction; a guarded INSERT allows just one first user.

Random 256-bit session tokens are stored as SHA-256 hashes in `user_sessions`.
Tokens expire after seven days, survive service restarts, and are revoked on logout.
Disabled users and expired/revoked sessions cannot authenticate. Authentication
attempts are limited to 20 per minute per backend remote address, including the
web server's shared address. The rate limit is in memory and resets on restart.
Passwords, password hashes, and raw session tokens are never returned by account
reads. Signup/login return raw tokens only for the client to establish a session.

The web server keeps the session in an HttpOnly, SameSite=Lax cookie, with Secure
on HTTPS. It forwards bearer authentication to Go; no management token is shipped
in browser assets. Set the public `ORIGIN` correctly when deploying behind HTTPS.

Admins manage all router configuration and can create member accounts. Members
can read shared configuration/history and change their own display name. API key
creator IDs are attribution, not ownership; keys remain installation-wide.
Optional legacy `ADMIN_TOKEN` access remains available for existing scripts but
is not required. The frontend never uses it.

`user_tokens` already reserves one-use verification/reset tokens, but email
verification, password resets, password changes, invitations, MFA, and account
role/status management are not implemented. Emails are not verified or delivered.

Run `make db-up`, `make db-generate`, and `make api-check` for schema/codegen/auth
validation. The existing user/session migration is reused; no new tables needed.
