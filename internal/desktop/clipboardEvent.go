package desktop

import "github.com/m4n5ter/lindows/internal/desktop/clipboard"

func (manager *Manager) WriteTextToClipboard(content string) error {
	return clipboard.WriteTextToClipboard(content)
}

func (manager *Manager) ReaderTextToClipboard() (string, error) {
	return clipboard.ReaderTextToClipboard()
}
