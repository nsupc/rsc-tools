from __future__ import print_function
from apiclient import discovery
from bs4 import BeautifulSoup as bs
from dataclasses import dataclass
from typing import List, Dict

import argparse
import httplib2
import os
import requests
import time

arghelp = {
    "u": "Script user, required for NS API compliance",
    "d": "Delegate nation, default = le_libertia",
    "r": "Region, default = europeia",
    "x": "Excluded nations -- VD, RSC, etc. Use once per nation (-x nation1 -x nation2...)",
    "b": "Base endocap, default = 10",
    "e": "Standard endocap, default = 25",
    "c": "Citizen endocap, default = 50",
    "k": "Google Sheets API key",
    "v": "Verbose output - will print all violators that a nation is endorsing"
}


class Args:
    user: str
    delegate: str
    region: str
    exclude: List[str]
    endocap: int
    citcap: int
    key: str
    verbose: bool


@dataclass
class Endorser:
    name: str
    percentage: float
    endorsing: List[str]


def parse_args() -> Args:
    parser = argparse.ArgumentParser(
        prog="Endorsers",
        description="A script to find nations that are endorsing endocap violators."
    )
    parser.add_argument('-u', '--user', help=arghelp['u'], type=str, required=True)
    parser.add_argument('-d', '--delegate', help=arghelp['d'], type=str, default="le_libertia", required=False)
    parser.add_argument('-r', '--region', help=arghelp['r'], type=str, default="europeia", required=False)
    parser.add_argument('-x', '--exclude', help=arghelp['x'], type=str, action='append', default=[], required=False)
    parser.add_argument('-b', '--basecap', help=arghelp['b'], type=int, default=10, required=False)
    parser.add_argument('-e', '--endocap', help=arghelp['e'], type=int, default=25, required=False)
    parser.add_argument('-c', '--citcap', help=arghelp['c'], type=int, default=50, required=False)
    parser.add_argument('-k', '--key', help=arghelp['k'], type=str, required=True)
    parser.add_argument('-v', '--verbose', help=arghelp['v'], action='store_true', required=False)

    args = Args()

    parser.parse_args(namespace=args)

    return args


def get_citizen_nations(args: Args) -> List[str]:
    discoveryUrl ='https://sheets.googleapis.com/$discovery/rest?version=v4'
    service = discovery.build(
        'sheets',
        'v4',
        http=httplib2.Http(),
        discoveryServiceUrl=discoveryUrl,
        developerKey=args.key)

    spreadsheetId = '1Zi2HtQuykoWV2P36B61J_eBnhSgj3VyDWFUbtbYWyTo'
    rangeName = 'Citizens!C1:C'
    result = service.spreadsheets().values().get(
        spreadsheetId=spreadsheetId, range=rangeName).execute()
    values = result.get('values', [])

    return [citizen[0] for citizen in values[1:]]


def get_delegate_endorsements(user: str, delegate: str) -> List[str]:
    headers = {
        'User-Agent': user
    }

    delegate_url = "https://www.nationstates.net/cgi-bin/api.cgi?nation={}&q=endorsements"

    endorsements = bs(requests.get(delegate_url.format(delegate), headers=headers).text, "xml").find("ENDORSEMENTS").text.split(",")
    time.sleep(1)

    return endorsements


def get_top_violators(args: Args, citizens: List[str], endorsements: List[str]) -> List[str]:
    # dict of nations that are violating their endocap band, and how many endorsements they have over their cap
    violators: Dict[str, int] = {}

    headers = {
        'User-Agent': args.user
    }

    endorsements_url = "https://www.nationstates.net/cgi-bin/api.cgi?region={}&q=censusranks;scale=66;start={}"

    offset = 0
    search = True
    while search:
        print("Checking nations {} through {}...".format((20 * offset) + 1, (20 * offset) + 20))

        nations = bs(requests.get(endorsements_url.format(args.region, (20 * offset) + 1), headers=headers).text, "xml").find_all("NATION")

        for nation in nations:
            name = nation.NAME.text
            numendos = int(nation.SCORE.text)

            if int(numendos) == 0:
                search = False
                break
            elif name in args.exclude:
                continue
            elif name not in endorsements and numendos > args.basecap:
                violators[name] = numendos - args.basecap
            elif name in endorsements and name not in citizens and numendos > args.endocap:
                violators[name] = numendos - args.endocap
            elif name in endorsements and name in citizens and numendos > args.citcap:
                violators[name] = numendos - args.citcap

        offset += 1
        time.sleep(1)

    sorted_violators = {k: v for k, v in sorted(violators.items(), key=lambda item: item[1], reverse=True)}

    return list(sorted_violators.keys())[:20]


def get_violator_endorsements(user: str, violators: List[str]) -> List[Endorser]:
    endorsers: Dict[str, Endorser] = {}
    percent = 100 / len(violators)

    headers = {
        'User-Agent': user
    }

    violator_url = "https://www.nationstates.net/cgi-bin/api.cgi?nation={}&q=endorsements"

    for violator in violators:
        print("Checking endorsements for {}...".format(violator))
        endorsements = bs(requests.get(violator_url.format(violator), headers=headers).text, "xml").find("ENDORSEMENTS").text.split(",")

        for endorser in endorsements:
            if endorser not in endorsers:
                endorsers[endorser] = Endorser(endorser, 0, [])
            endorsers[endorser].endorsing.append(violator)
            endorsers[endorser].percentage += percent
        
        time.sleep(1)

    sorted_endorsements = sorted(endorsers.values(), key=lambda endorser: endorser.percentage, reverse=True)

    return sorted_endorsements


def output_results(endorsers: List[Endorser], verbose: bool) -> None:
    if verbose:
        print("Writing output to output.txt")
        with open(f"{os.path.dirname(os.path.realpath(__file__))}/output.txt", "w") as out_file:
            for endorser in endorsers:
                out_file.write(f"{endorser.name}: {int(endorser.percentage)}%\n{','.join(endorser.endorsing)}\n\n")
    else:
        print("Writing output to output.csv")
        with open(f"{os.path.dirname(os.path.realpath(__file__))}/output.csv", "w") as out_file:
            for endorser in endorsers:
                out_file.write(f"{endorser.name},{int(endorser.percentage)}\n")
    

def main():
    args = parse_args()

    cit_nations = get_citizen_nations(args)

    endorsements = get_delegate_endorsements(args.user, args.delegate)

    violators = get_top_violators(args, cit_nations, endorsements)

    if not violators:
        print("No endocap violators.")
        return
    
    endorsers = get_violator_endorsements(args.user, violators)

    output_results(endorsers, args.verbose)
    

if __name__ == '__main__':
    main()
