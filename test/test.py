#!/usr/bin/env python2.7

import subprocess
import signal
import urllib2
import time
import json
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

    # def call_server():
    #     print(urllib2.urlopen('http://127.0.0.1:9000').read())

    # thread1 = threading.Thread(target=call_server)
    # thread1.start()
    # thread2 = threading.Thread(target=call_server)
    # thread2.start()
    # thread1.join()
    # thread2.join()
    print("replayed", urllib2.urlopen('http://127.0.0.1:9001', data="""
{
  "InboundTalk": {
    "Peer": {
      "IP": "127.0.0.1",
      "Port": 36594,
      "Zone": ""
    },
    "RequestTime": 1502532391803245893,
    "Request": "R0VUIC8gSFRUUC8xLjENCkFjY2VwdC1FbmNvZGluZzogaWRlbnRpdHkNCkhvc3Q6IDEyNy4wLjAuMTo5MDAwDQpDb25uZWN0aW9uOiBjbG9zZQ0KVXNlci1BZ2VudDogUHl0aG9uLXVybGxpYi8yLjcNCg0K",
    "ResponseTime": 1502532392497692179,
    "Response": "SFRUUC8xLjAgMjAwIE9LDQpTZXJ2ZXI6IEJhc2VIVFRQLzAuMyBQeXRob24vMi43LjEzDQpEYXRlOiBTYXQsIDEyIEF1ZyAyMDE3IDEwOjA2OjMyIEdNVA0KDQpnb29kIGRheQ=="
  },
  "OutboundTalks": [
    {
      "Peer": {
        "IP": "52.44.234.173",
        "Port": 80,
        "Zone": ""
      },
      "RequestTime": 1502532392243103569,
      "Request": "R0VUIC9nZXQgSFRUUC8xLjENCkFjY2VwdC1FbmNvZGluZzogaWRlbnRpdHkNCkhvc3Q6IHBvc3RtYW4tZWNoby5jb20NCkNvbm5lY3Rpb246IGNsb3NlDQpVc2VyLUFnZW50OiBQeXRob24tdXJsbGliLzIuNw0KDQo=",
      "ResponseTime": 1502532392496271817,
      "Response": "SFRUUC8xLjEgMjAwIE9LDQpTZXJ2ZXI6IG5naW54LzEuMTAuMg0KRGF0ZTogU2F0LCAxMiBBdWcgMjAxNyAxMDowNjozMiBHTVQNCkNvbnRlbnQtVHlwZTogYXBwbGljYXRpb24vanNvbjsgY2hhcnNldD11dGYtOA0KQ29udGVudC1MZW5ndGg6IDE0Nw0KQ29ubmVjdGlvbjogY2xvc2UNCkFjY2Vzcy1Db250cm9sLUFsbG93LU9yaWdpbjogDQpBY2Nlc3MtQ29udHJvbC1BbGxvdy1DcmVkZW50aWFsczogDQpBY2Nlc3MtQ29udHJvbC1BbGxvdy1NZXRob2RzOiANCkFjY2Vzcy1Db250cm9sLUFsbG93LUhlYWRlcnM6IA0KQWNjZXNzLUNvbnRyb2wtRXhwb3NlLUhlYWRlcnM6IA0KRVRhZzogVy8iOTMtTVZTdXY1bW9JMzZDUGRndUFhZkZnZyINClZhcnk6IEFjY2VwdC1FbmNvZGluZw0Kc2V0LWNvb2tpZTogc2FpbHMuc2lkPXMlM0FnTGhRRW9IWWxBbmtKZ2kyUWdUaHdaUmVPY24wbWVvdC5VcXlrYlBBcm9teFZGJTJCbzQ5V2pOWWFnYXFWY0hPNGMwTDMyUHA0YUpHS0k7IFBhdGg9LzsgSHR0cE9ubHkNCg0KeyJhcmdzIjp7fSwiaGVhZGVycyI6eyJob3N0IjoicG9zdG1hbi1lY2hvLmNvbSIsImFjY2VwdC1lbmNvZGluZyI6ImlkZW50aXR5IiwidXNlci1hZ2VudCI6IlB5dGhvbi11cmxsaWIvMi43In0sInVybCI6Imh0dHA6Ly9wb3N0bWFuLWVjaG8uY29tL2dldCJ9"
    }
  ]
}
    """).read())
    print('send SIGTERM')
    server.send_signal(signal.SIGTERM)
    print(server.communicate()[0])


def shell_execute(cmd):
    print(cmd)
    subprocess.check_call(cmd, shell=True)


main()
