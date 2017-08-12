import BaseHTTPServer
import SocketServer
import urllib2

PORT = 9000


class MyHandler(BaseHTTPServer.BaseHTTPRequestHandler):
    def do_GET(self):
        urllib2.urlopen('http://postman-echo.com/get')
        self.send_response(200, 'hello')


SocketServer.TCPServer.allow_reuse_address = True
httpd = SocketServer.TCPServer(("", PORT), MyHandler)

print "serving at port", PORT
httpd.serve_forever()
