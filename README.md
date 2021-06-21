
https://user-images.githubusercontent.com/7650862/122788869-4b9e4600-d2a6-11eb-98c7-b33f8b2706bc.mp4

<br />

![web image](https://imgur.com/5qrWJx1.png)


Usage:

After cloning the repo run server.go
Run 'create LHOST 8001' to generate a beacon. This will be placed in the out directory.
Execute the beacon on the target.

Inside the command server you can reference beacons using either their list id or their unique id.
For example if the output of 'list' is '[0] LuTNluGL@10.255.52.1 last seen 2021-05-23 21:07:17.208101927 -0400 EDT m=+179.667762638', you could refer to this beacon by LuTNluGL or 0.

Execute commands or download files from beacons:
exec BEACON cmd
download BEACON file

After selecting an active beacon with user shellcode may be injected into another process. List processes with plist and provide the shellcode command with the path to the data as well as the PID.

However if you select a beacon with the use command you can exempt the beacon identification.
Beacon identification can be replaced with * to schedule the operation for all live beacons.

Downloaded files will be saved to downloads/BEACON-IP/BEACON-ID/

This project is in development with limited functionality.

Commands:

```
listeners 
httplistener <iface> <hostname> <port>
exec <beacon id OR index> <command>
upload <beacon id OR index> <local file> OR <local file>
use <beacon id OR index>
script <beacon id OR index> <local file path> <remote executor path>
shellcode <path to shellcode> <PID>
list 
plist 
create <listener>
download <beacon id OR index> <remote file> OR <remote file>
migrate <PID>
help <command>
```

Todo list:
- Exec:
    - HTTP: finished
- Upload
    - HTTP: finished
- Download: 
    - HTTP: finished
- Migration:
    - Windows: finished
    - Linux: WIP
- Shellcode injection
    - Windows: finished
    - Linux: WIP
- Proxychaining beacons OR P2P beacons
- Automatic proxychains configuration depending on beacon route
- Add handling of multiple clients with random unique IDs: FINISHED
- RSA encrypt data instead of base 64 encode
- Randomly timed sending
- Web interface - WIP
