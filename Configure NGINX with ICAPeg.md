<h1 align="center">Configure NGINX with ICAPeg</h1>
<p align="center">
    <em>Secure ICAP.</em>
     
</p>

# Secure ICAP

Secure ICAP is ICAP over TLS.
By using the **NGINX** we can handle TLS. NGINX negotiates a secure connection with the client and serves ICAP content authenticated by the SSL/TLS certificate.

## What is NGINX?

**NGINX** is **open-source software** for web serving, reverse proxying, caching, load balancing, media streaming, and more. It started out as a web server designed for maximum performance and stability. In addition to its HTTP server capabilities, NGINX can also function as a proxy server for email (IMAP, POP3, and SMTP) and a reverse proxy and load balancer for HTTP, TCP, and UDP servers.

To know more about NGINX, [check this out](https://www.nginx.com/resources/glossary/nginx/).

## Table of Contents

- [Prerequisites](#prerequisites)
- [Install and Configure NGINX with ICAP](#install-and-configure-nginx-with-icap)
- [Testing](#testing)
- [Things to keep in mind](#things-to-keep-in-mind)

## Prerequisites

Before starting to use ICAP with NGINX, make sure you have the following prerequisites on your machine:

**1. NGINX**

     You should install the latest stable version of NGINX Open Source **nginx-1.24.0**.

     You can download **NGINX** from [here](https://nginx.org/en/download.html).

**2. Operating System: Windows or Linux**

    In this scenario, we will use the Windows 10 (64-bit) operating system.

**3. ICAP Server**

    ICAP Server is running on the same machine that NGINX runs on.

    In this scenario, we will use **ICAPeg**.

**4. SSL Certificate and Key**

    Generate an SSL/TLS certificate. You’ll need to install OpenSSL.

    You can download **OpensSSL** from [here](https://www.openssl.org/source/).

    If you already have OpenSSL installed on your system, you can use the following command to generate an SSL/TLS certificate and key.

    Run the command below to generate a self-signed certificate and key:

    **Note:** Change ***name*** to any specific name you want.

 ```bash
    openssl req -x509 -nodes -newkey rsa:2048 -keyout C:\nginx\name.key -out C:\nginx\name.crt days 365
```
After running the command, you will be prompted to provide some information:

```bash
    Country Name (2 letter code): IN
    State or Province Name: TN
    Locality Name: Your city name
    Organization Name: Your organization
    Organizational Unit Name: localhost
    Common Name: localhost
    Email Address: Your email address
```


## Install and Configure NGINX with ICAP

**1. Install NGINX**

 - Download NGINX from the NGINX site [here](https://nginx.org/en/download.html).
 - Extract the NGINX zip file in the path **C:\\** for Windows.
 - Rename the extracted folder to **nginx**.
 - From the Windows search box, search for **Edit the System Environment Variables**, Select:            
    - **Advanced**
    - **Environment Variables**
        - Under **System Variables**, select **Path**, then click **Edit**.
        - Click **New** to add NGINX path **c:\nginx**.
        - Click **OK**

    - Open Windows CMD to check the NGINX version and start NGINX.
        - For NGINX version.
            ```bash 
            nginx -version
            ```
        - Navigate to the NGINX path **c:\nginx**.
            ```bash
            cd c:\nginx
            ```
        - Run the NGINX.
            ```bash
            c:\nginx>nginx
            ```
        - Navigate to your browser to check if NGINX is running. Write this in the browser tab:
            ```bash
            http://localhost/
            ```   
            You should see the **"Welcome to nginx!"** page.

 **2. Configure NGINX with ICAP**

You can change the default configuration file **nginx.conf** to whatever you want to customize the application.

In this scenario, we will add a new section in **nginx.conf** to enable NGINX to negotiates a secure connection with the client and serves ICAP content authenticated by the SSL/TLS certificate.
- Add this section as a separate section in the **nginx.conf** file, not under the http section in the nginx.conf file.

    **Note:**
    - Change ***name*** of ssl_certificate and ssl_certificate_key to the name you created before.
    - **Listen** : We use **1345** as a port for using ICAP with SSL/TLS.

```bash
    stream {
        upstream stream_backend {
             server localhost:1344;
                 
        }
           
        server {
            listen                1345 ssl;
            proxy_pass            stream_backend;
            ssl_certificate       c:/nginx/name.crt;
            ssl_certificate_key   c:/nginx/name.key;
            ssl_protocols         SSLv3 TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
            ssl_ciphers           HIGH:!aNULL:!MD5;
            ssl_session_cache     shared:SSL:20m;
            ssl_session_timeout   4h;
            ssl_handshake_timeout 30s;
                       
         }
    }
    
```
 - Reload NGINX in a new CMD window using this command:
    
    ```bash
    cd C:\nginx
    c:\nginx>nginx -s reload
    ```


## Testing

1. Run the ICAP server.
2. Run the NGINX.
3. Open the CMD to test using Telnet.

```bash
C:\Users>telnet localhost 1345
```
  You will see that this port is accepting the connection.


## Things to keep in mind

- You will have to reload the **NGINX** each time you change anything in the nginx.conf file.

# Configure NGINX to run as a Windows Service

You can't run NGINX as a Windows service without third-party software.

They mentioned it will be a possible future enhancement [check this out](https://nginx.org/en/docs/windows.html).

So, we will use **NSSM** as 3rd party software in this scenario.

## What is NSSM?

NSSM stands for the **Non-Sucking Service Manager**. NSSM is a service helper that doesn't suck. servant and other service helper programs suck because they don't handle the failure of the application running as a service. If you use such a program, you may see a service listed as started when, in fact, the application has died. NSSM monitors the running service and will restart it if it dies. With NSSM, you know that if a service says it's running, it really is. Alternatively, if your application is well-behaved, you can configure NSSM to absolve all responsibility for restarting it and let Windows take care of recovery actions. To know more about the NSSM, [check this out](https://nssm.cc/).



  1. Download **NSSM** for Windows latest release [here](https://nssm.cc/download)

  2. Extract the NSSM zip file in the path **C:\\** for Windows, then open it and navigate to:

- **win64** folder in the NSSM extracted folder, then copy **nssm.exe** to the NGINX folder.

- Copy the **src** folder from the NSSM folder to the NGINX folder.

- Open Windows PowerShell as **Administrator**.

- Navigate to the NGINX folder.

``` bash
cd C:\nginx
```
- install NSSM
```bash
.\nssm.exe install
```
After running the command, you will be prompted to provide some information:

- In the **Path** tab, select **nginx.exe**.

- In the **Service Name** field, enter the name you need; we use **nginx** as a service name.

- Click the **Install Service** button, then OK.

- Open **Windows Services** from the Windows search box.

- You should find the service as you named it here.

- Right-click on it, then **Start**.

  3. Navigate to your browser to check if NGINX is running. Write this in the browser tab:

 ```bash
 http://localhost/
 ```
You should see the **"Welcome to nginx!"** page.

**Note**
- You will have to restart the **NGINX** service each time you change anything in the nginx.conf file from **Windows Services**.
    - Right-click on the service name, then **Restart**.






