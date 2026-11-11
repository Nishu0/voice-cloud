"""Run inside the backend image to check startup and child supervision."""

import os
import signal
import sqlite3
import subprocess
import tempfile
import time
import unittest
from pathlib import Path


class BackendTest(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.directory = Path(temporary.name)
        binary_dir = self.directory / 'bin'
        binary_dir.mkdir()
        for service in ('api', 'worker'):
            binary = binary_dir / service
            binary.write_text('''#!/usr/bin/python3
import os, signal, sys
from pathlib import Path
name = Path(sys.argv[0]).name
signal.signal(signal.SIGTERM, lambda *_: sys.exit(0))
Path(os.environ['TEST_STATE'], name + '.pid').write_text(str(os.getpid()))
if os.environ.get('FAIL_SERVICE') == name:
    sys.exit(7)
signal.pause()
''')
            binary.chmod(0o755)
        self.environment = dict(os.environ, PATH=f'{binary_dir}:' + os.environ['PATH'],
                                ADMIN_TOKEN='',
                                DB_PATH=str(self.directory / 'router.db'),
                                TEST_STATE=str(self.directory))

    def start(self, **environment):
        process = subprocess.Popen(['/bin/bash', '/app/docker/backend-entrypoint.sh'],
                                   cwd='/app', env=dict(self.environment, **environment),
                                   stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
        self.addCleanup(self.cleanup_process, process)
        return process

    def cleanup_process(self, process):
        for pid_file in self.directory.glob('*.pid'):
            try:
                os.kill(int(pid_file.read_text()), signal.SIGKILL)
            except ProcessLookupError:
                pass
        if process.poll() is None:
            process.kill()
        process.communicate(timeout=5)

    def assert_children_stopped(self):
        for pid_file in self.directory.glob('*.pid'):
            self.assertFalse(Path('/proc', pid_file.read_text()).exists(), pid_file.name)

    def test_stop_terminates_both_children_and_seeds_database(self):
        process = self.start()
        deadline = time.monotonic() + 15
        while len(list(self.directory.glob('*.pid'))) < 2 and time.monotonic() < deadline:
            self.assertIsNone(process.poll())
            time.sleep(0.05)
        self.assertEqual(len(list(self.directory.glob('*.pid'))), 2)
        with sqlite3.connect(self.environment['DB_PATH']) as database:
            self.assertEqual(database.execute('SELECT count(*) FROM connections').fetchone(), (38,))
        process.terminate()
        output, _ = process.communicate(timeout=10)
        self.assertEqual(process.returncode, 0, output)
        self.assert_children_stopped()

    def test_child_failure_stops_backend_and_sibling(self):
        process = self.start(FAIL_SERVICE='worker')
        output, _ = process.communicate(timeout=15)
        self.assertEqual(process.returncode, 7, output)
        self.assert_children_stopped()

    def test_short_token_fails_before_creating_database(self):
        process = self.start(ADMIN_TOKEN='short')
        output, _ = process.communicate(timeout=10)
        self.assertNotEqual(process.returncode, 0, output)
        self.assertFalse(Path(self.environment['DB_PATH']).exists())


if __name__ == '__main__':
    unittest.main()
