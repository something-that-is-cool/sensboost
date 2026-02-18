if not exist app.syso (
    rsrc -manifest app.exe.manifest -o app.syso
)
go build -ldflags="-s -w -H=windowsgui" -o sensboost.exe .