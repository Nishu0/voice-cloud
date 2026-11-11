# User accounts

One self-hosted installation shares connections, models, deployments, routing
policies, keys, and request history. Users have installation-wide `admin` or
`member` roles. There are no organizations, workspaces, memberships, or billing
tables.

Migration `00002_users.sql` adds:

| Table | Purpose |
| --- | --- |
| `users` | Unique email, display name, password hash, role, active/disabled status, email verification timestamp |
| `user_sessions` | User-owned session token hashes, creation/expiry timestamps, revocation |
| `user_tokens` | Expiring single-use token hashes for password reset or email verification |

Email uniqueness is case-insensitive for ASCII through SQLite's NOCASE collation.
The future authentication layer must validate and normalize email addresses and
store encoded password hashes, never plaintext passwords. `ListUsers` deliberately
excludes password hashes; credential lookups remain internal and must not be
serialized into public responses.

Session and action-token hashes are 64 lowercase hexadecimal characters, intended
for SHA-256 hashes of cryptographically random tokens. Raw tokens belong with the
client, not in SQLite. Expiry is exclusive: a token is invalid at `expires_at`.
Active-session lookup excludes revoked, expired, not-yet-valid, and disabled-user
sessions. Action-token consumption performs a guarded UPDATE/RETURNING so a token
can be used once, for its intended purpose, while its user is active.

`api_keys.created_by_user_id` records who created a router key. It is attribution,
not ownership or an access boundary. Existing/bootstrap keys have NULL creators;
the current admin-token API still creates such keys. Deleting a user removes
their sessions and action tokens, clears key attribution, and preserves router
keys and request history. Disabling a user blocks user sessions; shared router
keys remain valid until explicitly revoked.

The authentication layer must use transactions to consume a reset token, update
the password, and delete the user's sessions/remaining reset tokens together.
Disabling an account should also delete its sessions so re-enabling it does not
reactivate old sessions. Last-admin protection and role enforcement belong in
that layer and are not provided by these tables alone.

Run `make db-up` to apply the migration and `make db-generate` after query changes.
No user or default password is seeded. Login, logout, account-management routes,
email delivery, and authorization enforcement are not implemented yet; the API
continues to use `ADMIN_TOKEN`. Models for invitations, SSO, MFA, or billing can be
added when those features are selected.

Rolling back this migration drops account/session/token data and creator
attribution, while keeping the original router configuration and history.
