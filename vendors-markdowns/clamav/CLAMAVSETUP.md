
## For Linux

Update your package list

 ```bash
 $ sudo apt-get update
 ```

Install ClamAV

  ```bash
  $ sudo apt-get install clamav clamav-daemon -y
  ```

Stop ClamAV process

 ```bash
 $ sudo systemctl stop clamav-freshclam
 ```

Manually update the ClamAV signature database

```bash
$ sudo freshclam
```

Restart the service to update the database in the background

```bash
$ sudo systemctl start clamav-freshclam
```

Reconfigure **clamav-deomen** to create clamd(the daemon interface) socket file

```bash
$ sudo dpkg-reconfigure clamav-daemon
```

After hitting enter you will be asked a lot of questions, Answer as the following order:

1- Yes

2- UNIX

3- Ok

4- Ok

5- Ok

6- Yes

7- Yes

8- Yes

9- Ok

10- Ok

11- Yes

12- Ok

13- Ok

14- Ok

15- Yes

16- Ok

17- Yes

18- Yes

19- Ok

20- Ok

21- Ok

22- Yes

23- TrustSigned

24- Ok

By default the clamd(the daemon interface) socket file path should be ```/var/run/clamav/clamd.ctl```. This is the path you use in the config.toml file.

## For MAC

Make sure you have homebrew installed.

Install clamav using homebrew like this

 ```bash
   $ brew install clamav
 ```

This will install clamav along with all the dependencies needed.

Next, create two configuration files named, **freshclam.conf** & **clamd.conf** in the directory ```/usr/local/etc/clamav/``` (created when clamav is installed)

Put the following in the **freshclam.conf** file:

 ```bash
   # /usr/local/etc/clamav/freshclam.conf
   DatabaseMirror database.clamav.net
 ```

Next, put the following in the **clamd.conf** file:

```bash
  # /usr/local/etc/clamav/clamd.conf
  LocalSocket /usr/local/var/run/clamav/clamd.sock
```

You need to ensure the directory ```/usr/local/var/run/clamav``` exists, create it manually if doesn't exist.

Next, start the freshclam service like this

 ```bash
   $ freshclam -v
 ```

You should something like this:

 ```bash
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

 ```bash
   $ clamd
 ```

 And thats it.



## For Windows

Couldn't find a better way than whats in the [official documentation](https://www.clamav.net/documents/installing-clamav-on-windows)


The setups mentioned here is enough to get the clamav service up and running and ready to be used by ICAPeg, but if you want to have more control over clamav, you need to modify the clamd.conf files more, [You can find it here somewhere](https://www.clamav.net/documents/clam-antivirus-user-manual)
