prevent windows auto sleep when receive wol magic package continuously

## build

```cmd
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o DoNotSleep.exe .
```

## Usage

open powershell as Administrator

`.\DoNotSleep.exe [install, start, stop, uninstall, run]`

1. `.\DoNotSleep.exe install`
2. `.\DoNotSleep.exe start`