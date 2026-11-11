import sqlite3
import subprocess
import tempfile
import unittest
from pathlib import Path


class SchemaTest(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory(prefix='voice-cloud-schema-')
        self.addCleanup(temporary.cleanup)
        path = Path(temporary.name) / 'router.db'
        subprocess.run(
            ['make', 'db-up', f'DB_PATH={path}'],
            cwd=Path(__file__).resolve().parent.parent,
            capture_output=True, text=True, check=True,
        )
        self.db = sqlite3.connect(path)
        self.addCleanup(self.db.close)
        self.db.execute('PRAGMA foreign_keys = ON')
        self.db.executescript("""
            INSERT INTO connections (id, name, adapter, location_kind, base_url)
            VALUES (1, 'Local', 'whisper', 'local', 'http://localhost:9000');
            INSERT INTO models (id, name, publisher, task)
            VALUES (1, 'large-v3', 'openai', 'stt');
            INSERT INTO deployments (id, name, model_id, connection_id, upstream_model_id)
            VALUES (1, 'Local Whisper', 1, 1, 'large-v3');
            INSERT INTO routing_policies (id, name, task) VALUES (1, 'Speech', 'stt');
            INSERT INTO policy_targets (policy_id, deployment_id) VALUES (1, 1);
            INSERT INTO api_keys (id, name, key_hash, key_prefix, default_policy_id)
            VALUES (1, 'App', 'hash-one', 'vc_one', 1),
                   (2, 'Other app', 'hash-two', 'vc_two', 1);
            INSERT INTO requests (id, api_key_id, task, transport, policy_id,
                                  routing_snapshot_json, started_at,
                                  idempotency_key, request_fingerprint)
            VALUES ('r1', 1, 'stt', 'websocket', 1, '{}', 1000, 'retry-key', 'fingerprint');
            INSERT INTO attempts (id, request_id, deployment_id, attempt_number,
                                  started_at, selection_reason, status, finished_at)
            VALUES ('a1', 'r1', 1, 1, 1000, 'first eligible', 'failed', 1100),
                   ('a2', 'r1', 1, 2, 1100, 'retry', 'succeeded', 1500);
        """)

    def test_failed_attempts_can_have_usage_and_unknown_cost_is_not_zero(self):
        self.db.execute("""
            INSERT INTO attempt_usage (attempt_id, metric, quantity, unit,
                                       cost_nanos, cost_status)
            VALUES ('a1', 'input_audio', 100, 'audio_ms', 1000, 'final'),
                   ('a2', 'input_audio', 1000, 'audio_ms', NULL, 'unknown')
        """)
        self.assertEqual(self.db.execute(
            'SELECT cost_nanos FROM attempt_usage ORDER BY attempt_id'
        ).fetchall(), [(1000,), (None,)])
        self.assertEqual(self.db.execute('PRAGMA foreign_key_check').fetchall(), [])

    def test_idempotency_is_scoped_to_the_api_key(self):
        query = """
            INSERT INTO requests (id, api_key_id, task, transport,
                                  routing_snapshot_json, started_at,
                                  idempotency_key, request_fingerprint)
            VALUES (?, ?, 'stt', 'http', '{}', 2000, 'retry-key', 'fingerprint')
        """
        with self.assertRaises(sqlite3.IntegrityError):
            self.db.execute(query, ('duplicate', 1))
        self.db.execute(query, ('other-client', 2))

    def test_invalid_data_is_rejected(self):
        statements = [
            "UPDATE deployments SET capabilities_json = '[]'",
            "UPDATE deployments SET capabilities_json = 'invalid'",
            "UPDATE deployments SET connection_id = 999",
            "UPDATE connections SET enabled = 2",
            "UPDATE routing_policies SET max_attempts = 0",
            "UPDATE requests SET task = 'invalid'",
            "UPDATE requests SET finished_at = 999",
            "UPDATE requests SET request_fingerprint = NULL",
            "UPDATE attempts SET attempt_number = 1 WHERE id = 'a2'",
            "UPDATE attempts SET first_output_at = 1600 WHERE id = 'a2'",
            "DELETE FROM deployments WHERE id = 1",
            "DELETE FROM api_keys WHERE id = 1",
        ]
        for statement in statements:
            with self.subTest(statement=statement), self.assertRaises(sqlite3.DatabaseError):
                self.db.execute(statement)
        for quantity, cost, status in [
            (-1, None, 'unknown'), (0.5, None, 'unknown'),
            (1, -1, 'final'), (1, 0.5, 'final'),
            (1, None, 'final'), (1, 0, 'unknown'),
        ]:
            with self.subTest(quantity=quantity, cost=cost, status=status):
                with self.assertRaises(sqlite3.IntegrityError):
                    self.db.execute("""
                        INSERT INTO attempt_usage (attempt_id, metric, quantity, unit,
                                                   cost_nanos, cost_status)
                        VALUES ('a1', 'input_audio', ?, 'audio_ms', ?, ?)
                    """, (quantity, cost, status))


if __name__ == '__main__':
    unittest.main()
