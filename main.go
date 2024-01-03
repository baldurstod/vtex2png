package main


import (
	"fmt"
	"flag"
	"os"
	"log"
	"runtime/debug"
)

func main() {
	var inputFile string
	var outputFile string
	var frame int


	flag.StringVar(&inputFile, "i", "", "Input file")
	flag.StringVar(&outputFile, "o", "", "Output file")
	flag.IntVar(&frame, "f", 0, "Frame number")
	flag.Parse()


	if inputFile == "" {
		fmt.Println("No input file provided. Use the flag -i")
		os.Exit(1)
	}

	if outputFile == "" {
		fmt.Println("No output file provided. Use the flag -o")
		os.Exit(1)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error while processing", inputFile, r, string(debug.Stack()))
		}
	}()

	convertToPNG(inputFile, outputFile, frame)
}


func convertToPNG(inputFile string, outputFile string, frame int) {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	vtexFile := VtexFile{}
	vtexFile.SetData(data)
	buf := vtexFile.GetVtexData()

	if (buf != nil) {
		error := os.WriteFile(outputFile, buf, 0666)
		if error != nil {
			fmt.Println(error)
		}
	}
}
