package tui

import (
	"fmt"
	"net"
	"os"

	"github.com/cheesyphoenix/nat-udp-go/src"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// func PromptForAddress() (*net.UDPAddr, error) {
// 	fmt.Print("Enter remote address: ")

// 	var addrStr string
// 	_, err := fmt.Scanln(&addrStr)
// 	if err != nil {
// 		fmt.Println()
// 		return nil, fmt.Errorf("got error trying to read line: %v", err.Error())
// 	}

// 	addr, err := net.ResolveUDPAddr("udp4", addrStr)
// 	if err != nil {
// 		fmt.Println("Address is not valid. Please try again. Error:", err)
// 		return PromptForAddress()
// 	}
// 	return addr, nil
// }

func StartTUI() {
	log := ""
	logFile, err := os.Create("log.txt")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	app := tview.NewApplication().EnableMouse(false).EnablePaste(true)

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
							str := fmt.Sprintf(format, a...) + "\n"
							logView.Write([]byte(str))
							log += str
							logFile.Write([]byte(str))
						})
						log += "Client stopped"
						logFile.Write([]byte("Client stopped"))
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
			conn, err := net.ListenUDP("udp4", &net.UDPAddr{
				IP:   net.IPv4(0, 0, 0, 0),
				Port: src.ServerUDPPort,
				Zone: "",
			})
			if err != nil {
				logView.Write([]byte(err.Error() + "\n"))
				return
			}

			serverPage := tview.NewFlex()
			serverPage.SetBorder(true).SetTitle("Server running")

			serverInfo := tview.NewTextView()
			serverInfo.Write([]byte("Clients:\n"))

			serverActions := tview.NewList()
			serverActions.AddItem("New Client", "Press to add a new client", 'n', func() {
				clientAddress := ""

				form := tview.NewForm()
				form.AddInputField("Client address", clientAddress, 22, nil, func(text string) {
					clientAddress = text
				})
				form.AddButton("Save", func() {
					addr, err := net.ResolveUDPAddr("udp4", clientAddress)
					if err != nil {
						app.SetFocus(form.GetFormItemByLabel("Server address"))
						logView.Write([]byte(err.Error() + "\n"))
						return
					}

					src.StartHolePunch(*conn, *addr, func(format string, a ...any) {
						str := fmt.Sprintf(format, a...) + "\n"
						logView.Write([]byte(str))
						log += str
						logFile.Write([]byte(str))
					})

					serverInfo.Write([]byte(fmt.Sprintf("- %v:%v", addr.IP, addr.Port) + "\n"))

					actionPages.SwitchToPage("server-page")
				})
				form.AddButton("Cancel", func() {
					actionPages.SwitchToPage("server-page")
				})

				actionPages.AddAndSwitchToPage("new-client-form", form, true)
			})
			serverActions.AddItem("Quit", "Press to exit", 'q', func() {
				app.Stop()
			})

			serverPage.AddItem(serverInfo, 0, 1, false)
			serverPage.AddItem(serverActions, 0, 1, true)

			actionPages.AddAndSwitchToPage("server-page", serverPage, true)

			go func() {
				src.Server(*conn, func(format string, a ...any) {
					str := fmt.Sprintf(format, a...) + "\n"
					logView.Write([]byte(str))
					log += str
					logFile.Write([]byte(str))
				})
				conn.Close()
				app.Stop()
			}()
		}).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		}).SetBorder(true)

	stunAndLogFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	stunView := tview.NewTextView()
	stunView.SetRegions(true).SetBorder(true).SetTitle("Your address")

	addr, err := src.GetIPAndPort(&net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: src.ClientUDPPort,
	})
	if err != nil {
		panic(err)
	}

	stunView.Write([]byte(fmt.Sprintf("[\"addr\"]%v:%v[\"\"]", addr.IP, addr.Port)))

	stunAndLogFlex.AddItem(stunView, 3, 0, true)
	stunAndLogFlex.AddItem(logView, 0, 1, false)

	mainFlex.AddItem(actionPages, 0, 2, true).AddItem(stunAndLogFlex, 0, 1, false)

	if err := app.SetRoot(mainFlex, true).Run(); err != nil {
		app.Stop()
		fmt.Println(log)
		panic(err)
	}
	fmt.Println(log)
}
