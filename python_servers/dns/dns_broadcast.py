# dns_broadcast.py
from dnslib.server import DNSServer, BaseResolver
from dnslib import RR, A, QTYPE
import argparse

class BroadcastResolver(BaseResolver):
    def __init__(self, ip="255.255.255.255", ttl=60):
        self.ip = ip
        self.ttl = ttl

    def resolve(self, request, handler):
        reply = request.reply()
        q = request.q
        qname = q.qname
        qtype = QTYPE[q.qtype]
        # Añadimos un A record con 255.255.255.255 para cualquier consulta
        # (siempre que tenga sentido; incluso si la consulta no es tipo A)
        reply.add_answer(RR(rname=qname, rtype=1, rclass=1, ttl=self.ttl, rdata=A(self.ip)))
        # marcar como authoritative
        reply.auth = True
        return reply

if __name__ == "__main__":
    p = argparse.ArgumentParser()
    p.add_argument("--port", type=int, default=53, help="Puerto UDP/TCP a escuchar (por defecto 53, puede requerir root)")
    p.add_argument("--address", default="64.225.77.122", help="Interfaz a escuchar")
    p.add_argument("--ip", default="255.255.255.255", help="IP a devolver en el A record")
    p.add_argument("--ttl", type=int, default=60)
    args = p.parse_args()

    resolver = BroadcastResolver(ip=args.ip, ttl=args.ttl)
    server = DNSServer(resolver, port=args.port, address=args.address, tcp=True)
    print(f"Iniciando DNS que responde {args.ip} en {args.address}:{args.port} (ctrl-c para salir)")
    server.start_thread()

    try:
        while True:
            import time
            time.sleep(1)
    except KeyboardInterrupt:
        print("Deteniendo")
        server.stop()
