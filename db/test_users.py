import sqlite3
import unittest

import test_queries
import test_schema


class UserTest(unittest.TestCase):
    setUp = test_schema.SchemaTest.setUp
    query = test_queries.QueryTest.query

    def create_user(self, email='alice@example.com'):
        return self.query('CreateUser', email, 'Alice', 'encoded-password-hash', 'member').fetchone()[0]

    def test_email_uniqueness_roles_and_safe_user_listing(self):
        user_id = self.create_user()
        self.assertEqual(self.query('GetUserByEmail', 'ALICE@example.com').fetchone()[0], user_id)
        with self.assertRaises(sqlite3.IntegrityError):
            self.create_user('ALICE@EXAMPLE.COM')
        for statement in ["UPDATE users SET role = 'owner'", "UPDATE users SET status = 'pending'",
                          "UPDATE users SET email = ' '", "UPDATE users SET password_hash = ''"]:
            with self.subTest(statement=statement), self.assertRaises(sqlite3.IntegrityError):
                self.db.execute(statement)
        cursor = self.query('ListUsers', 10, 0)
        self.assertNotIn('password_hash', [column[0] for column in cursor.description])
        self.assertEqual(cursor.fetchone()[0], user_id)

    def test_sessions_obey_expiry_revocation_and_user_status(self):
        user_id = self.create_user()
        token = 'a' * 64
        self.query('CreateUserSession', token, user_id, 1000, 2000)
        self.assertIsNone(self.query('GetActiveUserSession', token, 999).fetchone())
        self.assertIsNotNone(self.query('GetActiveUserSession', token, 1000).fetchone())
        self.assertIsNone(self.query('GetActiveUserSession', token, 2000).fetchone())
        self.query('SetUserStatus', 'disabled', 1100, user_id)
        self.assertIsNone(self.query('GetActiveUserSession', token, 1200).fetchone())
        self.query('SetUserStatus', 'active', 1300, user_id)
        self.assertEqual(self.query('RevokeUserSession', 1400, token).rowcount, 1)
        self.assertIsNone(self.query('GetActiveUserSession', token, 1500).fetchone())
        self.assertEqual(self.query('RevokeUserSession', 1600, token).rowcount, 0)

    def test_reset_and_verification_tokens_are_single_use_and_purpose_bound(self):
        user_id = self.create_user()
        token = 'b' * 64
        self.query('CreateUserToken', token, user_id, 'password_reset', 1000, 2000)
        self.assertIsNone(self.query('ConsumeUserToken', 1200, token, 'email_verification').fetchone())
        self.assertIsNone(self.query('ConsumeUserToken', 2000, token, 'password_reset').fetchone())
        self.query('SetUserStatus', 'disabled', 1100, user_id)
        self.assertIsNone(self.query('ConsumeUserToken', 1200, token, 'password_reset').fetchone())
        self.query('SetUserStatus', 'active', 1300, user_id)
        row = self.query('ConsumeUserToken', 1500, token, 'password_reset').fetchone()
        self.assertEqual(row[1:3], (user_id, 'password_reset'))
        self.assertIsNone(self.query('ConsumeUserToken', 1600, token, 'password_reset').fetchone())
        verification = 'c' * 64
        self.query('CreateUserToken', verification, user_id, 'email_verification', 1000, 2000)
        self.assertIsNotNone(self.query('ConsumeUserToken', 1500, verification, 'email_verification').fetchone())
        self.assertEqual(self.query('VerifyUserEmail', 1500, user_id).rowcount, 1)
        self.assertEqual(self.query('VerifyUserEmail', 1600, user_id).rowcount, 0)

    def test_invalid_tokens_and_orphan_references_are_rejected(self):
        user_id = self.create_user()
        for token, owner, created, expires in [('plaintext', user_id, 1000, 2000),
                                               ('G' * 64, user_id, 1000, 2000),
                                               ('d' * 64, 9999, 1000, 2000),
                                               ('e' * 64, user_id, 1000, 1000)]:
            with self.subTest(token=token, owner=owner), self.assertRaises(sqlite3.IntegrityError):
                self.query('CreateUserSession', token, owner, created, expires)
            with self.subTest(token=token, owner=owner), self.assertRaises(sqlite3.IntegrityError):
                self.query('CreateUserToken', token, owner, 'password_reset', created, expires)

    def test_user_deletion_cleans_auth_records_and_preserves_router_history(self):
        user_id = self.create_user()
        self.db.execute('UPDATE api_keys SET created_by_user_id = ? WHERE id = 1', (user_id,))
        self.query('CreateUserSession', 'f' * 64, user_id, 1000, 2000)
        self.query('CreateUserToken', '0' * 64, user_id, 'password_reset', 1000, 2000)
        self.assertEqual(self.query('DeleteUser', user_id).rowcount, 1)
        self.assertEqual(self.db.execute('SELECT count(*) FROM user_sessions').fetchone(), (0,))
        self.assertEqual(self.db.execute('SELECT count(*) FROM user_tokens').fetchone(), (0,))
        self.assertEqual(self.db.execute('SELECT created_by_user_id FROM api_keys WHERE id = 1').fetchone(), (None,))
        self.assertEqual(self.db.execute('SELECT count(*) FROM requests').fetchone(), (1,))
        self.assertEqual(self.db.execute('PRAGMA foreign_key_check').fetchall(), [])
