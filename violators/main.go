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

type Violator struct {
	name string
	over int
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

	ctx := context.Background()

	httpClient := option.WithAPIKey(key)

	service, err := sheets.NewService(ctx, httpClient)
	if err != nil {
		log.Fatal(err)
	}

	spreadsheetID := "1Zi2HtQuykoWV2P36B61J_eBnhSgj3VyDWFUbtbYWyTo"
	readRange := "Citizens!C2:C"

	response, err := service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatal(err)
	}

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

	response, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making the API request:", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading the response body:", err)
	}

	var nation Nation
	err = xml.Unmarshal(body, &nation)
	if err != nil {
		log.Fatal("Error parsing the XML response:", err)
	}

	time.Sleep(time.Second)

	return strings.Split(nation.Endorsements, ",")
}

func getTopViolators(client *http.Client, args Args, citizens []string, delendos []string) []Violator {
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

		response, err := client.Do(req)
		if err != nil {
			log.Fatal("Error making the API request:", err)
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal("Error reading the response body:", err)
		}

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

	violators := make([]Violator, 0, len(endorsements))

	for k, v := range endorsements {
		violators = append(violators, Violator{k, v})
	}

	sort.Slice(violators, func(i, j int) bool {
		return violators[i].over > violators[j].over
	})

	if len(violators) > 20 {
		return violators[:20]
	} else {
		return violators
	}

}

func outputResults(args Args, violators []Violator) {
	file, err := os.Create("output.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for _, v := range violators {
		file.WriteString(fmt.Sprintf("%s: %d\n", v.name, v.over))
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

	fmt.Println("Writing results to output.txt")
	outputResults(args, violators)
}
