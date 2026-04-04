package ui

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"passit/internal/clipboard"
	"passit/internal/models"
	"passit/internal/security"
	"passit/internal/storage"
)

type MainWindow struct {
	window         fyne.Window
	storage        *storage.Storage
	clipboard      *clipboard.Manager
	searchEntry    *widget.Entry
	sitesContainer *container.Scroll
	sitesContent   *fyne.Container
	sitesData      []*models.Site
	statusLabel    *widget.Label
	onLock         func()
	onQuit         func()
}

func NewMainWindow(app fyne.App, storage *storage.Storage, clipboardMgr *clipboard.Manager, onLock, onQuit func()) *MainWindow {
	window := app.NewWindow("PassIt — Password Manager")
	window.Resize(fyne.NewSize(820, 640))
	window.CenterOnScreen()
	clipboardMgr.SetWindow(window)

	mw := &MainWindow{
		window:    window,
		storage:   storage,
		clipboard: clipboardMgr,
		onLock:    onLock,
		onQuit:    onQuit,
	}

	mw.setupUI()
	mw.refreshSites()
	return mw
}

func (mw *MainWindow) setupUI() {
	toolbar := mw.createToolbar()

	mw.searchEntry = widget.NewEntry()
	mw.searchEntry.SetPlaceHolder("Search sites or usernames...")
	mw.searchEntry.OnChanged = func(text string) {
		mw.filterSites(text)
	}
	searchBar := container.NewBorder(
		nil, nil,
		widget.NewIcon(theme.SearchIcon()),
		nil,
		mw.searchEntry,
	)

	mw.statusLabel = widget.NewLabel("")
	mw.statusLabel.Alignment = fyne.TextAlignCenter
	mw.statusLabel.Hide()

	mw.sitesContent = container.NewVBox()
	mw.sitesContainer = container.NewScroll(container.NewPadded(mw.sitesContent))

	topSection := container.NewVBox(toolbar, mw.statusLabel, searchBar)

	content := container.NewBorder(
		topSection,
		nil, nil, nil,
		mw.sitesContainer,
	)

	mw.window.SetContent(content)
	mw.window.SetCloseIntercept(func() {
		if mw.onQuit != nil {
			mw.onQuit()
		}
	})
}

func (mw *MainWindow) createToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			mw.showAddSiteDialog()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			mw.showExportDialog()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			mw.showImportDialog()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			mw.showChangePasswordDialog()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			mw.openDocs()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.LogoutIcon(), func() {
			mw.lock()
		}),
	)
}

func (mw *MainWindow) createSiteCard(site *models.Site) fyne.CanvasObject {
	accountsContainer := container.NewVBox()

	for i, account := range site.Accounts {
		accountsContainer.Add(mw.createAccountItemForSite(site, account))
		if i < len(site.Accounts)-1 {
			accountsContainer.Add(widget.NewSeparator())
		}
	}

	if len(site.Accounts) == 0 {
		noAccounts := widget.NewLabel("No accounts yet")
		noAccounts.Alignment = fyne.TextAlignCenter
		accountsContainer.Add(noAccounts)
	}

	addBtn := widget.NewButtonWithIcon("Add Account", theme.ContentAddIcon(), func() {
		mw.showAddAccountDialog(site)
	})
	addBtn.Importance = widget.HighImportance

	deleteBtn := widget.NewButtonWithIcon("Delete Site", theme.DeleteIcon(), func() {
		mw.deleteSite(site)
	})
	deleteBtn.Importance = widget.DangerImportance

	buttonRow := container.NewHBox(addBtn, deleteBtn)

	content := container.NewVBox(
		accountsContainer,
		widget.NewSeparator(),
		buttonRow,
	)

	subtitle := fmt.Sprintf("%d account(s)", len(site.Accounts))
	return widget.NewCard(site.Name, subtitle, content)
}

