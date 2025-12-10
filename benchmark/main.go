package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("[*] Simulador de presión de memoria iniciado")
	
	var blocks [][]byte
	blockSize := 100 * 1024 * 1024 // 100 MB por bloque

	for {
		// Dormimos un poco para no saturar de golpe
		time.Sleep(1 * time.Second)

		// Reservamos otro bloque de memoria
		block := make([]byte, blockSize)
		for i := range block {
			block[i] = 1
		}

		blocks = append(blocks, block)

		// Mostrar memoria aproximada consumida
		totalMB := len(blocks) * (blockSize / (1024 * 1024))
		fmt.Printf("[*] Memoria reservada: %d MB\n", totalMB)

		// Verificamos presión de memoria desde Windows
		// Si detecta bajo recurso, Windows debería generar Event ID 2004 automáticamente.
		// Nota: Go no tiene una API estándar para leer memoria libre sin CGO,
		// pero esta presión constante es suficiente para que Windows active WDI.
	}
}
