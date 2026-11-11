import importlib.util
import sqlite3
import unittest
from urllib.parse import urlsplit

import test_migrations


spec = importlib.util.spec_from_file_location('seed', test_migrations.ROOT / 'scripts/seed.py')
seed = importlib.util.module_from_spec(spec)
spec.loader.exec_module(seed)


class SeedTest(unittest.TestCase):
    setUp = test_migrations.MigrationTest.setUp
    make = test_migrations.MigrationTest.make

    def test_full_catalog_is_repeatable_and_preserves_added_providers(self):
        self.make('db-up')
        adapters = [provider[1] for provider in seed.PROVIDERS]
        self.assertEqual(len(adapters), len(set(adapters)))
        self.assertTrue({'openai', 'deepgram', 'anthropic', 'speechmatics', 'gemini',
                         'gladia', 'hume', 'maya', 'custom', 'local'} <= set(adapters))
        self.assertEqual(len([provider for provider in seed.PROVIDERS if provider[1] not in ('custom', 'local')]), 36)
        for name, adapter, location, base_url, credential_ref in seed.PROVIDERS:
            with self.subTest(provider=name):
                url = urlsplit(base_url)
                self.assertIn(url.scheme, ('http', 'https'))
                self.assertTrue(url.hostname)
                self.assertIsNone(url.username)
                self.assertFalse(url.query)
                if location == 'cloud':
                    self.assertEqual(url.scheme, 'https')
                    self.assertTrue(credential_ref.endswith('_API_KEY'))
        self.assertIn(f'Added {len(seed.PROVIDERS)} provider connection(s)',
                      self.make('db-seed').stdout)
        with sqlite3.connect(self.path) as database:
            database.execute('''
                INSERT INTO connections (name, adapter, location_kind, base_url)
                VALUES ('My custom runtime', 'my-runtime', 'local', 'http://localhost:9876')
            ''')
        self.assertIn('Added 0 provider connection(s)', self.make('db-seed').stdout)
        with sqlite3.connect(self.path) as database:
            self.assertEqual(database.execute('SELECT count(*) FROM connections').fetchone(),
                             (len(seed.PROVIDERS) + 1,))
            self.assertEqual(database.execute(
                "SELECT name FROM connections WHERE adapter = 'my-runtime'"
            ).fetchone(), ('My custom runtime',))
            self.assertEqual(database.execute(
                "SELECT count(*) FROM connections WHERE name != 'My custom runtime' AND enabled != 0"
            ).fetchone(), (0,))

    def test_repeat_seed_preserves_custom_settings_and_secrets_stay_external(self):
        self.make('db-up')
        providers = [('Local Whisper', 'whisper', 'local', 'http://localhost:9000', None),
                     ('Example cloud', 'example', 'cloud', 'https://example.com', 'EXAMPLE_API_KEY')]
        self.assertEqual(seed.seed(self.path, providers), 2)
        with sqlite3.connect(self.path) as database:
            self.assertEqual(database.execute(
                'SELECT enabled FROM connections ORDER BY id'
            ).fetchall(), [(0,), (0,)])
            database.execute("UPDATE connections SET base_url = 'http://custom:9000', enabled = 1 WHERE adapter = 'whisper'")
        self.assertEqual(seed.seed(self.path, providers), 0)
        with sqlite3.connect(self.path) as database:
            self.assertEqual(database.execute(
                "SELECT base_url, enabled FROM connections WHERE adapter = 'whisper'"
            ).fetchone(), ('http://custom:9000', 1))
            self.assertEqual(database.execute(
                "SELECT credential_ref FROM connections WHERE adapter = 'example'"
            ).fetchone(), ('EXAMPLE_API_KEY',))

    def test_failed_seed_rolls_back_all_rows(self):
        self.make('db-up')
        with self.assertRaises(sqlite3.IntegrityError):
            seed.seed(self.path, [('Valid', 'example', 'cloud', 'https://example.com', None),
                                  ('Invalid', 'example', 'invalid', 'https://example.com', None)])
        with sqlite3.connect(self.path) as database:
            self.assertEqual(database.execute('SELECT count(*) FROM connections').fetchone(), (0,))

    def test_missing_database_is_not_created(self):
        with self.assertRaises(sqlite3.OperationalError):
            seed.seed(self.path, [('Example', 'example', 'cloud', 'https://example.com', None)])
        self.assertFalse(self.path.exists())

    def test_empty_catalog_is_rejected(self):
        with self.assertRaises(ValueError):
            seed.seed(self.path, [])
