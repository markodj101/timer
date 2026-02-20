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
		Title: "Timer Control Panel",
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

		displayWindow.OnWindowEvent(events.Common.WindowRuntimeReady, func(e *application.WindowEvent) {
        log.Printf("WindowIsReady – sada postavljam fullscreen na odabrani monitor.")

        
        if !targetScreen.IsPrimary {
            log.Printf("Postavljam prozor na ekran: %s (primary: %t)", targetScreen.Name, targetScreen.IsPrimary)
            log.Printf("Pozicija: (%d, %d)", targetScreen.Bounds.X, targetScreen.Bounds.Y)
            displayWindow.SetPosition(targetScreen.Bounds.X, targetScreen.Bounds.Y)
        } else {
            displayWindow.SetPosition(0, 0)
        }

      
        time.Sleep(50 * time.Millisecond)

       
        displayWindow.Fullscreen()
        displayWindow.DisableSizeConstraints()
        
    })

    
    displayWindow.Show()

	
		
	})

	

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	
	

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