func (mw *MainWindow) createAccountItemForSite(site *models.Site, account *models.Account) fyne.CanvasObject {
	usernameLabel := widget.NewLabel(account.Username)
	usernameLabel.TextStyle = fyne.TextStyle{Bold: true}

	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		mw.clipboard.Copy(account.Password)
		mw.showCopyFeedback()
	})
	copyBtn.Importance = widget.HighImportance

	showBtn := widget.NewButtonWithIcon("", theme.VisibilityIcon(), func() {
		dialog.ShowInformation("Password — "+account.Username, account.Password, mw.window)
	})

	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		mw.showEditAccountDialog(site, account)
	})

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		mw.deleteAccount(site, account)
	})
	deleteBtn.Importance = widget.DangerImportance

	buttons := container.NewHBox(copyBtn, showBtn, editBtn, deleteBtn)
	return container.NewBorder(nil, nil, usernameLabel, buttons)
}

func (mw *MainWindow) createEmptyState() fyne.CanvasObject {
	title := widget.NewLabel("No passwords saved yet")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Click the + button in the toolbar to add your first site.")
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.Wrapping = fyne.TextWrapWord

	return container.NewCenter(container.NewVBox(
		container.NewCenter(widget.NewIcon(theme.StorageIcon())),
		title,
		subtitle,
	))
}

func (mw *MainWindow) showCopyFeedback() {
	mw.statusLabel.SetText("✓ Copied to clipboard — clears in 30 seconds")
	mw.statusLabel.Show()
	time.AfterFunc(3*time.Second, func() {
		mw.statusLabel.Hide()
	})
}

func (mw *MainWindow) refreshSites() {
	vault := mw.storage.GetVault()
	if vault != nil {
		mw.sitesData = vault.Sites
	} else {
		mw.sitesData = []*models.Site{}
	}

	mw.sitesContent.Objects = nil

	if len(mw.sitesData) == 0 {
		mw.sitesContent.Add(mw.createEmptyState())
	} else {
		for _, site := range mw.sitesData {
			mw.sitesContent.Add(mw.createSiteCard(site))
		}
	}

	mw.sitesContent.Refresh()
}

func (mw *MainWindow) filterSites(query string) {
	if query == "" {
		mw.refreshSites()
		return
	}

	vault := mw.storage.GetVault()
	if vault == nil {
		return
	}

	sites, _ := vault.Search(query)
	mw.sitesData = sites

	mw.sitesContent.Objects = nil

	if len(mw.sitesData) == 0 {
		noResults := widget.NewLabel("No results for \"" + query + "\"")
		noResults.Alignment = fyne.TextAlignCenter
		mw.sitesContent.Add(container.NewCenter(noResults))
	} else {
		for _, site := range mw.sitesData {
			mw.sitesContent.Add(mw.createSiteCard(site))
		}
	}

	mw.sitesContent.Refresh()
}

func (mw *MainWindow) showAddSiteDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("e.g. GitHub, Gmail, Netflix")
	nameEntry.Resize(fyne.NewSize(400, nameEntry.MinSize().Height))

	form := container.NewVBox(
		widget.NewLabel("Site Name:"),
		nameEntry,
	)

	var d *dialog.CustomDialog

	addBtn := widget.NewButton("Add Site", func() {
		if nameEntry.Text != "" {
			site := models.NewSite(nameEntry.Text, "")
			mw.storage.GetVault().AddSite(site)
			mw.storage.Save()
			mw.refreshSites()
			d.Hide()
		}
	})
	addBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})
	cancelBtn.Importance = widget.LowImportance

	buttons := container.NewHBox(addBtn, cancelBtn)
	content := container.NewVBox(form, buttons)

	d = dialog.NewCustom("Add Site", "Close", content, mw.window)
	d.Resize(fyne.NewSize(450, 160))
	d.Show()
}

