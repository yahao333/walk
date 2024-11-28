//go:build windows
// +build windows

package walk

import (
	"sync"
	"sync/atomic"

	"github.com/lxn/win"
)

// Windows hotkey modifiers
const (
	modAlt      = 0x0001
	modControl  = 0x0002
	modShift    = 0x0004
	modWin      = 0x0008
	modNoRepeat = 0x4000
)

var (
	nextHotkeyID uint32
	hotkeys      sync.Map // map[uint32]*GlobalHotKey
)

// GlobalHotKey represents a registered global hotkey
type GlobalHotKey struct {
	id       uint32
	shortcut Shortcut
	window   *WindowBase
	handler  func()
}

// RegisterGlobalHotKey registers a global hotkey with Windows
func RegisterGlobalHotKey(window *WindowBase, shortcut Shortcut, handler func()) (*GlobalHotKey, error) {
	id := atomic.AddUint32(&nextHotkeyID, 1)

	modifiers := uint32(0)
	if shortcut.Modifiers&ModShift != 0 {
		modifiers |= modShift
	}
	if shortcut.Modifiers&ModControl != 0 {
		modifiers |= modControl
	}
	if shortcut.Modifiers&ModAlt != 0 {
		modifiers |= modAlt
	}

	if !win.RegisterHotKey(window.hWnd, int32(id), modifiers, uint32(shortcut.Key)) {
		return nil, newError("RegisterHotKey failed")
	}

	hotkey := &GlobalHotKey{
		id:       id,
		shortcut: shortcut,
		window:   window,
		handler:  handler,
	}
	hotkeys.Store(id, hotkey)

	return hotkey, nil
}

// Unregister unregisters the global hotkey
func (h *GlobalHotKey) Unregister() error {
	if !win.UnregisterHotKey(h.window.hWnd, int32(h.id)) {
		return newError("UnregisterHotKey failed")
	}
	hotkeys.Delete(h.id)
	return nil
}
