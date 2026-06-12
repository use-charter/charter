# Charter demo — VHS tape (charmbracelet/vhs). Render with:
#   vhs docs/internal/demo/charter-demo.tape
# Produces docs/internal/demo/charter-demo.gif for the README / launch.
#
# Requires a built `charter` on PATH (e.g. `go build -o /tmp/charter ./cmd/charter`).

Output docs/internal/demo/charter-demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 700
Set Padding 24
Set Theme "Catppuccin Mocha"

Type "charter doctor --path ."
Sleep 500ms
Enter
Sleep 4s

Type "charter explain AE-MCP-001"
Sleep 500ms
Enter
Sleep 3s

Type "charter report --format html --out charter-report.html"
Sleep 500ms
Enter
Sleep 2s
