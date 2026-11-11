# Provider catalog

Run `make db-seed` to migrate the database and seed provider connections, or run
`python3 scripts/seed.py --db /path/to/router.db` against a migrated database.
`DB_PATH` selects the database for either command. Without it, the script defaults
to `data/voice-cloud.db` relative to the repository root.

The seed is a checked-in snapshot reviewed on 2026-11-11. It takes the union of
[Speko's public model catalog](https://speko.ai/models) and its
[Gateway catalog](https://github.com/SpekoAI/gateway/blob/main/gateway/catalog.go).
It does not fetch third-party data when run. Speko's authenticated live catalog
can vary by organization, region, and route health; the snapshot is not a claim
that every provider is currently routable there.

There are 36 cloud provider entries: Alibaba, Anthropic, AssemblyAI, Azure,
Baseten, Cartesia, Cerebras, Deepgram, DeepInfra, ElevenLabs, Fish Audio, Gemini,
Gladia, Google, Gradium, Hamsa, Hopper, Hume, Inworld, Maya Research, Meta,
MiniMax, Modulate, Munsit, Nari Labs, Omi Health, OpenAI, Palabra, Paxa Labs,
Rime, Smallest AI, Soniox, Speechify, Speechmatics, Together, and xAI.
Google and Gemini retain separate adapter IDs, matching Speko's distinct routes.

These entries occupy `connections`; a connection is a configured provider account
or endpoint, not a model. Seeded `adapter` values identify the intended future
adapter, and `credential_ref` contains an external secret name, never a key value.
All seed entries are disabled. Provider inference and credential resolution are
not implemented yet; adding or enabling a connection does not implement them.

Base URLs identify provider origins. They are not complete speech, LLM, or
WebSocket routes; adapters will select the appropriate path and may need separate
hosts for different modalities. Azure's URL requires your resource name. Local
and Custom URLs are placeholders to replace with your actual endpoint.
Review provider-specific authentication and deployment settings before use.

Seeding inserts missing names in one transaction. Repeating it preserves existing
URLs, credential references, enabled flags, timestamps, and user-added connections.
Seed names are the identity for this operation: renaming a seeded connection lets
the next run create another one under its original name. Update the seed explicitly
when expanding the built-in catalog; routine runs do not synchronize it with Speko.

## Add custom or local models

Custom and Local are starter connections. You can add any number of additional
ones using `POST /v1/connections`, without editing a fixed provider allowlist:

```json
{
  "name": "My local runtime",
  "adapter": "local",
  "location_kind": "local",
  "base_url": "http://127.0.0.1:8081",
  "enabled": false
}
```

For a custom hosted endpoint, use your own name and base URL, `adapter: "custom"`,
and `location_kind: "cloud"`. An optional `credential_ref` can name the external
credential. Adapter identifiers are not restricted to the seeded list.

Create the model through `POST /v1/models`, then link it to the connection through
`POST /v1/deployments` with its actual `upstream_model_id`. STT, TTS, and LLM are
the current schema tasks. Local models follow the same relationships as cloud
models; they do not need a separate table. See [the API guide](api.md) for requests.

The seed adds provider connections only. Models, deployments, voices, languages,
capabilities, and pricing remain explicit configuration; no compatibility or
pricing is inferred from a provider name.
