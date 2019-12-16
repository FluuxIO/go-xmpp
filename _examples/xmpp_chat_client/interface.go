package main

import (
	"fmt"
	"github.com/awesome-gocui/gocui"
	"log"
)

const (
	chatLogWindow = "clw"
	inputWindow   = "iw"
	menuWindow    = "menw"
)

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView(chatLogWindow, maxX/5, 0, maxX-1, 5*maxY/6-1, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Title = "Chat log"
		v.Wrap = true
		v.Autoscroll = true
	}

	if v, err := g.SetView(menuWindow, 0, 0, maxX/5-1, 5*maxY/6-1, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Title = "Contacts"
		v.Wrap = true
		v.Autoscroll = true
	}

	if v, err := g.SetView(inputWindow, 0, 5*maxY/6-1, maxX/1-1, maxY-1, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Title = "Write a message :"
		v.Editable = true
		v.Wrap = true

		if _, err = setCurrentViewOnTop(g, inputWindow); err != nil {
			return err
		}
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// Sends an input line from the user to the backend while also printing it in the chatlog window.
func writeInput(g *gocui.Gui, v *gocui.View) error {
	log, _ := g.View(chatLogWindow)
	for _, line := range v.ViewBufferLines() {
		textChan <- line
		fmt.Fprintln(log, "Me : ", line)
	}
	v.Clear()
	v.EditDeleteToStartOfLine()
	return nil
}

func setKeyBindings(g *gocui.Gui) {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(inputWindow, gocui.KeyEnter, gocui.ModNone, writeInput); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(inputWindow, gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)

	}
	if err := g.SetKeybinding(menuWindow, gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(menuWindow, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)

	}
	if err := g.SetKeybinding(menuWindow, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)

	}
	if err := g.SetKeybinding(menuWindow, gocui.KeyEnter, gocui.ModNone, getLine); err != nil {
		log.Panicln(err)
	}

}

// When we select a new correspondent, we change it in the client, and we display a message window confirming the change.
func getLine(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	// Updating the current correspondent, back-end side.
	CorrespChan <- l

	// Showing a message to the user, and switching back to input after the new contact is selected.
	message := "Now sending messages to : " + l + " in a private conversation"
	clv, _ := g.View(chatLogWindow)
	fmt.Fprintln(clv, infoFormat+message)
	g.SetCurrentView(inputWindow)
	return nil
}

// Changing view between input and "menu" (= basically contacts only right now) when pressing the specific key.
func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == inputWindow {
		_, err := g.SetCurrentView(menuWindow)
		return err
	}
	_, err := g.SetCurrentView(inputWindow)
	return err
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		// Avoid going below the list of contacts
		cv := g.CurrentView()
		h := cv.LinesHeight()
		if cy+1 >= h-1 {
			return nil
		}
		// Lower cursor
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}
