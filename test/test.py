#!/usr/bin/env python2.7

import subprocess
import signal
import urllib2
import time
import os


def main():
    shell_execute('go install -buildmode=c-shared github.com/v2pro/koala')
    shell_execute('go build -buildmode=c-shared -o koala-recoder.so github.com/v2pro/koala')
    env = os.environ.copy()
    env['LD_PRELOAD'] = '%s/koala-recoder.so' % os.path.abspath('.')
    server = subprocess.Popen(['python', 'server.py'], env=env)
    time.sleep(1)
    print(urllib2.urlopen('http://127.0.0.1:9000').read())
    server.send_signal(signal.SIGTERM)
    server.communicate()


def shell_execute(cmd):
    print(cmd)
    subprocess.check_call(cmd, shell=True)


main()
