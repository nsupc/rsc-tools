package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var arguments struct {
	User     string   `arg:"-u,--user,required" help:"Script user"`
	Key      string   `arg:"-k,--key,required" help:"Google Sheets API key"`
	Delegate string   `arg:"-d,--delegate" help:"Delegate nation" default:"le_libertia"`
	Region   string   `arg:"-r,--region" help:"Region" default:"europeia"`
	Excluded []string `arg:"-x,--excluded,separate" help:"Excluded nations -- VD, RSC, etc. Use once per nation (-x nation1 -x nation2...)"`
	Base     int      `arg:"-b,--base" help:"Base endocap" default:"10"`
	Standard int      `arg:"-e,--standard" help:"Standard endocap" default:"25"`
	Citizen  int      `arg:"-c,--citizen" help:"Citizen endocap" default:"50"`
	Verbose  bool     `arg:"-v,--verbose" help:"Verbose output"`
}

type Args struct {
	User     string
	Key      string
	Delegate string
	Region   string
	Excluded []string
	Base     int
	Standard int
	Citizen  int
	Verbose  bool
}

type Endorser struct {
	name       string
	percentage float64
	endorsing  []string
}

type Nation struct {
	ID           string `xml:"id,attr"`
	Endorsements string `xml:"ENDORSEMENTS"`
}

type Region struct {
	ID          string      `xml:"id,attr"`
	CensusRanks CensusRanks `xml:"CENSUSRANKS"`
}

type CensusRanks struct {
	ID      string         `xml:"id,attr"`
	Nations []CensusNation `xml:"NATIONS>NATION"`
}

type CensusNation struct {
	Name  string `xml:"NAME"`
	Rank  int    `xml:"RANK"`
	Score int    `xml:"SCORE"`
}

func contains(s []string, e string) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}

func getCitizenNations(key string) []string {

	// Create a new context and set the API key
	ctx := context.Background()

	// Create a new HTTP client with the API key
	httpClient := option.WithAPIKey(key)

	// Create a new service client
	service, err := sheets.NewService(ctx, httpClient)
	if err != nil {
		log.Fatal(err)
	}

	// Spreadsheet parameters
	spreadsheetID := "1Zi2HtQuykoWV2P36B61J_eBnhSgj3VyDWFUbtbYWyTo"
	readRange := "Citizens!C2:C"

	// Read the data from the spreadsheet
	response, err := service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatal(err)
	}

	// Create a string slice from the response data
	var data []string
	for _, i := range response.Values {
		data = append(data, i[0].(string))
	}

	return data
}

func getDelegateEndorsements(client *http.Client, user string, del string) []string {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.nationstates.net/cgi-bin/api.cgi?nation=%s&q=endorsements", del), nil)
	if err != nil {
		log.Fatal("Error creating request:", err)

	}

	req.Header.Set("User-Agent", fmt.Sprintf("Tarters/1.0 (%s)", user))

	// Make the API request
	response, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making the API request:", err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading the response body:", err)
	}

	// Parse the XML response
	var nation Nation
	err = xml.Unmarshal(body, &nation)
	if err != nil {
		log.Fatal("Error parsing the XML response:", err)
	}

	time.Sleep(time.Second)

	return strings.Split(nation.Endorsements, ",")
}

func getTopViolators(client *http.Client, args Args, citizens []string, delendos []string) map[string]int {
	endorsements := make(map[string]int)

	offset := 1
outer:
	for {

		fmt.Printf("Checking nations %v through %v\n", offset, offset+20)

		req, err := http.NewRequest("GET", fmt.Sprintf("https://www.nationstates.net/cgi-bin/api.cgi?region=%s&q=censusranks;scale=66;start=%v", args.Region, offset), nil)
		if err != nil {
			log.Fatal("Error creating request:", err)

		}

		req.Header.Set("User-Agent", fmt.Sprintf("Tarters/1.0 (%s)", args.User))

		// Make the API request
		response, err := client.Do(req)
		if err != nil {
			log.Fatal("Error making the API request:", err)
		}
		defer response.Body.Close()

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal("Error reading the response body:", err)
		}

		// Parse the XML response
		var region Region
		err = xml.Unmarshal(body, &region)
		if err != nil {
			log.Fatal("Error parsing the XML response:", err)
		}

		for _, nation := range region.CensusRanks.Nations {
			if nation.Score <= 0 {
				break outer
			} else if contains(args.Excluded, nation.Name) || nation.Name == args.Delegate {
				continue
			} else if contains(citizens, nation.Name) && contains(delendos, nation.Name) {
				if nation.Score <= args.Citizen {
					continue
				} else {
					endorsements[nation.Name] = nation.Score - args.Citizen
				}
			} else if contains(delendos, nation.Name) {
				if nation.Score <= args.Standard {
					continue
				} else {
					endorsements[nation.Name] = nation.Score - args.Standard
				}
			} else {
				if nation.Score <= args.Base {
					continue
				} else {
					endorsements[nation.Name] = nation.Score - args.Base
				}
			}
		}

		offset += 20
		time.Sleep(time.Second)
	}

	return endorsements
}

