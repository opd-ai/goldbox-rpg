//go:build js && wasm

package wasmui

import (
	"fmt"
	"syscall/js"
)

// Soft keyboard state — only accessed from the main goroutine (WASM is single-threaded).
var (
	skInput       js.Value
	skInitialized bool
	skVisible     bool
	skEnterFlag   bool
	skKeydownCb   js.Func
	skTouchCache  *bool
)

// initSoftKeyboard creates a transparent HTML input element for mobile soft keyboard support.
// The input is positioned over the name text box on the canvas so that tapping
// the area focuses it and triggers the mobile on-screen keyboard.
func initSoftKeyboard() {
	doc := js.Global().Get("document")

	input := doc.Call("createElement", "input")
	input.Set("type", "text")
	input.Set("id", "goldbox-soft-keyboard")
	input.Set("autocomplete", "off")
	input.Call("setAttribute", "autocorrect", "off")
	input.Call("setAttribute", "autocapitalize", "off")
	input.Call("setAttribute", "spellcheck", "false")
	input.Set("maxLength", 30)
	input.Call("setAttribute", "enterkeyhint", "done")

	style := input.Get("style")
	style.Set("position", "fixed")
	style.Set("zIndex", "10000")
	style.Set("opacity", "0.01")
	style.Set("fontSize", "16px") // prevents iOS auto-zoom on focus
	style.Set("color", "transparent")
	style.Set("background", "transparent")
	style.Set("border", "none")
	style.Set("outline", "none")
	style.Set("caretColor", "transparent")
	style.Set("display", "none")
	style.Set("padding", "0")
	style.Set("margin", "0")

	skKeydownCb = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if args[0].Get("key").String() == "Enter" {
			skEnterFlag = true
			args[0].Call("preventDefault")
		}
		return nil
	})
	input.Call("addEventListener", "keydown", skKeydownCb)

	doc.Get("body").Call("appendChild", input)
	skInput = input
	skInitialized = true
}

// positionSoftKeyboard positions the transparent input over the name text box on the canvas.
func positionSoftKeyboard() {
	canvas := js.Global().Get("document").Call("querySelector", "canvas")
	if !canvas.Truthy() {
		return
	}

	rect := canvas.Call("getBoundingClientRect")
	cLeft := rect.Get("left").Float()
	cTop := rect.Get("top").Float()
	cWidth := rect.Get("width").Float()
	cHeight := rect.Get("height").Float()

	scaleX := cWidth / float64(ScreenWidth)
	scaleY := cHeight / float64(ScreenHeight)

	// Name box coordinates match drawCharCreationName: (250, 150) size (300, 30)
	x := cLeft + 250.0*scaleX
	y := cTop + 150.0*scaleY
	w := 300.0 * scaleX
	h := 30.0 * scaleY

	style := skInput.Get("style")
	style.Set("left", fmt.Sprintf("%.0fpx", x))
	style.Set("top", fmt.Sprintf("%.0fpx", y))
	style.Set("width", fmt.Sprintf("%.0fpx", w))
	style.Set("height", fmt.Sprintf("%.0fpx", h))
}

// showSoftKeyboard makes the soft keyboard input overlay visible and positions it.
// The initial value is set only the first time it becomes visible, so that
// subsequent frames do not overwrite characters the user is actively typing.
func showSoftKeyboard(initialValue string) {
	if !skInitialized {
		initSoftKeyboard()
	}

	if !skVisible {
		skInput.Set("value", initialValue)
		positionSoftKeyboard()
		skInput.Get("style").Set("display", "block")
		skVisible = true
	}
}

// hideSoftKeyboard hides the soft keyboard input overlay and dismisses the keyboard.
func hideSoftKeyboard() {
	if skInitialized && skInput.Truthy() {
		skInput.Call("blur")
		skInput.Get("style").Set("display", "none")
	}
	skVisible = false
	skEnterFlag = false
}

// softKeyboardValue returns the current text in the soft keyboard input.
func softKeyboardValue() string {
	if !skInitialized || !skInput.Truthy() {
		return ""
	}
	return skInput.Get("value").String()
}

// setSoftKeyboardValue updates the soft keyboard input text, keeping it
// in sync when the desktop physical keyboard modifies the name.
func setSoftKeyboardValue(v string) {
	if skInitialized && skInput.Truthy() {
		skInput.Set("value", v)
	}
}

// softKeyboardEnterPressed returns true if Enter was pressed on the soft
// keyboard since the last call. The flag is reset after reading.
func softKeyboardEnterPressed() bool {
	v := skEnterFlag
	skEnterFlag = false
	return v
}

// isSoftKeyboardFocused reports whether the soft keyboard input currently has focus.
func isSoftKeyboardFocused() bool {
	if !skInitialized || !skInput.Truthy() {
		return false
	}
	active := js.Global().Get("document").Get("activeElement")
	return active.Truthy() && active.Equal(skInput)
}

// hasTouchSupport reports whether the browser supports touch input.
// The result is cached after the first call.
func hasTouchSupport() bool {
	if skTouchCache != nil {
		return *skTouchCache
	}
	nav := js.Global().Get("navigator")
	result := false
	if nav.Truthy() {
		mtp := nav.Get("maxTouchPoints")
		if mtp.Truthy() {
			result = mtp.Int() > 0
		}
	}
	skTouchCache = &result
	return result
}
