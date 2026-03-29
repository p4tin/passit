package clipboard

import (
	"time"

	"fyne.io/fyne/v2"
)

type Manager struct {
	timer       *time.Timer
	currentText string
	window      fyne.Window
}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) SetWindow(w fyne.Window) {
	m.window = w
}

func (m *Manager) Copy(text string) {
	if m.timer != nil {
		m.timer.Stop()
	}

	m.currentText = text
	m.copyToSystemClipboard(text)

	m.timer = time.AfterFunc(30*time.Second, func() {
		m.currentText = ""
		m.copyToSystemClipboard("")
	})
}

func (m *Manager) ClearClipboard() {
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
	m.currentText = ""
	m.copyToSystemClipboard("")
}

func (m *Manager) StopTimer() {
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
}

func (m *Manager) copyToSystemClipboard(text string) {
	if m.window == nil {
		return
	}
	m.window.Clipboard().SetContent(text)
}