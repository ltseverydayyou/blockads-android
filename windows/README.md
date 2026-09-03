# BlockAds for Windows

The Windows frontend uses Compose Desktop and follows the Android app's Material 3 UI, color palette, typography, navigation model, cards, switches, and BlockAds icon artwork.

The filtering backend is the repository's existing Go tunnel/filter engine running as a loopback-only local service. Android-only OS hooks are replaced with Windows integrations where available.

## Build

From the repository root:

```powershell
powershell -ExecutionPolicy Bypass -File .\windows\build.ps1
```

The build script:

1. Tests and compiles the Go BlockAds backend.
2. Bundles the backend into the Compose Desktop application.
3. Builds the Windows installer.
4. Builds a portable ZIP.

Outputs are written to `windows/dist/`.

System-wide protection creates a Wintun `BlockAds` virtual adapter and routes IPv4 and IPv6 traffic through the repository's existing full-tunnel gVisor/filter engine. The virtual adapter advertises internal DNS addresses (`10.254.0.1` and `fd00:ad:beef::1`), so DNS is intercepted inside the tunnel like Android; physical adapter DNS settings are never rewritten. The backend binds its own upstream sockets to the physical interface to avoid VPN routing loops. Creating the adapter and routes requires administrator privileges.

