# Illumio ServiceNow Sync App

## Description
A command line utility that syncs ServiceNow CMDB with Illumio. For a routine sync, set to run on a cron job (.e.g, daily).

## Third-party Packages
Uses a Go package that provides the plumbing for accessing the Illumio APIs. The location of the package is here: https://stash.ilabs.io/users/brian.pitta/repos/illumioapi/. You can add this to your environment by running `go get stash.ilabs.io/scm/~brian.pitta/illumioapi.git`

## Parameters
*_-config_*: path to config file

## Configuration File
A sample config file (`config.toml`) is included in the repository. An annotated version is below:
```
[illumio]
fqdn = "pce-snc.illumioeval.com"
port = 8443 # Integer (no quotes).
org = 1 # Integer (no quotes). Value is 1 for on-prem PCE installations.
user = "api_1e51a1dda1b7b0c6e"
key = "46ae2d7c7c9301b375ec4bdb39f0d471ec9289768bbaa702ae070dc9934a5223"
match_field = "host_name" # Matches to ServniceNow match_field. Must either be "host_name" or "name".

[serviceNow]
table_url = "https://dev61812.service-now.com/cmdb_ci_server_list.do"
user =  "admin"
password = "U9%*#Mkf#8l7"
match_field = "host_name"

[labelMapping]
## To ignore a field (e.g., not sync the app field) comment the line.
app = "u_application"
enviornment = "used_for"
location = "u_data_center"
role = "u_type"

[logging]
log_only = false # True will make no changes to PCE. Log will reflect what will be updated if set to false.
log_directory = "" # Blank value stores log in same folder where tool is run.


### UNMANAGED WORKLOADS IS BETA ###

[unmanagedWorkloads]
## Set enable to true to create unmanaged workloads for any server in CMDB table that is not in PCE.
## Current option supports two methods:
# 1) using the cmdb_ci_server_list table to create an unmanaged workload with the ip_address and host_name fields
# 2) using the cmdb_ci_network_adapter table to create an unmanaged workload with an interface for each entry
enable = true 
table = "cmdb_ci_server_list" # "cmdb_ci_server_list.do" or "cmdb_ci_network_adapter"
 ```

## Logging
Each run will create a log file. The file naming convention is `IllumioServiceNowSync__YYYYMMDD_HHMMSS.log` with the time stamp based on the start of execution. Logging captures each entry and the API response status. An example output of the logging is below.

```
2018/06/28 09:52:13 INFO - Log only mode set to false
2018/06/28 00:41:52 INFO - db.prod.illumioeval.com - env label updated from PROD to Production
2018/06/28 00:41:52 INFO - db.prod.illumioeval.com - loc label updated from US to US-22
2018/06/28 09:52:17 INFO - Processed 179 servers; 175 not in PCE as workloads
```

## Executables
The binaries for Windows, Mac, and Linux are included in the `bin` folder.
