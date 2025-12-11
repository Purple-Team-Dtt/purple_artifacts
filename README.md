

# **purple_artifacts**

Este repositorio contiene una colección de artefactos utilizados para pruebas, compilación, firma y empaquetado. Los archivos están organizados por carpetas, pero su función puede entenderse de forma genérica según su nombre.

---

## 📄 **Tipos de archivos**

A continuación se describe el propósito general de cada tipo de archivo según su nombre o extensión:

### **Archivos fuente**

* **`main.go`**
  Código fuente en Go que genera los binarios correspondientes dentro de cada carpeta.

### **Binarios compilados**

* **`*.exe`**
  Ejecutables compilados a partir del código fuente para diferentes pruebas o utilidades.
* **`*.dll`**
  Bibliotecas dinámicas utilizadas en pruebas de carga o ejecución desde binarios.

### **Binarios firmados**

* **`*_signed.exe` / `*_signed.dll`**
  Versiones firmadas de los binarios originales, utilizadas para validar procesos de firma o ejecución con integridad verificada.

### **Binarios codificados**

* **`*.b64`**
  Binarios codificados en Base64, normalmente para transporte, incrustación o almacenamiento seguro dentro de texto plano.
* **`*.rev`**
  Variantes revertidas para pruebas específicas.



### **Herramientas internas**

* **`signer.exe`**
  Herramienta para re-firmar binarios dentro del flujo de pruebas. 

---

### Archivos `.b64`

Los archivos con extensión `.b64` son binarios codificados en Base64 para facilitar su transporte o almacenamiento. Para recuperar el archivo original en **Linux**, basta con decodificarlo:

```bash
base64 -d archivo.exe.b64 > archivo.exe
```

En **PowerShell**, puede restaurarse con:

```powershell
[IO.File]::WriteAllBytes("archivo.exe", [Convert]::FromBase64String((Get-Content "archivo.exe.b64" -Raw)))
```

---

### Archivos `.rev` (Base64 invertido)

Los archivos `.rev` contienen binarios que fueron codificados en Base64 y luego invertidos con `rev` (ejemplo de creación: `cat archivo.exe | base64 -w0 | rev > archivo.exe.b64.rev`). Para recuperar el archivo original en **Linux**, simplemente se revierte la cadena y se decodifica:

```bash
rev archivo.exe.b64.rev | base64 -d > archivo.exe
```

En **PowerShell**, la cadena debe convertirse en un arreglo, invertirse y decodificarse:

```powershell
$rev = Get-Content "archivo.exe.b64.rev" -Raw
$arr = $rev.ToCharArray()
[Array]::Reverse($arr)
$normal = -join $arr
[IO.File]::WriteAllBytes("archivo.exe", [Convert]::FromBase64String($normal))
```

---
## 🧪 Propósito del repositorio

Este repositorio sirve como colección de artefactos para:

* Probar compilación y firma de binarios
* Validar carga de DLLs y ejecución desde distintos tipos de ejecutables

* Probar servicios o servidores simples
* Mantener ejemplos de referencia para experimentos o entornos de laboratorio


### Recursos utiles
- [RegexChecker](https://regex101.com/)
- [WebHook](https://webhook.site/)
