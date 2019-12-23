package main

import (
	"errors"
	"fmt"
	"github.com/awesome-gocui/gocui"
	"log"
	"strings"
)

const (
	// Windows
	chatLogWindow      = "clw" // Where (received and sent) messages are logged
	chatInputWindow    = "iw"  // Where messages are written
	rawInputWindow     = "rw"  // Where raw stanzas are written
	contactsListWindow = "cl"  // Where the contacts list is shown, and contacts are selectable
	menuWindow         = "mw"  // Where the menu is shown
	disconnectMsg      = "msg"

	// Menu options
	disconnect         = "Disconnect"
	askServerForRoster = "Ask server for roster"
	rawMode            = "Switch to Send Raw Mode"
	messageMode        = "Switch to Send Message Mode"
	contactList        = "Contacts list"
	backFromContacts   = "<- Go back"
)

// To store names of views on top
type viewsState struct {
	input          string   // Which input view is on top
	side           string   // Which side view is on top
	contacts       []string // Contacts list
	currentContact string   // Contact we are currently messaging
}

var (
	// Which window is on top currently on top of the other.
	// This is the init setup
	viewState = viewsState{
		input: chatInputWindow,
		side:  menuWindow,
	}
	menuOptions = []string{contactList, rawMode, askServerForRoster, disconnect}
	// Errors
	servConnFail = errors.New("failed to connect to server. Check your configuration ? Exiting")
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

	if v, err := g.SetView(contactsListWindow, 0, 0, maxX/5-1, 5*maxY/6-2, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Title = "Contacts"
		v.Wrap = true
		v.Autoscroll = true
	}

	if v, err := g.SetView(menuWindow, 0, 0, maxX/5-1, 5*maxY/6-2, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Title = "Menu"
		v.Wrap = true
		v.Autoscroll = true
		fmt.Fprint(v, strings.Join(menuOptions, "\n"))
		if _, err = setCurrentViewOnTop(g, menuWindow); err != nil {
			return err
		}
	}

	if v, err := g.SetView(rawInputWindow, 0, 5*maxY/6-1, maxX/1-1, maxY-1, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Title = "Write or paste a raw stanza. Press \"Ctrl+E\" to send :"
		v.Editable = true
		v.Wrap = true
	}

	if v, err := g.SetView(chatInputWindow, 0, 5*maxY/6-1, maxX/1-1, maxY-1, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		v.Title = "Write a message :"
		v.Editable = true
		v.Wrap = true

		if _, err = setCurrentViewOnTop(g, chatInputWindow); err != nil {
			return err
		}
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// Sends an input text from the user to the backend while also printing it in the chatlog window.
// KeyEnter is viewed as "\n" by gocui, so messages should only be one line, whereas raw sending has a different key
// binding and therefor should work with this too (for multiple lines stanzas)
func writeInput(g *gocui.Gui, v *gocui.View) error {
	chatLogWindow, _ := g.View(chatLogWindow)

	input := strings.Join(v.ViewBufferLines(), "\n")

	fmt.Fprintln(chatLogWindow, "Me : ", input)
	textChan <- input

	v.Clear()
	v.EditDeleteToStartOfLine()
	return nil
}

func setKeyBindings(g *gocui.Gui) {
	// ==========================
	// All views
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	// ==========================
	// Chat input
	if err := g.SetKeybinding(chatInputWindow, gocui.KeyEnter, gocui.ModNone, writeInput); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(chatInputWindow, gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}

	// ==========================
	// Raw input
	if err := g.SetKeybinding(rawInputWindow, gocui.KeyCtrlE, gocui.ModNone, writeInput); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(rawInputWindow, gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}

	// ==========================
	// Menu
	if err := g.SetKeybinding(menuWindow, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(menuWindow, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(menuWindow, gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(menuWindow, gocui.KeyEnter, gocui.ModNone, getLine); err != nil {
		log.Panicln(err)
	}

	// ==========================
	// Contacts list
	if err := g.SetKeybinding(contactsListWindow, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(contactsListWindow, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(contactsListWindow, gocui.KeyEnter, gocui.ModNone, getLine); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(contactsListWindow, gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}

	// ==========================
	// Disconnect message
	if err := g.SetKeybinding(disconnectMsg, gocui.KeyEnter, gocui.ModNone, delMsg); err != nil {
		log.Panicln(err)
	}

}

// General
// Used to handle menu selections and navigations
func getLine(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	if viewState.side == menuWindow {
		if l == contactList {
			cv, _ := g.View(contactsListWindow)
			viewState.side = contactsListWindow
			g.SetViewOnTop(contactsListWindow)
			g.SetCurrentView(contactsListWindow)
			if len(cv.ViewBufferLines()) == 0 {
				printContactsToWindow(g, viewState.contacts)
			}
		} else if l == disconnect {
			maxX, maxY := g.Size()
			msg := "You disconnected from the server. Press enter to quit."
			if v, err := g.SetView(disconnectMsg, maxX/2-30, maxY/2, maxX/2-29+len(msg), maxY/2+2, 0); err != nil {
				if !gocui.IsUnknownView(err) {
					return err
				}
				fmt.Fprintln(v, msg)
				if _, err := g.SetCurrentView(disconnectMsg); err != nil {
					return err
				}
			}
			killChan <- disconnectErr
		} else if l == askServerForRoster {
			chlw, _ := g.View(chatLogWindow)
			fmt.Fprintln(chlw, infoFormat+" Not yet implemented !")
		} else if l == rawMode {
			mw, _ := g.View(menuWindow)
			viewState.input = rawInputWindow
			g.SetViewOnTop(rawInputWindow)
			g.SetCurrentView(rawInputWindow)
			menuOptions[1] = messageMode
			v.Clear()
			v.EditDeleteToStartOfLine()
			fmt.Fprintln(mw, strings.Join(menuOptions, "\n"))
			message := "Now sending in raw stanza mode"
			clv, _ := g.View(chatLogWindow)
			fmt.Fprintln(clv, infoFormat+message)
		} else if l == messageMode {
			mw, _ := g.View(menuWindow)
			viewState.input = chatInputWindow
			g.SetViewOnTop(chatInputWindow)
			g.SetCurrentView(chatInputWindow)
			menuOptions[1] = rawMode
			v.Clear()
			v.EditDeleteToStartOfLine()
			fmt.Fprintln(mw, strings.Join(menuOptions, "\n"))
			message := "Now sending in messages mode"
			clv, _ := g.View(chatLogWindow)
			fmt.Fprintln(clv, infoFormat+message)
		}
	} else if viewState.side == contactsListWindow {
		if l == backFromContacts {
			viewState.side = menuWindow
			g.SetViewOnTop(menuWindow)
			g.SetCurrentView(menuWindow)
		} else if l == "" {
			return nil
		} else {
			// Updating the current correspondent, back-end side.
			CorrespChan <- l
			viewState.currentContact = l
			// Showing the selected contact in contacts list
			cl, _ := g.View(contactsListWindow)
			cts := cl.ViewBufferLines()
			cl.Clear()
			printContactsToWindow(g, cts)
			// Showing a message to the user, and switching back to input after the new contact is selected.
			message := "Now sending messages to : " + l + " in a private conversation"
			clv, _ := g.View(chatLogWindow)
			fmt.Fprintln(clv, infoFormat+message)
			g.SetCurrentView(chatInputWindow)
		}
	}

	return nil
}

func printContactsToWindow(g *gocui.Gui, contactsList []string) {
	cl, _ := g.View(contactsListWindow)
	for _, c := range contactsList {
		c = strings.ReplaceAll(c, " *", "")
		if c == viewState.currentContact {
			fmt.Fprintf(cl, c+" *\n")
		} else {
			fmt.Fprintf(cl, c+"\n")
		}
	}
}

// Changing view between input and "menu/contacts" when pressing the specific key.
func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == chatInputWindow || v.Name() == rawInputWindow {
		_, err := g.SetCurrentView(viewState.side)
		return err
	} else if v.Name() == menuWindow || v.Name() == contactsListWindow {
		_, err := g.SetCurrentView(viewState.input)
		return err
	}

	// Should not be reached right now
	_, err := g.SetCurrentView(chatInputWindow)
	return err
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		// Avoid going below the list of contacts. Although lines are stored in the view as a slice
		// in the used lib. Therefor, if the number of lines is too big, the cursor will go past the last line since
		// increasing slice capacity is done by doubling it. Last lines will be "nil" and reachable by the cursor
		// in a dynamic context (such as contacts list)
		cv := g.CurrentView()
		h := cv.LinesHeight()
		if cy+1 >= h {
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

func delMsg(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(disconnectMsg); err != nil {
		return err
	}
	errChan <- gocui.ErrQuit // Quit the program
	return nil
}
