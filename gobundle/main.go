package main 

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var filebits = []string{
`package main

import (
	"bytes"
	"compress/gzip"
	"github.com/firelizzard18/gobundle"
)

var _data = []byte{`,`
}

func init() {
	var err error
	gobundle.Setup.Application.ResourceTar, err = gzip.NewReader(bytes.NewReader(_data))
	if err != nil { panic(err) }
}
`}

type GoBundleWriter struct {
	*os.File
	count int
}

func Open(path string) (bw *GoBundleWriter, err error) {
	file, err := os.OpenFile(path, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
	if err != nil { return }
	
	bw = &GoBundleWriter{File: file}
	
	_, err = bw.File.WriteString(filebits[0])
	return
}

func (bw *GoBundleWriter) Close() (err error) {
	_, err = bw.File.WriteString(filebits[1])
	if err != nil { return }
	
	return bw.File.Close()
}

func (bw *GoBundleWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		err = bw.WriteByte(b)
		if err != nil { return }
		n++
	}
	return
}

func (bw *GoBundleWriter) WriteByte(b byte) (err error) {
	if bw.count > 0 && bw.count % 256 == 0 {
		_, err = fmt.Fprint(bw.File, "\n")
		if err != nil { return }
	}
	if bw.count % 16 == 0 {
		_, err = fmt.Fprint(bw.File, "\n\t")
		if err != nil { return }
	} else if bw.count % 4 == 0 {
		_, err = fmt.Fprint(bw.File, "   ")
		if err != nil { return }
	}
	_, err = fmt.Fprintf(bw.File, "0x%02X, ", b)
	bw.count++
	return err
}

func main() {
	file := flag.String("file", "", "Specify a file to tar to instead of writing a bundle.go file in the project.")
	
	flag.Parse()
	
	wd, err := os.Getwd()
	if err != nil { panic(err) }
	
	for _, arg := range flag.Args() {
		pkg, err := build.Import(arg, wd, 0)
		if err != nil { panic(err) }
		
		if pkg.IsCommand() {
			fmt.Println("Processing package", pkg.ImportPath)
		} else {
			fmt.Println("Skipping non-command package", pkg.ImportPath)
			continue
		}
		
		bundlepath := filepath.Join(pkg.Dir, "bundle.go")
		fmt.Println("\tWriting file ", bundlepath)
		
		var dest io.WriteCloser
		
		if *file == "" {
			dest, err = Open(bundlepath)
		} else {
			dest, err = os.OpenFile(*file, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
		}
		if err != nil { panic(err) }
		
		bundledir := filepath.Join(pkg.Dir, "bundle")
		zipwr, err := gzip.NewWriterLevel(dest, gzip.BestCompression)
		if err != nil { panic(err) }
		
		tarwr := tar.NewWriter(zipwr)
		writeDir(tarwr, bundledir, bundledir)
		
		err = tarwr.Close()
		if err != nil { panic(err) }
		
		err = zipwr.Close()
		if err != nil { panic(err) }
		
		err = dest.Close()
		if err != nil { panic(err) }
	}
}

func writeDir(wr *tar.Writer, prefix, directory string) {
	fmt.Println("\tProcessing directory ", directory)
	
	matches, err := filepath.Glob(filepath.Join(directory, "*"))
	if err != nil { panic(err) }
	
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	
	for _, match := range matches {
		if !strings.HasPrefix(match, prefix) {
			panic(fmt.Sprint("Path ", match, " does not have prefix ", prefix))
		}
		
		name := match[len(prefix):]
		
		file, err := os.Open(match)
		if err != nil { panic(err) }
		
		stat, err := file.Stat()
		if err != nil { panic(err) }
		
		if stat.IsDir() {
			writeDir(wr, prefix, match)
			continue
		}
		
		fmt.Println("\tArchiving bundle file", name)
		
		err = wr.WriteHeader(&tar.Header{
			Name: name,
			Size: stat.Size(),
		})
		if err != nil { panic(err) }
		
		var buf [1024]byte
		for {
			n, err := file.Read(buf[:])
			if err == io.EOF { break }
			if err != nil { panic(err) }
			
			_, err = wr.Write(buf[:n])
			if err != nil { panic(err) }
		}
	}
}
















