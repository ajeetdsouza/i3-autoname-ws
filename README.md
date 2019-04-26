# i3-autoname-ws

Original script by [justbuchanan](http://github.com/justbuchanan/i3scripts)

This script listens for i3 events and updates workspace names to show icons for running programs. It contains icons for a few programs, but more can easily be added by adding them to icons below.
It also re-numbers workspaces in ascending order with one skipped number between monitors (leaving a gap for a new workspace to be created). By default, i3 workspace numbers are sticky, so they quickly get out of order.

## Dependencies

- xorg-xprop - install through system package manager
- i3ipc      - install with pip
- nerdfonts  - install with pip

## Configuration

The default i3 config's keybindings reference workspaces by name, which is an issue when using this script because the "names" are constantaly changing to include window icons. Instead, you'll need to change the keybindings to reference workspaces by number.
Change lines like: `bindsym $mod+1 workspace 1` to `bindsym $mod+1 workspace number 1`.

Add icons for common programs you use to `config.py`. The keys are the X window class `WM_CLASS` names (lowercase) and the icons can be any text you want to display.

If you're not sure what the WM_CLASS is for your application, you can use [`xprop`](https://linux.die.net/man/1/xprop). Run `xprop | grep WM_CLASS` then click on the application you want to inspect.

The icon character codes can be found on the [Nerd Font cheatsheet](https://nerdfonts.com)
