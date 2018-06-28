# Illumio ServiceNow Sync App

## Description
A command line utility that imports labels into the Illumio PCE from a CSV file.

## Third-party Packages
Uses a Go package that provides the plumbing for accessing the Illumio APIs. The location of the package is here: https://stash.ilabs.io/users/brian.pitta/repos/illumioapi/. You can add this to your environment by running `go get stash.ilabs.io/scm/~brian.pitta/illumioapi.git`

## Parameters
*_-config_*: path to JSON config file

## Configuration File
A sample config file (`sampleConfig.json`) is included in the repository. An annotated version is below:
```
{
    "illumioFQDN": "pce-snc.illumioeval.com",
    "illumioPort": 8443,
    "illumioOrg": 1,
    "illumioUser": "api_1e51a1dda1b7b0c6e",
    "illumioKey":"46ae2d7c7c9301b375ec4bdb39f0d471ec9289768bbaa702ae070dc9934a5223",
    "illumioMatchField": "host_name",
    "serviceNowURL": "https://dev61812.service-now.com/cmdb_ci_server_list.do",
    "serviceNowUser": "admin",
    "serviceNowPwd": "U9%*#Mkf#8l7",
    "serviceNowMatchField":"host_name",
    "appField":"u_application",
    "envField":"used_for",
    "locField": "u_data_center",
    "roleField": "u_type",
    "loggingOnly": false
}
 ```
**Notes:**
* `illumioMatchField` must either be `name` or `host_name`. This is what is used to match to the `serviceNowMatchField`
* `appField`, `envField`, `locField`, and `roleField` are the ServiceNow fields that map to their respective Illumio labels.
* Set `loggingOnly` to `true` if you want to run this tool to see what changes *would* happen - no updates will happen in the PCE.

## Logging
Each run will create a log file. The file naming convention is `"IllumioServiceNowSync__YYYYMMDD_HHMMSS.log` with the time stamp based on the start of execution. Logging captures each entry and the API response status. An example output of the logging is below.

```
2018/06/28 00:41:52 INFO - db.prod.illumioeval.com - env label updated from PROD to Production
2018/06/28 00:41:52 INFO - db.prod.illumioeval.com - loc label updated from US to US-22
```

## Executables
The binaries for Windows, Mac, and Linux are included in the `bin` folder.
