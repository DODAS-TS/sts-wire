@echo off

set "issuer="
set "client_name="
set "pw_file="

:parseArgs
if "%~1" == "" goto :checkArgs

if /I "%~1" == "-h" (
    call :usage
    exit /B
)

if /I "%~1" == "-i" (
    set "issuer=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1" == "--issuer" (
    set "issuer=%2"
    shift
    shift
    goto :parseArgs
)

if /I "%~1" == "-c" (
    set "client_name=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1" == "--client-name" (
    set "client_name=%2"
    shift
    shift
    goto :parseArgs
)

if /I "%~1" == "-p" (
    set "pw_file=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1" == "--pw-file" (
    set "pw_file=%2"
    shift
    shift
    goto :parseArgs
)

goto :parseArgs

:checkArgs
if "%issuer%" == "" goto :missingArgs
if "%client_name%" == "" goto :missingArgs
if "%pw_file%" == "" goto :missingArgs

if not exist "%pw_file%" (
    echo Error: The specified pw-file '%pw_file%' does not exist.
    exit /B 1
)

oidc-gen --pw-file="%pw_file%" --scope-all --confirm-default --iss="%issuer%" --flow=device "%client_name%"
exit /B

:missingArgs
echo Error: Issuer, client name, and pw-file must be provided.
exit /B 1

:usage
echo Usage: %~n0 -i ISSUER -c CLIENT_NAME -p PW_FILE
echo   -i, --issuer       Specify the issuer URL (e.g., https://iam.example.com)
echo   -c, --client-name  Specify the client name (e.g., iam-client)
echo   -p, --pw-file      Specify the password file
echo   -h, --help         Display this help message
exit /B 1
