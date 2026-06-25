@echo off
set WAS_RUNNING=0

echo Checking if Clash Node Rover is running...
tasklist /FI "IMAGENAME eq clash-node-rover.exe" 2>NUL | find /I /N "clash-node-rover.exe">NUL
if "%ERRORLEVEL%"=="0" (
    echo Clash Node Rover is running. Stopping it...
    taskkill /F /IM clash-node-rover.exe >NUL 2>&1
    set WAS_RUNNING=1
)

echo Building Frontend...
cd frontend
call npm install
call npm run build
cd ..

echo Building Clash Node Rover (Background Mode)...
go build -ldflags="-H windowsgui -s -w" -o clash-node-rover.exe
echo Compilation Done!

if "%WAS_RUNNING%"=="1" (
    echo Restarting Clash Node Rover...
    start "" clash-node-rover.exe
    echo Restarted successfully!
)
