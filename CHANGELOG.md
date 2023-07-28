# 0.1.16 Release

## New Features
 - Support for MCR contract term.
 - Oracle partner support and OCI VXC connectivity. Credit @aszynkow

# 0.1.15 Release

## New Features
 - Support for extracting AWS Hosted Connection ID's.

# 0.1.14 Release

## New Features
 - Support for filtering Partner Megaports by Diversity Zone.

# 0.1.13 Release

## New Features
 - Support for MCR Prefix Filter Lists.
 - Support for BGP Peer Filters and BGP Prefix Filters on VXC's.

# 0.1.12 Release

## New Features
 - Support for manually supplying peering values for Azure VXC's.

# 0.1.11 Release

## New Features
 - Add GetPorts functionality. Credit @daniel-noland

# 0.1.10 Release

## Changes
 - Move MarshallMcrAEndConfig to megaport/terraform-provider-megaport. Credit @daniel-noland

# 0.1.9 Release

## New Features
 - Add optional connection name attribute for AWS connections. Credit @ngarratt
 - Add import support for pre-existing AWS connections. Credit @ngarratt

# 0.1.8 Release

## New Features
 - PartnerConfigInterface (MCR A end configuration in VXC) handles static ip routes

# 0.1.7 Release

## New Features
 - Add NAT support for MCR VXC Connections.

# 0.1.6 Release

## New Features
 - Migrate BGP Connection support from AWS-VXC to all VXC Connections.
 
## Changes
 - VXC type resources now expect aEnd and bEnd configuration objects to represent these configurations.

# 0.1.5 Release

## New Features
 - Add BGP Connection support for AWS VXC Connections
 
## Changes
 - Rewrite and rename BuyAWSHostedVIF to BuyAWSVXC to handle new parameters and use cases

# 0.1.4 Release

## Changes
 - Fix marshalling issue with VirtualRouter in VXCResource.

# 0.1.3-beta Release

## New Features
 - Optionally Specify the Google Interconnect Location when creating a GCP Connection. Credit @kdw174

# 0.1.2-beta Release

Welcome to the first release of the `megaportgo` library!

## New Features
 - Support for the following Megaport Products has been added:
   - Ports (Single and LAG).
   - VXC.
   - MCR2.
   - AWS Hosted VIF and Hosted Connection.
   - Google Cloud Interconnect.
   - Azure ExpressRoute
   
 - The following lookup functionality is available:
   - Locations.
   - Partners Ports.

## Changes
 - (added in 0.1.2) Changed the `WaitForPortProvisioning` function so that it considers
   "LIVE" or "CONFIGURED" as an active status.
   
## Notes
This product is a `beta` release, please test all your changes in the
Megaport staging environment before running on Production. Details can
be found in the documentation. If you run into any issues, please log 
an issue on GitHub issues.
