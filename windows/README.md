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

System-wide protection changes Windows DNS settings and therefore requires administrator privileges when protection is started.
