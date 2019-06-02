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
	io.WriteString(*P, "/save\n")
	io.WriteString(*P, "/quit\n")
	s.ChannelMessageSend(support.Config.FactorioChannelID, "Saved server, now restarting!")
	time.Sleep(3 * time.Second)
	*R = false
	RestartCount = RestartCount + 1
	return
}
