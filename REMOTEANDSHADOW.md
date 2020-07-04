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

## Shadowing Any External Service

ICAPeg now lets the users **shadow** any of the external services including both virus scanners & remote ICAP servers. What does this mean ? So basically a remote call will be made without that having any actual impact on the work flow of the ICAPeg. For example: if you select  ``virustotal`` as your virus scanner vendor to be **shadowed**, ICAPeg is gonna scan the requested file using ``virustotal`` but the results of the scan won't be considered at all rather, just logged for the user to see. So whether the virus scanner judges the file as malicious or an OK one, ICAPeg is just gonna log the results and not consider it within its work flow.

Similarly, if a remote ICAP server ``abcd`` is selected to be shadowed, the traffic will be passed on to ``abcd`` but its response is not going to effect the work flow of the ICAPeg, but just log the response received and continue doing what its suppose to do.

The services can be **shadowed** along with a primary ``resp/req_scanner_vendor`` ,in which case, both the services will be called(including the shadow one), while the primary one will effect the work flow and the **shadow** will just log the results. Also, a service can be **shadowed** independant of any primary services, in which case, ICAPeg is always gonna return ``204`` and just log the results of the shadow vendor.

### How To Do This?

Notice the two fields called ``resp_scanner_vendor_shadow`` & ``req_scanner_vendor_shadow`` in the ``app`` block of the ``config.toml`` file. Simply put the name of the service in the respective fields to make them your **shadow** vendor. For example:

  ```bash
    resp_scanner_vendor_shadow = 'virustotal'
  ```
Is gonna make ``virustotal`` your shadow vendor, that is virustotal will be called for scanning RESPMOD requests but the results will just be logged and won't have any impact. And:

  ```bash
    resp_scanner_vendor_shadow = 'icap_abcd'
  ```
Will make the external ICAP server ``abcd``(just for example) to just log its responses rather than having any impact.


**NOTE**: For remote ICAP servers, the prefix **icap_** must be used with the name of the vendors to be used in the config file, both in their indiviual blocks (like ``icap_something``) and in the ``app`` block for ``resp/req`` scanner vendors.
