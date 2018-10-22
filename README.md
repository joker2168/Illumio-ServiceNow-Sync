# Illumio ServiceNow Sync App

## Description
A command line utility that syncs ServiceNow CMDB with Illumio. For a routine sync, set to run on a cron job (.e.g, daily).

## Third-party Packages
Uses a Go package that provides the plumbing for accessing the Illumio APIs. The location of the package is here: https://stash.ilabs.io/users/brian.pitta/repos/illumioapi/. You can add this to your environment by running `go get stash.ilabs.io/scm/~brian.pitta/illumioapi.git`

## Parameters
*_-config_*: path to config file (default: config.toml)
*_-v_*: runs verbose logging for debugging

## Configuration File
A sample config file (`config.toml`) is included in the repository. An annotated version is below:
```
[illumio]
fqdn = "demo4.illum.io"
port = 443 # Integer (no quotes).
org = 14 # Integer (no quotes). Value is 1 for on-prem PCE installations.
user = "api_1994a3a47d8e4bbb1"
key = "c15c446e8067c348f6ba66c3c9ee1e254e887e607e62fa0ea52f6ae3841f8a79"
match_field = "host_name" # Matches to ServniceNow match_field. Must either be "host_name" or "name".

[serviceNow]
table_url = "https://dev68954.service-now.com/cmdb_ci_server_list.do"
user =  "admin"
password = "UGk6gsKaK6jWa34"
match_field = "host_name"

[labelMapping]
## To ignore a field (e.g., not sync the app field) comment the line.
app = "u_application"
enviornment = "u_environment"
location = "u_data_center"
role = "u_type"

[logging]
log_only = false # True will make no changes to PCE. Log will reflect what will be updated if set to false.
log_directory = "" # Blank value stores log in same folder where tool is run.

### DO NOT USE THIS PIECE - KEEP SET TO FALSE ###
[unmanagedWorkloads]
enable = false 
table = "cmdb_ci_server_list" # "cmdb_ci_server_list.do" or "cmdb_ci_network_adapter"
 ```

## Logging
Each run will create a log file. The file naming convention is `IllumioServiceNowSync__YYYYMMDD_HHMMSS.log` with the time stamp based on the start of execution. Logging captures each entry. An example output of the logging is below. Using the *_-v_* flag provides more verbose logging for debugging.

```
2018/06/28 09:52:13 INFO - Log only mode set to false
2018/06/28 00:41:52 INFO - db.prod.illumioeval.com - env label updated from PROD to Production
2018/06/28 00:41:52 INFO - db.prod.illumioeval.com - loc label updated from US to US-22
2018/06/28 09:52:17 INFO - Processed 179 servers; 175 not in PCE as workloads
```

## Executables
The binaries for Windows, Mac, and Linux are included in the `bin` folder.
