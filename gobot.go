// gobot is an attempt to do a single, monolithic Go application which deals with autonomous agents in OpenSimulator.
package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/op/go-logging" // more complete package to log to different outputs; we start with file, syslog, and stderr; later: WebSockets?
	"github.com/Pallinder/go-randomdata"
	"github.com/spf13/viper" // to read config files
	"golang.org/x/net/websocket"
	"gopkg.in/natefinch/lumberjack.v2" // rolling file logs
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"path/filepath"
//	"sync/atomic" // this is weird since we USE sync/atomic, but the Go compiler complains...
	"syscall"
)

var (
	// Default configurations, hopefully exported to other files and packages
	// we probably should have a struct for this (or even several)
	Host, GoBotDSN, URLPathPrefix, PDO_Prefix, PathToStaticFiles,
	ServerPort, FrontEnd, MapURL, LSLSignaturePIN string
	logFileName string = "log/gobot.log"
	logMaxSize, logMaxBackups, logMaxAge int // configuration for the go-logging logger
	logSeverityStderr, logSeverityFile, logSeveritySyslog logging.Level // more configuration for the go-logging logger
	ShowPopulation bool = true
	Log = logging.MustGetLogger("gobot")	// configuration for the go-logging logger, must be available everywhere
	logFormat logging.Formatter	// must be initialised or all hell breaks loose
)

const NullUUID = "00000000-0000-0000-0000-000000000000" // always useful when we deal with SL/OpenSimulator...

//type templateParameters map[string]string
type templateParameters map[string]interface{}

