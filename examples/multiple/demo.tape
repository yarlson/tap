Output ../../assets/demo.gif

Set Shell zsh
Set FontSize 28
Set Width 1400
Set Height 800
Set TypingSpeed 100ms

Sleep 1s
Type "go run ./main.go"
Enter
Sleep 2s

# First prompt: Enter name
Type "Alice"
Enter
Sleep 1.5s

# Second prompt: Programming language (accept default "Go")
Enter
Sleep 1.5s

# Third prompt: Select project type (navigate to "Web Application")
Down
Sleep 0.5s
Enter
Sleep 1.5s

# Fourth prompt: Years of experience (change from default)
Type "none"
Enter
Sleep 3s
Backspace
Backspace
Backspace
Backspace
Type "3"
Enter
Sleep 1.5s

# Fifth prompt: Show summary (accept default "Yes")
Enter
Sleep 1.5s

# Sixth prompt: Detailed format (change to "Yes")
Right
Sleep 0.5s
Enter
Sleep 3s

# Let the summary display for a moment
Sleep 5s
