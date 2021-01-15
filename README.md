# radiko-server

Unofficial radiko recording server.

**Please do not make the server publicly available.  Use just for private.**

## Features

- Record radio programs by keywords
- Simple web UI to play the recorded radio programs
- Download timeshift program by URL

## How to use

### Run

```sh
$ docker run --rm -v $(pwd)/data:/data -p 8080:8080 uphy/radiko-server -data /data
```

### Registering keywords

Edit the `keywords.json` file located on your data directory after the initial server boot.  
After edit it, please restart server.