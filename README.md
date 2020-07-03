# i3-autoname-ws

Listens for [i3](https://github.com/i3/i3) events and renames workspaces to show the icons of running programs.

# Installation

- To build and install the binary, use:
  ```sh
  go install
  ```
- This uses [Nerd Fonts](https://github.com/ryanoasis/nerd-fonts) for icons, so make sure you are using one of them in your status bar.
- Add `i3-autoname-ws` to your i3 configuration to make it start up by default:
  ```
  exec --no-startup-id i3-autoname-ws
  ```

## Configuration

Unknown entries show up as a question mark by default. To add new entries:

- Use [xprop](https://gitlab.freedesktop.org/xorg/app/xprop) to find the class and instance of the window.
- Find an appropriate icon using the [Nerd Fonts Cheat Sheet](https://www.nerdfonts.com/cheat-sheet).
- Add it to the code and rebuild it!
