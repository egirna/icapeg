# icapeg
Open Source ICAP server

Scan any file requested over the internet using the icapeg. This project is still in progress. If you don't know about the ICAP protocol, here is a
bit about icap:

## What is ICAP?

**ICAP** stands for **Internet Content Adaptation Protocol**. If a **content**(for example: file) you've requested over the internet
to download or whatever, needs **adaptation**(some kind of modification or analysis), the proxy server sends the content to the icap server for adaptation and after performing the required tasks on the content, the icap server sends it back to the proxy server so that it may return the adapted content back to the destination. This can occur both during request and response.

To know more about the icap protocol, [check this out](https://tools.ietf.org/html/rfc3507).

## How do i turn this thing on!!

To turn on the icapeg server, proceed with the following steps(assuming you have golang installed in you system):

1. Enable go mod by
 ```
    export GO111MODULE=on

 ```
1. Add the dependencies in the vendor file by
 ```
   go mod vendor

 ```
1. Build the icapeg binary  by
  ```
    go build .

  ```

1. Finally execute the file like you would for any other executable according to your OS, for unix based users though
  ```
    ./icapeg

  ```
   You should see something like, ```Staring the icap server ...```
OR, you can do none of the above and simply execute the **run.sh** shell file provided, by
 ```
  ./run.sh
```
That should do the trick.

1. Now that the server is up and running the next thing to do is setup a proxy server which can send the request body to the icapeg server for adaptation. [Squid](http://www.squid-cache.org/) looks like just the thing for the job, go to the site provided and set it up like you want. Here is a sample conf file for squid:
 ```
  acl hoka dstdomain .<the domain name>
  http_access allow hoka
  http_access deny all

  icap_enable on
  icap_service service_resp respmod_precache icap://127.0.0.1:1344/respmod-icapeg
  adaptation_access service_resp allow all
  http_port 3128
  cache deny all
```
1. Now that you have squid running as well, you can test it out by trying to download/access a file on the domain you've provided on the conf file of the squid and see the magic happen! You'll be able to download/access the file if its alright, but something like a malicious file, you are gonna see something like this:
![error_page](img/error_page.png)

Oh, and do not forget to setup your Browser or Machine 's  proxy settings according to the squid.


### Contributing

This project is still a WIP. So you can contribute as well. See the contributions guide [here](CONTRIBUTING.md).

### License

Icapeg is licensed under the [MIT License](LICENSE).