// loadConfiguration loads all the configuration from the config.toml file.
// It's a separate function because we want to be able to do a killall -HUP gobot to force the configuration to be read again.
// Also, if the configuration file changes, this ought to read it back in again without the need of a HUP signal (20170811).
func loadConfiguration() {
	fmt.Print("Reading Gobot configuration:")	// note that we might not have go-logging active as yet, so we use fmt
	// Open our config file and extract relevant data from there
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return	// we might still get away with this!
	}
	// Without these set, we cannot do anything
	viper.SetDefault("gobot.EngineRunning", true) // try to set this as quickly as possible, or the engine WILL run!
	EngineRunning.Store(viper.GetBool("gobot.EngineRunning")); fmt.Print(".")
	viper.SetDefault("gobot.Host", "localhost") // to prevent bombing out with panics
	Host = viper.GetString("gobot.Host"); fmt.Print(".")
	viper.SetDefault("gobot.URLPathPrefix", "/go")
	URLPathPrefix = viper.GetString("gobot.URLPathPrefix"); fmt.Print(".")
	GoBotDSN = viper.GetString("gobot.GoBotDSN"); fmt.Print(".")
	viper.SetDefault("PDO_Prefix", "mysql") // for now, nothing else will work anyway...
	PDO_Prefix = viper.GetString("gobot.PDO_Prefix"); fmt.Print(".")
	viper.SetDefault("gobot.PathToStaticFiles", "~/go/src/gobot")
	path, err := expandPath(viper.GetString("gobot.PathToStaticFiles")); fmt.Print(".")
	if err != nil {
		fmt.Println("Error expanding path:", err)
		path = ""	// we might get away with this as well
	}
	PathToStaticFiles = path
	viper.SetDefault("gobot.ServerPort", ":3000")
	ServerPort = viper.GetString("gobot.ServerPort"); fmt.Print(".")
	FrontEnd = viper.GetString("gobot.FrontEnd"); fmt.Print(".")
	MapURL = viper.GetString("opensim.MapURL"); fmt.Print(".")
	viper.SetDefault("gobot.LSLSignaturePIN", "9876") // better than no signature at all
	LSLSignaturePIN = viper.GetString("opensim.LSLSignaturePIN"); fmt.Print(".")
	viper.SetDefault("gobot.ShowPopulation", true) // try to set this as quickly as possible, or the engine WILL run!
	ShowPopulation = viper.GetBool("gobot.ShowPopulation"); fmt.Print(".")
	// logging options
	viper.SetDefault("log.FileName", "log/gobot.log")
	logFileName = viper.GetString("log.FileName"); fmt.Print(".")
	viper.SetDefault("log.Format", `%{color}%{time:2006/01/02 15:04:05.0} %{shortfile} - %{shortfunc} ▶ %{level:.4s}%{color:reset} %{message}`)
	logFormat = logging.MustStringFormatter(viper.GetString("log.Format")); fmt.Print(".")
	viper.SetDefault("log.MaxSize", 500)
	logMaxSize = viper.GetInt("log.MaxSize"); fmt.Print(".")
	viper.SetDefault("log.MaxBackups", 3)
	logMaxBackups = viper.GetInt("log.MaxBackups"); fmt.Print(".")
	viper.SetDefault("log.MaxAge", 28)
	logMaxAge = viper.GetInt("log.MaxAge"); fmt.Print(".")
	viper.SetDefault("log.SeverityStderr", logging.DEBUG)
	switch viper.GetString("log.SeverityStderr") {
		case "CRITICAL":
			logSeverityStderr = logging.CRITICAL
    	case "ERROR":
			logSeverityStderr = logging.ERROR
    	case "WARNING":
			logSeverityStderr = logging.WARNING
    	case "NOTICE":
			logSeverityStderr = logging.NOTICE
    	case "INFO":
			logSeverityStderr = logging.INFO
    	case "DEBUG":
			logSeverityStderr = logging.DEBUG
		// default case is handled directly by viper
	}
	fmt.Print(".")
	viper.SetDefault("log.SeverityFile", logging.DEBUG)
	switch viper.GetString("log.SeverityFile") {
		case "CRITICAL":
			logSeverityFile = logging.CRITICAL
    	case "ERROR":
			logSeverityFile = logging.ERROR
    	case "WARNING":
			logSeverityFile = logging.WARNING
    	case "NOTICE":
			logSeverityFile = logging.NOTICE
    	case "INFO":
			logSeverityFile = logging.INFO
    	case "DEBUG":
			logSeverityFile = logging.DEBUG
	}
	fmt.Print(".")
	viper.SetDefault("log.SeveritySyslog", logging.CRITICAL) // we don't want to swamp syslog with debugging messages!!
	switch viper.GetString("log.SeveritySyslog") {
		case "CRITICAL":
			logSeveritySyslog = logging.CRITICAL
    	case "ERROR":
			logSeveritySyslog = logging.ERROR
    	case "WARNING":
			logSeveritySyslog = logging.WARNING
    	case "NOTICE":
			logSeveritySyslog = logging.NOTICE
    	case "INFO":
			logSeveritySyslog = logging.INFO
    	case "DEBUG":
			logSeveritySyslog = logging.DEBUG
	}
	fmt.Print(".")
	fmt.Println("read!")	// note that we might not have go-logging active as yet, so we use fmt
	
	// Setup the lumberjack rotating logger. This is because we need it for the go-logging logger when writing to files. (20170813)
	rotatingLogger := &lumberjack.Logger{
	    Filename:   logFileName,	// this is an option set on the config.yaml file, eventually the others will be so, too.
	    MaxSize:    logMaxSize, // megabytes
	    MaxBackups: logMaxBackups,
	    MaxAge:     logMaxAge, //days
	}
	// Setup the go-logging Logger. (20170812) We have three loggers: one to stderr, one to a logfile, one to syslog for critical stuff. (20170813
	backendStderr	:= logging.NewLogBackend(os.Stderr, "", 0)
	backendFile		:= logging.NewLogBackend(rotatingLogger, "", 0)
	backendSyslog,_	:= logging.NewSyslogBackend("")

	// Set formatting for stderr and file (basically the same). I'm assuming syslog has its own format, but I'll have to see what happens (20170813).
	backendStderrFormatter	:= logging.NewBackendFormatter(backendStderr, logFormat)
	backendFileFormatter	:= logging.NewBackendFormatter(backendFile, logFormat)

	// Check if we're overriding the default severity for each backend. This is user-configurable. By default: DEBUG, DEBUG, CRITICAL.
	// TODO(gwyneth): What about a WebSocket backend using https://github.com/cryptix/exp/wslog ? (20170813)
	backendStderrLeveled := logging.AddModuleLevel(backendStderrFormatter)
	backendStderrLeveled.SetLevel(logSeverityStderr, "gobot")
	backendFileLeveled := logging.AddModuleLevel(backendFileFormatter)
	backendFileLeveled.SetLevel(logSeverityFile, "gobot")
	backendSyslogLeveled := logging.AddModuleLevel(backendSyslog)
	backendSyslogLeveled.SetLevel(logSeveritySyslog, "gobot")

	// Set the backends to be used. Logging should commence now.
	logging.SetBackend(backendStderrLeveled, backendFileLeveled, backendSyslogLeveled)
	fmt.Println("Logging set up.")
}

