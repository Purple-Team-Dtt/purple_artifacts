package main

import (
 "bytes"
 "encoding/base64"
 "fmt"
 "net"
 "net/http"
 "net/url"
 "strconv"
 "sync"
 "time"

 utls "github.com/refraction-networking/utls"
)

func main() {
 targetURL := ""
 proxyStr := ""
 proxyUser := ""
 proxyPass := ""
 
 // Es mantenen les 300 peticions totals
 totalPeticions := 300 

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
  MaxIdleConns:        150, 
  MaxIdleConnsPerHost: 150,
 }

 client := &http.Client{
  Transport: transport,
  Timeout:   15 * time.Second, // Augmentat a 15s per a major estabilitat
  CheckRedirect: func(req *http.Request, via []*http.Request) error {
   return http.ErrUseLastResponse
  },
 }

 var wg sync.WaitGroup

 fmt.Printf("Iniciando el envío de %d peticiones (una cada 22 segundos) hacia %s...\n", totalPeticions, targetURL)

 for i := 1; i <= totalPeticions; i++ {
  wg.Add(1)

  // Llançament de la Goroutine
  go func(id int) {
   defer wg.Done()

   formData := url.Values{}
   formData.Set("data", strconv.Itoa(id))
   payload := bytes.NewBufferString(formData.Encode())

   req, err := http.NewRequest("POST", targetURL, payload)
   if err != nil {
    fmt.Printf("[Petición #%d] Error creando petición: %v\n", id, err)
    return
   }

   req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
   req.Header.Set("Proxy-Authorization", basicAuth)
   req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
   req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
   req.Header.Set("Accept-Language", "es-ES,es;q=0.9")

   fmt.Printf("[Petición #%d] Enviando POST...\n", id)
   resp, err := client.Do(req)
   if err != nil {
    fmt.Printf("[Petición #%d] Error en el envío: %v\n", id, err)
   } else {
    fmt.Printf("[Petición #%d] Respuesta recibida - Estado: %s\n", id, resp.Status)
    resp.Body.Close()
   }
  }(i)

  // MODIFICACIÓ CLAU: Espera exacta de 22 segons abans de llançar la següent petició
  time.Sleep(22 * time.Second)
 }

 fmt.Println("Todas las peticiones han sido disparadas. Esperando que terminen las últimas activas...")
 wg.Wait()

 fmt.Println("Proceso completado. Se han procesado las 300 peticiones.")
}
