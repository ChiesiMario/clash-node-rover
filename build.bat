@echo off
echo Building Frontend...
cd frontend
call npm install
call npm run build
cd ..

echo Building Clash Node Rover (Background Mode)...
go build -ldflags="-H windowsgui -s -w" -o clash-node-rover.exe
echo Done!
