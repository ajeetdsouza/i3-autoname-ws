#!/usr/bin/env python3

import collections
import i3ipc
import re
import signal
import sys
import subprocess
from typing import Optional, List

from config import DEFAULT_ICON
from config import ICONS

# A type that represents a parsed workspace "name"
NameParts = collections.namedtuple('NameParts', ['num', 'short_name', 'icons'])


def icon_for_window(con: i3ipc.Con) -> str:
    """
    Try all window classes and use the first one we have an icon for
    """
    wm_classes = xprop(con.window, 'WM_CLASS') or []

    for wm_class in wm_classes:
        try:
            return ICONS[wm_class.casefold()]
        except KeyError:
            pass

    print(f'No icon available for window with classes: {wm_classes}')
    return DEFAULT_ICON


def rename_workspaces(i3_conn: i3ipc.Connection, renumber_workspaces: bool = False) -> None:
    """
    Renames all workspaces based on the windows present.
    Renumbers them in ascending order, with one gap left between monitors.
    Eg. workspace numbering on two monitors: [1, 2, 3], [5, 6]
    """
    prev_output = None
    n = 1

    for ws_info, ws in zip(i3_conn.get_workspaces(), i3_conn.get_tree().workspaces()):
        name_parts = parse_workspace_name(ws.name)
        new_icons = " ".join([icon_for_window(w) for w in ws.leaves()]) + " "

        # As we enumerate, leave one gap in workspace numbers between each monitor.
        # This leaves a space to insert a new one later.
        if prev_output is not None and ws_info.output != prev_output:
            n += 1
        prev_output = ws_info.output

        # renumber workspace
        new_num = n
        n += 1

        new_name = construct_workspace_name(
            NameParts(num=new_num if renumber_workspaces else ws.num,
                      short_name=name_parts.short_name,
                      icons=new_icons)
        )

        if ws.name != new_name:
            i3_conn.command(f'rename workspace "{ws.name}" to "{new_name}"')


def on_exit(i3_conn: i3ipc.Connection) -> None:
    """
    Rename workspaces to just numbers and short_names, removing the icons.
    """
    for workspace in i3_conn.get_tree().workspaces():
        name_parts = parse_workspace_name(workspace.name)
        new_name = construct_workspace_name(
            NameParts(num=name_parts.num,
                      short_name=name_parts.short_name,
                      icons=None)
        )

        if workspace.name is not new_name:
            i3_conn.command(f'rename workspace "{workspace.name}" to "{new_name}"')

    i3_conn.main_quit()
    sys.exit(0)


def focused_workspace(i3_conn: i3ipc.Connection) -> i3ipc.WorkspaceReply:
    for ws in i3_conn.get_workspaces():
        if ws.focused:
            return ws


def parse_workspace_name(name: str) -> NameParts:
    """
    Takes a workspace name and splits it into three parts (set to None if not found):
    - num: the workspace number
    - short_name: the workspace name (assumed to have no spaces)
    - icons: the string that comes after the name
    """
    match = re.match(r'(?P<num>\d+):?(?P<short_name>\w+)? ?(?P<icons>.+)?', name).groupdict()
    return NameParts(**match)


def construct_workspace_name(parts: NameParts) -> str:
    """
    Given a NameParts object, returns the formatted name by concatenating them together.
    """
    new_name = str(parts.num)

    if parts.short_name or parts.icons:
        new_name += ':'

        if parts.short_name:
            new_name += parts.short_name

        if parts.icons:
            new_name += ' ' + parts.icons

    return new_name


def xprop(win_id, xproperty: str) -> Optional[List[str]]:
    """
    Return an array of values for the X property on the given window.
    Requires xorg-xprop to be installed.
    """
    try:
        prop = subprocess.check_output(
            ['xprop', '-id', str(win_id), xproperty],
            stderr=subprocess.DEVNULL
        ).decode('utf-8')

        return re.findall('"([^"]+)"', prop)

    except subprocess.CalledProcessError:
        print("Unable to get property for window '%d'" % win_id)
        return None


def main() -> None:
    i3_conn = i3ipc.Connection()

    # Exit gracefully when Ctrl+C is pressed
    for sig in (signal.SIGINT, signal.SIGTERM):
        signal.signal(sig, lambda *_: on_exit(i3_conn))

    rename_workspaces(i3_conn)

    def event_handler(i3_conn: i3ipc.Connection, evt: i3ipc.WindowEvent) -> None:
        """
        Call rename_workspaces() for relevant window events
        """
        if evt.change in ('new', 'close', 'move'):
            rename_workspaces(i3_conn)

    i3_conn.on('window', event_handler)
    i3_conn.on('workspace::move', event_handler)
    i3_conn.main()


if __name__ == '__main__':
    main()
