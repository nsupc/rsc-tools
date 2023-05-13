use std::collections::HashMap;
use reqwest::blocking::Client;
use serde::Deserialize;
use serde_xml_rs::from_str;
use std::thread;
use std::time::Duration;
use structopt::StructOpt;
use thiserror::Error;

/* 

A script where you can insert a set of, say, 20 nations (these would be the biggest endocap violators) and then show which players are endorsing the highest percentage of that set of 20 nations
Do you want to have to put the names in manually? I can make this automatic in a similar way to the script that took us forever to set up last week does. Otherwise doable

A script that lists the number of endocap violators endorsed by each nation, and ranks them by number
I can tweak the output of the script I gave you last week to do this, yeah. Do you want to keep the list of endocap violators as well or do you just want the number?

A script where you can insert 1 nation's name and then specifically show the endocap violators they are endorsing for only that 1 nation
Less sure about the utility of this one, the script I gave you last week does this for every nation in one go. I would just run through that list and check for whatever nation you're looking for. 

For #2, I think having both a numbers-only and a more in-depth one with names would be great if possible
For #3, that is fair. I guess I was wondering if something roughly similar to our old WA endotarting tool where you could just enter a nation name and the specific, most up-to-date results would instantly pop up. But if that's not possible, I can work with what I have

 */

#[derive(Debug, StructOpt)]
#[structopt(name = "endorsers", about = "A tool to find nations endorsing endocap violators")]
struct Opt {
    /// User-Agent, for identifying the script user to NS
    #[structopt(name="user", short, long)]
    user_agent: String,
    /// API rate limit, the number of requests that the script will make in a 30 second period (max 45)
    #[structopt(short="l", long, default_value = "30")]
    rate_limit: usize,
    /// Target endorsement cap
    #[structopt(short, long, default_value = "25")]
    endocap: usize,
    /// Target region, leave blank for Europeia
    #[structopt(short, long, default_value = "europeia")]
    region: String,
    /// Nations that are allowed to exceed the endocap. Can be passed as a list (i.e. -i le_libertia pland_adanna) or multiple times (i.e. -i le_libertia -i pland_adanna)
    #[structopt(short, long)]
    ignore: Vec<String>,
}

#[derive(Debug, Error)]
enum EndorseError {
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),
    #[error("XML parsing error -- most likely an invalid region")]
    Xml(#[from] serde_xml_rs::Error),
}

#[derive(Debug, Deserialize, PartialEq)]
struct Ranks {
    #[serde(rename = "CENSUSRANKS")]
    census_ranks: CensusRanks,
}

#[derive(Debug, Deserialize, PartialEq)]
struct CensusRanks {
    #[serde(rename = "NATIONS")]
    nations: Nations,
}

#[derive(Debug, Deserialize, PartialEq)]
struct Nations {
    #[serde(rename = "NATION")]
    nations: Vec<Nation>,
}

#[derive(Debug, Deserialize, PartialEq)]
struct Nation {
    #[serde(rename = "NAME")]
    name: String,
    #[serde(rename = "RANK")]
    rank: u32,
    #[serde(rename = "SCORE")]
    score: u32,
}

fn request(client: &Client, url: &str, limit: &u64) -> Result<reqwest::blocking::Response, reqwest::Error>{
    let response = client.get(url)
        .send()?;

    thread::sleep(Duration::from_millis((30 / limit) * 1000));

    Ok(response)
}

fn get_top_violators(client: &Client, opt: &Opt, limit: u64) -> Result<Vec<String>, EndorseError> {
    let mut violators: Vec<String> = Vec::new();

    let mut offset = 0;

    let mut reached_zero = false;

    while violators.len() < 20 && !reached_zero {
        println!("Checking nations {} to {}", 1 + offset, 20 + offset);

        let nations: Ranks = from_str(&request(&client, &format!("https://www.nationstates.net/cgi-bin/api.cgi?region={}&q=censusranks;scale=66;start={}", opt.region,1 + offset), &limit)?.text()?)?;

        for nation in nations.census_ranks.nations.nations {
            if violators.len() >= 20 {
                break;
            }

            if nation.score == 0 {
                reached_zero = true;
                break;
            }

            if nation.score > opt.endocap as u32 && !opt.ignore.contains(&nation.name) {
                violators.push(nation.name);
            }
        }

        offset += 20;
    }

    Ok(violators)
}

fn get_endorsements(client: &Client, limit: u64, violators: Vec<String>) -> Result<HashMap<String, u32>, EndorseError> {
    let mut endorsers: HashMap<String, u32> = HashMap::new();

    Ok(endorsers)
}

fn main() {
    let opt = Opt::from_args();

    let limit: u64 = match opt.rate_limit {
        0 => 1,
        1..=45 => opt.rate_limit as u64,
        _ => 30
    };

    let client = Client::builder()
        .user_agent(format!("Endocap violator endorsers, developed by nation=UPC, used by {}", opt.user_agent))
        .build()
        .unwrap();

    let violators = match get_top_violators(&client, &opt, limit) {
        Ok(violators) => violators,
        Err(e) => {
            println!("Could not get violators: {}", e);
            return;
        }
    };

    if violators.len() == 0 {
        println!("No violators found");
        return;
    }

    println!("{:#?}", violators);
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn deserialize_census() {
        let test_xml = r#"
            <REGION id="europeia">
                <CENSUSRANKS id="66">
                    <NATIONS>
                        <NATION>
                            <NAME>le_libertia</NAME>
                            <RANK>1</RANK>
                            <SCORE>193</SCORE>
                        </NATION>
                        <NATION>
                            <NAME>mancheseva_city</NAME>
                            <RANK>2</RANK>
                            <SCORE>145</SCORE>
                        </NATION>
                    </NATIONS>
                </CENSUSRANKS>
            </REGION>
        "#;

        let parsed: Ranks = from_str(test_xml).expect("Failed to parse XML");

        let expected = Ranks {
            census_ranks: CensusRanks {
                nations: Nations {
                    nations: vec![
                        Nation {
                            name: "le_libertia".to_string(),
                            rank: 1,
                            score: 193,
                        },
                        Nation {
                            name: "mancheseva_city".to_string(),
                            rank: 2,
                            score: 145,
                        },
                    ],
                },
            },
        };

        assert_eq!(parsed, expected);
    }
}