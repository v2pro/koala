#!/usr/bin/env python2.7

import subprocess
import signal
import urllib2
import time
import threading
import os


def main():
    shell_execute('go install -buildmode=c-shared github.com/v2pro/koala')
    shell_execute('go build -buildmode=c-shared -o koala-recoder.so github.com/v2pro/koala')
    env = os.environ.copy()
    env['LD_PRELOAD'] = '%s/koala-recoder.so' % os.path.abspath('.')
    server = subprocess.Popen(
        [
            # 'strace', '-e', 'trace=network',
            'python', 'server.py'
        ],
        env=env, stdout=subprocess.PIPE)
    time.sleep(1)

    def call_server():
        print(urllib2.urlopen('http://127.0.0.1:9000').read())

    thread1 = threading.Thread(target=call_server)
    thread1.start()
    thread2 = threading.Thread(target=call_server)
    thread2.start()
    thread1.join()
    thread2.join()
    print('send SIGTERM')
    server.send_signal(signal.SIGTERM)
    print(server.communicate()[0])


def shell_execute(cmd):
    print(cmd)
    subprocess.check_call(cmd, shell=True)


main()
