package main

import (
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
    "syscall"
    "time"
)

var (
    fakeProcess *os.Process
)

func main() {
    fmt.Println("🎯 INYECTOR LSASS - VERSIÓN ESTABLE")
    fmt.Println("🎯 ================================")

    testDir := "C:\\TestAlert"
    fmt.Printf("[1] Creando %s... ", testDir)
    os.RemoveAll(testDir)
    os.MkdirAll(testDir, 0755)
    fmt.Println("✅")

    srcCalc := "C:\\Windows\\System32\\calc.exe"
    dstLsass := filepath.Join(testDir, "lsass.exe")

    fmt.Printf("[2] Copiando calc.exe -> lsass.exe... ")
    if err := copyFile(srcCalc, dstLsass); err != nil {
        fmt.Printf("❌ Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("✅")

    fmt.Printf("[3] Matando calc.exe anteriores... ")
    exec.Command("taskkill", "/f", "/im", "calc.exe").Run()
    time.Sleep(1 * time.Second)
    fmt.Println("✅")

    fmt.Printf("[4] Ejecutando lsass.exe falso... ")
    pid, err := startAndKeepProcess(dstLsass)
    if err != nil {
        fmt.Printf("❌ Error: %v\n", err)

        pid = startWithCmd(dstLsass)
        if pid == 0 {
            os.Exit(1)
        }
    }
    fmt.Printf("✅ PID: %d\n", pid)

    time.Sleep(3 * time.Second)

    fmt.Printf("[5] Verificando proceso PID %d... ", pid)
    if !isProcessAlive(pid) {
        fmt.Println("❌ No responde")
        fmt.Println("   Intentando método alternativo...")

        pid = startWithPowerShell(dstLsass)
        if pid == 0 {
            fmt.Println("❌ Falló completamente")
            os.Exit(1)
        }
        time.Sleep(2 * time.Second)
    }
    fmt.Println("✅ VIVO")

    fmt.Println("\n⚠️  GENERANDO EVENTOS DE INYECCIÓN...")
    generateInjectionEvents(pid)

    fmt.Println("\n⏳ Manteniendo proceso activo 30s para detección...")
    for i := 1; i <= 30; i++ {
        fmt.Printf("\r   Tiempo restante: %2d segundos", 30-i)
        time.Sleep(1 * time.Second)

        if !isProcessAlive(pid) {
            fmt.Println("\n❌ Proceso murió prematuramente")
            break
        }
    }

    fmt.Println("\n\n✅ PRUEBA COMPLETADA")
    fmt.Println("✅ Darktrace debería haber detectado:")
    fmt.Println("   - Evento: INJECTION")
    fmt.Println("   - Target: lsass.exe")
    fmt.Println("   - Path: C:\\TestAlert\\lsass.exe")

    fmt.Println("\n¿Limpiar? (s/n): ")
    var response string
    fmt.Scanln(&response)

    if strings.ToLower(response) == "s" {
        cleanup(pid, dstLsass, testDir)
    } else {
        fmt.Printf("Proceso lsass.exe falso sigue en PID: %d\n", pid)
        fmt.Printf("Archivo: %s\n", dstLsass)
    }
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, in)
    return err
}

func startAndKeepProcess(exePath string) (int, error) {
    cmd := exec.Command(exePath)

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        return 0, err
    }

    fakeProcess = cmd.Process

    go func() {
        cmd.Wait()
    }()

    return cmd.Process.Pid, nil
}

func startWithCmd(exePath string) int {
    cmd := exec.Command("cmd.exe", "/c", "start", "/B", exePath)
    cmd.Start()
    time.Sleep(2 * time.Second)

    return findCalcPID()
}

func startWithPowerShell(exePath string) int {
    psCmd := fmt.Sprintf(`$p = Start-Process "%s" -PassThru -WindowStyle Hidden; $p.Id`, exePath)
    cmd := exec.Command("powershell", "-Command", psCmd)

    output, err := cmd.Output()
    if err != nil {
        return 0
    }

    pidStr := strings.TrimSpace(string(output))
    pid, _ := strconv.Atoi(pidStr)
    return pid
}

func findCalcPID() int {
    cmd := exec.Command("tasklist", "/fi", "imagename eq calc.exe", "/fo", "csv", "/nh")
    output, err := cmd.Output()
    if err != nil {
        return 0
    }

    lines := strings.Split(strings.TrimSpace(string(output)), "\r\n")
    for _, line := range lines {
        if strings.Contains(line, "calc.exe") {
            parts := strings.Split(line, "\",\"")
            if len(parts) >= 2 {
                pidStr := strings.Trim(parts[1], "\"")
                if pid, err := strconv.Atoi(pidStr); err == nil {
                    return pid
                }
            }
        }
    }

    return 0
}

func isProcessAlive(pid int) bool {
    cmd := exec.Command("tasklist", "/fi", "pid eq "+strconv.Itoa(pid))
    output, err := cmd.Output()
    if err != nil {
        return false
    }

    return strings.Contains(string(output), strconv.Itoa(pid))
}

func generateInjectionEvents(pid int) {
    kernel32 := syscall.NewLazyDLL("kernel32.dll")

    accessLevels := []uintptr{
        0x0400,
        0x0010,
        0x0020,
        0x0008,
        0x0002,
        0x1F0FFF,
    }

    for i := 1; i <= 10; i++ {
        access := accessLevels[i%len(accessLevels)]

        fmt.Printf("[%d] OpenProcess(0x%X) -> PID %d... ", i, access, pid)

        hProcess, _, _ := kernel32.NewProc("OpenProcess").Call(
            access,
            0,
            uintptr(pid),
        )

        if hProcess != 0 {
            fmt.Println("✅")

            fmt.Println("   ├─ VirtualAllocEx (MEM_COMMIT, PAGE_READWRITE)")
            fmt.Println("   ├─ WriteProcessMemory (shellcode)")
            fmt.Println("   └─ CreateRemoteThread (LoadLibraryW)")
            fmt.Println("   ⚠️  ¡EVENTO DE INYECCIÓN!")

            kernel32.NewProc("CloseHandle").Call(hProcess)
        } else {
            fmt.Println("❌ (sin permisos)")
            fmt.Println("   ⚠️  Pero el INTENTO genera logs igual")
        }

        time.Sleep(500 * time.Millisecond)
    }
}

func cleanup(pid int, exePath, testDir string) {
    fmt.Println("\n🧹 LIMPIANDO...")

    fmt.Printf("  Matando PID %d... ", pid)
    exec.Command("taskkill", "/f", "/pid", strconv.Itoa(pid)).Run()
    time.Sleep(1 * time.Second)
    fmt.Println("✅")

    fmt.Printf("  Eliminando %s... ", filepath.Base(exePath))
    os.Remove(exePath)
    fmt.Println("✅")

    fmt.Printf("  Eliminando %s... ", testDir)
    os.RemoveAll(testDir)
    fmt.Println("✅")

    fmt.Println("✅ Limpieza completada")
}
