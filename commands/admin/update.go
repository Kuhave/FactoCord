package admin

import (
	"io"
	"time"

	"../../support"
	"github.com/bwmarrin/discordgo"
)

// Restart saves and restarts the server
func Update(s *discordgo.Session, m *discordgo.MessageCreate) {
	if *R == false {
		s.ChannelMessageSend(support.Config.FactorioChannelID, "Server is not running!")
		return
	}

	s.ChannelMessageSend(support.Config.FactorioChannelID, "Server received factorio client update command.")
	*QuitFlag = 1
	io.WriteString(*P, "/quit\n")
	time.Sleep(600 * time.Millisecond)
	for {
		if *QuitFlag == 2 {
			s.ChannelMessageSend(support.Config.FactorioChannelID, "server is closed.")
			*QuitFlag = 0
			break
		}
	}

	*R = false
	UpdateCmd = 1

	return
}

func UpdateExp(s *discordgo.Session, m *discordgo.MessageCreate) {
	if *R == false {
		s.ChannelMessageSend(support.Config.FactorioChannelID, "Server is not running!")
		return
	}

	s.ChannelMessageSend(support.Config.FactorioChannelID, "Server received factorio client experimental update command.")
	*QuitFlag = 1
	io.WriteString(*P, "/quit\n")
	time.Sleep(600 * time.Millisecond)
	for {
		if *QuitFlag == 2 {
			s.ChannelMessageSend(support.Config.FactorioChannelID, "server is closed.")
			*QuitFlag = 0
			break
		}
	}

	*R = false
	UpdateCmd = 3

	return
}
