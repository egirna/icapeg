<h1 align="center">Go ICAP Server</h1>
<p align="center">
    <em> k8-go-icap server.</em>
</p>

<p align="center">
    <a href="https://github.com/k8-proxy/go-icap-server/actions/workflows/build.yml">
        <img src="https://github.com/k8-proxy/go-icap-server/actions/workflows/build.yml/badge.svg"/>
    </a>
    <a href="https://codecov.io/gh/k8-proxy/go-icap-server">
        <img src="https://codecov.io/gh/k8-proxy/go-icap-server/branch/main/graph/badge.svg"/>
    </a>	    
    <a href="https://goreportcard.com/report/github.com/k8-proxy/go-icap-server">
      <img src="https://goreportcard.com/badge/k8-proxy/go-icap-server" alt="Go Report Card">
    </a>
	<a href="https://github.com/k8-proxy/go-icap-server/pulls">
        <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat" alt="Contributions welcome">
    </a>
    <a href="https://opensource.org/licenses/Apache-2.0">
        <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="Apache License, Version 2.0">
    </a>
</p>

# ICAPeg

Open Source multi-vendor ICAP server

Scan files requested via a proxy server using ICAPeg ICAP server, ICAPeg is an ICAP server connecting web proxies with API based scanning services and more soon!. ICAPeg currently supports [Glasswall]() , [VirusTotal](https://www.virustotal.com/gui/home/upload),[VMRAY](https://www.vmray.com/) , [MetaDefender](https://metadefender.opswat.com/?lang=en) & [Clamav](https://www.clamav.net/)  for scanning the files following the ICAP protocol. If you don't know about the ICAP protocol, here is a bit about it:

## What is ICAP?

**ICAP** stands for **Internet Content Adaptation Protocol**. If a **content** (for example: file) you've requested over the internet
to download or whatever, needs **adaptation**(some kind of modification or analysis), the proxy server sends the content to the ICAP server for adaptation and after performing the required tasks on the content, the ICAP server sends it back to the proxy server so that it may return the adapted content back to the destination. This can occur both during request and response.

To know more about the ICAP protocol, [check this out](https://tools.ietf.org/html/rfc3507).

## Prerequisites

Before starting to play with ICAPeg, make sure you have the following things in your machine:

1. **Golang**(latest enough to be able to use go mod)

     You should install at least **go version 1.16**.

     You can get how to **install Golang** from [here](https://go.dev/doc/install).

2. A **proxy** server

​		You should configure a Proxy server with ICAPeg, you can get an example from [here](https://www.egirna.com/post/configure-squid-4-17-with-icap-ssl).

3. Clone the **ICAPeg** repository

  ```bash
  $ git clone https://github.com/egirna/icapeg.git
  ```

4. Run the project with the default configuration in **config.toml** file 

  1. Run the following command in the directory where you've installed **ICAPeg**

```bash
$ cd ~/ICAPeg
```

  2. Run the following command to start **ICAPeg**

```bash
$ go run main.go
```

You should see something like, ```ICAP server is running on localhost:1344 ...```. This tells you the ICAP server is up and running.

## Configuration

​	You can change the default configuration file whatever you want to customize the application.

- ### Mapping a variable of config.toml file with an environment variable

  This feeature is supported for **strings, int, bool, time.duration and string slices** only **(every type used in this project)**.

  Let's have an example to explain how to map, assume that there is an env variable called LOG_LEVEL and you want to assign LOG_LEVEL value to app.log_level. You should change the value of (log_level= "debug") to (log_level= "$_LOG_LEVEL").

  > **Note**: before you use this feature please make sure that the env variable that you want to use is globally in your machine and not just exported in a local session.

  - ### Config.toml file sections

    - #### **[app] section** 

      - This section includes the configuration variables that are responsible for the application overall like the port number and current services integrated with **ICAPeg**.

        ```toml
        [app]
        log_level = "debug" # the log levels for tha app, available values: info-->logging the overall progress of the app, debug --> log everything including errors, error --> log infos and just errors
        write_logs_to_console= false
        log_flush_duration = 2
        port = 1344
        services= ["glasswall" , "echo", "clamav"]
        verify_server_cert=false
        ```

        - **log_level**

          The log levels for the app, possible values:

          - **info**: Logging the overall progress of the app.
          - **debug**: Log everything including errors.
          - **error**: Log info and just errors.

        - **write_logs_to_console**

          It's used to enable writing logs to **ICAPeg** console window or not, possible values:

          - **true**: Writing logs to **ICAPeg** console window and **log.txt** file.
          - **false**: Writing logs to **ICAPeg** **log.txt** file only.

        - **log_flush_duration**

          Deleting logs that were written in **ICAPeg** log.txt from **n** hours (2 hours for example), possible values:

          - Any integer value.

        - **port**

          The port number that **ICAPeg** runs on. The default port number for any ICAP server is **1344**, possible values:

          - Any port number that isn't used in your machine.

        - **services**

          The array that contains integrated services names with **ICAPeg**, possible values:

          - Integrated services names with **ICAPeg** (ex: ["echo"]).

      - **[echo] section** 

        >  **Note**: Variables explained in **echo** service are mandatory with any service integrated with **ICAPeg**.

        ```toml
        [echo]
        vendor = "echo"
        service_caption= "echo service"   #Service
        service_tag = "ECHO ICAP"  #ISTAG
        req_mode=true
        resp_mode=true
        shadow_service=false
        preview_enabled = true# options send preview header or not
        preview_bytes = "1024" #byte
        timeout  = 300 #seconds , ICAP will return 408 - Request timeout
        fail_threshold = 2
        #max file size value from 1 to 9223372036854775807, and value of zero means unlimited
        max_filesize = 0 #bytes
        return_original_if_max_file_size_exceeded=false
        bypass_extensions = []
        process_extensions = ["*"] # * = everything except the ones in bypass, unknown = system couldn't find out the type of the file
        base_url = "$_CLOUDAPI_URL" #
        scan_endpoint = "/api/rebuild/file"
        api_key = "$_AUTH_TOKENS"
        ```
        
        - ### **Mandatory variables (Variables that should any service has)**
        
          - **vendor**
        
            The name of the vendor's service, possible values:
        
            - The vendor of that service (ex: **"echo"**)
        
          - **service_caption**
        
            Service caption header value.
        
          - **service_tag**
        
            Service caption header value.
        
          - **req_mode**
        
            Boolean variable that indicates to wether **request mode** is enabled or not, possible values:
        
            - **true**: Request mode is enabled.
            - **false**: Request mode is disabled.
        
            Get more details about **request mode** from [here](https://datatracker.ietf.org/doc/html/rfc3507#section-3.1).
        
          - **resp_mode**
        
            Boolean variable that indicates to wether **response mode** is enabled or not, possible values:
        
            - **true**: Response mode is enabled.
            - **false**: Response mode is disabled.
        
            Get more details about **response mode** from [here](https://datatracker.ietf.org/doc/html/rfc3507#section-3.2).
        
          - **shadow_service**
        
            Boolean variable that indicates to wether **shadow service mode** is enabled or not, possible values:
        
            - **true**: Shadow service mode is enabled.
            - **false**: Shadow service mode is disabled.
        
            > **Note**: Shadow servie mode is used for debugging purposes. it means that when user/client sent a request to **ICAPeg**, **ICAPeg** will send an **ICAP** response with **204 (No modifications) ICAP status code** incase **ICAP** request has (**Allow: 204**) header or with **200 (OK) ICAP status code** with the **original HTTP message** incase **ICAP** request hasn't (**Allow: 204**) header.
        
          - **preview_enabled**
        
            Boolean variable that indicates to wether **message preview** is enabled or not, possible values:
        
            - **true**: Message Preview is enabled.
            - **false**: Message Preview is disabled.
        
            Get more details about **Message Preview** from [here](https://datatracker.ietf.org/doc/html/rfc3507#section-4.5).
        
          - **preview_bytes**
        
            It indicates how many bytes of preview are needed for a particular **ICAP** application on a per-resource basis, possible values:
        
            - Any string numeric value
        
          - **timeout**
        
            It indicates to How many seconds that **ICAP will return 408 - Request timeout** after, possible value:
        
            - Any integer value.
        
        - ### **Optional variables** (Variables that depends on the service)
        
          > **Note:** 
          >
          > - You may not use these variables in your service and you may use, It depends on your service and It's up to you.
          > - We will pretend that this service is for file processing and it sends that file to an external api to process it then it gets it back again, So all optional variables depends on that scenario in this service. (It's just a fake scenario service can do any thing not just for processing files).
        
          - **max_filesize**
        
            It's the maximum **HTTP** message file size that service can process , possible values:
        
            - Any valid integer value.
        
          - **return_original_if_max_file_size_exceeded**
        
            Boolean variable that indicates to wether service should return the original file if the file size exceeds the maximum file size or not, possible values:
        
            - **true**: Returning the original file.
            - **false**: Returning **400 Bad request**.
        
            Get more details about **request mode** from [here](https://datatracker.ietf.org/doc/html/rfc3507#section-3.1).
        
          - **bypass_extensions**
        
            An Array that contains the extensions of that service can't process if the **HTTP** message contains a file.
        
            - Any valid file extenstions.
        
          - **process_extensions**
        
            An Array that contains the extensions of that service canprocess if the **HTTP** message contains a file.
        
            - Any valid file extenstions.
        
          - **base_url**
        
            The external **API URL** that service sends the files through a request to it.
        
          - **scan_endpoint**
        
            Endoint of the exyernal **API URL** that service sends the files through a request to it..
        
          - **api_key**
        
            The key of the external **API** that service sends the files through a request to it.
        

## Adding a new service ot ICAPeg

- [How to add a new service for a new vendor](How%20to%20add%20a%20new%20service%20for%20a%20new%20vendor.md)

## Things to keep in mind

1. You will have to restart the ICAP server each time you change anything in the config file.

2. You will have to restart squid whenever you restart the ICAP.

3. You need to configure your network(or your browser)'s proxy settings to go through squid.

## More on ICAPeg

1. [Remote ICAP Servers & Shadowing](REMOTEANDSHADOW.md)

2. [Logging](LOGS.md)


### Contributing

This project is still a WIP. So you can contribute as well. See the contributions guide [here](CONTRIBUTING.md).

### License

ICAPeg is licensed under the [Apache License 2.0](LICENSE).
