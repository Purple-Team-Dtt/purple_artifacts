from http.server import BaseHTTPRequestHandler, HTTPServer
import ssl

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

def run(server_class=HTTPServer, handler_class=SimpleHandler, port=443):
    server_address = ('0.0.0.0', port)
    httpd = server_class(server_address, handler_class)

    # ---- SSLContext moderno ----
    context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
    context.load_cert_chain(certfile="cert.pem", keyfile="key.pem")

    httpd.socket = context.wrap_socket(httpd.socket, server_side=True)
    # -----------------------------

    print(f"[*] HTTPS listening on port {port}...")
    httpd.serve_forever()

if __name__ == '__main__':
    run()
