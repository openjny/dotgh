# dotgh Demo - VHS Recording Script
# Run: vhs .vhs/demo.tape

Output assets/demo.gif

# Terminal Settings
Set Shell "bash"
Set FontSize 14
Set Width 700
Set Height 350
Set Padding 10
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set PlaybackSpeed 1.5

# Start demo
Sleep 300ms

# 1. List available templates
Type "dotgh list"
Sleep 200ms
Enter
Sleep 1.5s

# 2. Pull template to current project
Type "dotgh pull web-app"
Sleep 200ms
Enter
Sleep 1.5s

# 3. Push as new template
Type "dotgh push my-new-template"
Sleep 200ms
Enter
Sleep 1.5s

# End
Sleep 500ms
