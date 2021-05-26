# Provider Data

Each provider has their own list of regions and VM sizes.  There is very little standardization between providers.  Eezhee solves this by having standard names and mapping rules.  The data in this directory contains the provider details and how they map to Eezhee values.

## Getting Data

You can get the latest provider data by running the `fetch_cloud_details.sh` script in the `raw` subdirectory.  This will generate a number of JSON files.  Note each file is in the provider format and needs to be converted into the format that Eezhee uses.

## Processing the data
