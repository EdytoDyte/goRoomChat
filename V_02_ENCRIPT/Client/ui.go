package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/jroimartin/gocui"
)

// Variable on UI.GO

var connec net.Conn
var GoCui *gocui.Gui

func IniGu(conecct net.Conn) (*gocui.Gui, error) {
	connec = conecct
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		fmt.Print(err)
	}
	GoCui = g
	defer g.Close()
	g.SetManagerFunc(layout)
	g.Mouse = true
	msg := keys{
		Protocol: []byte("Ok"),
	}
	public, _ := json.Marshal(msg)
	conecct.Write(public)
	conecct.Write([]byte("\n"))
	g.Mouse = true

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		fmt.Print(err)
	}
	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, sendMessage); err != nil {
		fmt.Print(err)
	}
	if err := g.SetKeybinding("button", gocui.MouseLeft, gocui.ModNone, close); err != nil {
		fmt.Print(err)
	}
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, close); err != nil {
		fmt.Print(err)
	}
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		fmt.Print(err)

	}
	return g, nil
}
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if input, err := g.SetView("input", 0, maxY-5, maxX-20, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		input.Title = "input"
		input.Wrap = true
		input.Editable = true
		input.Autoscroll = true
		if _, err = setCurrentViewOnTop(g, "input"); err != nil {
			return err
		}

	}
	if v, err := g.SetView("messages", 0, 0, maxX-20, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "::: Room: " + NameRoom + " --- Messages :::"
		v.Wrap = true
		v.Autoscroll = true

	}
	if v, err := g.SetView("button", maxX-24, 0, maxX-20, +2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.SelBgColor = gocui.ColorWhite
		v.FgColor = gocui.ColorBlack
		v.BgColor = gocui.ColorWhite
		v.Wrap = true
		fmt.Fprintln(v, "<--")

	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}
func sendMessage(g *gocui.Gui, v *gocui.View) error {

	inp, _ := g.View("input")
	if len(inp.Buffer()) == 0 {
		inp.SetCursor(0, 0)
		inp.Clear()
		return nil
	}
	msg := inp.Buffer()
	msg = strings.TrimRight(msg, "\n")
	msg = strings.TrimSpace(msg)
	msgEn, _ := encriptar([]byte(msg), publicKey)
	message := msges{
		Mensaje: msgEn,
	}
	msgJson, _ := json.Marshal(message)
	connec.Write(msgJson)
	connec.Write([]byte("\n"))
	inp.SetCursor(0, 0)
	inp.Clear()

	return nil
}
func updateView(g *gocui.Gui, mesDesen []byte) {

	g.Update(func(g *gocui.Gui) error {
		v, err := GoCui.View("messages")
		if err != nil {
			fmt.Print(err)
			return err
		}
		fmt.Fprint(v, string(mesDesen))

		return layout(g)
	})
}
func close(g *gocui.Gui, v *gocui.View) error {
	connec.Close()
	return gocui.ErrQuit
}
