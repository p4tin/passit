package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"passit/internal/security"
	"passit/internal/storage"
)

type SetupWindow struct {
	window  fyne.Window
	storage *storage.Storage
	onSetup func()
	onQuit  func()
}

func NewSetupWindow(app fyne.App, storage *storage.Storage, onSetup, onQuit func()) *SetupWindow {
	window := app.NewWindow("PassIt — First Run Setup")
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(460, 360))
	window.CenterOnScreen()

	sw := &SetupWindow{
		window:  window,
		storage: storage,
		onSetup: onSetup,
		onQuit:  onQuit,
	}

	sw.setupUI()
	return sw
}

func (sw *SetupWindow) setupUI() {
	intro := widget.NewLabel("Create a strong master password to protect your vault.\nYou will need it every time you open PassIt.")
	intro.Alignment = fyne.TextAlignCenter
	intro.Wrapping = fyne.TextWrapWord

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Master password")

	confirmEntry := widget.NewPasswordEntry()
	confirmEntry.SetPlaceHolder("Confirm password")

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter
	statusLabel.Wrapping = fyne.TextWrapWord

	var createButton *widget.Button
	createButton = widget.NewButton("Create Vault", func() {
		password := passwordEntry.Text
		confirm := confirmEntry.Text

		feedback := security.ValidatePasswordStrength(password)
		if !feedback.Valid {
			statusLabel.SetText(strings.Join(feedback.Issues, "; "))
			return
		}

		if password != confirm {
			statusLabel.SetText("Passwords do not match")
			return
		}

		createButton.SetText("Creating...")
		createButton.Disable()

		go func() {
			err := sw.storage.CreateVault(password)
			if err != nil {
				statusLabel.SetText("Failed to create vault: " + err.Error())
				createButton.SetText("Create Vault")
				createButton.Enable()
			} else {
				sw.window.Hide()
				if sw.onSetup != nil {
					sw.onSetup()
				}
			}
		}()
	})
	createButton.Importance = widget.HighImportance

	formContent := container.NewVBox(
		intro,
		widget.NewSeparator(),
		passwordEntry,
		confirmEntry,
		createButton,
		statusLabel,
	)

	card := widget.NewCard("PassIt", "First run setup", formContent)

	sw.window.SetContent(container.NewPadded(card))
	sw.window.SetCloseIntercept(func() {
		if sw.onQuit != nil {
			sw.onQuit()
		}
	})
}

func (sw *SetupWindow) Show() {
	sw.window.Show()
}

func (sw *SetupWindow) Hide() {
	sw.window.Hide()
}
