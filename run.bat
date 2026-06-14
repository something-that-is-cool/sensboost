for /f "tokens=*" %%i in ('git describe --tags --always') do set GIT_VERSION=%%i
for /f "tokens=*" %%i in ('git rev-parse --short HEAD') do set GIT_COMMIT=%%i
for /f "tokens=*" %%i in ('powershell -Command "[DateTimeOffset]::Now.ToUnixTimeSeconds()"') do set BUILD_TIME=%%i

set PKG=github.com/something-that-is-cool/zutil/internal/version
set LDFLAGS=-s -w -X '%PKG%.Version=%GIT_VERSION%' -X '%PKG%.Commit=%GIT_COMMIT%' -X '%PKG%.BuildTime=%BUILD_TIME%'

go run -ldflags="%LDFLAGS%" .