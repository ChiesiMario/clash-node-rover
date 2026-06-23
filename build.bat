@echo off
echo Building Clash Node Rover (Background Mode)...
go build -ldflags="-H windowsgui -s -w" -o clash-node-rover.exe
echo Done!
