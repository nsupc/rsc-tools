# rsc-tools
A collection (tbd) of tools for Europeia's Regional Security Council. 
# Installation
1. Click the green 'Code' button on the top right of this page, and select 'Download ZIP'.
2. Extract the ZIP file to a folder of your choice.
3. Run the file 'install.bat'.
# Usage (Windows)
## endorsers
1. Open the file 'run.txt' in a text editor.
2. Replace the text '<nation_name>' with the name of your main nation.
3. Replace the text '<api_key>' with a Google API key.
4. Add or remove excluded nations as necessary.
5. (Optional) Add the -v flag to the end of the command to enable verbose output.
6. Save the file as 'run.bat'.
7. Run 'run.bat'.
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