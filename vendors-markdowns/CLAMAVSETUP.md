
## For Linux

Install the clamav and the clamav daemon like this

 ```code
  $ sudo apt-get install clamav clamav-daemon
 ```

This will install the dependencies as well.

Next , start the freshclam service(the database clamav uses to scan files, a must dependency) like this

  ```code
    $ sudo service clamav-freshclam restart
    $ sudo service clamav-freshclam status
  ```

The status should be **active**

Finally, start the clamav daemon like this

 ```code
  $ sudo service clamav-daemon start
  $ sudo service clamav-daemon status
 ```

This creates the socket file needed. Again, the status should be **active**.

By default the clamd(the daemon interface) socket file path should be ```/var/run/clamav/clamd.ctl```. This is the path you use in the config.toml file.

## For MAC

Make sure you have homebrew installed.

Install clamav using homebrew like this

 ```code
   $ brew install clamav
 ```

This will install clamav along with all the dependencies needed.

Next, create two configuration files named, **freshclam.conf** & **clamd.conf** in the directory ```/usr/local/etc/clamav/``` (created when clamav is installed)

Put the following in the **freshclam.conf** file:

 ```code
   # /usr/local/etc/clamav/freshclam.conf
   DatabaseMirror database.clamav.net
 ```

Next, put the following in the **clamd.conf** file:

```code
  # /usr/local/etc/clamav/clamd.conf
  LocalSocket /usr/local/var/run/clamav/clamd.sock
```

You need to ensure the directory ```/usr/local/var/run/clamav``` exists, create it manually if doesn't exist.

Next, start the freshclam service like this

 ```code
   $ freshclam -v
 ```

You should something like this:

 ```code
    Current working dir is /usr/local/Cellar/clamav/0.98.1/share/clamav
    Max retries == 3
    ClamAV update process started at Tue Feb  4 11:31:22 2014
    Using IPv6 aware code
    Querying current.cvd.clamav.net
    TTL: 1694
    Software version from DNS: 0.98.1
    Retrieving http://database.clamav.net/main.cvd
    Trying to download http://database.clamav.net/main.cvd (IP: 81.91.100.173)
    Downloading main.cvd [100%]
    Loading signatures from main.cvd
    Properly loaded 2424225 signatures from new main.cvd
    main.cvd updated (version: 55, sigs: 2424225, f-level: 60, builder: neo)
    Querying main.55.76.1.0.515B64AD.ping.clamav.net
...

```

Finally start the clamav daemon, by executing the **clamd** command

 ```code
   $ clamd
 ```

 And thats it.



## For Windows

Couldn't find a better way than whats in the [official documentation](https://www.clamav.net/documents/installing-clamav-on-windows)


The setups mentioned here is enough to get the clamav service up and running and ready to be used by ICAPeg, but if you want to have more control over clamav, you need to modify the clamd.conf files more, [You can find it here somewhere](https://www.clamav.net/documents/clam-antivirus-user-manual)
