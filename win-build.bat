@echo off

set "psCommand=powershell -Command ""$env:CPATH='C:\Program Files (x86)\WinFsp\inc\fuse'; $env:INCLUDE='C:\Program Files\mingw64\include'; $env:PATH='C:\Program Files\mingw64\bin;C:\Program Files\mingw64\lib;' + $env:PATH; $pattern = '(?s)(UNAME_S = \$\(shell uname -s\).*?\r?\nendif)'; $content = Get-Content -Raw -Path 'Makefile'; $editedText = $content -replace $pattern, ''; Set-Content -Value $editedText -Path 'Makefile'; $pattern = 'BUILTTIME =.*'; $content = Get-Content -Raw -Path 'Makefile'; $editedText = $content -replace $pattern, 'BUILTTIME := $(shell powershell -Command "Get-Date -Format MM-dd-yy")'; Set-Content -Value $editedText -Path 'Makefile'; mingw32-make.exe build-windows-with-rclone;""

%psCommand%
