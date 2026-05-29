package main

import (
 "bytes"
 "encoding/base64"
 "fmt"
 "net"
 "net/http"
 "net/url"
 "strconv"
 "time"

 utls "github.com/refraction-networking/utls"
)

func main() {
 targetURL := "https://29052026-1256-1.cyberghost.es/api"
 proxyStr := ""
 proxyUser := ""
 proxyPass := ""

 proxyURL, err := url.Parse(proxyStr)
 if err != nil {
  fmt.Printf("Error configurando proxy: %v\n", err)
  return
 }

 auth := proxyUser + ":" + proxyPass
 basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

 transport := &http.Transport{
  Proxy: http.ProxyURL(proxyURL),
  DialTLS: func(network, addr string) (net.Conn, error) {
   dialer := &net.Dialer{Timeout: 15 * time.Second}
   conn, err := dialer.Dial(network, addr)
   if err != nil {
    return nil, err
   }

   config := &utls.Config{InsecureSkipVerify: true}
   uConn := utls.UClient(conn, config, utls.HelloChrome_102)
   
   if err := uConn.Handshake(); err != nil {
    uConn.Close()
    return nil, err
   }
   return uConn, nil
  },
 }

 client := &http.Client{
  Transport: transport,
  Timeout:   15 * time.Second,
  CheckRedirect: func(req *http.Request, via []*http.Request) error {
   return http.ErrUseLastResponse
  },
 }

 for i := 0; i < 60; i++ {
  fmt.Printf("Iniciando intento %d de 60...\n", i+1)

  formData := url.Values{}
  formData.Set("data", strconv.Itoa(i))
  payload := bytes.NewBufferString(formData.Encode())

  req, err := http.NewRequest("POST", targetURL, payload)
  if err != nil {
   fmt.Printf("Error creando peticion: %v\n", err)
   continue
  }

  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  req.Header.Set("Proxy-Authorization", basicAuth)
  req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
  req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
  req.Header.Set("Accept-Language", "es-ES,es;q=0.9")

  fmt.Printf("Enviando POST a %s con JA3 falsificado...\n", targetURL)
  resp, err := client.Do(req)
  if err != nil {
   fmt.Printf("Error en la iteracion %d: %v\n", i, err)
  } else {
   fmt.Printf("Respuesta recibida - Estado: %s\n", resp.Status)
   resp.Body.Close()
  }

  fmt.Println("Esperando 6 segundos para el siguiente envio...\n")
  time.Sleep(6 * time.Second)
 }

 fmt.Println("Proceso finalizado.")
}
