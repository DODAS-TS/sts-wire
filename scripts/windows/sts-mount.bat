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

REM Initialize variables
set "PW_FILE="
set "CLIENT_NAME="
set "IAM_URL="
set "RGW_URL="
set "ROLE="
set "AUDIENCE="
set "BUCKET="
set "MOUNTPOINT="

:parseArgs
if "%~1"=="" goto :executeScript

if /I "%~1"== "-p" (
    set "PW_FILE=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "--pw-file" (
    set "PW_FILE=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "-c" (
    set "CLIENT_NAME=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "--client-name" (
    set "CLIENT_NAME=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "-i" (
    set "IAM_URL=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "--iam-url" (
    set "IAM_URL=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "-r" (
    set "RGW_URL=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "--rgw-url" (
    set "RGW_URL=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "-o" (
    set "ROLE=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "--role" (
    set "ROLE=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "-a" (
    set "AUDIENCE=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "--audience" (
    set "AUDIENCE=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "-b" (
    set "BUCKET=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "--bucket" (
    set "BUCKET=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "-m" (
    set "MOUNTPOINT=%2"
    shift
    shift
    goto :parseArgs
)
if /I "%~1"== "--mountpoint" (
    set "MOUNTPOINT=%2"
    shift
    shift
    goto :parseArgs
)
shift
goto :parseArgs

:executeScript
REM Verify required parameters
if "%PW_FILE%"=="" if "%CLIENT_NAME%"=="" if "%IAM_URL%"=="" if "%RGW_URL%"=="" if "%ROLE%"=="" if "%AUDIENCE%"=="" if "%BUCKET%"=="" if "%MOUNTPOINT%"=="" goto :usage

REM Read the temporary file content and set the value to a variable
for /f "delims=" %%A in ('powershell -Command "oidc-gen -p %CLIENT_NAME% --pw-file='%PW_FILE%' | ConvertFrom-Json | Select -ExpandProperty client_id"') do (
    set "IAM_CLIENT_ID=%%A"
)

for /f "delims=" %%B in ('powershell -Command "oidc-gen -p %CLIENT_NAME% --pw-file='%PW_FILE%' | ConvertFrom-Json | Select -ExpandProperty client_secret"') do (
    set "IAM_CLIENT_SECRET=%%B"
)

for /f "delims=" %%C in ('powershell -Command "oidc-gen -p %CLIENT_NAME% --pw-file='%PW_FILE%' | ConvertFrom-Json | Select -ExpandProperty refresh_token"') do (
    set "REFRESH_TOKEN=%%C"
)

oidc-add --pw-file="%PW_FILE%" %CLIENT_NAME%
oidc-token %CLIENT_NAME% --aud=%AUDIENCE% > "token"
set /p ACCESS_TOKEN=<"token"

REM Optional delay (emulated)
timeout /t 5 >nul

REM Execute sts-wire command with variables
"%ROOTDIR%\sts-wire_windows.exe" "%IAM_URL%" myRGW "%RGW_URL%" "%ROLE%" "%AUDIENCE%" "%BUCKET%" "%MOUNTPOINT%" --tryRemount --noDummyFileCheck
exit /B

:usage
echo Usage: %0 -p PW_FILE -c CLIENT_NAME -i IAM_URL -r RGW_URL -o ROLE -a AUDIENCE -b BUCKET -m MOUNTPOINT
echo   -p, --pw-file       OIDC password file
echo   -c, --client-name   OIDC client name
echo   -i, --iam-url       IAM URL (e.g., https://iam.example.com/)
echo   -r, --rgw-url       RGW URL (e.g., https://rgw.example.com/)
echo   -o, --role          Role (e.g., S3AccessIAMRole)
echo   -a, --audience      Audience (e.g., https://wlcg.cern.ch/jwt/v1/any)
echo   -b, --bucket        Bucket (e.g., /bucket)
echo   -m, --mountpoint    Mountpoint (e.g., ./mountpoint)
exit /B 1
