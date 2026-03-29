package tray

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type Manager struct {
	app      fyne.App
	onOpen   func()
	onLock   func()
	onQuit   func()
}

func New(app fyne.App, onOpen, onLock, onQuit func()) *Manager {
	return &Manager{
		app:    app,
		onOpen: onOpen,
		onLock: onLock,
		onQuit: onQuit,
	}
}

func (tm *Manager) SetupSystemTray() {
	if desk, ok := tm.app.(desktop.App); ok {
		menu := fyne.NewMenu("PassIt",
			fyne.NewMenuItem("Open", func() {
				if tm.onOpen != nil {
					tm.onOpen()
				}
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Lock", func() {
				if tm.onLock != nil {
					tm.onLock()
				}
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				if tm.onQuit != nil {
					tm.onQuit()
				}
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}
}