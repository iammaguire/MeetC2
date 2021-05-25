![c2 image](https://i.imgur.com/x6sQ3Dd.png)

Usage:

After cloning the repo run server.go
Run 'create LHOST 8001' to generate a beacon. This will be placed in the out directory.
Execute the beacon on the target.

Inside the command server you can reference beacons using either their list id or their unique id.
For example if the output of 'list' is '[0] LuTNluGL@10.255.52.1 last seen 2021-05-23 21:07:17.208101927 -0400 EDT m=+179.667762638', you could refer to this beacon by LuTNluGL or 0.

Execute commands or download files from beacons:
exec BEACON cmd
download BEACON file

However if you select a beacon with the use command you can exempt the beacon identification.
Beacon identification can be replaced with * to schedule the operation for all live beacons.

Downloaded files will be saved to downloads/BEACON-IP/BEACON-ID/

Commands:

create &lt;listener&gt;<br />
download &lt;beacon id OR index&gt; &lt;remote file&gt; OR &lt;remote file&gt;<br />
upload &lt;beacon id OR index&gt; &lt;local file&gt; OR &lt;local file&gt;<br />
use &lt;beacon id OR index&gt;<br />
script &lt;beacon id OR index&gt; &lt;local file path&gt; &lt;remote executor path&gt;<br />
help &lt;command&gt;<br />
listeners <br />
exec &lt;beacon id OR index&gt; &lt;command&gt;<br />
list <br />
httplistener &lt;iface&gt; &lt;hostname&gt; &lt;port&gt;<br />

Todo list:
- Add communication over DNS, HTTPS, ICMP etc
- Exec:
    - HTTP: finished
- Upload
- Download: 
    - HTTP: finished
- Proxychaining beacons OR P2P beacons
- Automatic proxychains configuration depending on beacon route
- Add handling of multiple clients with random unique IDs: FINISHED
- RSA encrypt data instead of base 64
- Randomly timed sending of fragments
- Web interface
