prevent windows auto sleep when receive wol magic package continuously

## build
```cmd
go build -o DoNotSleep.exe .
```

## create service



1. go to [winsw](https://github.com/winsw/winsw) and download release file `WinSW-x64.exe`
2. create directory `DoNotSleep`, move `WinSW-x64.exe` to here and rename to `DoNotSleep_Service.exe`
3. create xml file `DoNotSleep_Service.xml` and copy below example to this file

```xml
<service>
  
  <!-- ID of the service. It should be unique across the Windows system-->
  <id>DoNotSleep</id>
  <!-- Display name of the service -->
  <name>DoNotSleep</name>
  <!-- Service description -->
  <description>prevent windows auto sleep when receive wol magic package continuously</description>
  
  <!-- Path to the executable, which should be started -->
  <executable>C:\Program Files\DoNotSleep\DoNotSleep.exe</executable>
  <onfailure action="restart" delay="10 sec"/>
  <startmode>Automatic</startmode>
</service>
```

4. move go program `DoNotSleep.exe` to dir `DoNotSleep`
4. open powershell, type ` .\DoNotSleep_Service.exe install`
5. find service(taskManager -> open service) and start


Don't use sc.exe directly
~~move DoNotSleep.exe to where you want and don't delete it.~~

~~To create a Windows Service from an executable, you can use sc.exe:~~

```cmd
sc.exe create DoNotSleep binPath= "<path_to_the_service_executable>"
```
~~You must have quotation marks around the actual exe path, and a space after the binPath=.~~