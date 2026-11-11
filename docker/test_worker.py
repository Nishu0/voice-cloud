"""Run inside the backend image to exercise the real API and worker together."""

import json
import os
import sqlite3
import subprocess
import tempfile
import threading
import time
import unittest
import urllib.error
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path


class WorkerIntegrationTest(unittest.TestCase):
    def test_real_worker_probes_cleans_up_and_stops_with_api(self):
        requests = []

        class HealthHandler(BaseHTTPRequestHandler):
            def do_GET(self):
                requests.append((self.path, self.headers.get('Authorization')))
                self.send_response(200)
                self.end_headers()
                self.wfile.write(b'ok')

            def log_message(self, *_):
                pass

        health_server = ThreadingHTTPServer(('127.0.0.1', 0), HealthHandler)
        self.addCleanup(health_server.server_close)
        self.addCleanup(health_server.shutdown)
        threading.Thread(target=health_server.serve_forever, daemon=True).start()
        origin = f'http://127.0.0.1:{health_server.server_port}'
        token = 'worker-integration-admin-token-at-least-32'
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'router.db'
            subprocess.run(['goose', '-env=none', '-dir', '/app/db/migrations',
                            'sqlite3', str(path), 'up'], check=True, capture_output=True)
            now = int(time.time() * 1000)
            with sqlite3.connect(path) as database:
                database.execute('''INSERT INTO connections
                    (id, name, adapter, location_kind, base_url, credential_ref)
                    VALUES (1, 'Integration runtime', 'custom', 'local', ?, 'NEVER_SEND')''', (origin,))
                database.execute("INSERT INTO models (id, name, publisher, task) VALUES (1, 'test', 'test', 'stt')")
                database.execute('''INSERT INTO deployments
                    (id, name, model_id, connection_id, upstream_model_id, settings_json)
                    VALUES (1, 'test', 1, 1, 'test', ?)''',
                    (json.dumps({'healthcheck': {'url': origin + '/healthz'}}),))
                database.execute("INSERT INTO users (id, email, name, password_hash) VALUES (1, 'test@example.com', 'Test', 'encoded-hash')")
                for index, expires in enumerate((now - 1000, now + 3600000)):
                    for table in ('user_sessions', 'user_tokens'):
                        fields = 'token_hash, user_id, created_at, expires_at'
                        values = [str(index) * 64, 1, now - 3600000, expires]
                        if table == 'user_tokens':
                            fields += ', purpose'
                            values.append('password_reset')
                        database.execute(f'INSERT INTO {table} ({fields}) VALUES ({",".join("?" for _ in values)})', values)
            environment = dict(os.environ, ADMIN_TOKEN=token, DB_PATH=str(path),
                               API_ADDR='127.0.0.1:8080', NEVER_SEND='secret-value')
            process = subprocess.Popen(['/bin/bash', '/app/docker/backend-entrypoint.sh'],
                                       cwd='/app', env=environment, stdout=subprocess.PIPE,
                                       stderr=subprocess.STDOUT, text=True)
            try:
                deadline = time.monotonic() + 10
                result = None
                while time.monotonic() < deadline:
                    self.assertIsNone(process.poll(), 'backend exited before health was available')
                    request = urllib.request.Request('http://127.0.0.1:8080/v1/deployments/1/health',
                                                     headers={'Authorization': 'Bearer ' + token})
                    try:
                        with urllib.request.urlopen(request, timeout=1) as response:
                            result = json.load(response)
                        if result['status'] == 'healthy':
                            break
                    except urllib.error.URLError:
                        pass
                    time.sleep(0.05)
                self.assertIsNotNone(result)
                self.assertEqual(result['status'], 'healthy', result)
                self.assertIsNotNone(result['latency_ms'])
                self.assertEqual(requests, [('/healthz', None)])
                with sqlite3.connect(path) as database:
                    for table in ('user_sessions', 'user_tokens'):
                        self.assertEqual(database.execute(f'SELECT count(*) FROM {table}').fetchone(), (1,))
                process.terminate()
                output, _ = process.communicate(timeout=10)
                self.assertEqual(process.returncode, 0, output)
                self.assertIn('worker stopped', output)
            finally:
                if process.poll() is None:
                    process.kill()
                    process.communicate(timeout=5)


if __name__ == '__main__':
    unittest.main()
