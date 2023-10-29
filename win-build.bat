@echo off

set "psCommand=powershell -Command ""$pattern = '(?s)(UNAME_S = \$\(shell uname -s\).*?\r?\nendif)'; $content = Get-Content -Raw -Path 'Makefile'; $editedText = $content -replace $pattern, ''; Set-Content -Value $editedText -Path 'Makefile'; $pattern = 'BUILTTIME =.*'; $content = Get-Content -Raw -Path 'Makefile'; $editedText = $content -replace $pattern, 'BUILTTIME := $(shell powershell -Command "Get-Date -Format MM-dd-yy")'; Set-Content -Value $editedText -Path 'Makefile'; & '${MAKE}' build-windows-with-rclone;""

%psCommand%
