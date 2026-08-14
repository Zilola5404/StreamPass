# StreamPass Windows native core (Wintun)

Requires Go 1.22+ and Administrator to create the Wintun adapter.

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\build_core.ps1
```

Produces `streampasscore.exe` and `wintun.dll` in this folder. Flutter CMake
copies them next to `streampass.exe`.

`wintun.dll` is the official redistributable from https://www.wintun.net/
(GPLv2). Do not replace it with an unofficial build.
