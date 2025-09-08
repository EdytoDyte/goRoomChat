package client

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

type UI struct {
	g         *gocui.Gui
	client    *Client
	serverKey bool
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
	if v, err := g.SetView("messages", 0, 0, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Messages"
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

	if !ui.serverKey {
		if v, err := g.SetView("popup", maxX/2-15, maxY/2-1, maxX/2+15, maxY/2+1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			fmt.Fprintln(v, "Waiting for server key...")
		}
	} else {
		g.DeleteView("popup")
	}

	return nil
}

func (ui *UI) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
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
	ui.g.Update(func(g *gocui.Gui) error {
		return ui.layout(g)
	})
}

func (ui *UI) ShowRoomPrompt() {
	// This will be implemented in a later step
}

func (ui *UI) ShowUsernamePrompt() {
	// This will be implemented in a later step
}
