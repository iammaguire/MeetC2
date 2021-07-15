It's great to see interest in the project growing! I've spent a lot of time on this but it has been only a hobby project until now. When I finish up the certification I'm currently working on I'll be cleaning the repo up and making usage more user friendly. 

https://user-images.githubusercontent.com/7650862/123324333-38011280-d526-11eb-8765-786f0fe7e1ee.mp4

https://user-images.githubusercontent.com/7650862/122882317-91e9b880-d32b-11eb-82da-da6bf98a7044.mp4

Usage:

The default HttpListener is instantiated with the first two command line arguments - the first being the interface to listen on and the second being the hostname.

To listen on wlp2s0 with command.com being the hostname, run
`./c2 wlp2s0 command.com`

Inside the command server you can reference beacons using either their list id or their unique id.
For example if the output of 'list' is '[0] LuTNluGL@10.255.52.1 last seen 2021-05-23 21:07:17.208101927 -0400 EDT m=+179.667762638', you could refer to this beacon by LuTNluGL or 0.

After selecting an active beacon with user shellcode may be injected into another process. List processes with plist and provide the shellcode command with the path to the data as well as the PID.

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
- Mimikatz support: finished
- Generic modules:
    - C#: WIP
    - Go: finished
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
