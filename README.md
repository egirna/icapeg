<h1 align="center">Go ICAP Server</h1>
<p align="center">
    <em>ICAPeg Server.</em>
</p>

<p align="center">
    <a href="https://github.com/egirna/icapeg/actions/workflows/build.yml">
        <img src="https://github.com/egirna/icapeg/actions/workflows/build.yml/badge.svg"/>
    </a> 
    <a href="https://codecov.io/gh/egirna/icapeg">
        <img src="https://codecov.io/gh/egirna/icapeg/branch/master/graph/badge.svg?token=HRMICTHXBQ)"/>
    </a>	    
    <!-- <a href="https://goreportcard.com/report/github.com/k8-proxy/go-icap-server">
      <img src="https://goreportcard.com/badge/k8-proxy/go-icap-server" alt="Go Report Card">
    </a> -->
	<a href="https://github.com/egirna/icapeg/pulls">
        <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat" alt="Contributions welcome">
    </a>
    <a href="https://opensource.org/licenses/Apache-2.0">
        <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="Apache License, Version 2.0">
    </a>
</p>

# ICAPeg

Open Source multi-vendor ICAP server

Scan files requested via a proxy server using ICAPeg ICAP server, ICAPeg is an ICAP server connecting web proxies with API-based scanning services and more soon. ICAPeg currently supports [**VirusTotal**](https://www.virustotal.com/gui/home/upload), [**Cloudmersive**](https://cloudmersive.com/) & [**Clamav**](https://www.clamav.net/)  for scanning the files following the ICAP protocol. If you don't know about the ICAP protocol, here is a bit about it: 

## What is ICAP?

**ICAP** stands for **Internet Content Adaptation Protocol**. If a **content** (for example: file) you've requested over the internet
to download or whatever, needs **adaptation**(some kind of modification or analysis), the proxy server sends the content to the ICAP server for adaptation and after performing the required tasks on the content, the ICAP server sends it back to the proxy server so that it may return the adapted content to the destination. This can occur both during request and response.

To know more about the ICAP protocol, [check this out](https://tools.ietf.org/html/rfc3507).

## Table of Contents

- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Adding a new vendor to ICAPeg](#adding-a-new-vendor-to-ICAPeg)
- [Developer Guide](#developer-guide)
- [How to Setup Existed Services](how-to-setup-existed-services)
  - [Echo](#echo)
  - [Virustotal](#virustotal)
  - [ClamAV](#clamav)
  - [Cloudmersive](#cloudmersive)
  - [Gray Images](#grayimages)

- [Things to keep in mind](things-to-keep-in-mind)
- [More on ICAPeg](#more-on-icapeg)
- [Contributing](#contributing)
- [License](#license)

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

6. Build **ICAPeg** binary

```bash
$ go build .
```

6. Finally execute the file like you would for any other executable according to your OS, for Unix-based users though

```bash
$ ./icapeg
```

You should see something like, ```ICAP server is running on localhost:1344 ...```. This tells you the ICAP server is up and running.

## Configuration

​	You can change the default configuration file whatever you want to customize the application.

- ### Mapping a variable of config.toml file with an environment variable

  This feature is supported for **strings, int, bool, time.duration and string slices** only **(every type used in this project)**.

  Let's have an example to explain how to map, assume that there is an environment variable in your machine called PORT and you want to assign PORT value to app.port. You should change the value of (port= 1344) to (port= "$_PORT").

  > **Note**: before you use this feature please make sure that the env variable that you want to use is globally in your machine and not just exported in a local session.

  - ### Config.toml file sections

    - #### **[app] section** 

      - This section includes the configuration variables that are responsible for the application overall like the port number and current services integrated with **ICAPeg**.

        ```toml
        [app]
        port = 1344
        services= ["echo", "virustotal", "clamav", "cloudmersive"]
        debugging_headers=true
        ```
        
        - **port**
        
          The port number that **ICAPeg** runs on. The default port number for any ICAP server is **1344**, possible values:
        
          - Any port number that isn't used in your machine.
        
        - **services**
        
          The array that contains integrated services names with **ICAPeg**, possible values:
        
          - Integrated services names with **ICAPeg** (ex: ["echo"]).
        
        - **debugging_headers**
        
          A boolean variable which indicates if debugging headers should be displayed with ICAP headers or not. Debugging headers tell the client shadow service is enabled for example. they start with **X-ICAPeg-{{HEADER_NAME}}**. possible values:
        
          - **true**: Debugging headers should be displayed with ICAP headers.
          - **false**: Debugging headers should not be displayed with ICAP headers.
        
          - Any port number that isn't used in your machine.
        
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
        process_extensions = ["pdf", "zip", "com"] 
        # * = everything except the ones in bypass, unknown = system couldn't find out the type of the file
        reject_extensions = ["docx"]
        bypass_extensions = ["*"]
        #max file size value from 1 to 9223372036854775807, and value of zero means unlimited
        max_filesize = 0 #bytes
        return_original_if_max_file_size_exceeded=false
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
        
            A boolean variable that indicates whether **request mode** is enabled or not, possible values:
        
            - **true**: Request mode is enabled.
            - **false**: Request mode is disabled.
        
            Get more details about **request mode** from [here](https://datatracker.ietf.org/doc/html/rfc3507#section-3.1).
        
          - **resp_mode**
        
            A boolean variable that indicates whether **response mode** is enabled or not, possible values:
        
            - **true**: Response mode is enabled.
            - **false**: Response mode is disabled.
        
            Get more details about **response mode** from [here](https://datatracker.ietf.org/doc/html/rfc3507#section-3.2).
        
          - **shadow_service**
        
            A boolean variable that indicates whether **shadow service mode** is enabled or not, possible values:
        
            - **true**: Shadow service mode is enabled.
            - **false**: Shadow service mode is disabled.
        
            > **Note**: Shadow service mode is used for debugging purposes. it means that when user/client sent a request to **ICAPeg**, **ICAPeg** will send an **ICAP** response with **204 (No modifications) ICAP status code** in case **ICAP** request has (**Allow: 204**) header or with **200 (OK) ICAP status code** with the **original HTTP message** in case **ICAP** request hasn't (**Allow: 204**) header.
        
          - **preview_enabled**
        
            A boolean variable that indicates whether **message preview** is enabled or not, possible values:
        
            - **true**: Message Preview is enabled.
            - **false**: Message Preview is disabled.
        
            Get more details about **Message Preview** from [here](https://datatracker.ietf.org/doc/html/rfc3507#section-4.5).
        
          - **preview_bytes**
        
            It indicates how many bytes of preview are needed for a particular **ICAP** application on a per-resource basis, possible values:
        
            - Any string numeric value
        
          - **process_extensions**
        
            It indicates the file types that should be processed and scanned from the service, and possible values:
        
            - Any string file types.
        
          - **reject_extensions**
        
            It indicates the file types that should be rejected by the service, and possible values:
        
            - Any string file types.
        
          - **bypass_extensions**
        
            It indicates the file types that should be bypassed by the service and that means the files with these types will not be processed and will not be rejected. They will just be returned to the client as they were sent to **ICAPeg** from him, and possible values:
        
            - Any string file types.
        
            > **Notes about extensions arrays:**
            >
            > - **Asterisk** sign (*****) means every file type except the ones in other arrays. example:
            >
            >   process_extensions = ["pdf", "zip", "com"] 
            >   reject_extensions = ["docx"]
            >   bypass_extensions = ["*"]
            >
            >   this example means that any file type will be bypassed except **docx** type which in **reject_extensions**, **pdf**, **zip**, **com** types which in **process_extensions**.
            >
            > - Only one array from (**process_extensions**, **reject_extensions**, **bypass_extensions**)  arrays should has **Asterisk** sign (*****).
            >
            >   process_extensions = ["pdf", "zip", "com"] 
            >   reject_extensions = ["docx", "pdf", "\*"]
            >   bypass_extensions = ["*"]
            >
            >   this configuration is not valid and **ICAPeg** will not run. another example:
            >
            >   process_extensions = ["pdf", "zip", "com"] 
            >   reject_extensions = ["docx"]
            >   bypass_extensions = ["*"]
            >
            >   this configuration is valid and **ICAPeg** will run normally.
            >
            > - Two arrays can't have the same file type. example:
            >
            >   process_extensions = ["pdf", "zip", "com"] 
            >   reject_extensions = ["docx", "pdf"]
            >   bypass_extensions = ["*"]
            >
            >   this configuration is not valid and **ICAPeg** will not run. another example:
            >
            >   process_extensions = ["pdf", "zip", "com"] 
            >   reject_extensions = ["docx"]
            >   bypass_extensions = ["*"]
            >
            >   this configuration is valid and **ICAPeg** will run normally.
          
        - ### **Optional variables** (Variables that depends on the service)
        
          > **Notes:** 
          >
          > - You may not use these variables in your service and you may use them, It depends on your service and It's up to you.
          > - We will pretend that this service is for file processing and it sends that file to an external API to process it then it gets it back again, So all optional variables depend on that scenario in this service. (It's just a fake scenario service that can do anything not just for processing files).
        
          - **max_filesize**
        
            It's the maximum **HTTP** message file size that the service can process, possible values:
        
            - Any valid integer value.
        
          - **return_original_if_max_file_size_exceeded**
        
            A boolean variable that indicates to wether service should return the original file if the file size exceeds the maximum file size or not, possible values:
        
            - **true**: Returning the original file.
            - **false**: Returning **400 Bad request**.
        
            Get more details about **request mode** from [here](https://datatracker.ietf.org/doc/html/rfc3507#section-3.1).
        

## Adding a new vendor to ICAPeg

- [How to add a new service for a new vendor](ADDING-NEW-VENDOR.md).

  After reading the above markdown, read the next section because it may help you while implementing your new service.

## Developer Guide

- [Developer Guide](developer-guide.md).

  This is a developer guide which includes a lot of functions to help the developer while implementing his new service.

## How to Setup Existed Services

- #### **Echo**: It doesn't need setup, it takes the HTTP message and returns it as it is. **Echo** is just an example service.

- #### [**Virustotal**](/vendors-markdowns/virustotal/VIRUSTOTALAPI.md).

- #### [**ClamAV**](/vendors-markdowns/clamav/CLAMAVSETUP.md).

- #### [**Cloudmersive**](/vendors-markdowns/cloudmersive/CLOUDMERSIVEAPI.md).

- #### [**Gray Images**](/vendors-markdowns/greyimages/GRAYIMAGESSETUP.md).

## Testing

- [How to test **ICAPeg**](Testing.md)

## Things to keep in mind

1. You will have to restart the ICAP server each time you change anything in the config file.

2. You will have to restart squid whenever you restart the ICAP.

3. You need to configure your network(or your browser)'s proxy settings to go through squid.

## More on ICAPeg

1. [Remote ICAP Servers & Shadowing](REMOTEANDSHADOW.md)

2. [Logging](LOGS.md)


## Contributing

This project is still a WIP. So you can contribute as well. See the contributions guide [here](CONTRIBUTING.md).

## License

ICAPeg is licensed under the [Apache License 2.0](LICENSE).

