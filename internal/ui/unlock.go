package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"passit/internal/security"
	"passit/internal/storage"
)

type UnlockWindow struct {
	window        fyne.Window
	storage       *storage.Storage
	auth          *security.AuthManager
	passwordEntry *widget.Entry
	unlockButton  *widget.Button
	statusLabel   *widget.Label
	onUnlock      func()
	onQuit        func()
}

func NewUnlockWindow(app fyne.App, storage *storage.Storage, auth *security.AuthManager, onUnlock, onQuit func()) *UnlockWindow {
	window := app.NewWindow("PassIt — Unlock")
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(420, 280))
	window.CenterOnScreen()

	uw := &UnlockWindow{
		window:   window,
		storage:  storage,
		auth:     auth,
		onUnlock: onUnlock,
		onQuit:   onQuit,
	}

	uw.setupUI()
	return uw
}

func (uw *UnlockWindow) setupUI() {
	infoLabel := widget.NewLabel("Enter your master password to access your vault.")
	infoLabel.Alignment = fyne.TextAlignCenter
	infoLabel.Wrapping = fyne.TextWrapWord

	uw.passwordEntry = widget.NewPasswordEntry()
	uw.passwordEntry.SetPlaceHolder("Master password")
	uw.passwordEntry.OnSubmitted = func(_ string) {
		uw.attemptUnlock()
	}

	uw.unlockButton = widget.NewButton("Unlock", func() {
		uw.attemptUnlock()
	})
	uw.unlockButton.Importance = widget.HighImportance

	uw.statusLabel = widget.NewLabel("")
	uw.statusLabel.Alignment = fyne.TextAlignCenter
	uw.statusLabel.Wrapping = fyne.TextWrapWord

	formContent := container.NewVBox(
		infoLabel,
		widget.NewSeparator(),
		uw.passwordEntry,
		uw.unlockButton,
		uw.statusLabel,
	)

	card := widget.NewCard("PassIt", "Vault is locked", formContent)

	uw.window.SetContent(container.NewPadded(card))
	uw.window.SetCloseIntercept(func() {
		if uw.onQuit != nil {
			uw.onQuit()
		}
	})
	uw.updateLockoutStatus()
}

func (uw *UnlockWindow) attemptUnlock() {
	isLocked, duration := uw.auth.IsLockedOut()
	if isLocked {
		uw.statusLabel.SetText(fmt.Sprintf("Locked out for %d seconds", int(duration.Seconds())))
		return
	}

	password := uw.passwordEntry.Text
	if password == "" {
		uw.statusLabel.SetText("Please enter your password")
		return
	}

	uw.unlockButton.SetText("Unlocking...")
	uw.unlockButton.Disable()

	go func() {
		err := uw.storage.UnlockVault(password)

		time.Sleep(100 * time.Millisecond)

		if err != nil {
			uw.auth.RecordFailedAttempt()

			failedAttempts := uw.auth.GetFailedAttempts()
			var message string

			if failedAttempts >= 3 {
				remaining := uw.auth.GetRemainingLockoutTime()
				if remaining > 0 {
					message = fmt.Sprintf("Wrong password. Locked out for %d seconds.", int(remaining.Seconds()))
				} else {
					message = "Wrong password."
				}
			} else {
				message = "Wrong password."
			}

			uw.statusLabel.SetText(message)
			uw.unlockButton.SetText("Unlock")
			uw.unlockButton.Enable()
			uw.passwordEntry.SetText("")
		} else {
			uw.auth.RecordSuccessfulAttempt()
			uw.window.Hide()
			if uw.onUnlock != nil {
				uw.onUnlock()
			}
		}
	}()
}

func (uw *UnlockWindow) updateLockoutStatus() {
	isLocked, duration := uw.auth.IsLockedOut()
	if isLocked {
		uw.statusLabel.SetText(fmt.Sprintf("Locked out for %d seconds", int(duration.Seconds())))
		uw.unlockButton.Disable()
		uw.passwordEntry.Disable()

		time.AfterFunc(1*time.Second, func() {
			uw.updateLockoutStatus()
		})
	} else {
		uw.statusLabel.SetText("")
		uw.unlockButton.Enable()
		uw.passwordEntry.Enable()
	}
}

func (uw *UnlockWindow) Show() {
	uw.passwordEntry.SetText("")
	uw.statusLabel.SetText("")
	uw.unlockButton.SetText("Unlock")
	uw.updateLockoutStatus()
	uw.window.Show()
	uw.passwordEntry.FocusGained()
}

func (uw *UnlockWindow) Hide() {
	uw.window.Hide()
}
