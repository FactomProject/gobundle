package gobundle

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

func systemConfigBase() string {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("PROGRAMFILES")
		
	case "darwin":
		return filepath.Join("/Library", "Application Support")
		
	default:
		return filepath.Join("/etc", "opt")
	}
}

func systemDataBase() string {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("PROGRAMFILES")
		
	case "darwin":
		return filepath.Join("/Library", "Application Support")
		
	default:
		return filepath.Join("/opt")
	}
}

func userHome() string {
	var homeDir string
	usr, err := user.Current()
	if err == nil {
		homeDir = usr.HomeDir
	}
	
	if err != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}
	
	return homeDir
}

func windowsAppData(roaming bool) string {
	appData := os.Getenv("LOCALAPPDATA")
	if roaming || appData == "" {
		appData = os.Getenv("APPDATA")
	}
	
	return appData
}

func userConfigBase(roaming bool) string {
	switch runtime.GOOS {
	case "windows":
		return windowsAppData(roaming)
		
	case "darwin":
		return filepath.Join(userHome(), "Library", "Application Support")
		
	default:
		return filepath.Join(userHome())
	}
}

func userDataBase(roaming bool) string {
	switch runtime.GOOS {
	case "windows":
		return windowsAppData(roaming)
		
	case "darwin":
		return filepath.Join(userHome(), "Library", "Application Support")
		
	default:
		return filepath.Join(userHome())
	}
}

func AppDir(appName string, config, sys, roaming bool) string {
	// if the caller added a . prefx, remove it
	if strings.HasPrefix(appName, ".") {
		appName = appName[1:]
	}
	
	var base string
	if config && sys {
		base = systemConfigBase()
	} else if sys {
		base = systemDataBase()
	} else if config {
		base = userConfigBase(roaming)
	} else {
		base = userDataBase(roaming)
	}
	
	var dir string
	if config {
		dir = "config"
	} else {
		dir = "data"
	}
	
	switch runtime.GOOS {
	case "windows", "darwin":
		r, n := utf8.DecodeRune([]byte(appName))
		appName = string(unicode.ToUpper(r)) + appName[n:]
		
	default:
		r, n := utf8.DecodeRune([]byte(appName))
		appName = string(unicode.ToLower(r)) + appName[n:]
		
		if !sys {
			appName = "." + appName
		}
	}
	
	return filepath.Join(base, appName, dir)
}