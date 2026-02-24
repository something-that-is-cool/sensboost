if not exist app.syso (
    rsrc -manifest app.exe.manifest -o app.syso -ico icon.ico
)
go build -ldflags="-s -w -H=windowsgui" -tags no_emoji -o zutil.exe .