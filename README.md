![noun_Keyboard_3563522](https://user-images.githubusercontent.com/17187770/97167613-7d81fb80-175d-11eb-9f0c-2c335d666d85.png)


# kboard
Terminal game to practice keyboard typing. Built with [bubbletea](https://github.com/charmbracelet/bubbletea)

### Install
```bash
go get github.com/CamiloGarciaLaRotta/kboard
```

### How to use

```bash
kboard [number] [time]

number: the number of words to generate. Must be a non-zero positive integer.
        defaults to 1 word.
time:   the number of seconds that the game will last.
        If none is passed, tha game finishes after the first word.

Examples:
 - kboard 2
 - kboard 1 30
 ```

### Untimed mode
![untimed demo](https://user-images.githubusercontent.com/17187770/97325058-f611b680-1848-11eb-8b3b-d80660a8ded4.gif)

### Timed mode
![timed demo](https://user-images.githubusercontent.com/17187770/97336421-5e669500-1855-11eb-8d4c-1683a771c0e6.gif)