func (mw *MainWindow) showAddAccountDialog(site *models.Site) {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username or email")
	usernameEntry.Resize(fyne.NewSize(400, usernameEntry.MinSize().Height))

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(security.GeneratePassword(16))
	passwordEntry.Resize(fyne.NewSize(400, passwordEntry.MinSize().Height))

	generateBtn := widget.NewButton("Generate New", func() {
		passwordEntry.SetText(security.GeneratePassword(16))
	})

	var showPasswordBtn *widget.Button
	showPasswordBtn = widget.NewButton("Show", func() {
		if passwordEntry.Password {
			passwordEntry.Password = false
			showPasswordBtn.SetText("Hide")
		} else {
			passwordEntry.Password = true
			showPasswordBtn.SetText("Show")
		}
		passwordEntry.Refresh()
	})

	passwordContainer := container.NewBorder(
		nil, nil, nil,
		container.NewHBox(generateBtn, showPasswordBtn),
		passwordEntry,
	)

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Notes (optional)")
	notesEntry.Resize(fyne.NewSize(400, 80))

	form := container.NewVBox(
		widget.NewLabel("Username:"),
		usernameEntry,
		widget.NewLabel("Password:"),
		passwordContainer,
		widget.NewLabel("Notes:"),
		notesEntry,
	)

	var d *dialog.CustomDialog

	addBtn := widget.NewButton("Add Account", func() {
		if usernameEntry.Text != "" && passwordEntry.Text != "" {
			account := models.NewAccount(usernameEntry.Text, passwordEntry.Text, notesEntry.Text)
			site.AddAccount(account)
			mw.storage.Save()
			mw.refreshSites()
			d.Hide()
		}
	})
	addBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})
	cancelBtn.Importance = widget.LowImportance

	buttons := container.NewHBox(addBtn, cancelBtn)
	content := container.NewVBox(form, buttons)

	d = dialog.NewCustom("Add Account", "Close", content, mw.window)
	d.Resize(fyne.NewSize(500, 370))
	d.Show()
}

func (mw *MainWindow) showEditAccountDialog(site *models.Site, account *models.Account) {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetText(account.Username)
	usernameEntry.Resize(fyne.NewSize(400, usernameEntry.MinSize().Height))

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(account.Password)
	passwordEntry.Resize(fyne.NewSize(400, passwordEntry.MinSize().Height))

	generateBtn := widget.NewButton("Generate New", func() {
		passwordEntry.SetText(security.GeneratePassword(16))
	})

	var showPasswordBtn *widget.Button
	showPasswordBtn = widget.NewButton("Show", func() {
		if passwordEntry.Password {
			passwordEntry.Password = false
			showPasswordBtn.SetText("Hide")
		} else {
			passwordEntry.Password = true
			showPasswordBtn.SetText("Show")
		}
		passwordEntry.Refresh()
	})

	passwordContainer := container.NewBorder(
		nil, nil, nil,
		container.NewHBox(generateBtn, showPasswordBtn),
		passwordEntry,
	)

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetText(account.Notes)
	notesEntry.Resize(fyne.NewSize(400, 80))

	form := container.NewVBox(
		widget.NewLabel("Username:"),
		usernameEntry,
		widget.NewLabel("Password:"),
		passwordContainer,
		widget.NewLabel("Notes:"),
		notesEntry,
	)

	var d *dialog.CustomDialog

	saveBtn := widget.NewButton("Save Changes", func() {
		if usernameEntry.Text != "" && passwordEntry.Text != "" {
			account.Username = usernameEntry.Text
			account.Password = passwordEntry.Text
			account.Notes = notesEntry.Text
			mw.storage.Save()
			mw.refreshSites()
			d.Hide()
		}
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})
	cancelBtn.Importance = widget.LowImportance

	buttons := container.NewHBox(saveBtn, cancelBtn)
	content := container.NewVBox(form, buttons)

	d = dialog.NewCustom("Edit Account", "Close", content, mw.window)
	d.Resize(fyne.NewSize(500, 370))
	d.Show()
}

func (mw *MainWindow) deleteAccount(site *models.Site, account *models.Account) {
	dialog.ShowConfirm("Delete Account",
		fmt.Sprintf("Are you sure you want to delete the account for '%s'?", account.Username),
		func(confirmed bool) {
			if confirmed {
				site.RemoveAccount(account.ID)
				mw.storage.Save()
				mw.refreshSites()
			}
		}, mw.window)
}

