import sqlite3
import unittest

import test_queries
import test_schema


class HealthTest(unittest.TestCase):
    setUp = test_schema.SchemaTest.setUp
    query = test_queries.QueryTest.query

    def test_latest_sample_wins_and_invalid_states_are_rejected(self):
        values = (1, 'reachability', 'healthy', 1000, 2000, 10, None, 'a' * 64)
        self.assertEqual(self.query('SetDeploymentHealth', *values).rowcount, 1)
        self.assertEqual(self.query('SetDeploymentHealth',
                                   1, 'reachability', 'unhealthy', 900, 1900, 20,
                                   'network_error', 'b' * 64).rowcount, 0)
        self.assertEqual(self.query('GetDeploymentHealth', 1).fetchone()[2], 'healthy')
        for invalid in [
            (999, 'reachability', 'healthy', 1000, 2000, 1, None, 'a' * 64),
            (1, 'reachability', 'invalid', 1000, 2000, 1, None, 'a' * 64),
            (1, 'reachability', 'healthy', 1000, 1000, 1, None, 'a' * 64),
            (1, 'reachability', 'healthy', 1000, 2000, -1, None, 'a' * 64),
        ]:
            with self.subTest(values=invalid), self.assertRaises(sqlite3.IntegrityError):
                self.query('SetDeploymentHealth', *invalid)

    def test_only_enabled_connections_and_deployments_are_selected(self):
        self.assertEqual(self.query('ListHealthDeploymentIDs').fetchall(), [(1,)])
        self.db.execute('UPDATE connections SET enabled = 0')
        self.assertEqual(self.query('ListHealthDeploymentIDs').fetchall(), [])
        self.db.execute('UPDATE connections SET enabled = 1')
        self.db.execute('UPDATE deployments SET enabled = 0')
        self.assertEqual(self.query('ListHealthDeploymentIDs').fetchall(), [])
