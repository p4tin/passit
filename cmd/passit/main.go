package main

import (
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"passit/assets"
	"passit/internal/clipboard"
	"passit/internal/security"
	"passit/internal/storage"
	"passit/internal/tray"
	"passit/internal/ui"
)

type App struct {
	fyneApp     fyne.App
	storage     *storage.Storage
	auth        *security.AuthManager
	clipboard   *clipboard.Manager
	trayManager *tray.Manager

	setupWindow  *ui.SetupWindow
	unlockWindow *ui.UnlockWindow
	mainWindow   *ui.MainWindow

	isLocked bool
}

func main() {
	a := &App{}
	a.init()
	a.run()
}

func (a *App) init() {
	iconResource := fyne.NewStaticResource("icon.png", assets.Icon)
	a.fyneApp = app.NewWithID("com.p4tin.passit")
	a.fyneApp.SetIcon(iconResource)
	a.fyneApp.Settings().SetTheme(&passItTheme{})

	// Re-apply the icon once the app is fully started so the Dock tile updates.
	// On macOS the initial SetIcon call can be a no-op before NSApp is running.
	a.fyneApp.Lifecycle().SetOnStarted(func() {
		a.fyneApp.SetIcon(iconResource)
	})

	a.storage = storage.New()
	a.auth = security.NewAuthManager()
	a.clipboard = clipboard.New()

	if err := a.auth.LoadLockoutState(); err != nil {
		log.Printf("Warning: Failed to load lockout state: %v", err)
	}

	a.trayManager = tray.New(a.fyneApp, a.onTrayOpen, a.onTrayLock, a.onTrayQuit)
	a.trayManager.SetupSystemTray()

	a.setupWindow = ui.NewSetupWindow(a.fyneApp, a.storage, a.onVaultCreated, a.onTrayQuit)
	a.unlockWindow = ui.NewUnlockWindow(a.fyneApp, a.storage, a.auth, a.onVaultUnlocked, a.onTrayQuit)
	a.mainWindow = ui.NewMainWindow(a.fyneApp, a.storage, a.clipboard, a.onLock, a.onTrayQuit)

	a.isLocked = true
}

func (a *App) run() {
	if !a.storage.VaultExists() {
		a.setupWindow.Show()
	} else {
		a.unlockWindow.Show()
	}

	a.fyneApp.Run()
}

func (a *App) onVaultCreated() {
	a.unlockWindow.Show()
}

func (a *App) onVaultUnlocked() {
	a.isLocked = false
	a.mainWindow.Show()
}

func (a *App) onLock() {
	a.isLocked = true
	a.storage.Lock()
	a.mainWindow.Hide()
	a.unlockWindow.Show()
}

func (a *App) onTrayOpen() {
	if a.isLocked {
		a.unlockWindow.Show()
	} else {
		a.mainWindow.Show()
	}
}

func (a *App) onTrayLock() {
	if !a.isLocked {
		a.onLock()
	}
}

func (a *App) onTrayQuit() {
	a.fyneApp.Quit()
}

// passItTheme applies a consistent accent color across all platforms while
// delegating everything else to the system default (honours dark/light mode).
type passItTheme struct{}

func (t *passItTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 74, G: 144, B: 217, A: 255} // #4A90D9
	case theme.ColorNameFocus:
		return color.NRGBA{R: 74, G: 144, B: 217, A: 180}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *passItTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *passItTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *passItTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}