package main


import (
	"fmt"
	"flag"
	"os"
	"log"
)

func main() {
	var inputFile string
	var outputFile string
	var frame int

	flag.StringVar(&inputFile, "i", "", "Input file")
	flag.StringVar(&outputFile, "o", "", "Output file")
	flag.IntVar(&frame, "f", 0, "Frame number")
	flag.Parse()

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