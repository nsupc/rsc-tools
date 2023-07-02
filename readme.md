# rsc-tools

A collection of tools for Europeia's Regional Security Council.

- endorsers: Reports nations that are endorsing endocap violators. Optionally returns the violators that each nation is endorsing.
- tarters: An endotarting tool designed with Europeia's endocap system in mind. Gives the user a list of nations to endorse or unendorse.
- violators: Reports nations that are exceeding their endocap and by how much.

# Installation

1. Download the latest [release](https://github.com/nsupc/rsc-tools/releases) for your operating system.
2. Save the tool or tools of your choosing to a folder.
3. For detailed instructions on using each particular tool, see below.

# Usage (Windows)

## endorsers

1. Create a new text file in the same folder as the tool and call it 'endorsers.txt'.
2. Open the file in a text editor and copy the template from this repository's [example](https://github.com/nsupc/rsc-tools/blob/main/scripts/endorsers.txt).
3. Replace the text 'nation_name' with the name of your main nation.
4. Replace the text 'api_key' with a Google API key.
5. Add or remove excluded nations as necessary.
6. (Optional) Add the -v flag to the end of the command to enable verbose output.
7. Save the file in that same folder as 'endorsers.bat'.
8. Run 'endorsers.bat'.

### Configuration Options

The script contains a number of required and optional configuration options. These can be set by editing the file 'run.txt' in a text editor. The following options are available:

- -u: The name of your main nation. [Required]
  - Usage: -u upc
- -k: A Google API key. [Required]
  - Usage: -k 1234567890abcdef
- -d: The name of the delegate nation. [Optional]
  - Default: le_libertia
  - Usage: -d mancheseva_city
- -x: A nation to exclude from endocap checking. [Optional]
  - Usage: -x mancheseva_city -x pichtonia
- -r: The region to check. [Optional]
  - Default: europeia
  - Usage: -r the_north_pacific
- -b: The base endocap -- the endocap for nations that are not endorsing the delegate. [Optional]
  - Default: 10
  - Usage: -b 1
- -e: The standard endocap -- the endocap for nations that are not citizens but are endorsing the delegate. [Optional]
  - Default: 25
  - Usage: -e 10
- -c: The citizen endocap -- the endocap for nations that are citizens and are endorsing the delegate. [Optional]
  - Default: 50
  - Usage: -c 25
- -v: Enable verbose output. [Optional]
  - Usage: -v

## tarters

1. Create a new text file in the same folder as the tool and call it 'tarters.txt'.
2. Open the file in a text editor and copy the template from this repository's [example](https://github.com/nsupc/rsc-tools/blob/main/scripts/tarters.txt).
3. Replace the text 'nation_name' with the name of your main nation.
4. Replace the text 'api_key' with a Google API key.
5. Add or remove excluded nations as necessary.
6. Save the file as 'tarters.bat'.
7. Run 'tarters.bat'.

### Configuration Options

The script contains a number of required and optional configuration options. These can be set by editing the file 'run.txt' in a text editor. The following options are available:

- -u: The name of your main nation. [Required]
  - Usage: -u upc
- -k: A Google API key. [Required]
  - Usage: -k 1234567890abcdef
- -d: The name of the delegate nation. [Optional]
  - Default: le_libertia
  - Usage: -d mancheseva_city
- -x: A nation to exclude from endocap checking. [Optional]
  - Usage: -x mancheseva_city -x pichtonia
- -r: The region to check. [Optional]
  - Default: europeia
  - Usage: -r the_north_pacific
- -b: The base endocap -- the endocap for nations that are not endorsing the delegate. [Optional]
  - Default: 10
  - Usage: -b 1
- -e: The standard endocap -- the endocap for nations that are not citizens but are endorsing the delegate. [Optional]
  - Default: 25
  - Usage: -e 10
- -c: The citizen endocap -- the endocap for nations that are citizens and are endorsing the delegate. [Optional]
  - Default: 50
  - Usage: -c 25
- -l: The limit -- the number of endorsements below a nation's cap that qualify it for endotarting. [Optional]
  - Default: 5
  - Usage: -l 10

## violators

1. Create a new text file in the same folder as the tool and call it 'violators.txt'.
2. Open the file in a text editor and copy the template from this repository's [example](https://github.com/nsupc/rsc-tools/blob/main/scripts/violators.txt).
3. Replace the text 'nation_name' with the name of your main nation.
4. Replace the text 'api_key' with a Google API key.
5. Add or remove excluded nations as necessary.
6. Save the file as 'violators.bat'.
7. Run 'violators.bat'.

### Configuration Options

The script contains a number of required and optional configuration options. These can be set by editing the file 'run.txt' in a text editor. The following options are available:

- -u: The name of your main nation. [Required]
  - Usage: -u upc
- -k: A Google API key. [Required]
  - Usage: -k 1234567890abcdef
- -d: The name of the delegate nation. [Optional]
  - Default: le_libertia
  - Usage: -d mancheseva_city
- -x: A nation to exclude from endocap checking. [Optional]
  - Usage: -x mancheseva_city -x pichtonia
- -r: The region to check. [Optional]
  - Default: europeia
  - Usage: -r the_north_pacific
- -b: The base endocap -- the endocap for nations that are not endorsing the delegate. [Optional]
  - Default: 10
  - Usage: -b 1
- -e: The standard endocap -- the endocap for nations that are not citizens but are endorsing the delegate. [Optional]
  - Default: 25
  - Usage: -e 10
- -c: The citizen endocap -- the endocap for nations that are citizens and are endorsing the delegate. [Optional]
  - Default: 50
  - Usage: -c 25
