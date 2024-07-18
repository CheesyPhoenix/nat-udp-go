package tui

import (
	"fmt"
	"net"

	"github.com/cheesyphoenix/nat-udp-go/src"
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
	logView.SetChangedFunc(func() { app.Draw() })
	logView.Write([]byte{})

	actionPages := tview.NewPages()

	list := tview.NewList()
	actionPages.AddAndSwitchToPage("start-list", list, true)
	list.
		AddItem("Start Client", "", 'c', func() {
			addressStr := ""

			form := tview.NewForm()
			form.
				AddInputField("Server address", addressStr, 22, nil, func(text string) { addressStr = text }).
				AddButton("Save", func() {
					addr, err := net.ResolveUDPAddr("udp4", addressStr)
					if err != nil {
						app.SetFocus(form.GetFormItemByLabel("Server address"))
						logView.Write([]byte(err.Error() + "\n"))
						return
					}

					clientPage := tview.NewFlex()
					clientPage.SetBorder(true).SetTitle("Client running")

					clientInfo := tview.NewTextView()
					clientInfo.Write([]byte(fmt.Sprintf("Server IP: %v\n", addr.IP)))
					clientInfo.Write([]byte(fmt.Sprintf("Server port: %v\n", addr.Port)))

					clientActions := tview.NewList()
					clientActions.AddItem("Quit", "Press to exit", 'q', func() {
						app.Stop()
					})

					clientPage.AddItem(clientInfo, 0, 1, false)
					clientPage.AddItem(clientActions, 0, 1, true)

					actionPages.AddAndSwitchToPage("client-page", clientPage, true)

					go func() {
						src.Client(*addr, func(format string, a ...any) {
							logView.Write([]byte(fmt.Sprintf(format, a...) + "\n"))
						})
						app.Stop()
					}()
				}).
				AddButton("Cancel", func() {
					actionPages.SwitchToPage("start-list")
				})
			form.SetBorder(true).SetTitle("Start Client").SetTitleAlign(tview.AlignLeft)
			actionPages.AddAndSwitchToPage("start-client-form", form, true)
		}).
		AddItem("Start Server", "", 's', func() {
			serverPage := tview.NewFlex()
			serverPage.SetBorder(true).SetTitle("Server running")

			serverInfo := tview.NewTextView()

			serverActions := tview.NewList()

			serverPage.AddItem(serverInfo, 0, 1, false)
			serverPage.AddItem(serverActions, 0, 1, true)

			actionPages.AddAndSwitchToPage("server-page", serverPage, true)

			go func() {
				src.Server(func(format string, a ...any) {
					logView.Write([]byte(fmt.Sprintf(format, a...) + "\n"))
				})
				app.Stop()
			}()
		}).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		}).SetBorder(true)

	mainFlex.AddItem(actionPages, 0, 2, true).AddItem(logView, 0, 1, false)

	if err := app.SetRoot(mainFlex, true).Run(); err != nil {
		panic(err)
	}
}
