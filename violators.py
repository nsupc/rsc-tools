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


@dataclass
class Violator:
    name: str
    over_by: int


def parse_args() -> Args:
    parser = argparse.ArgumentParser(
        prog="Violators",
        description="A script to find nations that are violating an endocap."
    )
    parser.add_argument('-u', '--user', help=arghelp['u'], type=str, required=True)
    parser.add_argument('-d', '--delegate', help=arghelp['d'], type=str, default="le_libertia", required=False)
    parser.add_argument('-r', '--region', help=arghelp['r'], type=str, default="europeia", required=False)
    parser.add_argument('-x', '--exclude', help=arghelp['x'], type=str, action='append', default=[], required=False)
    parser.add_argument('-b', '--basecap', help=arghelp['b'], type=int, default=10, required=False)
    parser.add_argument('-e', '--endocap', help=arghelp['e'], type=int, default=25, required=False)
    parser.add_argument('-c', '--citcap', help=arghelp['c'], type=int, default=50, required=False)
    parser.add_argument('-k', '--key', help=arghelp['k'], type=str, required=True)

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
    violators: List[Violator] = []

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
            elif name == args.delegate or name in args.exclude:
                continue
            elif name not in endorsements and numendos > args.basecap:
                violators.append(Violator(name, numendos - args.basecap))
            elif name in endorsements and name not in citizens and numendos > args.endocap:
                violators.append(Violator(name, numendos - args.endocap))
            elif name in endorsements and name in citizens and numendos > args.citcap:
                violators.append(Violator(name, numendos - args.citcap))

        offset += 1
        time.sleep(1)

    violators.sort(key=lambda x: x.over_by, reverse=True)

    return violators[:20]


def output_results(violators: List[Violator]) -> None:
    print("Writing output to output.txt")
    with open(f"{os.path.dirname(os.path.realpath(__file__))}/output.txt", "w") as out_file:
        for violator in violators:
            out_file.write(f"{violator.name}: {violator.over_by}\n")
    

def main():
    args = parse_args()

    cit_nations = get_citizen_nations(args)

    endorsements = get_delegate_endorsements(args.user, args.delegate)

    violators = get_top_violators(args, cit_nations, endorsements)

    if not violators:
        print("No endocap violators.")
        return

    output_results(violators)
    

if __name__ == '__main__':
    main()