func (mw *MainWindow) deleteSite(site *models.Site) {
	dialog.ShowConfirm("Delete Site",
		fmt.Sprintf("Delete '%s' and all its accounts? This cannot be undone.", site.Name),
		func(confirmed bool) {
			if confirmed {
				mw.storage.GetVault().RemoveSite(site.ID)
				mw.storage.Save()
				mw.refreshSites()
			}
		}, mw.window)
}

func (mw *MainWindow) showChangePasswordDialog() {
	oldPasswordEntry := widget.NewPasswordEntry()
	oldPasswordEntry.SetPlaceHolder("Current master password")
	oldPasswordEntry.Resize(fyne.NewSize(400, oldPasswordEntry.MinSize().Height))

	newPasswordEntry := widget.NewPasswordEntry()
	newPasswordEntry.SetPlaceHolder("New master password")
	newPasswordEntry.Resize(fyne.NewSize(400, newPasswordEntry.MinSize().Height))

	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry.SetPlaceHolder("Confirm new password")
	confirmPasswordEntry.Resize(fyne.NewSize(400, confirmPasswordEntry.MinSize().Height))

	form := container.NewVBox(
		widget.NewLabel("Current Password:"),
		oldPasswordEntry,
		widget.NewLabel("New Password:"),
		newPasswordEntry,
		widget.NewLabel("Confirm Password:"),
		confirmPasswordEntry,
	)

	var d *dialog.CustomDialog

	changeBtn := widget.NewButton("Change Password", func() {
		if newPasswordEntry.Text != confirmPasswordEntry.Text {
			dialog.ShowError(fmt.Errorf("passwords do not match"), mw.window)
			return
		}

		feedback := security.ValidatePasswordStrength(newPasswordEntry.Text)
		if !feedback.Valid {
			dialog.ShowError(fmt.Errorf("weak password: %s", strings.Join(feedback.Issues, "; ")), mw.window)
			return
		}

		if err := mw.storage.ChangePassword(oldPasswordEntry.Text, newPasswordEntry.Text); err != nil {
			dialog.ShowError(err, mw.window)
		} else {
			dialog.ShowInformation("Success", "Master password changed successfully.", mw.window)
			d.Hide()
		}
	})
	changeBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})
	cancelBtn.Importance = widget.LowImportance

	buttons := container.NewHBox(changeBtn, cancelBtn)
	content := container.NewVBox(form, buttons)

	d = dialog.NewCustom("Change Master Password", "Close", content, mw.window)
	d.Resize(fyne.NewSize(450, 270))
	d.Show()
}

func (mw *MainWindow) lock() {
	if mw.onLock != nil {
		mw.onLock()
	}
}

func (mw *MainWindow) Show() {
	mw.refreshSites()
	mw.window.Show()
}

func (mw *MainWindow) Hide() {
	mw.window.Hide()
}

func (mw *MainWindow) showExportDialog() {
	dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, mw.window)
			return
		}
		if uc == nil {
			return // cancelled
		}
		defer uc.Close()
		if err := mw.storage.GetVault().ExportCSV(uc); err != nil {
			dialog.ShowError(err, mw.window)
			return
		}
		dialog.ShowInformation("Export", "Vault exported successfully.", mw.window)
	}, mw.window)
}

func (mw *MainWindow) showImportDialog() {
	dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, mw.window)
			return
		}
		if uc == nil {
			return // cancelled
		}
		defer uc.Close()
		if err := mw.storage.GetVault().ImportCSV(uc); err != nil {
			dialog.ShowError(err, mw.window)
			return
		}
		mw.storage.Save()
		mw.refreshSites()
		dialog.ShowInformation("Import", "Vault imported successfully.", mw.window)
	}, mw.window)
}

func (mw *MainWindow) openDocs() {
	docsURL, _ := url.Parse("https://p4tin.github.io/passit/")
	fyne.CurrentApp().OpenURL(docsURL)
}
