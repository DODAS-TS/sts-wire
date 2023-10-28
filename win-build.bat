@echo off

set "psCommand=powershell -Command ""$pattern = '(?s)(UNAME_S = \$\(shell uname -s\).*?\r?\nendif)'; $content = Get-Content -Raw -Path 'Makefile'; $editedText = $content -replace $pattern, ''; Set-Content -Value $editedText -Path 'Makefile'; mingw32-make.exe build-windows-with-rclone;""

%psCommand%
