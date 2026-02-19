package main

import (
	"embed"
	_ "embed"
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)


var assets embed.FS

func init() {
	
	application.RegisterEvent[string]("time")
}


func main() {

	
	app := application.New(application.Options{
		Name:        "timer-main",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&GreetService{}),
			application.NewService(&FileReading{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		
	})

	
	mainWin := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Window 1",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
		MinWidth: 1024,
		MinHeight: 800,
		Width:  1024,          // Početna širina
        Height: 800,
	})



	displayWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Timer Display",
		AlwaysOnTop:  true,
		Frameless: true,
		URL: "/display",
		Hidden: true,
		
	})


	mainWin.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		log.Printf("Glavni prozor se zatvara, izlazim.")
		if displayWindow != nil {
			displayWindow.Close()
		}
		
		app.Quit()
	})


	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(event *application.ApplicationEvent){
		screens := app.Screen.GetAll();
		log.Printf("Broj ekrana: %d", len(screens))

		if len(screens) == 0 {
			log.Printf("Nema dostupnih ekrana, izlazim.")
			displayWindow.SetPosition(0,0)
			displayWindow.Fullscreen()
			displayWindow.Show()
			displayWindow.SetAlwaysOnTop(false)
			return
		}

		var primary *application.Screen
		var secondary *application.Screen

		for i := range screens {
			screen := screens[i]
			if screen.IsPrimary {
				primary = screen
			} else {
				if secondary == nil {
					secondary = screen
				}
			}
		}

		targetScreen := secondary
		if targetScreen == nil {
			log.Printf("Nema sekundarnog ekrana, koristim primarni.")
			targetScreen = primary
		}

		if !targetScreen.IsPrimary {
			log.Printf("Postavljam prozor na ekran: %s and is Primary: %t", targetScreen.Name,targetScreen.IsPrimary)
			log.Printf("Ekran dimenzije: %dx%d", targetScreen.Bounds.Width, targetScreen.Bounds.Height)
			log.Printf("Ekran pozicija: (%d, %d)", targetScreen.Bounds.X, targetScreen.Bounds.Y)
			displayWindow.SetPosition(targetScreen.Bounds.X, targetScreen.Bounds.Y)
			
			
		}else{
			displayWindow.SetPosition(0,0)
		}
		time.Sleep(500 * time.Millisecond) // Kratko čekanje da se prozor postavi prije fullscreen
		displayWindow.Fullscreen()
		displayWindow.DisableSizeConstraints()
		displayWindow.Hide()
	//	displayWindow.Show()
	//	displayWindow.Focus()
		
	})

	

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	
	

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
