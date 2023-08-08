package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
)

var arguments struct {
	User     string `arg:"-u,--user,required" help:"Your main nation"`
	Region   string `arg:"-r,--region" help:"Target region" default:"europeia"`
	Count    int    `arg:"-c,--count" help:"Telegram batch size (1-8)" default:"8"`
	Template string `arg:"-t,--template" help:"Telegram template"`
}

type Args struct {
	User     string
	Region   string
	Count    int
	Template string
}

type Nation struct {
	ID           string `xml:"id,attr"`
	Endorsements string `xml:"ENDORSEMENTS"`
	Region       string `xml:"REGION"`
}

type Region struct {
	WANations string `xml:"UNNATIONS"`
}

func contains(s []string, e string) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}

func get_nation_details(client *http.Client, nation string) Nation {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.nationstates.net/cgi-bin/api.cgi?nation=%s&q=region+endorsements", nation), nil)
	if err != nil {
		log.Fatal("Error creating request:", err)

	}

	req.Header.Set("User-Agent", fmt.Sprintf("Nopers/1.0 (%s)", nation))

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
	var nat Nation
	err = xml.Unmarshal(body, &nat)
	if err != nil {
		log.Fatal("Error parsing the XML response:", err)
	}

	time.Sleep(time.Second)

	return nat
}

func get_wa_nations(client *http.Client, region string) []string {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.nationstates.net/cgi-bin/api.cgi?region=%s&q=wanations", region), nil)
	if err != nil {
		log.Fatal("Error creating request:", err)

	}

	req.Header.Set("User-Agent", fmt.Sprintf("Nopers/1.0 (%s)", arguments.User))

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
	var reg Region
	err = xml.Unmarshal(body, &reg)
	if err != nil {
		log.Fatal("Error parsing the XML response:", err)
	}

	time.Sleep(time.Second)

	return strings.Split(reg.WANations, ",")
}

func output_results(targets []string, template string, batchSize int) {
	// create output html file
	// break targets into batches of args.Count
	// create link to send telegram to each batch
	// append to list on html file

	f, err := os.Create("output.html")
	if err != nil {
		log.Fatal("Error creating output file:", err)
	}
	defer f.Close()

	_, err = f.WriteString("<html><head><title>Telegram Targets</title></head><body><h1>Telegram Targets</h1><ul>")
	if err != nil {
		log.Fatal("Error writing to output file:", err)
	}

	for i := 0; i < len(targets); i += batchSize {
		batchName := fmt.Sprintf("Batch %d", i/batchSize+1)

		if template != "" {
			_, err = f.WriteString(fmt.Sprintf("<li><a href=\"https://www.nationstates.net/page=compose_telegram?tgto=%s&message=%s\">%s</a></li>", strings.TrimRight(strings.Join(targets[i:i+batchSize], ","), ","), template, batchName))
		} else {
			_, err = f.WriteString(fmt.Sprintf("<li><a href=\"https://www.nationstates.net/page=compose_telegram?tgto=%s\">%s</a></li>", strings.Join(targets[i:i+batchSize], ","), batchName))
		}
	}
}

func main() {
	arg.MustParse(&arguments)

	if arguments.Region != "europeia" {
		arguments.Region = strings.ReplaceAll(strings.ToLower(arguments.Region), " ", "_")
	}

	if arguments.Count < 1 || arguments.Count > 8 {
		arguments.Count = 8
	}

	if arguments.Template != "" {
		arguments.Template = strings.ReplaceAll(arguments.Template, "%", "%25")
	}

	args := Args{
		User:     arguments.User,
		Region:   arguments.Region,
		Count:    arguments.Count,
		Template: arguments.Template,
	}

	client := &http.Client{}

	fmt.Println("Checking your endorsements")
	nation := get_nation_details(client, args.User)

	fmt.Println("Getting all WA nations")
	wa_nations := get_wa_nations(client, args.Region)

	var targets []string

	for _, n := range wa_nations {
		if !contains(strings.Split(nation.Endorsements, ","), n) {
			targets = append(targets, n)
		}
	}

	fmt.Println("Writing targets to output.html")
	output_results(targets, args.Template, args.Count)
}
