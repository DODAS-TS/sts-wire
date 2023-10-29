@echo off
setlocal

REM Bypass "Terminate Batch Job" prompt.
if "%~1"=="-FIXED_CTRL_C" (
   REM Remove the -FIXED_CTRL_C parameter
   SHIFT
) ELSE (
   REM Run the batch with <NUL and -FIXED_CTRL_C
   CALL <NUL %0 -FIXED_CTRL_C %*
   GOTO :EOF
)

for /f "delims=" %%A in ('git rev-parse --show-toplevel 2^>nul') do (
    set "ROOTDIR=%%A"
)

if "%ROOTDIR%" == "" (
    echo Error: Unable to retrieve the root directory from Git.
    exit /B 2
)

REM Execute sts-wire command with variables
"%ROOTDIR%\sts-wire_windows.exe" clean
IF EXIST "token" (
    del "token"
)
exit /B
