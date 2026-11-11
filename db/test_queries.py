import os
import re
import shutil
import sqlite3
import subprocess
import tempfile
import unittest
from pathlib import Path

import test_schema


ROOT = Path(__file__).resolve().parent.parent


class QueryTest(unittest.TestCase):
    setUp = test_schema.SchemaTest.setUp

    def query(self, name, *parameters):
        generated = (ROOT / 'db/query.sql.go').read_text()
        match = re.search(r'const \w+ = `(-- name: ' + name + r' :\w+\n.*?)`', generated, re.S)
        self.assertIsNotNone(match, name)
        # Python's SQLite wrapper cannot report rowcount with a leading SQL comment.
        return self.db.execute(match[1].split('\n', 1)[1], parameters)

    def test_every_generated_statement_is_valid_sqlite(self):
        generated = (ROOT / 'db/query.sql.go').read_text()
        queries = re.findall(r'const \w+ = `(-- name: (\w+) :\w+\n.*?)`', generated, re.S)
        source_names = re.findall(r'-- name: (\w+)', (ROOT / 'db/query.sql').read_text())
        self.assertCountEqual([name for _, name in queries], source_names)
        for sql, name in queries:
            with self.subTest(query=name):
                self.assertNotIn('sqlc.', sql)
                count = 0
                for number in re.findall(r'\?(\d*)', sql):
                    count = max(count, int(number)) if number else count + 1
                self.db.execute('EXPLAIN ' + sql, [None] * count)

    def test_api_key_expiration_and_revocation(self):
        self.assertIsNotNone(self.query('GetActiveAPIKeyByHash', 'hash-one', 1000).fetchone())
        self.db.execute('UPDATE api_keys SET expires_at = 2000 WHERE id = 1')
        self.assertIsNone(self.query('GetActiveAPIKeyByHash', 'hash-one', 2000).fetchone())
        self.assertEqual(self.query('RevokeAPIKey', 1500, 1).rowcount, 1)
        self.assertEqual(self.db.execute(
            'SELECT revoked_at, updated_at FROM api_keys WHERE id = 1'
        ).fetchone(), (1500, 1500))
        self.assertIsNone(self.query('GetActiveAPIKeyByHash', 'hash-one', 1000).fetchone())

    def test_request_attempt_and_usage_lifecycle(self):
        self.query('CreateRequest', 'r2', 1, 'stt', 'http', 'en', 1, '{}', 2000,
                   'trace', 'retry-2', 'fingerprint').fetchone()
        self.assertEqual(self.query('GetRequestByIdempotencyKey', 1, 'retry-2').fetchone()[0], 'r2')
        with self.assertRaises(sqlite3.IntegrityError):
            self.query('CreateRequest', 'r3', 1, 'stt', 'http', 'en', 1, '{}', 2000,
                       None, 'retry-2', 'fingerprint').fetchone()
        self.assertEqual(self.query('StartRequest', 'r2').rowcount, 1)
        self.assertEqual(self.query('StartRequest', 'r2').rowcount, 0)
        self.query('CreateAttempt', 'a3', 'r2', 1, 1, 2000, 'priority').fetchone()
        self.query('SetAttemptProviderRequestID', 'provider-123', 'a3')
        self.assertEqual(self.query('SetAttemptFirstOutput', 2100, 'a3').rowcount, 1)
        self.assertEqual(self.query('SetAttemptFirstOutput', 2200, 'a3').rowcount, 0)
        self.query('CreateAttemptUsage', 'a3', 'input_audio', 100, 'audio_ms',
                   '{"rate": 10}', None, 'USD', 'unknown').fetchone()
        self.query('UpdateAttemptUsage', 200, 2000, 'final', 2300, 'a3', 'input_audio').fetchone()
        usage = self.query('ListRequestUsage', 'r2').fetchone()
        self.assertEqual(usage[2:6], (200, 'audio_ms', '{"rate": 10}', 2000))
        self.assertEqual(self.query('FinishAttempt', 'running', 2300, None, 'a3').rowcount, 0)
        self.assertEqual(self.query('FinishAttempt', 'succeeded', 2300, None, 'a3').rowcount, 1)
        self.assertEqual(self.query('FinishAttempt', 'failed', 2400, 'late', 'a3').rowcount, 0)
        self.assertEqual(self.query('FinishRequest', 'succeeded', 2300, None, 'r2').rowcount, 1)
        self.assertEqual(self.query('FinishRequest', 'failed', 2400, 'late', 'r2').rowcount, 0)
        with self.assertRaises(sqlite3.IntegrityError):
            self.query('DeleteRequest', 'r2')
        self.query('DeleteAttemptUsage', 'a3', 'input_audio')
        self.query('DeleteAttempt', 'a3')
        self.assertEqual(self.query('DeleteRequest', 'r2').rowcount, 1)

    def test_policy_candidates_exclude_disabled_and_wrong_task_deployments(self):
        self.assertEqual(len(self.query('ListPolicyDeployments', 1).fetchall()), 1)
        self.db.execute('UPDATE connections SET enabled = 0 WHERE id = 1')
        self.assertEqual(self.query('ListPolicyDeployments', 1).fetchall(), [])
        self.db.execute('UPDATE connections SET enabled = 1 WHERE id = 1')
        self.db.execute('UPDATE deployments SET enabled = 0 WHERE id = 1')
        self.assertEqual(self.query('ListPolicyDeployments', 1).fetchall(), [])
        self.db.execute('UPDATE deployments SET enabled = 1 WHERE id = 1')
        self.db.execute("UPDATE models SET task = 'tts' WHERE id = 1")
        self.assertEqual(self.query('ListPolicyDeployments', 1).fetchall(), [])


class CodeGenerationTest(unittest.TestCase):
    def test_fresh_generation_matches_checked_in_code_and_compiles(self):
        with tempfile.TemporaryDirectory(prefix='voice-cloud-codegen-') as directory:
            temporary = Path(directory)
            shutil.copytree(ROOT / 'db/migrations', temporary / 'db/migrations')
            for name in ('query.sql', 'go.mod'):
                shutil.copy(ROOT / 'db' / name, temporary / 'db' / name)
            shutil.copy(ROOT / 'sqlc.yml', temporary / 'sqlc.yml')
            result = subprocess.run(['sqlc', 'generate'], cwd=temporary,
                                    capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            for generated in (temporary / 'db').glob('*.go'):
                self.assertEqual(generated.read_text(), (ROOT / 'db' / generated.name).read_text(),
                                 f'{generated.name} is stale; run make db-generate')
            environment = dict(os.environ, GOCACHE=str(temporary / 'go-cache'), GOWORK='off')
            result = subprocess.run(['go', 'test', './...'], cwd=temporary / 'db',
                                    env=environment, capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)


if __name__ == '__main__':
    unittest.main()
