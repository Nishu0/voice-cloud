"""Real console + Go API integration; run against an isolated, fresh Compose stack."""
import http.cookiejar
import json
import os
import unittest
import urllib.error
import urllib.parse
import urllib.request


class ConsoleTest(unittest.TestCase):
    def test_signup_configuration_sessions_and_authorization(self):
        web = os.environ.get('WEB_URL', 'http://web:3000')
        backend = os.environ.get('API_URL', 'http://backend:8080')
        origin = os.environ.get('ORIGIN', 'http://127.0.0.1:13000')
        cookies = http.cookiejar.CookieJar()
        client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cookies))

        def get(path):
            with client.open(web + path, timeout=15) as response:
                return response.geturl(), response.read().decode(), response.headers

        def post(path, values, status=200, request_origin=origin):
            request = urllib.request.Request(web + path, urllib.parse.urlencode(values).encode(),
                                             headers={'Origin': request_origin})
            try:
                response = client.open(request, timeout=15)
            except urllib.error.HTTPError as response:
                self.assertEqual(response.code, status, response.read().decode())
                return ''
            with response:
                body = response.read().decode()
                self.assertEqual(response.status, status, body)
                return body

        def api(path, token=None, method='GET', body=None, status=200):
            headers = {'Content-Type': 'application/json'}
            if token:
                headers['Authorization'] = 'Bearer ' + token
            request = urllib.request.Request(backend + '/v1/' + path,
                None if body is None else json.dumps(body).encode(), headers, method=method)
            try:
                response = urllib.request.urlopen(request, timeout=15)
            except urllib.error.HTTPError as response:
                self.assertEqual(response.code, status)
                return None
            with response:
                self.assertEqual(response.status, status)
                return json.load(response) if response.status != 204 else None

        self.assertTrue(api('auth/status')['needs_setup'])
        self.assertTrue(get('/')[0].endswith('/signup'))
        post('/signup', {'name': 'Console Admin', 'email': 'admin@example.test', 'password': 'console-test-password'})
        self.assertFalse(api('auth/status')['needs_setup'])
        session = next(cookie for cookie in cookies if cookie.name == 'voicecloud_session')
        self.assertIn('HttpOnly', session._rest)
        self.assertEqual(session._rest.get('SameSite'), 'Lax')
        token = session.value
        self.assertTrue(get('/signup')[0].endswith('/login'))
        post('/models?/save', {'name': 'Integration Model', 'publisher': 'test', 'task': 'stt', 'version': '1'})
        model = next(row for row in api('models', token) if row['name'] == 'Integration Model')
        post('/manage/connections?/save', {'name': 'Integration Local', 'adapter': 'custom',
            'location_kind': 'local', 'base_url': 'http://localhost:9000', 'enabled': 'on'})
        connection = next(row for row in api('connections', token) if row['name'] == 'Integration Local')
        post('/manage/deployments?/save', {'name': 'Integration Deployment', 'model_id': model['id'],
            'connection_id': connection['id'], 'upstream_model_id': 'test-model', 'enabled': 'on',
            'capabilities_json': '{}', 'pricing_json': '{}', 'settings_json': '{}'})
        deployment = next(row for row in api('deployments', token) if row['name'] == 'Integration Deployment')
        post('/manage/deployments?/language', {'id': deployment['id'], 'language': 'en-US'})
        self.assertEqual(api(f"deployments/{deployment['id']}/languages", token)[0]['language_tag'], 'en-US')
        post('/manage/policies?/save', {'name': 'Integration Policy', 'task': 'stt', 'objective': 'latency',
            'max_attempts': 2, 'fallback_enabled': 'on', 'constraints_json': '{}'})
        policy = next(row for row in api('policies', token) if row['name'] == 'Integration Policy')
        post('/manage/policies?/target', {'id': policy['id'], 'deployment_id': deployment['id'], 'priority': 0})
        self.assertEqual(api(f"policies/{policy['id']}/targets", token)[0]['deployment_id'], deployment['id'])
        post('/api-keys?/create', {'name': 'Integration Key', 'days': 30, 'policy': policy['id']})
        key = next(row for row in api('api-keys', token) if row['name'] == 'Integration Key')
        self.assertEqual(key['created_by_user_id'], api('auth/me', token)['id'])
        post('/api-keys?/revoke', {'id': key['id']})
        self.assertIsNotNone(api(f"api-keys/{key['id']}", token)['revoked_at'])
        post('/settings/account?/profile', {'name': 'Updated Admin'})
        self.assertEqual(api('auth/me', token)['name'], 'Updated Admin')
        post('/settings/account?/member', {'name': 'Member', 'email': 'member@example.test', 'password': 'member-test-password'})
        member = api('auth/login', method='POST', body={'email': 'member@example.test', 'password': 'member-test-password'})['token']
        api('models', member, method='POST', body={'name': 'Blocked', 'publisher': 'test', 'task': 'llm'}, status=403)
        post('/settings/account?/profile', {'name': 'CSRF'}, status=403, request_origin='https://untrusted.example')
        self.assertEqual(api('auth/me', token)['name'], 'Updated Admin')
        for path in ['/', '/models', '/api-keys', '/activity', '/logs', '/quickstart', '/settings/account',
                     '/manage/connections', '/manage/deployments', '/manage/policies']:
            _, body, headers = get(path)
            self.assertNotIn('Internal Error', body, path)
            self.assertNotIn(token, body, path)
            self.assertNotIn('console-test-password', body, path)
            self.assertIn('no-store', headers.get('Cache-Control', ''))
        post('/logout', {})
        api('auth/me', token, status=401)
        self.assertTrue(get('/models')[0].endswith('/login'))
        post('/login', {'email': 'admin@example.test', 'password': 'wrong-password'}, status=401)
        post('/login', {'email': 'ADMIN@example.test', 'password': 'console-test-password'})
        self.assertIn('Integration Model', get('/models')[1])


if __name__ == '__main__':
    unittest.main()
