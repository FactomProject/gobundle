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
`} // filebits end

type ByteWriter struct {
	count int
	other io.Writer
}

func (bw *ByteWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		err = bw.WriteByte(b)
		if err != nil { return }
		n++
	}
	return
}

func (bw *ByteWriter) WriteByte(b byte) (err error) {
	if bw.count > 0 && bw.count % 256 == 0 {
		_, err = fmt.Fprint(bw.other, "\n")
		if err != nil { return }
	}
	if bw.count % 16 == 0 {
		_, err = fmt.Fprint(bw.other, "\n\t")
		if err != nil { return }
	} else if bw.count % 4 == 0 {
		_, err = fmt.Fprint(bw.other, "   ")
		if err != nil { return }
	}
	_, err = fmt.Fprintf(bw.other, "0x%02X, ", b)
	bw.count++
	return err
}

func main() {
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
		
		bundle, err := os.OpenFile(bundlepath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
		if err != nil { panic(err) }
		
		_, err = bundle.WriteString(filebits[0])
		if err != nil { panic(err) }
		
		bundledir := filepath.Join(pkg.Dir, "bundle")
		zipwr, err := gzip.NewWriterLevel(&ByteWriter{other: bundle}, gzip.BestCompression)
		if err != nil { panic(err) }
		
		tarwr := tar.NewWriter(zipwr)
		writeDir(tarwr, bundledir, bundledir)
		
		err = tarwr.Close()
		if err != nil { panic(err) }
		
		err = zipwr.Close()
		if err != nil { panic(err) }
		
		_, err = bundle.WriteString(filebits[1])
		if err != nil { panic(err) }
		
		err = bundle.Close()
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
















