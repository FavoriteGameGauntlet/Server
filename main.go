package main

import (
	genauth "FGG-Service/api/generated/auth"
	geneffects "FGG-Service/api/generated/effects"
	genexchanges "FGG-Service/api/generated/exchanges"
	gengames "FGG-Service/api/generated/games"
	gengrants "FGG-Service/api/generated/grants"
	genitems "FGG-Service/api/generated/items"
	genparties "FGG-Service/api/generated/parties"
	genperks "FGG-Service/api/generated/perks"
	genpoints "FGG-Service/api/generated/points"
	gensysparams "FGG-Service/api/generated/system_parameters"
	gentimers "FGG-Service/api/generated/timers"
	genwheeleffects "FGG-Service/api/generated/wheel_effects"
	"FGG-Service/src/auth/ctrlauth"
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/effects/ctrleffects"
	"FGG-Service/src/exchanges/ctrlexchanges"
	"FGG-Service/src/games/ctrlgames"
	"FGG-Service/src/grants/ctrlgrants"
	"FGG-Service/src/items/ctrlitems"
	"FGG-Service/src/parties/ctrlparties"
	"FGG-Service/src/perks/ctrlperks"
	"FGG-Service/src/points/ctrlpoints"
	"FGG-Service/src/sysparams/ctrlsysparams"
	"FGG-Service/src/timers/ctrltimers"
	"FGG-Service/src/timers/srvtimers"
	"FGG-Service/src/wheeleffects/ctrlwheeleffects"
	"embed"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed index.html
//go:embed api/specification
var scalarUI embed.FS

func main() {
	e := echo.New()
	e.HideBanner = true

	dbCloseFunc := dbaccess.Init()
	defer dbCloseFunc()

	registerHandlers(e)
	addScalarRoutes(e)
	fixCORS(e)

	f := createFileAndStartLogger()
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	startLogScheduler(f)

	defer func(e *echo.Echo) {
		_ = e.Close()
	}(e)

	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}

func registerHandlers(e *echo.Echo) {
	ts := srvtimers.NewService()

	genauth.RegisterHandlers(e, ctrlauth.NewController())
	gengames.RegisterHandlers(e, ctrlgames.NewController(ts))
	genitems.RegisterHandlers(e, ctrlitems.NewController())
	geneffects.RegisterHandlers(e, ctrleffects.NewController())
	genperks.RegisterHandlers(e, ctrlperks.NewController())
	genexchanges.RegisterHandlers(e, ctrlexchanges.NewController())
	gengrants.RegisterHandlers(e, ctrlgrants.NewController())
	genparties.RegisterHandlers(e, ctrlparties.NewController())
	genpoints.RegisterHandlers(e, ctrlpoints.NewController())
	gensysparams.RegisterHandlers(e, ctrlsysparams.NewController())
	gentimers.RegisterHandlers(e, ctrltimers.NewController(ts))
	genwheeleffects.RegisterHandlers(e, ctrlwheeleffects.NewController())
}

func createFileAndStartLogger() *os.File {
	logsDir := getLogsDir()
	filename := filepath.Join(logsDir, time.Now().Format("2006-01-02")+".txt")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)

	if err != nil {
		panic(err)
	}

	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	slog.SetDefault(logger)

	return file
}

func getLogsDir() string {
	root := os.Getenv("APP_ROOT")

	if root != "" {
		return filepath.Join(root, "logs")
	}

	return "logs"
}

// startLogScheduler rotates the log file every midnight. The file the logger already writes to is
// passed in so that the first rotation closes it too.
func startLogScheduler(currentFile *os.File) {
	scheduler, err := gocron.NewScheduler()

	if err != nil {
		panic(err)
	}

	_, err = scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(0, 0, 0))),
		gocron.NewTask(func() {
			previousFile := currentFile
			currentFile = createFileAndStartLogger()

			_ = previousFile.Close()
		}),
	)

	if err != nil {
		panic(err)
	}

	scheduler.Start()
}

func addScalarRoutes(e *echo.Echo) {
	fileServer := http.FileServer(http.FS(scalarUI))
	e.GET("/api/specification/*", echo.WrapHandler(fileServer))
	e.GET("/scalar/*", echo.WrapHandler(http.StripPrefix("/scalar", fileServer)))
	e.GET("/scalar", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/scalar/")
	})
}

func fixCORS(e *echo.Echo) {
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{"*"}}))
}