// main starts here.
func main() {
	// to change the flags on the default logger
	// see https://stackoverflow.com/a/24809859/1035977
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Config viper, which reads in the configuration file every time it's needed.
	// Note that we need some hard-coded variables for the path and config file name.
	viper.SetConfigName("config")
	viper.SetConfigType("toml") // just to make sure; it's the same format as OpenSimulator (or MySQL) config files
	viper.AddConfigPath("$HOME/go/src/gobot/") // that's how I have it
	viper.AddConfigPath("$HOME/go/src/github.com/GwynethLlewelyn/gobot/") // that's how you'll have it
	viper.AddConfigPath(".")               // optionally look for config in the working directory

	loadConfiguration() // this gets loaded always, on the first time it runs
	viper.WatchConfig() // if the config file is changed, this is supposed to reload it (20170811)
	viper.OnConfigChange(func(e fsnotify.Event) {
		if (Log == nil) {
			fmt.Println("Config file changed:", e.Name) // if we couldn't configure the logging subsystem, it's better to print it to the console
		} else {
			Log.Info("Config file changed:", e.Name)
		}
		loadConfiguration() // I think that this needs to be here, or else, how does Viper know what to call?
	})
	
	// prepares a special channel to look for termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGCONT)

	// goroutine which listens to signals and calls the loadConfiguration() function if someone sends us a HUP
	go func() {
		for {
	        sig := <-sigs
	        Log.Notice("Got signal", sig)
	        switch sig {
		        case syscall.SIGUSR1:
		        	sendMessageToBrowser("status", "", randomdata.FullName(randomdata.Female) + "<br />", "") // defined on engine.go for now
		        case syscall.SIGUSR2:
		        	sendMessageToBrowser("status", "", randomdata.Country(randomdata.FullCountry) + "<br />", "") // defined on engine.go for now
		        case syscall.SIGHUP:
		        	// HACK(gwyneth): if the engine dies, send a SIGHUP to get it running again (20170811).
		        	//  Moved to HUP (instead of CONT done 20170723) because the configuration is now automatically re-read.
		        	// NOTE(gwyneth): Noticed that when the app is suspended with a Ctrl-Z from the shell, it will get a HUP
		        	//  and after that, when putting it in the background, it will get a CONT. So we change the logic appropriately!
		        	sendMessageToBrowser("status", "warning", "<code>SIGHUP</code> caught, stopping engine.", "")
		        	EngineRunning.Store(false)
		        case syscall.SIGCONT:
		        	sendMessageToBrowser("status", "warning", "<code>SIGCONT</code> caught, restarting engine.", "")
		        	EngineRunning.Store(true)
				default:
		        	Log.Warning("Unknown UNIX signal", sig, "caught!! Ignoring...")
	        }
        }
    }()

	// do some database tests. If it fails, it means the database is broken or corrupted and it's worthless
	//  to run this application anyway!
	Log.Info("Testing opening database connection at ", GoBotDSN, "\nPath to static files is:", PathToStaticFiles)

	db, err := sql.Open(PDO_Prefix, GoBotDSN) // presumes mysql for now (supercedes old sqlite3)
	checkErrPanic(err) // abort if it cannot even open the database

	// query
	rows, err := db.Query("SELECT UUID, Name, Location, Position FROM Agents")
	checkErrPanic(err) // if select fails, probably the table doesn't even exist; we abort because the database is corrupted!

 	var agent AgentType // type defined on ui.go to be used on database requests

	for rows.Next() {
		err = rows.Scan(&agent.UUID, &agent.Name, &agent.Location, &agent.Position)
		checkErr(err) // if we get some errors here, we will get in trouble later on; but we might have an empty database, so that's ok.
		Log.Debug("Agent '", *agent.Name.Ptr(), "' (", *agent.UUID.Ptr(), ") at", *agent.Location.Ptr(), "Position:", *agent.Position.Ptr())
	}
	rows.Close()
	db.Close()

	Log.Infof("\n\nDatabase tests ended, last error was %v:\n\nStarting gobot application at: http://%s%v%s\n\n", err, Host, ServerPort, URLPathPrefix)

	// this was just to make tests; now start the engine as a separate goroutine in the background
	
	go engine() // run everything but the kitchen sink in parallel; yay goroutines!
	
	go garbageCollector() // this will periodically remove from the database all old items that are 'dead' (20170730)

	// Now prepare the web interface

	// Load all templates
	err = GobotTemplates.init(PathToStaticFiles + "/templates/*.tpl")
	checkErr(err) // abort if templates are not found

	// Configure routers for our many inworld scripts
	// In my case, paths with /go will be served by gobot, the rest by nginx as before
	// Exception is for static files
	http.HandleFunc(URLPathPrefix + "/update-inventory/",	updateInventory)
	http.HandleFunc(URLPathPrefix + "/update-sensor/",		updateSensor)
	http.HandleFunc(URLPathPrefix + "/register-position/",	registerPosition)
	http.HandleFunc(URLPathPrefix + "/register-agent/",		registerAgent)
	http.HandleFunc(URLPathPrefix + "/configure-cube/",		configureCube)
	http.HandleFunc(URLPathPrefix + "/process-cube/",		processCube)

	// Static files. This should be handled directly by nginx, but we include it here
	//  for a standalone version...
	fslib := http.FileServer(http.Dir(PathToStaticFiles + "/lib"))
	http.Handle(URLPathPrefix + "/lib/", http.StripPrefix(URLPathPrefix + "/lib/", fslib))

	templatelib := http.FileServer(http.Dir(PathToStaticFiles + "/templates"))
	http.Handle(URLPathPrefix + "/templates/",
		http.StripPrefix(URLPathPrefix + "/templates/", templatelib)) // not sure if this is needed

	// Deal with templated output for the admin back office, defined on backoffice.go
	// For now this is crude, each page is really very similar, but there are not many so each will get its own handler function for now
	http.HandleFunc(URLPathPrefix + "/admin/agents/",					backofficeAgents)
	http.HandleFunc(URLPathPrefix + "/admin/logout/",					backofficeLogout)
	http.HandleFunc(URLPathPrefix + "/admin/login/",					backofficeLogin) // probably not necessary
	http.HandleFunc(URLPathPrefix + "/admin/objects/",					backofficeObjects)
	http.HandleFunc(URLPathPrefix + "/admin/positions/",				backofficePositions)
	http.HandleFunc(URLPathPrefix + "/admin/inventory/",				backofficeInventory)
	http.HandleFunc(URLPathPrefix + "/admin/user-management/",			backofficeUserManagement)
	http.HandleFunc(URLPathPrefix + "/admin/commands/exec/",			backofficeCommandsExec)
	http.HandleFunc(URLPathPrefix + "/admin/commands/",					backofficeCommands)
	http.HandleFunc(URLPathPrefix + "/admin/controller-commands/exec/",	backofficeControllerCommandsExec)
	http.HandleFunc(URLPathPrefix + "/admin/controller-commands/",		backofficeControllerCommands)
	http.HandleFunc(URLPathPrefix + "/admin/engine/",					backofficeEngine)
	// LSL Template Generator
	http.HandleFunc(URLPathPrefix + "/admin/lsl-register-object/",		backofficeLSLRegisterObject)
	http.HandleFunc(URLPathPrefix + "/admin/lsl-bot-controller/",		backofficeLSLBotController)
	http.HandleFunc(URLPathPrefix + "/admin/lsl-agent-scripts/",		backofficeLSLAgentScripts)
	// fallthrough for admin
	http.HandleFunc(URLPathPrefix + "/admin/",							backofficeMain)

	// deal with agGrid UI elements
	http.HandleFunc(URLPathPrefix + "/uiObjects/",						uiObjects)
	http.HandleFunc(URLPathPrefix + "/uiObjectsUpdate/",				uiObjectsUpdate) // to change the database manually
	http.HandleFunc(URLPathPrefix + "/uiObjectsRemove/",				uiObjectsRemove) // to remove rows of the database manually
	http.HandleFunc(URLPathPrefix + "/uiAgents/",						uiAgents)
	http.HandleFunc(URLPathPrefix + "/uiAgentsUpdate/",					uiAgentsUpdate)
	http.HandleFunc(URLPathPrefix + "/uiAgentsRemove/",					uiAgentsRemove)
	http.HandleFunc(URLPathPrefix + "/uiPositions/",					uiPositions)
	http.HandleFunc(URLPathPrefix + "/uiPositionsUpdate/",				uiPositionsUpdate)
	http.HandleFunc(URLPathPrefix + "/uiPositionsRemove/",				uiPositionsRemove)
	http.HandleFunc(URLPathPrefix + "/uiInventory/",					uiInventory)
	http.HandleFunc(URLPathPrefix + "/uiInventoryUpdate/",				uiInventoryUpdate)
	http.HandleFunc(URLPathPrefix + "/uiInventoryRemove/",				uiInventoryRemove)
	http.HandleFunc(URLPathPrefix + "/uiUserManagement/",				uiUserManagement)
	http.HandleFunc(URLPathPrefix + "/uiUserManagementUpdate/",			uiUserManagementUpdate)
	http.HandleFunc(URLPathPrefix + "/uiUserManagementRemove/",			uiUserManagementRemove)

	// Handle Websockets on Engine
	http.Handle(URLPathPrefix + "/wsEngine/",							websocket.Handler(serveWs))

	http.HandleFunc(URLPathPrefix + "/",								backofficeLogin) // if not auth, then get auth

    err = http.ListenAndServe(ServerPort, nil) // set listen port
    checkErr(err) // if it can't listen to all the above, then it has to abort anyway
}

// checkErrPanic logs a fatal error and panics.
func checkErrPanic(err error) {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		Log.Panic(filepath.Base(file), ":", line, ":", pc, ok, " - panic:", err)
	}
}

// checkErr checks if there is an error, and if yes, it logs it out and continues.
//  this is for 'normal' situations when we want to get a log if something goes wrong but do not need to panic
func checkErr(err error) {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		Log.Error(filepath.Base(file), ":", line, ":", pc, ok, " - error:", err)
	}
}

// expandPath expands the tilde as the user's home directory.
//  found at http://stackoverflow.com/a/43578461/1035977
func expandPath(path string) (string, error) {
    if len(path) == 0 || path[0] != '~' {
        return path, nil
    }

    usr, err := user.Current()
    if err != nil {
        return "", err
    }
    return filepath.Join(usr.HomeDir, path[1:]), nil
}