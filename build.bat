:: This script builds the server.
@ECHO OFF

SET build_dir=_build_
SET exe_dir=cmd
SET server_dir=proxy
SET settings_file=whitelist.txt

MKDIR "%build_dir%"

:: Build the server.
CD "%exe_dir%\%server_dir%"
go build
MOVE "%server_dir%.exe" ".\..\..\%build_dir%\"
CD ".\..\..\"

:: Copy some additional files for the server.
COPY "%exe_dir%\%server_dir%\%settings_file%" "%build_dir%\"
