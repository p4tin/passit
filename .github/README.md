# PassIt

A local-only, encrypted password manager for macOS, Linux, and Windows. No cloud sync, no telemetry, no accounts required.

Built with Go and [Fyne](https://fyne.io).

## Features

- **AES-256-GCM encryption** — vault is encrypted with your master password using PBKDF2 key derivation (100,000 iterations)
- **Multi-account per site** — store multiple username/password pairs under each site
- **Password generator** — generates strong, pronounceable passwords with entropy validation (min 60 bits)
- **Clipboard safety** — copied passwords are automatically cleared from clipboard after 30 seconds
- **Lockout protection** — progressive lockout after failed unlock attempts (10 s → 30 s → 60 s)
- **System tray** — quick lock/unlock access from the menu bar without opening the main window
- **Real-time search** — filters across site names and usernames as you type

## Requirements

- Go 1.24+
- A C compiler (Fyne uses CGo for its OpenGL backend):

| OS      | Requirement |
|---------|-------------|
| macOS   | `xcode-select --install` |
| Linux   | `sudo apt-get install gcc libgl1-mesa-dev xorg-dev` |
| Windows | [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) |

## Build & Run

```bash
# Run directly
go run ./cmd/passit

# Build a native binary
make build        # outputs ./PassIt
```

## Vault location

| OS      | Path                             |
|---------|----------------------------------|
| macOS   | `~/.config/passit/vault.enc`     |
| Linux   | `~/.config/passit/vault.enc`     |
| Windows | `%AppData%\passit\vault.enc`     |

## Usage

| Action                  | How                                                              |
|-------------------------|------------------------------------------------------------------|
| Add a site              | Click **+** in the toolbar                                       |
| Add an account          | Open a site card → **Add Account**                               |
| Copy a password         | Click the blue **Copy** button on any account row                |
| Show a password         | Click the eye icon on any account row                            |
| Edit an account         | Click the pencil icon on any account row                         |
| Delete an account       | Click the red trash icon on any account row                      |
| Search                  | Type in the search bar — filters sites and usernames in real time |
| Change master password  | Click the ⚙ icon in the toolbar                                  |
| Lock vault              | Click the lock icon in the toolbar, or use the system tray       |

## First run

On first launch you will be prompted to create a master password. Choose a strong password — **it cannot be recovered if lost**, as it is the only key to your vault.

## Packaging & Distribution

Install the packaging tools first:

```bash
make install-tools
# installs: fyne CLI + fyne-cross
```

### Native build (run on the target OS)

```bash
make package-darwin    # → PassIt.app
make package-windows   # → PassIt.exe
make package-linux     # → PassIt-linux
```

### Cross-platform build (requires Docker)

[fyne-cross](https://github.com/fyne-io/fyne-cross) uses Docker to cross-compile for all platforms from any machine:

```bash
make cross-darwin    # → fyne-cross/dist/darwin/
make cross-windows   # → fyne-cross/dist/windows/
make cross-linux     # → fyne-cross/dist/linux/
make cross-all       # all three
```

### Automated GitHub releases

Push a version tag to trigger a release build across all three platforms via GitHub Actions:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow (`.github/workflows/release.yml`) builds native packages on macOS, Linux, and Windows runners and publishes them as GitHub Release assets automatically.

## Security notes

- The vault file is encrypted with AES-256-GCM. Without the master password it is unreadable.
- Passwords copied to clipboard are cleared automatically after 30 seconds.
- After 3 failed unlock attempts the app enforces a progressive lockout (up to 60 seconds per attempt).
- The password generator enforces a minimum entropy of 60 bits and rejects single dictionary words.
