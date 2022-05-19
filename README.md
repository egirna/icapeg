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

## Configuration

​	You can change the default configuration file whatever you want to customize the application.

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
        
        

### Insert `Glasswall` as your scanner vendor in the config.toml file

  ```code
    resp_scanner_vendor = "glasswall"
  ```

  Or,

  ```code
    req_scanner_vendor = "glasswall"
  ```

Setup **VirusTotal:**

Insert `VirusTotal` as your scanner vendor in the config.toml file

  ```code
    resp_scanner_vendor = "virustotal"
  ```

  Or,

  ```code
    req_scanner_vendor = "virustotal"
  ```

In that same file, add a **VirusTotal API key** in the `api_key` field of the `[virustotal]` section. [Here is how you can get it](VIRUSTOTALAPI.md).

Setup **MetaDefender:**

Insert `MetaDefender` as your scanner vendor in the config.toml file

  ```code
    resp_scanner_vendor = "metadefender"
  ```

  Or,

  ```code
    req_scanner_vendor = "metadefender"
  ```

In that same file, add a **MetaDefender API key** in the `api_key` field of the `[metadefender]` section. [Here is how you can get it](METADEFENDER.md).

Setup **VMRay:**

Insert `vmray` as your scanner vendor in the config.toml file

  ```code
    resp_scanner_vendor = "vmray"
  ```

  Or,

  ```code
    req_scanner_vendor = "vmray"
  ```

In that same file, add a **VMRay API key** in the `api_key` field of the `[vmray]` section. [Get your api key by requesting a free trial](https://www.vmray.com/analyzer-malware-sandbox-free-trial/).

Setup **Clamav:**

Insert `clamav` as your scanner vendor in the config.toml file

  ```code
    resp_scanner_vendor = "clamav"
  ```

Next, provide the **clamd socket file path**(getting back to this in a bit) in the config.toml file inside the clamav section

  ```code
    socket_path = "<path to clamd socket file>"
  ```

[Here is how you setup clamav and generate the socket file](CLAMAVSETUP.md)


**NOTE**: All the settings of ICAPeg is present in the **config.toml** file in the repo. Also before selecting your vendors as the scanners, keep in mind to check whether that certain vendor supports the modification mode or not. For example, when adding ``virustotal``  as the ``resp_scanner_vendor``, check under the configuration of ``virustotal`` if the ``resp_supported`` flag is true or not. Likewise for ``req_scanner_vendor`` and for any other vendors. Also you can provide `none` in the ``resp/req_scanner_vendor/vendor_shadow`` fields to indicate no vendor is provided & ICAPeg is just gonna avoid processing the requests.

## How do I turn this thing on!!

To turn on the ICAPeg server, proceed with the following steps (assuming you have golang installed in you system):

1. Clone the ICAPeg repository

  ```bash
    git clone https://github.com/egirna/icapeg.git

  ```


2. Enable `go mod`

  ```bash
    export GO114MODULE=on

  ```
>    In case not using go version 1.14, you could discover your version

  ```bash
    go version

  ```

>           You should use the corresponding export command

>           1.14 ===> export GO114MODULE=on

>           1.13 ===> export GO113MODULE=on

>           etc.

3.  Change the directory to the repository

  ```bash
    cd icapeg/
  ```

4. Add the dependencies in the vendor file

  ```bash
    go mod vendor
  ```

5. Build the ICAPeg binary by

  ```bash
    go build .
  ```

6. Finally execute the file like you would for any other executable according to your OS, for Unix-based users though

  ```bash
    ./icapeg
  ```

   You should see something like, ```ICAP server is running on localhost:1344 ...```. This tells you the ICAP server is up and running
OR, you can do none of the above and simply execute the **run.sh** shell file provided, by

  ```bash
   ./run.sh
  ```
That should do the trick.

2. Now that the server is up and running, the next thing to do is setup a proxy server which can send the request body to the ICAPeg server for adaptation. [Squid](http://www.squid-cache.org/) looks like just the thing for the job, go to the site provided and set it up like you want.
After setting up your proxy server for example squid, change its configuration file:

Open squid.conf file

  ```bash
    sudo nano /etc/squid/squid.conf
  ```
Add the following lines at the bottom of your ACLs configurations

  ```configuration
    icap_enable on
    icap_service service_resp respmod_precache icap://127.0.0.1:1344/respmod
    adaptation_access service_resp allow all
  ```

Add the following line at the end of the file

  ```configuration
    cache deny all
  ```

A sample conf file for squid exists in the repository in a file
   [squid.conf](https://github.com/mkaram007/icapeg/blob/fa4ce337b27a2583c93c5dc81d8c7310fdc38e3a/squid.conf)


Save and close the file
  Press CTRL + x, then press Y, then Enter

Restart squid:

  ```bash
    systemctl restart squid
  ```

## REQMOD

**Go-ICAP-server** in **REQMOD** supports different values of **Content-Type** HTTP header:

- First value which represents a regular file sent in body like and in this case the Content-Type value maybe **application/pdf**, **text/plain** .. etc. In this type the body of the **HTTP** request which is included inside the **ICAP** request contains the file which wanted to be scanned only, **[Filebin](https://filebin.net/)** is one of the websites which use this types of **Content-Type**s in uploading files.

  ![](./img/Filebin%20contentType.png)

  ![](./img/Filebin%20requestBody.png)

- In the second type the **Content-Type** value is **multipart/form-data**, In this type the HTTP request body contains multiple parts of fields and the files, **[File.io](https://www.file.io/)** is one of the websites which use **multipart/form-data** in uploading files.

  ![](./img/fileio%20contentType.png)

  This screenshot shows different parts in the HTTP request body and the boundary -----------------------------2145730498846892413066303047 differentiates between all parts.

  ![](./img/fileio%20requsetBody.png)

- In the third type the **Content-Type** value is **application/json**, In this type the HTTP request body contains the file which wanted to be scanned encoded in base64 or a normal JSON file. **[Glasswall solutions](https://www.glasswallsolutions.com/test-drive/)** is one of the websites which encodes the file in Base64 and put it in the HTTp request body as a JSON file.

  ![](./img/gw%20contentType.png)

  ![](./img/gw%20requestBody.png)

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
