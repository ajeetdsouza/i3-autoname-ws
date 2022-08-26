package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.i3wm.org/i3/v4"
)

func renameAllWs() error {
	ws, err := i3.GetWorkspaces()
	if err != nil {
		return err
	}

	wsNums := make(map[i3.WorkspaceID]int64)
	for _, w := range ws {
		// ignore named workspaces
		if w.Num == -1 {
			continue
		}

		wsNums[w.ID] = w.Num
	}

	tree, err := i3.GetTree()
	if err != nil {
		return err
	}

	var cmdBuilder strings.Builder

	var helper func(*i3.Node)
	helper = func(node *i3.Node) {
		if wsNum, ok := wsNums[i3.WorkspaceID(node.ID)]; ok {
			oldName := node.Name
			newName := getWsName(node, wsNum)

			cmd := fmt.Sprintf(`rename workspace "%s" to "%s";`, oldName, newName)
			cmdBuilder.WriteString(cmd)

			return
		}

		for _, child := range node.Nodes {
			helper(child)
		}
	}
	helper(tree.Root)

	cmd := cmdBuilder.String()

	_, err = i3.RunCommand(cmd)
	return err
}

func resetWsName() error {
	ws, err := i3.GetWorkspaces()
	if err != nil {
		return err
	}

	var cmdBuilder strings.Builder

	for _, w := range ws {
		cmd := fmt.Sprintf(`rename workspace "%s" to "%d";`, w.Name, w.Num)
		cmdBuilder.WriteString(cmd)
	}

	cmd := cmdBuilder.String()

	_, err = i3.RunCommand(cmd)
	return err
}

func getWsName(node *i3.Node, wsNum int64) string {
	winIcons := make([]string, 0)

	var helper func(*i3.Node)
	helper = func(node *i3.Node) {
		for _, child := range node.Nodes {
			if len(child.Nodes) == 0 {
				winIcon := getWinIcon(&child.WindowProperties)
				winIcons = append(winIcons, winIcon)

				continue
			}

			helper(child)
		}
	}

	helper(node)
	for _, child := range node.FloatingNodes {
		helper(child)
	}

	if len(winIcons) == 0 {
		return fmt.Sprintf("%d", wsNum)
	}

	return fmt.Sprintf("%d: %s", wsNum, strings.Join(winIcons, ""))
}

func getWinIcon(prop *i3.WindowProperties) string {
	switch strings.ToLower(prop.Class) {
	case "arandr":
		return "\uf878"
	case "atril":
		return "\uf725"
	case "audacity":
		return "\uf025"
	case "blueberry.py":
		return "\uf293"
	case "caja", "thunar":
		return "\uf07b"
	case "code", "codium", "vscodium":
		return "\ue70c"
	case "firefox":
		return "\uf738"
	case "google-chrome", "chromium-browser":
		switch prop.Instance {
		// WhatsApp
		case "crx_hnpfjngllnobngcgfapefoaidbinmjnm":
			return "\ufaa2"
		default:
			return "\uf268"
		}
	case "io.github.celluloid_player.celluloid":
		return "\uf144"
	case "jetbrains-idea":
		return "\ue7b5"
	case "keepassxc":
		return "\uf43d"
	case "kitty", "mate-terminal", "xfce4-terminal":
		return "\ue795"
	case "mate-screenshot":
		return "\uf5ff"
	case "mate-volume-control":
		return "\uf9c2"
	case "obsidian":
		return "\uf249"
	case "pavucontrol":
		return "\uf9c2"
	case "postman":
		return "\uf14c"
	case "rosaimagewriter":
		return "\ufaed"
	case "slack":
		return "\uf198"
	case "spotify":
		return "\uf1bc"
	case "transmission-gtk", "qbittorrent":
		return "\uf019"
	case "virtualbox", "virtualbox machine", "virtualbox manager", "virtualboxvm":
		return "\uf6a6"
	case "vlc":
		return "\ufa7b"
	case "woeusbgui":
		return "\uf287"
	case "xarchiver":
		return "\uf187"
	case "zoom":
		return "\uf03d"
	default:
		log.Printf("no icon found for: %+v", prop)
		return "\uf128"
	}
}

func main() {
	// The script relies on workspace ID which was added in i3wm v4.18.0
	if err := i3.AtLeast(4, 18); err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 5)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if err := resetWsName(); err != nil {
			log.Fatalf("failed to clear all workspaces: %s", err)
		}
		os.Exit(1)
	}()

	if err := renameAllWs(); err != nil {
		log.Printf("failed to rename all workspaces: %s", err)
	}

	subscriber := i3.Subscribe(
		i3.WindowEventType,
		i3.WorkspaceEventType,
	)

	for subscriber.Next() {
		switch event := subscriber.Event().(type) {
		case *i3.WindowEvent:
			switch event.Change {
			case "new", "close", "move", "floating":
			default:
				continue
			}
		case *i3.WorkspaceEvent:
			switch event.Change {
			case "move":
			default:
				continue
			}
		default:
			log.Printf("received event of type: %T", event)
		}

		if err := renameAllWs(); err != nil {
			log.Fatal(err)
		}
	}
}
