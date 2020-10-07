Copyright © 2020 CIS Maxwell, LLC. All rights reserved.
Copyright © 2020 The Calyx Institute

Build:
Install Go on your machine https://golang.org/doc/install

  On Bash:
    GOPATH="path-to-flasher-source" GOOS=[darwin|linux|windows] GOARCH=amd64 go build -tags [release|debug|parallel|debug,parallel] -ldflags "-X main.version="your-version"" -o [device|parallel]-flasher[-debug].[darwin|linux|exe]
  On Cmd:
    SET GOPATH="path-to-flasher-source"
    SET GOOS=[darwin|linux|windows]
    SET GOARCH=amd64
    go build -tags [release|debug|parallel|debug,parallel] -ldflags "-X main.version="your-version"" -o [device|parallel]-flasher[-debug].[darwin|linux|exe]
  On PowerShell:
    $Env:GOPATH="path-to-flasher-source"; $Env:GOOS = "[darwin|linux|windows]"; $Env:GOARCH = "amd64"; go build -tags [release|debug|parallel|debug,parallel] -ldflags "-X main.version="your-version"" -o [device|parallel]-flasher[-debug].[darwin|linux|exe]
  Via Make:
    make [[device|parallel]-flasher[-debug].[darwin|linux|exe]]

Execution:
Plug each device of a same model to a USB port

The following files must be available in the current directory:
    CalyxOS factory image

 On Windows:
    Double-click on CalyxOS-flasher_windows.exe (will not show error output)
    or 
    Open PowerShell or Command Line
    Type: .\[[device|parallel]-flasher[-debug].exe]
    Press enter
 On Linux:
    Open a terminal in the current directory
    Type: sudo ./[[device|parallel]-flasher[-debug].linux]
    Press enter
 On Mac:
    Open a terminal in the current directory
    Type: ./[[device|parallel]-flasher[-debug].darwin]
    Press enter