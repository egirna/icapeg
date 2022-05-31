## Testing

- ### C-ICAP client

  You can test unsing c-icap client in you linux machine, you can download it from [here](https://howtoinstall.co/en/c-icap).

  - #### Testing REQMOD

    You can test **REQMOD** by the following command

    ```bash
    c-icap-client -i 127.0.0.1  -p 1344 -s service_name -f ./name_of_the_file_you_want_to_test.pdf -o name_of_the_file_after_testing.pdf  -v -req http://www.example.com
    ```

    

  - #### Testing RESPMOD

    ```bash
    c-icap-client -i 127.0.0.1  -p 1344 -s service_name  -f ./name_of_the_file_you_want_to_test.pdf -o name_of_the_file_after_testing.pdf  -v
    ```

  You can check the documentation of how to test using c-icap client from [here](http://manpages.ubuntu.com/manpages/bionic/man8/c-icap-client.8.html).

- ### Proxy server

  You can test using Proxy server like squid, you can check how to install it in your machine from [here](https://www.egirna.com/post/configure-squid-4-17-with-icap-ssl), Then you should configure it with your favorite browser and enjoy. 
