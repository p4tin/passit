# PassIt - Password Manager

A simple, secure password manager for macOS, Linux, and Windows

## 1. Download & Install

### Where to download

Go to the [PassIt releases page on GitHub](https://github.com/p4tin/passit/releases) and download the package for your operating system from the latest release:

- `PassIt-macOS-<version>.dmg` - macOS
- `PassIt-Linux-<version>.tar.gz` - Linux (x86-64)
- `PassIt-Windows-<version>.msi` - Windows (x86-64)

### macOS

1. Download the `.dmg` file and open it.
2. Drag **PassIt.app** into the Applications folder.
3. Open Applications folder and double-click **PassIt** to launch it. On first open, if macOS shows a security warning, right-click the icon and choose Open.

> **Note:** The app is self-signed. You may see a security prompt on first launch.

### Linux

1. Extract the archive: `tar -xzf PassIt-Linux-<version>.tar.gz`
2. Move the binary to a directory on your PATH, e.g. `/usr/local/bin/PassIt`
3. Make it executable: `chmod +x /usr/local/bin/PassIt`
4. Run `PassIt` from a terminal or create a desktop shortcut.

### Windows

1. Download the `.msi` installer file.
2. Double-click the installer and follow the Windows installation wizard.
3. PassIt will be installed to `Program Files\PassIt` and a Start menu shortcut will be created.
4. Launch PassIt from the Start menu or Applications.

---

## 2. First Launch - Creating Your Vault

The first time PassIt starts it will prompt you to create a new vault. A vault is the encrypted file that stores all your passwords.

1. Enter a strong master password. This is the only password you need to remember.
2. Confirm the password and click **Create Vault**.
3. PassIt will open the main window, ready to use.

> **⚠️ Important:** PassIt has no password-reset mechanism. If you forget your master password your vault cannot be recovered. Store it somewhere safe.

---

## 3. Daily Use

### Unlocking the vault

Each time you start PassIt you will be asked for your master password. Enter it and click **Unlock**.

### Adding a site

1. Click the **+** button in the toolbar.
2. Type the site name (e.g. GitHub, Gmail) and click **Add Site**.

### Adding an account to a site

1. Find the site card and click **Add Account**.
2. Enter the username or email address.
3. A 16-character password is generated automatically. Click **Generate New** for a different one, or type your own.
4. Optionally add notes, then click **Add Account**.

### Copying a password

Click the **Copy** button next to any account. The password is placed on the clipboard and automatically cleared after 30 seconds.

### Viewing or editing an account

- **Eye icon** - Reveals the password in a pop-up dialog.
- **Pencil icon** - Opens the edit dialog to change username, password, or notes.

### Searching

Use the search bar at the top to filter by site name or username. Clear the field to show all entries.

### Locking the vault

Click the **lock** icon in the toolbar to lock PassIt immediately. You can also lock or quit from the system-tray icon.

### Changing your master password

1. Click the **settings (gear)** icon in the toolbar.
2. Enter your current password, then choose and confirm a new one.
3. Click **Change Password**.

---

## 4. System Tray

PassIt adds an icon to the system tray (Windows/Linux) or menu bar (macOS). Right-click it for quick access to:

- **Open** - Bring the main window to the front (unlocking first if needed).
- **Lock** - Lock the vault without quitting.
- **Quit** - Exit PassIt completely.