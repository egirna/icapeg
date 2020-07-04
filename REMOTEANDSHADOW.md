## Communicate With External ICAP Servers

ICAPeg now enables itself to work as a surrogate by sitting between an ICAP client and an external ICAP server to pass the traffic on to.

### How To Do This?

There are two blocks of configs placed in the ``config.toml`` file as example named: **icap_something** & **icap_somethingelse** both are of course almost identical. They represent the Remote ICAP configurations. For example: if the name of the remote ICAP server you want to communicate with using ICAPeg is **abcd**, change the **icap_something**(or put a new block) in the config file with the prefix **icap_abcd**. Fill in every fields provided there.
Next, depending on the **Request MOD** you want, place the name of the remote ICAP server in either ``resp_scanner_vendor`` (for RESPMOD)
or ``req_scanner_vendor`` (for REQMOD) under the ``app`` block just like you would with the virus scanners. For example:

  ```bash
    resp_scanner_vendor = 'icap_abcd'
    req_scanner_vendor = 'icap_abcd'

  ```

And thats about it, the ICAPeg will now work like a middleware receiving the ICAP requests from the clients and passing that on to the remote ICAP server and also pass on the response received from the remote back to the client, according to the name and other configurations you've provided in the ``config.toml`` file. You should see something like this in the logs(if the proper level provided):

  ```bash
  Passing request to the remote ICAP server...

  ```
