#!/usr/bin/env python3
"""Seed provider connections into an existing, migrated SQLite database."""

import argparse
import os
import sqlite3
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
# Public catalog + Gateway adapters, reviewed 2026-11-11:
# https://speko.ai/models
# https://github.com/SpekoAI/gateway/blob/main/gateway/catalog.go
# These are disabled configuration records; provider adapters are not implemented yet.
# Base URLs are origins, not modality-specific HTTP or WebSocket paths.
PROVIDERS = (
    ('Alibaba', 'alibaba', 'cloud', 'https://dashscope-intl.aliyuncs.com', 'DASHSCOPE_API_KEY'),
    ('Anthropic', 'anthropic', 'cloud', 'https://api.anthropic.com', 'ANTHROPIC_API_KEY'),
    ('AssemblyAI', 'assemblyai', 'cloud', 'https://streaming.assemblyai.com', 'ASSEMBLYAI_API_KEY'),
    ('Azure', 'azure', 'cloud', 'https://your-resource.openai.azure.com', 'AZURE_OPENAI_API_KEY'),
    ('Baseten', 'baseten', 'cloud', 'https://inference.baseten.co', 'BASETEN_API_KEY'),
    ('Cartesia', 'cartesia', 'cloud', 'https://api.cartesia.ai', 'CARTESIA_API_KEY'),
    ('Cerebras', 'cerebras', 'cloud', 'https://api.cerebras.ai', 'CEREBRAS_API_KEY'),
    ('Deepgram', 'deepgram', 'cloud', 'https://api.deepgram.com', 'DEEPGRAM_API_KEY'),
    ('DeepInfra', 'deepinfra', 'cloud', 'https://api.deepinfra.com', 'DEEPINFRA_API_KEY'),
    ('ElevenLabs', 'elevenlabs', 'cloud', 'https://api.elevenlabs.io', 'ELEVENLABS_API_KEY'),
    ('Fish Audio', 'fish', 'cloud', 'https://api.fish.audio', 'FISH_API_KEY'),
    ('Gemini', 'gemini', 'cloud', 'https://generativelanguage.googleapis.com', 'GOOGLE_API_KEY'),
    ('Gladia', 'gladia', 'cloud', 'https://api.gladia.io', 'GLADIA_API_KEY'),
    ('Google', 'google', 'cloud', 'https://generativelanguage.googleapis.com', 'GOOGLE_API_KEY'),
    ('Gradium', 'gradium', 'cloud', 'https://api.gradium.ai', 'GRADIUM_API_KEY'),
    ('Hamsa', 'hamsa', 'cloud', 'https://api.tryhamsa.com', 'HAMSA_API_KEY'),
    ('Hopper', 'hopper', 'cloud', 'https://api.withhopper.com', 'HOPPER_API_KEY'),
    ('Hume', 'hume', 'cloud', 'https://api.hume.ai', 'HUME_API_KEY'),
    ('Inworld', 'inworld', 'cloud', 'https://api.inworld.ai', 'INWORLD_API_KEY'),
    ('Maya Research', 'maya', 'cloud', 'https://tts.mayaresearch.ai', 'MAYA_API_KEY'),
    ('Meta', 'meta', 'cloud', 'https://api.meta.ai', 'META_API_KEY'),
    ('MiniMax', 'minimax', 'cloud', 'https://api.minimax.io', 'MINIMAX_API_KEY'),
    ('Modulate', 'modulate', 'cloud', 'https://platform.modulate.ai', 'MODULATE_API_KEY'),
    ('Munsit', 'munsit', 'cloud', 'https://api.munsit.com', 'MUNSIT_API_KEY'),
    ('Nari Labs', 'nari', 'cloud', 'https://api.narilabs.com', 'NARI_API_KEY'),
    ('Omi Health', 'omi', 'cloud', 'https://api.us.omi.health', 'OMI_API_KEY'),
    ('OpenAI', 'openai', 'cloud', 'https://api.openai.com', 'OPENAI_API_KEY'),
    ('Palabra', 'palabra', 'cloud', 'https://stream.palabra.ai', 'PALABRA_API_KEY'),
    ('Paxa Labs', 'paxa', 'cloud', 'https://api.paxalabs.com', 'PAXA_API_KEY'),
    ('Rime', 'rime', 'cloud', 'https://users-ws.rime.ai', 'RIME_API_KEY'),
    ('Smallest AI', 'smallest', 'cloud', 'https://api.smallest.ai', 'SMALLEST_API_KEY'),
    ('Soniox', 'soniox', 'cloud', 'https://stt-rt.soniox.com', 'SONIOX_API_KEY'),
    ('Speechify', 'speechify', 'cloud', 'https://api.speechify.ai', 'SPEECHIFY_API_KEY'),
    ('Speechmatics', 'speechmatics', 'cloud', 'https://global.rt.speechmatics.com', 'SPEECHMATICS_API_KEY'),
    ('Together', 'together', 'cloud', 'https://api.together.ai', 'TOGETHER_API_KEY'),
    ('xAI', 'xai', 'cloud', 'https://api.x.ai', 'XAI_API_KEY'),
    # Replace these template URLs before use; arbitrary additional connections are allowed.
    ('Custom', 'custom', 'cloud', 'https://your-provider.example', 'CUSTOM_API_KEY'),
    ('Local', 'local', 'local', 'http://127.0.0.1:8000', None),
)


def seed(path, providers):
    if not providers:
        raise ValueError('The supported-provider catalog has not been defined yet.')
    database = sqlite3.connect(Path(path).resolve().as_uri() + '?mode=rw', uri=True)
    try:
        database.execute('PRAGMA foreign_keys = ON')
        with database:
            before = database.total_changes
            database.executemany('''
                INSERT INTO connections
                    (name, adapter, location_kind, base_url, credential_ref, enabled)
                VALUES (?, ?, ?, ?, ?, 0)
                ON CONFLICT (name) DO NOTHING
            ''', providers)
            return database.total_changes - before
    finally:
        database.close()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--db', default=os.environ.get('DB_PATH') or ROOT / 'data/voice-cloud.db',
                        help='existing SQLite database (default: DB_PATH or data/voice-cloud.db)')
    args = parser.parse_args()
    try:
        inserted = seed(args.db, PROVIDERS)
    except (sqlite3.Error, ValueError) as error:
        parser.exit(1, f'Seed failed: {error}\nRun make db-up first if the database is not migrated.\n')
    print(f'Added {inserted} provider connection(s); existing connections were preserved.')


if __name__ == '__main__':
    main()
