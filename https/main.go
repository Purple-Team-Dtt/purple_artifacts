package main

import (
    "crypto/tls"
    "log"
    "net/http"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Servidor TLS 1.0/1.1 sin HTTP/2"))
    })

    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS10,
        MaxVersion: tls.VersionTLS12,

        CipherSuites: []uint16{
            tls.TLS_RSA_WITH_AES_128_CBC_SHA,
            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
            tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
            tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
        },

        PreferServerCipherSuites: true,

        // 👇 DESACTIVACIÓN ABSOLUTA DE HTTP/2
        NextProtos: []string{"http/1.1"},
    }

    server := &http.Server{
        Addr: ":8443",
        Handler: mux,
        TLSConfig: tlsConfig,

        // 👇 Esto es CRÍTICO: evita que Go inserte "h2" automáticamente
        TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
    }

    log.Println("Servidor HTTPS escuchando en https://localhost:8443 ...")
    log.Fatal(server.ListenAndServeTLS("cert.pem", "key.pem"))
}
