<h1 align="center">Configure NGINX with ICAPeg</h1>
<p align="center">
    <em>Secure ICAP.</em>
     
</p>

# Secure ICAP

Secure ICAP is ICAP over TLS. 
By using the **NGINX** we can convert the plain ICAP connection to secure connection using SSL/TLS certificate and key.
 
## What is NGINX?

**NGINX** is **open source software** for web serving, reverse proxying, caching, load balancing, media streaming, and more. It started out as a web server designed for maximum performance and stability. In addition to its HTTP server capabilities, NGINX can also function as a proxy server for email (IMAP, POP3, and SMTP) and a reverse proxy and load balancer for HTTP, TCP, and UDP servers.

To know more about the NGINX, [check this out](https://www.nginx.com/resources/glossary/nginx/).

## Table of Contents

- [Prerequisites](#prerequisites)
- [Configure NGINX](#configure-nginx)
- [Testing](#testing)
- [Things to keep in mind](#things-to-keep-in-mind)

## Prerequisites

Before starting using ICAP with NGINX, make sure you have the following things in your machine:

1. **NGINX**

     You should install the latest stable version of NGINX Open Source **nginx-1.24.0**.

     You can download **NGINX** from [here](https://nginx.org/en/download.html).

2. **Operating System Windows or Linux**

    In this scenario we will use Windows10 (64 bit) Operating System.

3. **ICAP Server**
    
    ICAP Server running on the same machine that NGINX runs on.
    In this scenario we will use **ICAPeg**

4. **SSL Certificate and Key**

    Generate an SSL/TLS Certificate. You’ll need to install OpenSSL.

    You can download **OpensSSL** from [here](https://www.openssl.org/source/).

    If you already have OpenSSL installed on your system, you can use the following command to generate an SSL/TLS certificate and key.

    Run the command below to generate a self-signed certificate and key: 

    **Note:** Change ***name*** to any specific name you want.

    ```bash
      openssl req -x509 -nodes -newkey rsa:2048 -keyout C:\nginx\name.key -out C:\nginx\name.crt   
      days 365
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




## Configure NGINX

1. **Install NGINX**

    - Download NGINX from NGINX site [here](https://nginx.org/en/download.html).
    - Extract NGINX zip file in path **C:\\** for Windows.
    - Rename the extracted folder to **nginx**.
    - From Windows search box; search for **Edit the System Environment Variables**, Select:
        - **Advanced**
        - **Environment Variables**
        - Under **System Variables**, select **Path** then click **Edit**
        - Click **New** to add NGINX path **c:\nginx** 
        - Click **OK** 
    - Open Windows CMD to check NGINX version and start the NGINX.
      ```bash
      nginx -version
      ```
      - Navigate to NGINX path **c:\nginx**.
      ```bash
      cd c:\nginx
      ```
      - Run the NGINX.
      ```bash
      c:\nginx>nginx
      ```
      - Navigate to your browser to check if NGINX is running. Write this in browser tab:

      ```bash
      http://localhost/
      ```
      You should see **"Welcome to nginx!"** page.

        You can change the default configuration file **nginx.conf** whatever you want to customize the application. 

        In this scenario we will add a new section in **nginx.conf** to enable the NGINX accept the ICAP server connection with SSL  
        Certificate and Key.

    - Add this section as a separate section in **nginx.conf** file, not under http section in the nginx.conf file:

       **Note:** 
          - Change ***name*** of ssl_certificate and ssl_certificate_key to the name you created before.
          - **Listen** : We use **1345** as a port for using ICAP with SSL/TLS.

         ```bash
            stream {
                upstream stream_backend {
                     server localhost:1344;
              
                }
        

                server {
                    listen                1345 ssl;
                    proxy_pass            stream_backend;
                    ssl_certificate       c:/nginx/name.crt;
                    ssl_certificate_key   c:/nginx/name.key;
                    ssl_protocols         SSLv3 TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
                    ssl_ciphers           HIGH:!aNULL:!MD5;
                    ssl_session_cache     shared:SSL:20m;
                    ssl_session_timeout   4h;
                    ssl_handshake_timeout 30s;
                    # #...
                }
            }
      ``` 

    - Reload NGINX in new CMD window using this command: 

    ```bash
      cd C:\nginx
      c:\nginx>nginx -s reload
      ``` 
  

## Testing

1. Run the ICAP server.
2. Run the NGINX.
3. Open the CMD to test using telnet.

  ```bash
    C:\Users>telnet localhost 1345
  ```
  You will see this port is accepting the connection. 


## Things to keep in mind

  - You will have to reload the **NGINX** server each time you change anything in the nginx.conf file.

# Configure NGINX to run as a Windows Service

You can't run the NGINX as a Windows Service without 3rd party software.

They mentioned it will be a Possible future enhancements [check this out](https://nginx.org/en/docs/windows.html).

So, we will use **NSSM** as a 3rd party software in this scenario.

## What is NSSM ?

NSSM stands for the **Non-Sucking Service Manager**. NSSM is a service helper which doesn't suck. servant and other service helper programs suck because they don't handle failure of the application running as a service. If you use such a program you may see a service listed as started when in fact the application has died. NSSM monitors the running service and will restart it if it dies. With NSSM you know that if a service says it's running, it really is. Alternatively, if your application is well-behaved you can configure NSSM to absolve all responsibility for restarting it and let Windows take care of recovery actions. To know more about the NSSM, [check this out](https://nssm.cc/).
 


  1. Download **NSSM** for Windows latest release [here](https://nssm.cc/download)

  2. Extract NSSM zip file in path **C:\\** for Windows, then open it and navigate to: 

      - **win64** folder in NSSM extracted folder then, Copy **nssm.exe** to NGINX folder.

      - Copy **src** folder from NSSM folder to  NGINX folder.

      - Open Windows PowerShell as **Administrator**.

      - Navigate to NGINX folder 
      ``` bash
      cd C:\nginx
      ```
      - install NSSM 
      ```bash
      .\nssm.exe install
      ```
      After running the command, you will be prompted to provide some information:

      - In **Path** tab select the **nginx.exe**.

      - In **Service Name** enter the name you need, we use **nginx** as a service name.

      - Click **Install Service** button, then OK.

      - Open **Windows Services** from Windows search box.

      - You should find the service as you named it here. 

      - Right click on it, then **Start**.
  3. Navigate to your browser to check if NGINX is running. Write this in browser tab :

      ```bash
      http://localhost/
      ```
      You should see **"Welcome to nginx!"** page.

**Note**
  - You will have to restart the **NGINX** service each time you change anything in the nginx.conf file from **Windows Services**.
  - Right click on the service name, then **Restart**.






