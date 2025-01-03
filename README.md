```
+(*-> |             | <-*)
+())| | Mauerspecht | |(()
+ \"| | thcepsreuaM | |"/
```

# Simple Probing Tool for Corporate Walled Garden Networks

The Problem: Network sensors such as
[Suricata](https://suricata-ids.org/) or [Zeek](https://zeek.org/)
have been successfully deployed in a large network, but the rate of
alarms or other useful information is suspiciously low -- not even the
usual background noise can be seen. Can we be sure that our sensors
are fed all the relevant traffic?

An attempt at a solution: Let's generate some network traffic and see
if we can transmit some magic strings to and from the outside world
beyond our walled garden network -- and if we are able to detect those
using our sensors.

## Operation

From a user perspective:

1. Generate a server configuration file that defines TCP ports and
   magic strings to exchange (see below for an example). Configure
   matching alerting rules in the network sensors.
2. Start the server on a publicly accessible host.
3. Start clients with the `-server` parameter pointing to one of the
   HTTP ports served by the server.
4. Analyze logs generated by the server and the network sensors.

The server writes its log output to standard error.

What happens behind the scenes:

1. On startup, both server and client generate private/public NaCL key
   pairs.
2. The client posts its public key to the server and receives the
   server's public key
3. The client requests the server's configuration. The configuration
   is signed/encrypted to circumvent tampering by middleboxes.
4. The client runs a few experiments, expecting every configured magic
   strings to be correctly transmitted via a special header, a Cookie
   or _Set-Cookie_ header, the message body.
5. The client posts its findings to the server.

### Example server configuration file

```
{
    "hostname": "mauerspecht.example.com",
    "http-ports": [8080, 18080],
    "magic-strings": [
        "unique-match-string-18475910",
        "START_KEYLOGGER",
        "X5O!P%@AP[4\\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*"
    ]
}
```

### Command line parameters

Client:
```
  -server string
    	Server URL (default "http://localhost:8080")
  -proxy string
    	Proxy URL
```
Server:
```
  -config string
    	Config file (default "mauerspecht.json")
```

## Building

For recent Go versions, simply running `make all` from the Git checkout is
sufficient.

The following binaries will be generated:
- `mauerspecht-server`: The server component, a Linux/x86-64 binary
- `mauerspecht-client-$ARCH`: The clients, for various architectures

## Limitations, possible future features

- HTTPS -- self-signed server certificates, possible use of client certificates
- Non-HTTP protocols (IRC?)
- The server stores session keys submitted by clients in memory and does
  not expire them yet. This is a denial-of-service vector.
- Bundled client configuration for easy single-binary deployment (see
  also: [spyre](https://github.com/spyre-project/spyre))

## Contact

Sascha Steinbiss <<sascha.steinbiss@dcso.de>>

Original Author: Hilko Bengen

## Copyright

Copyright 2019, 2024 Deutsche Cyber-Sicherheitsorganisation GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
