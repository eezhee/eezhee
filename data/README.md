# Provider Data

Each provider has their own list of regions and VM sizes.  There is very little standardization between providers.  Eezhee solves this by having standard names and mapping rules.  The data in this directory contains the provider details and how they map to Eezhee values.

## TODO

- read each of the 3 provider files
  - filter data to just have items that are mapped
  - produce a final data set that go can use
- need to be able to list options to a user (size and regions)
- need to be able to convert whats in config/defaults to provider version

## Adding a New Provider

There are several steps to adding a new provider to Eezhee. One part is to write the code to create and manage VMs. The other is to map the provider sizes and regions to the standard format that Eezhee uses.

### Adding Code

The provider code is in `pkg`.  The code needs to support the `code.VMManager` interface. This has all the function methods that Eezhee uses.  You can use one of the existing providers as a template.

## Adding Data

There are two types of provider data. First there is the list of VM sizes and regions that it supports.  Then there is mapping data that converts the provider data to the common Eezhee format.

Start by creating a script to download the provider data in JSON format.  There are bash scripts in `data/raw` that can be used as templates.

Then create a mapping JSON file.  This will map VM sizes and regions from the provider naming to the standard Eezhee naming.

Once you the data, you need to add code to `data\process_files` to load and convert the data. The code needs to support the `ProviderImporter` interface.  Use one of the existing importers as a template.

## Updating Provider Data

### Getting Latest Data

You can get the latest provider data by running the `fetch_cloud_details.sh` script in the `raw` subdirectory.  This will generate a number of JSON files.  Note each file is in the provider format and needs to be converted into the format that Eezhee uses.

### List Ubuntu Versions

To see what Ubuntu images a provider has available, run the `process_files -listUbuntuImages` with the appropriate provider flag. As an example, below is the output from Linode

```bash
./process_files -linode -listUbuntuImages
processing Linode files
  images file has 39 images
    ID: linode/ubuntu16.04lts  Label: Ubuntu 16.04 LTS    
    ID: linode/ubuntu18.04  Label: Ubuntu 18.04 LTS    
    ID: linode/ubuntu20.04  Label: Ubuntu 20.04 LTS    
    ID: linode/ubuntu20.10  Label: Ubuntu 20.10    
```

Update the mapping file with the appropriate ID for the given Ubuntu verison you want Eezhee to use

### Updating Sizes and Regions

Then check the mapping file and see if any of the mappings need to edited.  If so, you will need to run `data/process_files` to generate updates data for Eezhee.  Once you have done that, you can rebuild Eezhee and it should have the updated data