func getViolatorEndorsements(client *http.Client, args Args, violators map[string]int) []Endorser {
	endorsers := make(map[string]Endorser)
	percentage := 100 / float64(len(violators))

	for violator := range violators {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://www.nationstates.net/cgi-bin/api.cgi?nation=%s&q=endorsements", violator), nil)
		if err != nil {
			log.Fatal("Error creating request:", err)

		}

		req.Header.Set("User-Agent", fmt.Sprintf("Tarters/1.0 (%s)", args.User))

		// Make the API request
		response, err := client.Do(req)
		if err != nil {
			log.Fatal("Error making the API request:", err)
		}
		defer response.Body.Close()

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal("Error reading the response body:", err)
		}

		// Parse the XML response
		var nation Nation
		err = xml.Unmarshal(body, &nation)
		if err != nil {
			log.Fatal("Error parsing the XML response:", err)
		}

		for _, endorser := range strings.Split(nation.Endorsements, ",") {
			if entry, ok := endorsers[endorser]; ok {
				entry.percentage += percentage
				entry.endorsing = append(entry.endorsing, violator)
				endorsers[endorser] = entry
			} else {
				endorsers[endorser] = Endorser{endorser, percentage, []string{violator}}
			}
		}
	}

	var sortedEndorsers []Endorser = make([]Endorser, 0, len(endorsers))
	for _, endorser := range endorsers {
		sortedEndorsers = append(sortedEndorsers, endorser)
	}

	sort.Slice(sortedEndorsers, func(i, j int) bool {
		return sortedEndorsers[i].percentage > sortedEndorsers[j].percentage
	})

	return sortedEndorsers
}

func outputResults(args Args, endorsers []Endorser) {
	if args.Verbose {
		file, err := os.Create("output.txt")
		if err != nil {
			log.Fatal("Error creating output.txt:", err)
		}
		defer file.Close()

		for _, endorser := range endorsers {
			file.WriteString(fmt.Sprintf("%s: %.2f%%\n%s\n\n", endorser.name, endorser.percentage, strings.Join(endorser.endorsing, ",")))
		}
	} else {
		file, err := os.Create("output.txt")
		if err != nil {
			log.Fatal("Error creating output.txt:", err)
		}
		defer file.Close()

		for _, endorser := range endorsers {
			file.WriteString(fmt.Sprintf("%s,%.2f%%\n", endorser.name, endorser.percentage))
		}
	}
}

func main() {
	arg.MustParse(&arguments)

	args := Args{
		strings.ToLower(strings.ReplaceAll(arguments.User, " ", "_")),
		arguments.Key,
		strings.ToLower(strings.ReplaceAll(arguments.Delegate, " ", "_")),
		strings.ToLower(strings.ReplaceAll(arguments.Region, " ", "_")),
		arguments.Excluded,
		arguments.Base,
		arguments.Standard,
		arguments.Citizen,
		arguments.Verbose,
	}

	fmt.Println("Getting citizen nations")
	citizenNations := getCitizenNations(args.Key)

	client := &http.Client{}

	fmt.Println("Getting delegate endorsements")
	delegateEndorsements := getDelegateEndorsements(client, args.User, args.Delegate)

	fmt.Println("Getting nations and endorsement numbers")
	violators := getTopViolators(client, args, citizenNations, delegateEndorsements)

	fmt.Println("Getting violator endorsements")
	endorsers := getViolatorEndorsements(client, args, violators)

	fmt.Println("Writing results to output.txt")
	outputResults(args, endorsers)
}
