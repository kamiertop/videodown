# Build Directory

This directory contains platform build configuration and the application icon.
The project distributes portable archives and does not build platform installers.

- `appicon.png` - application icon used by the Wails application and release archives.
- `darwin/` - macOS build metadata.
- `linux/` - Linux Taskfile.
- `windows/` - Windows Taskfile and application manifest.
- `bin/` - local build output; it is ignored by Git.

Use the root `Taskfile.yml` to build or package portable artifacts.
