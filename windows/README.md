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

System-wide protection runs the filtering engine as a loopback DNS service on port 53 and uses Windows NRPT to redirect DNS queries to `127.0.0.1` / `::1`. It does not create a Wintun adapter, replace the default IPv4/IPv6 route, or force application traffic through a userspace VPN. Games and other applications keep using the active physical network adapter directly. Administrator privileges are required to install and remove the NRPT rule.
