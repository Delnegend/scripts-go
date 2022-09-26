@echo off
mkdir %1
cd %1
@REM create a file name $1 and write "package" + %1 to it
echo package main> %1.go
go mod init %1
cd ..
go work use %1