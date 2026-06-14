# Charter demo — VHS tape (charmbracelet/vhs ≥ 0.11).
#
# Render:
#   1. build the binary onto PATH:   go build -o /tmp/charter ./cmd/charter
#   2. stage the demo repo:          bash docs/internal/demo/make-fixture.sh
#   3. render:                       vhs docs/internal/demo/charter-demo.tape
#
# Produces docs/internal/demo/charter-demo.gif — the README hero.
# The look mirrors the marketing site: a floating terminal pane on the
# brand-dark backdrop, colourful window bar, rounded corners, blue prompt.

Output docs/internal/demo/charter-demo.gif

Set Shell "bash"
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 21
Set Width 1480
Set Height 920
Set Padding 36
Set Margin 56
Set MarginFill "#06080c"
Set BorderRadius 16
Set WindowBar Colorful
Set WindowBarSize 44
Set TypingSpeed 55ms
Set CursorBlink true
Set Theme { "name": "charter", "background": "#0b0e14", "foreground": "#eaedf3", "cursor": "#5aa2f7", "selection": "#243348", "black": "#0b0e14", "red": "#ec6464", "green": "#45dd92", "yellow": "#e8b045", "blue": "#5aa2f7", "magenta": "#a78bfa", "cyan": "#5aa2f7", "white": "#eaedf3", "brightBlack": "#59626f", "brightRed": "#ec6464", "brightGreen": "#45dd92", "brightYellow": "#e8b045", "brightBlue": "#7cb8ff", "brightMagenta": "#a78bfa", "brightCyan": "#7cb8ff", "brightWhite": "#ffffff" }

# --- hidden setup: PATH, demo repo, brand prompt, clear ---
Hide
Type "export PATH=/tmp:$PATH; cd /tmp/acme-api" Enter
Type "export PS1=$'\e[1;38;2;90;162;247macme-api\e[0m \e[38;2;124;184;255m❯\e[0m '" Enter
Type "clear" Enter
Show

Sleep 800ms

# 1 — measure
Type "charter doctor" Sleep 450ms Enter
Sleep 4.5s

# 2 — understand a finding
Type "charter explain AE-MCP-001" Sleep 450ms Enter
Sleep 3.5s

# 3 — offline, self-contained report
Type "charter report --format html --out charter-report.html" Sleep 450ms Enter
Sleep 2.8s

Sleep 1.2s
