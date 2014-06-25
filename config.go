package gobundle

import (
	"archive/tar"
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// setup/config info
var Setup struct {
	// info the consumer should set
	Application struct {
		// name of the app
		Name string
		
		// if the app should use system folders
		System bool
		
		// windows stuff
		Roaming bool
		
		// resource tar binary
		ResourceTar io.Reader
	}
	
	// directories where config and data files are stored
	Directories struct {
		Config, Data *string
	}
	
	// the unpack mode flag
	Unpack *string
	
	// the handler for non-data/-config resources
	Handler func(tarpath string, filedata io.Reader)
}

var unpackHelp =
`Controls when the config and data directories are unpacked:
    --unpack unpack[,force]    Unpack; synonyms are 1, t, true (all case insensitive)
    --unpack supress           Don't unpack; synonyms are 0, f, false (all case insensitive)
    --unpack detect[,force]    (default) Unpack if neither directory exists
    --unpack only[,force]      Unpack then exit
    
    Adding force will overwrite existing files`

// set up config info
//   unpacks if necessary
//   calls flag.Parse()
//   set up other flags before this call
func Init() {
	flags()
	flag.Parse()
	unpack()
}

func flags() {
	app := Setup.Application
	
	if app.Name == "" {
		panic("Empty app name!")
	}
	
	Setup.Directories.Config = flag.String("conf", AppDir(app.Name, true, app.System, app.Roaming), "Set the configuration directory")
	Setup.Directories.Data = flag.String("data", AppDir(app.Name, false, app.System, app.Roaming), "Set the data directory")
	
	Setup.Unpack = flag.String("unpack", "detect", unpackHelp)
}

func unpack() {
	unpack := strings.Split(*Setup.Unpack, ",")
	var force, exit bool
	
	// check for force
	switch len(unpack) {
	case 1:
		
	case 2:
		switch unpack[1] {
		case "force":
			force = true
			
		default:
			panic(fmt.Sprint("--unpack: bad second argument: ", unpack[1]))
		}
	
	default:
		panic(fmt.Sprint("--unpack: bad number of arguments: ", *Setup.Unpack))
	}
	
	// parse the flag argument
	switch unpack[0] {
	case "", "unpack":
		// continue
		
	case "suppress":
		// don't unpack
		return
		
	case "detect":
		// if either directory exists, return
		_, err := os.Stat(*Setup.Directories.Config)
		if err == nil { return }
		_, err = os.Stat(*Setup.Directories.Data)
		if err == nil { return }
		
	case "only":
		exit = true
		
	default:
		panic(fmt.Sprint("--unpack: bad argument: ", *Setup.Unpack))
	}
	
	// can't unpack a non-existant tar
	if Setup.Application.ResourceTar == nil {
		panic("Empty resource tar reader!")
	}
	
	// make the tar reader
	reader := tar.NewReader(Setup.Application.ResourceTar)
	
	// set up flags for writing
	//   if force, truncate
	//   else, create, don't overwrite
	flag := os.O_WRONLY | os.O_CREATE
	if force {
		flag |= os.O_TRUNC
	} else {
		flag |= os.O_EXCL
	}
	
	for {
		// get the next header
		//   if the tar is done, break
		//   panic on other errors
		header, err := reader.Next()
		if err == io.EOF { break }
		if err != nil { panic(err) }
		
		// get the index of the first slash
		//   if there isn't a slash, ignore/handoff
		index := strings.Index(header.Name, string(os.PathSeparator))
		if index < 0 {
			if Setup.Handler != nil {
				Setup.Handler(header.Name, reader)
			}
			continue
		}
		
		// get the config or data directory
		var dir string
		base, rest := header.Name[:index], header.Name[index+1:]
		switch base {
		case "config":
			dir = *Setup.Directories.Config
			
		case "data":
			dir = *Setup.Directories.Data
			
			// ignore/handoff other stuff
		default:
			if Setup.Handler != nil {
				Setup.Handler(header.Name, reader)
			}
			continue
		}
		
		// make any directories
		path := filepath.Join(dir, rest)
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil { panic(err) }
		
		// open the file
		file, err := os.OpenFile(path, flag, 0644)
		if os.IsExist(err) { continue }
		if err != nil { panic(err) }
		
		// write the file
		_, err = bufio.NewReader(reader).WriteTo(file)
		if err != nil { panic(err) }
	}
	
	if exit {
		fmt.Println("Done unpacking")
		os.Exit(0)
	}
}