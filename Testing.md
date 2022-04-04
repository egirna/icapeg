## Testing

- ### C-ICAP client

  You can test unsing c-icap client in you linux machine, you can download it from [here](https://howtoinstall.co/en/c-icap).

  - #### Testing REQMOD

    You can test **REQMOD** by the following command

    ```bash
    c-icap-client -i 127.0.0.1  -p 1344 -s gw_rebuild  -f ./name_of_the_file_you_want_to_test.pdf -o name_of_the_file_after testing.pdf  -v -req http://www.example.com
    ```

    

  - #### Testing RESPMOD

    ```bash
    c-icap-client -i 127.0.0.1  -p 1344 -s gw_rebuild  -f ./name_of_the_file_you_want_to_test.pdf -o name_of_the_file_after testing.pdf  -v
    ```

  You can check the documentation of how to test using c-icap client from [here](http://manpages.ubuntu.com/manpages/bionic/man8/c-icap-client.8.html).

- ### Proxy server

  You can test using Proxy server like squid, you can check how to install it in your machine from [here](https://www.egirna.com/post/configure-squid-4-17-with-icap-ssl), Then you should configure it with your favorite browser. 

  - ### REQMOD

    You can test **REQMOD** by uploading files yo any website like [**Filebin**](https://filebin.net/) and check if it's modified or not.

  - ### RESPMOD 

    You can test **RESPMOD** by googling "PDF sample download", then open any result which includes a PDF file and and check if it's modified or not.