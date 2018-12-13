package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/dyatlov/go-oembed/oembed"
	m "github.com/keighl/metabolize"
)

// MetaData structure
type MetaData struct {
	Title       string  `meta:"og:title"`
	Description string  `meta:"og:description,description"`
	Type        string  `meta:"og:type"`
	URL         url.URL `meta:"og:url"`
}

// AuthorData structure
type AuthorData struct {
	AuthorName string `json:"author_name,omitempty"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// oembed reads from providers.json so the scripts know what sites authors can be crawled from
	providers, err := ioutil.ReadFile("providers.json")
	if err != nil {
		panic(err)
	}

	// Creates an oembed instance
	oe := oembed.NewOembed()
	oe.ParseProviders(bytes.NewReader(providers))

	// Creates a Text file which is used as the input file GO-MetaCrawler reads from
	inputFile, err := os.Open("test.txt")
	if err != nil {
		log.Fatal(err)
	}

	// make sure it gets closed
	defer inputFile.Close()

	// Creates the CSV file
	outputFile, err := os.Create("result.csv")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// create a new scanner and read the file line by line
	scanner := bufio.NewScanner(inputFile)

	for scanner.Scan() {
		var dataStructured []string

		// keep original url name ( some sites don't send it back so we still need them )
		originalURL := scanner.Text()

		// Issues a GET to the specific URL
		siteMeta, err := http.Get(originalURL)

		if err != nil {
			dataStructured = []string{originalURL, err.Error(), "-", "-"}
			log.Printf("Metadata could not be extracted from %s because of %s.", originalURL, err.Error())
		} else {
			// Creates an empty structure for MetaData.
			data := new(MetaData)
			// Creates an empty structure for AuthorData.
			adata := new(AuthorData)
			adata.AuthorName = "not retrievable"
			item := oe.FindItem(originalURL)

			if item != nil {
				// Gets oembed data from URL
				info, err := item.FetchOembed(oembed.Options{URL: originalURL})
				if err != nil {
					log.Printf("An error occured: %s\n", err.Error())
				} else if info.Status >= 300 {
						log.Printf("Response status code is: %d\n", info.Status)
					} else {
						// connecting adata.AuthorName to info.AuthorName and printing the input URL to the console
						log.Println(originalURL)
						adata.AuthorName = info.AuthorName
					}
				}
			}

			// Grabs information from a sites body and puts it in to data.
			err = m.Metabolize(siteMeta.Body, data)
			if err != nil {
				log.Println(err)
			}
			// Structures the data for output to CSV
			dataStructured = []string{originalURL, data.Title, data.Type, data.Description, adata.AuthorName}
		}

		// Send meta  data title to CSV Writer
		// create array of strings with structure components
		CSVWriter(writer, dataStructured)
	}

	// check for errors
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// CSVWriter function to write to CSV
func CSVWriter(writer *csv.Writer, data []string) bool {

	// Writes input to CSV
	err := writer.Write(data)
	if err != nil {
		log.Printf("Can not write to file ; %s , notifying user and carrying on.", err)
		return false
	}
	return true
}
