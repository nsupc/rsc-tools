package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/codeclysm/extract/v3"
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
	Limit    int      `arg:"-l,--limit" help:"Number of endorsements under cap to qualify a nation for endorsing" default:"5"`
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
	Limit    int
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

type EndoRegion struct {
	Name    string `xml:"id,attr"`
	Nations string `xml:"UNNATIONS"`
}

type Nations struct {
	XMLName xml.Name     `xml:"NATIONS"`
	Nations []DumpNation `xml:"NATION"`
}

type DumpNation struct {
	Name         string `xml:"NAME"`
	Endorsements string `xml:"ENDORSEMENTS"`
}

type Targets struct {
	Endorse   []string
	Unendorse []string
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

	return strings.Split(nation.Endorsements, ",")
}

func getEndorsementNumbers(client *http.Client, user string, region string) map[string]int {
	endorsements := make(map[string]int)

	offset := 1
outer:
	for {

		fmt.Printf("Checking nations %v through %v\n", offset, offset+20)

		req, err := http.NewRequest("GET", fmt.Sprintf("https://www.nationstates.net/cgi-bin/api.cgi?region=%s&q=censusranks;scale=66;start=%v", region, offset), nil)
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

		var region Region
		err = xml.Unmarshal(body, &region)
		if err != nil {
			log.Fatal("Error parsing the XML response:", err)
		}

		for _, nation := range region.CensusRanks.Nations {
			if nation.Score > 0 {
				endorsements[nation.Name] = nation.Score
			} else {
				break outer
			}
		}

		offset += 20
		time.Sleep(time.Second)
	}

	return addAllWAs(client, user, region, endorsements)
}

func addAllWAs(client *http.Client, user string, region string, nations map[string]int) map[string]int {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.nationstates.net/cgi-bin/api.cgi?region=%s&q=wanations", region), nil)
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

	var reg EndoRegion
	err = xml.Unmarshal(body, &region)
	if err != nil {
		log.Fatal("Error parsing the XML response:", err)
	}

	for _, nation := range strings.Split(reg.Nations, ",") {
		if _, ok := nations[nation]; !ok {
			if nation != "" {
				nations[nation] = 0
			}
		}
	}

	return nations
}

func getDump(client *http.Client, user string) {
	req, err := http.NewRequest("GET", "https://www.nationstates.net/pages/nations.xml.gz", nil)
	if err != nil {
		log.Fatal("Error creating request:", err)

	}

	req.Header.Set("User-Agent", fmt.Sprintf("Tarters/1.0 (%s)", user))

	response, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making the API request:", err)
	}
	defer response.Body.Close()

	extract.Archive(context.TODO(), response.Body, "nations.xml", nil)
}

func getNationsEndorseBy(target string) []string {
	endorsing := []string{}

	xmlData, err := os.ReadFile("nations.xml")
	if err != nil {
		log.Fatal(err)
	}

	var nations Nations
	err = xml.Unmarshal(xmlData, &nations)
	if err != nil {
		log.Fatal(err)
	}

	for _, nation := range nations.Nations {
		endorsements := strings.Split(nation.Endorsements, ",")
		for _, endorsement := range endorsements {
			if endorsement == target {
				endorsing = append(endorsing, strings.ToLower(strings.ReplaceAll(nation.Name, " ", "_")))
				break
			}
		}

	}

	return endorsing
}

func DeleteDump() {
	err := os.Remove("nations.xml")
	if err != nil {
		log.Fatal(err)
	}
}

func getTargets(args Args, was map[string]int, citizens []string, delendos []string, self_endorsing []string) Targets {
	targets := Targets{}

	for nation, endorsements := range was {
		if contains(args.Excluded, nation) || nation == args.Delegate {
			continue
		}

		if contains(self_endorsing, nation) {
			if endorsements > args.Citizen {
				targets.Unendorse = append(targets.Unendorse, nation)
			} else if endorsements > args.Standard && !contains(citizens, nation) {
				targets.Unendorse = append(targets.Unendorse, nation)
			} else if endorsements > args.Base && !contains(delendos, nation) {
				targets.Unendorse = append(targets.Unendorse, nation)
			} else {
				continue
			}
		}

		if !contains(self_endorsing, nation) {
			if contains(citizens, nation) && contains(delendos, nation) {
				if endorsements < args.Citizen && args.Citizen-endorsements > args.Limit {
					targets.Endorse = append(targets.Endorse, nation)
				}
			} else if contains(delendos, nation) {
				if endorsements < args.Standard && args.Standard-endorsements > args.Limit {
					targets.Endorse = append(targets.Endorse, nation)
				}
			} else {
				if endorsements < args.Base && args.Base-endorsements > args.Limit {
					targets.Endorse = append(targets.Endorse, nation)
				}
			}
		}
	}

	return targets
}

func outputTargets(targets Targets) {
	// write targets to output.html
	f, err := os.Create("output.html")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString("<html><head><title>Targets</title></head><body><h1>Endorse</h1><ul>")
	if err != nil {
		log.Fatal(err)
	}

	for _, nation := range targets.Endorse {
		_, err = f.WriteString(fmt.Sprintf("<li><a href='https://www.nationstates.net/nation=%s#composebutton'>%s</a></li>\n", nation, nation))
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = f.WriteString("</ul><h1>Unendorse</h1><ul>")
	if err != nil {
		log.Fatal(err)
	}

	for _, nation := range targets.Unendorse {
		_, err = f.WriteString(fmt.Sprintf("<li><a href='https://www.nationstates.net/nation=%s#composebutton'>%s</a></li>\n", nation, nation))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	arg.MustParse(&arguments)

	args := Args{
		User:     strings.ToLower(strings.ReplaceAll(arguments.User, " ", "_")),
		Key:      arguments.Key,
		Delegate: strings.ToLower(strings.ReplaceAll(arguments.Delegate, " ", "_")),
		Region:   strings.ToLower(strings.ReplaceAll(arguments.Region, " ", "_")),
		Excluded: arguments.Excluded,
		Base:     arguments.Base,
		Standard: arguments.Standard,
		Citizen:  arguments.Citizen,
		Limit:    arguments.Limit,
	}

	fmt.Println("Getting citizen nations")
	citizenNations := getCitizenNations(args.Key)

	client := &http.Client{}

	fmt.Println("Getting delegate endorsements")
	delegateEndorsements := getDelegateEndorsements(client, args.User, args.Delegate)

	fmt.Println("Getting nations and endorsements")
	endorsements := getEndorsementNumbers(client, args.User, args.Region)

	fmt.Println("Getting nations that you are endorsing (this uses the daily dump and may take a minute to process)")
	getDump(client, args.User)

	endorsing := getNationsEndorseBy(strings.ToLower(strings.ReplaceAll(args.User, " ", "_")))

	DeleteDump()

	fmt.Println("Getting targets")
	targets := getTargets(
		args,
		endorsements,
		citizenNations,
		delegateEndorsements,
		endorsing,
	)

	fmt.Println("Writing targets to output.html")
	outputTargets(targets)

}
