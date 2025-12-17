from http.server import BaseHTTPRequestHandler, HTTPServer

class SimpleHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        print(f"Received GET request from: {self.client_address}")
        print(f"Headers:\n{self.headers}")
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"GET OK")

    def do_POST(self):
        content_length = int(self.headers.get('Content-Length', 0))
        post_data = self.rfile.read(content_length) if content_length else b''
        print(f"Received POST request from: {self.client_address}")
        print(f"Headers:\n{self.headers}")
        print(f"Body:\n{post_data.decode(errors='ignore')}")
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"POST OK")

# Configuración del servidor
def run(server_class=HTTPServer, handler_class=SimpleHandler, port=80):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print(f"[*] Listening on port {port}...")
    httpd.serve_forever()

if __name__ == '__main__':
    run()
