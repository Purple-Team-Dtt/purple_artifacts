package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Ejecuta whoami usando la ruta completa
	cmd := exec.Command("C:\\Windows\\System32\\cmd.exe", "/C", "move", "./1", "./2")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error ejecutando comando: %v\n%s", err, out)
		os.Exit(1)
	}
	fmt.Print(string(out))
}
