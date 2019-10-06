package main

import (
	"github.com/spf13/viper"
	"github.com/therecipe/qt/widgets"
	"log"
)

func buildSettingsWindow() *widgets.QMainWindow {
	// create a window
	// with a minimum size of 250*200
	// and sets the title to "Hello Widgets Example"
	window := widgets.NewQMainWindow(nil, 0)
	window.SetMinimumSize2(250, 250)
	window.SetWindowTitle("Credentials Inout")

	// create a regular widget
	// give it a QVBoxLayout
	// and make it the central widget of the window
	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())
	window.SetCentralWidget(widget)

	// create a line edit
	// with a custom placeholder text
	// and add it to the central widgets layout
	loginInput := widgets.NewQLineEdit(nil)
	loginInput.SetPlaceholderText("Type your login ...")
	widget.Layout().AddWidget(loginInput)

	passwordInput := widgets.NewQLineEdit(nil)
	passwordInput.SetPlaceholderText("Type your password ...")
	passwordInput.SetEchoMode(widgets.QLineEdit__Password)
	widget.Layout().AddWidget(passwordInput)

	targetHostInput := widgets.NewQLineEdit(nil)
	targetHostInput.SetPlaceholderText("Type hostname where http queries will be forwarded to ...")
	widget.Layout().AddWidget(targetHostInput)

	targetPortInput := widgets.NewQLineEdit(nil)
	targetPortInput.SetPlaceholderText("Type hostname port where http queries will be forwarded to ...")
	widget.Layout().AddWidget(targetPortInput)

	listeningInterface := widgets.NewQLineEdit(nil)
	listeningInterface.SetPlaceholderText("Type interface address on which the proxy will be listening ...")
	widget.Layout().AddWidget(listeningInterface)

	listeningPortInput := widgets.NewQLineEdit(nil)
	listeningPortInput.SetPlaceholderText("Type port on which the proxy will be listening ...")
	widget.Layout().AddWidget(listeningPortInput)

	logIsVerbose := widgets.NewQCheckBox2("Verbose Logging", nil)
	logIsVerbose.SetTristate(false)
	widget.Layout().AddWidget(logIsVerbose)

	// create a button
	// connect the clicked signal
	// and add it to the central widgets layout
	button := widgets.NewQPushButton2("Save", nil)
	button.ConnectClicked(func(bool) {
		viper.Set("Verbose", logIsVerbose.IsChecked())
		viper.Set(ListenAddr, listeningInterface.Text())
		viper.Set(ListenPort, listeningPortInput.Text())
		viper.Set(ProxyHost, targetHostInput.Text())
		viper.Set(ProxyPort, targetPortInput.Text())
		viper.Set(LOGIN, loginInput.Text())
		viper.Set(PASSWORD, passwordInput.Text())
		viper.Set(VerboseLog, false)

		err := viper.WriteConfig()
		if err != nil {
			log.Printf("Error: %s", err)
		}
		// widgets.QMessageBox_Information(nil, "OK", passwordInput.Text(), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	})
	widget.Layout().AddWidget(button)
	return window
}
