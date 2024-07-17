package tui

import (
	"fmt"
	"net"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func PromptForAddress() (*net.UDPAddr, error) {
	fmt.Print("Enter remote address: ")

	var addrStr string
	_, err := fmt.Scanln(&addrStr)
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("got error trying to read line: %v", err.Error())
	}

	addr, err := net.ResolveUDPAddr("udp4", addrStr)
	if err != nil {
		fmt.Println("Address is not valid. Please try again. Error:", err)
		return PromptForAddress()
	}
	return addr, nil
}

func StartTUI() {
	app := tview.NewApplication().EnableMouse(true).EnablePaste(true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			app.Stop()
		case tcell.KeyCtrlC:
			return nil
		}
		return event
	})

	mainFlex := tview.NewFlex()

	logView := tview.NewTextView()
	logView.SetBorder(true).SetTitle("Log")
	logView.Write([]byte("Hello Log\nNew Line"))

	actionPages := tview.NewPages()

	list := tview.NewList()
	actionPages.AddAndSwitchToPage("start-list", list, true)
	list.
		AddItem("Start Client", "", 'c', func() {
			form := tview.NewForm().
				AddInputField("Server address", "", 22, nil, nil).
				AddButton("Save", nil).
				AddButton("Cancel", func() {
					actionPages.SwitchToPage("start-list")
				})
			form.SetBorder(true).SetTitle("Start Client").SetTitleAlign(tview.AlignLeft)
			actionPages.AddAndSwitchToPage("start-client-form", form, true)
		}).
		AddItem("Start Server", "", 's', nil).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		}).SetBorder(true)

	mainFlex.AddItem(actionPages, 0, 2, true).AddItem(logView, 0, 1, false)

	if err := app.SetRoot(mainFlex, true).Run(); err != nil {
		panic(err)
	}
}
