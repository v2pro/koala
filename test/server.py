import BaseHTTPServer
import SocketServer
import socket
import threading
import urllib2

PORT = 9000

class MyHandler(BaseHTTPServer.BaseHTTPRequestHandler):
    def do_GET(self):
        urllib2.urlopen('http://postman-echo.com/get').read()
        self.send_response(200, 'OK')
        self.end_headers()
        self.wfile.write('good day')


class ThreadingMixIn:
    """Mix-in class to handle each request in a new thread."""

    # Decides how threads will act upon termination of the
    # main process
    daemon_threads = False

    def process_request_thread(self, request, client_address):
        """Same as in BaseServer but as a thread.

        In addition, exception handling is done here.

        """
        try:
            try:
                self.finish_request(request, client_address)
                self.shutdown_request(request)
            except:
                self.handle_error(request, client_address)
                self.shutdown_request(request)
        finally:
            socket.socket(socket.AF_INET, socket.SOCK_DGRAM).sendto(
                'to-koala:thread-shutdown||',
                ('127.127.127.127', 127))

    def process_request(self, request, client_address):
        """Start a new thread to process the request."""
        t = threading.Thread(target=self.process_request_thread,
                             args=(request, client_address))
        t.daemon = self.daemon_threads
        t.start()


class ThreadedTCPServer(ThreadingMixIn, SocketServer.TCPServer):
    pass


SocketServer.TCPServer.allow_reuse_address = True
httpd = ThreadedTCPServer(("", PORT), MyHandler)

print "serving at port", PORT
httpd.serve_forever()
