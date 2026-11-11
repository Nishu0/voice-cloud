import sqlite3
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent


def schema(db):
    return db.execute("""
        SELECT type, name, sql FROM sqlite_schema
        WHERE name NOT LIKE 'sqlite_%' AND name != 'goose_db_version'
        ORDER BY type, name
    """).fetchall()


class MigrationTest(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory(prefix='voice cloud db ')
        self.addCleanup(temporary.cleanup)
        self.path = Path(temporary.name) / 'nested' / 'router.db'

    def make(self, target, *args, check=True):
        return subprocess.run(
            ['make', target, f'DB_PATH={self.path}', *args], cwd=ROOT,
            capture_output=True, text=True, check=check,
        )

    def test_migration_matches_schema_and_can_roll_back_and_reapply(self):
        self.make('db-up')
        with sqlite3.connect(':memory:') as expected, sqlite3.connect(self.path) as actual:
            expected.executescript((ROOT / 'db/schema.sql').read_text())
            self.assertEqual(schema(actual), schema(expected))
            actual.execute("""
                INSERT INTO connections (name, adapter, location_kind, base_url)
                VALUES ('test', 'whisper', 'local', 'http://localhost:9000')
            """)
        self.make('db-up')
        self.make('db-status')
        self.make('db-version')
        with sqlite3.connect(self.path) as db:
            self.assertEqual(db.execute('SELECT count(*) FROM connections').fetchone(), (1,))
        self.make('db-down')
        with sqlite3.connect(self.path) as db:
            self.assertEqual(schema(db), [])
        self.make('db-up')
        with sqlite3.connect(':memory:') as expected, sqlite3.connect(self.path) as actual:
            expected.executescript((ROOT / 'db/schema.sql').read_text())
            self.assertEqual(schema(actual), schema(expected))

    def test_backup_is_valid_and_refuses_to_overwrite(self):
        self.make('db-up')
        backup = self.path.with_name('backup.db')
        self.make('db-backup', f'BACKUP={backup}')
        with sqlite3.connect(backup) as db:
            self.assertEqual(db.execute('PRAGMA integrity_check').fetchone(), ('ok',))
            self.assertEqual(len([row for row in schema(db) if row[0] == 'table']), 10)
        before = backup.read_bytes()
        self.assertNotEqual(
            self.make('db-backup', f'BACKUP={backup}', check=False).returncode, 0,
        )
        self.assertEqual(backup.read_bytes(), before)


if __name__ == '__main__':
    unittest.main()
