#!/usr/bin/env python2.7

import os
import signal
import subprocess
import threading
import time
import urllib2


def main():
    shell_execute('go install -buildmode=c-shared github.com/v2pro/koala')
    shell_execute('go build -buildmode=c-shared -o koala-recoder.so github.com/v2pro/koala')
    env = os.environ.copy()
    env['LD_PRELOAD'] = '%s/koala-recoder.so' % os.path.abspath('.')
    env['KOALA_MODE'] = 'REPLAYING'
    # env['KOALA_MODE'] = 'RECORDING'
    env['SERVER_MODE'] = 'SINGLE_THREAD'
    server = subprocess.Popen(
        [
            # 'strace', '-e', 'trace=network',
            'python', 'server.py'
        ],
        env=env, stdout=subprocess.PIPE)
    time.sleep(1)

    def call_server():
        print(urllib2.urlopen('http://127.0.0.1:9000').read())


    def replay():
        print("replayed", urllib2.urlopen('http://127.0.0.1:9001', data="""
{
  "InboundTalk": {
    "Peer": {
      "IP": "127.0.0.1",
      "Port": 37360,
      "Zone": ""
    },
    "RequestTime": 1502583648166694263,
    "Request": "R0VUIC8gSFRUUC8xLjENCkFjY2VwdC1FbmNvZGluZzogaWRlbnRpdHkNCkhvc3Q6IDEyNy4wLjAuMTo5MDAwDQpDb25uZWN0aW9uOiBjbG9zZQ0KVXNlci1BZ2VudDogUHl0aG9uLXVybGxpYi8yLjcNCg0K",
    "ResponseTime": 1502583648218118763,
    "Response": "Z29vZCBkYXk="
  },
  "OutboundTalks": [
    {
      "Peer": {
        "IP": "111.206.223.205",
        "Port": 80,
        "Zone": ""
      },
      "RequestTime": 1502583648179667868,
      "Request": "R0VUIC8gSFRUUC8xLjENCkhvc3Q6IHd3dy5iYWlkdS5jb20NCkNvbm5lY3Rpb246IGtlZXAtYWxpdmUNCkFjY2VwdC1FbmNvZGluZzogZ3ppcCwgZGVmbGF0ZQ0KQWNjZXB0OiAqLyoNClVzZXItQWdlbnQ6IHB5dGhvbi1yZXF1ZXN0cy8yLjE4LjMNCg0K",
      "ResponseTime": 1502583648187122110,
      "Response": "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"
    },
    {
      "Peer": {
        "IP": "111.206.223.205",
        "Port": 80,
        "Zone": ""
      },
      "RequestTime": 1502583648197354188,
      "Request": "R0VUIC8gSFRUUC8xLjENCkhvc3Q6IHd3dy5iYWlkdS5jb20NCkNvbm5lY3Rpb246IGtlZXAtYWxpdmUNCkFjY2VwdC1FbmNvZGluZzogZ3ppcCwgZGVmbGF0ZQ0KQWNjZXB0OiAqLyoNClVzZXItQWdlbnQ6IHB5dGhvbi1yZXF1ZXN0cy8yLjE4LjMNCkNvb2tpZTogQkRPUlo9MjczMTUNCg0K",
      "ResponseTime": 1502583648215846817,
      "Response": "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"
    }
  ]
}
        """).read())

    thread1 = threading.Thread(target=replay)
    thread1.start()
    thread1.join()
    thread2 = threading.Thread(target=replay)
    thread2.start()
    thread2.join()
    time.sleep(1)
    print('send SIGTERM')
    server.send_signal(signal.SIGTERM)
    print(server.communicate()[0])

def shell_execute(cmd):
    print(cmd)
    subprocess.check_call(cmd, shell=True)


main()
