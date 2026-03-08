Useful commands which I better record here so I don't forget later

We are using mediamtx as servermfor camera video streaming over rtsp:
(if not installed you should first)

```bash
./mediamtx 
```
now we start streaming using rpicam-vid (rotation is needed in my current setup, not needed if your orientation is the default):
```bash
rpicam-vid -t 0 --inline --rotation 180 --codec h264 -o - | ffmpeg -re -i - -c copy -f rtsp rtsp://localhost:8554/stream
```

You can visualize the streaming by opening VLC or similar and :
```bash
rtsp://192.168.1.18:8554/stream
```


Compile for RPI 4 boot:
```bash
GOOS=linux GOARCH=arm GOARM=7 go build -o boot .
```
Sync binay from local to remote:
```bash
rsync -av boot boot@192.168.1.18:/home/boot/   
```
Note: make sure to edit the IP address of your robot as needed.
