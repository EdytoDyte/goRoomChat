package client

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

type UI struct {
	g             *gocui.Gui
	client        *Client
	serverKey     bool
	roomName      string
	username      string
	waitingForKey bool
}

func NewUI(client *Client) *UI {
	return &UI{
		client: client,
	}
}

func (ui *UI) Run() error {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	ui.g = g
	defer ui.g.Close()

	ui.g.SetManagerFunc(ui.layout)

	if err := ui.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, ui.quit); err != nil {
		return err
	}
	if err := ui.g.SetKeybinding("room", gocui.KeyEnter, gocui.ModNone, ui.setRoom); err != nil {
		return err
	}
	if err := ui.g.SetKeybinding("username", gocui.KeyEnter, gocui.ModNone, ui.setUsername); err != nil {
		return err
	}
	if err := ui.g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, ui.sendMessage); err != nil {
		return err
	}

	if err := ui.g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (ui *UI) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if ui.roomName == "" {
		if v, err := g.SetView("room", maxX/2-15, maxY/2-1, maxX/2+15, maxY/2+1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Enter Room Name"
			v.Editable = true
			v.Wrap = true
			if _, err := g.SetCurrentView("room"); err != nil {
				return err
			}
		}
	} else if ui.username == "" {
		if v, err := g.SetView("username", maxX/2-15, maxY/2-1, maxX/2+15, maxY/2+1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Enter Username"
			v.Editable = true
			v.Wrap = true
			if _, err := g.SetCurrentView("username"); err != nil {
				return err
			}
		}
	} else {
		// "room" view is deleted in setRoom, "username" view is deleted in setUsername

		// Create chat views
		if v, err := g.SetView("messages", 0, 0, maxX-1, maxY-5); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = fmt.Sprintf("Room: %s | User: %s", ui.roomName, ui.username)
			v.Wrap = true
			v.Autoscroll = true
		}

		if v, err := g.SetView("input", 0, maxY-5, maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Input"
			v.Editable = true
			v.Wrap = true
			if _, err := g.SetCurrentView("input"); err != nil {
				return err
			}
		}
	}

	if ui.waitingForKey && !ui.serverKey {
		if v, err := g.SetView("popup", maxX/2-15, maxY/2-1, maxX/2+15, maxY/2+1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			fmt.Fprintln(v, "Waiting for server key...")
		}
	}

	return nil
}

func (ui *UI) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (ui *UI) setRoom(g *gocui.Gui, v *gocui.View) error {
	roomName := strings.TrimSpace(v.Buffer())
	if roomName != "" {
		ui.roomName = roomName
		ui.client.JoinRoom(roomName)
		ui.waitingForKey = true
		g.DeleteView("room")
		// Don't set current view yet, wait for key
	}
	return nil
}

func (ui *UI) setUsername(g *gocui.Gui, v *gocui.View) error {
	username := strings.TrimSpace(v.Buffer())
	if username != "" {
		ui.username = username
		ui.client.SendUsername(username)
		g.DeleteView("username")
	}
	return nil
}

func (ui *UI) sendMessage(g *gocui.Gui, v *gocui.View) error {
	message := strings.TrimSpace(v.Buffer())
	if message != "" {
		ui.client.SendMessage(message)
		v.Clear()
		v.SetCursor(0, 0)
	}
	return nil
}

func (ui *UI) UpdateMessages(message string) {
	ui.g.Update(func(g *gocui.Gui) error {
		v, err := g.View("messages")
		if err != nil {
			return err
		}
		fmt.Fprintln(v, message)
		return nil
	})
}

func (ui *UI) SetServerKey(status bool) {
	ui.serverKey = status
	ui.waitingForKey = false
	ui.g.Update(func(g *gocui.Gui) error {
		g.DeleteView("popup")
		if _, err := g.SetCurrentView("username"); err != nil {
			return err
		}
		return ui.layout(g)
	})
}
