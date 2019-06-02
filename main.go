package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"./commands"
	"./commands/admin"
	"./commands/utils"
	"./support"
	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

// Running is the boolean that tells if the server is running or not
var Running bool

// Pipe is an WriteCloser interface
var Pipe io.WriteCloser

// Session is a discordgo session
var Session *discordgo.Session

var UpdateCmd *int

func main() {
	support.Config.LoadEnv()
	Running = false
	admin.R = &Running
	admin.SaveResult = &support.SaveResult
	admin.QuitFlag = &support.QuitFlag
	UpdateCmd = &admin.UpdateCmd
	utils.UserList = &support.OnlineUserList
	utils.UserListResult = &support.UserListResult

	*UpdateCmd = 0

	// Do not exit the app on this error.
	if err := os.Remove("factorio.log"); err != nil {
		fmt.Println("Factorio.log doesn't exist, continuing anyway")
	}

	logging, err := os.OpenFile("factorio.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		support.ErrorLog(fmt.Errorf("%s: An error occurred when attempting to open factorio.log\nDetails: %s", time.Now(), err))
	}

	mwriter := io.MultiWriter(logging, os.Stdout)
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for {
			// If the process is already running DO NOT RUN IT AGAIN
			if !Running {
				if *UpdateCmd == 1 || *UpdateCmd == 3 { //factorio update
					if support.Config.UpdaterLocation == "" {
						Session.ChannelMessageSend(support.Config.FactorioChannelID, "factorio updater path not found.")
					} else {
						experimental := ""
						if *UpdateCmd == 3 {
							experimental = "x"
						}

						mod_param := fmt.Sprintf("%s -%sDa %s", support.Config.UpdaterLocation, experimental, support.Config.Executable)
						mod_params := strings.Split(mod_param, " ")

						cmd_mod := exec.Command("python3", mod_params...)
						cmd_mod.Stderr = os.Stderr
						cmd_mod.Stdout = mwriter

						err := cmd_mod.Run()
						fmt.Printf("%T", err)
						fmt.Println(err)

						if err != nil {
							if exitError, ok := err.(*exec.ExitError); ok {
								waitStatus := exitError.Sys().(syscall.WaitStatus)
								if waitStatus.ExitStatus() == 2 {
									Session.ChannelMessageSend(support.Config.FactorioChannelID, "There is no update. restarting server..")
								}
							} else {
								Session.ChannelMessageSend(support.Config.FactorioChannelID, "Factorio update failed. Restart server anyway..")
							}
						} else {
							Session.ChannelMessageSend(support.Config.FactorioChannelID, "Factorio updated successful! restarting server..")
						}
					}
				} else if *UpdateCmd == 2 { //mod update
					if support.Config.ModUpdaterLocation == "" {
						Session.ChannelMessageSend(support.Config.FactorioChannelID, "mods updater path not found.")
					} else {
						departed := strings.Split(support.Config.Executable, "/")
						pathsize := len(departed)
						test := strings.Join(departed[:pathsize-3], "/")

						mod_param := fmt.Sprintf("-p %s -s %s", test, support.Config.ModUpdaterServerSetting)
						mod_params := strings.Split(mod_param, " ")

						cmd_mod := exec.Command(support.Config.ModUpdaterLocation, mod_params...)
						cmd_mod.Stderr = os.Stderr
						cmd_mod.Stdout = mwriter

						err := cmd_mod.Run()
						if err != nil {
							Session.ChannelMessageSend(support.Config.FactorioChannelID, "mods update failed. Start server anyway..")
						} else {
							Session.ChannelMessageSend(support.Config.FactorioChannelID, "mods updated successful! starting server..")
						}
					}
				}
				Running = true
				cmd := exec.Command(support.Config.Executable, support.Config.LaunchParameters...)
				cmd.Stderr = os.Stderr
				cmd.Stdout = mwriter
				Pipe, err = cmd.StdinPipe()
				if err != nil {
					support.ErrorLog(fmt.Errorf("%s: An error occurred when attempting to execute cmd.StdinPipe()\nDetails: %s", time.Now(), err))
				}

				err := cmd.Start()

				if err != nil {
					support.ErrorLog(fmt.Errorf("%s: An error occurred when attempting to start the server\nDetails: %s", time.Now(), err))
				}
				if admin.RestartCount > 0 {
					time.Sleep(3 * time.Second)
					Session.ChannelMessageSend(support.Config.FactorioChannelID,
						"Server restarted successfully!")
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		Console := bufio.NewReader(os.Stdin)
		for {
			if *UpdateCmd == 0 {
				line, _, err := Console.ReadLine()
				if err != nil {
					support.ErrorLog(fmt.Errorf("%s: An error occurred when attempting to read the input to pass as input to the console\nDetails: %s", time.Now(), err))
				}
				_, err = io.WriteString(Pipe, fmt.Sprintf("%s\n", line))
				if err != nil {
					support.ErrorLog(fmt.Errorf("%s: An error occurred when attempting to pass input to the console\nDetails: %s", time.Now(), err))
				}
			}

		}
	}()

	go func() {
		// Wait 10 seconds on start up before continuing
		time.Sleep(10 * time.Second)

		for {
			support.CacheDiscordMembers(Session)
			//sleep for 4 hours (caches every 4 hours)
			time.Sleep(4 * time.Hour)
		}
	}()
	discord()
}

func discord() {
	// No hard coding the token }:<
	discordToken := support.Config.DiscordToken
	commands.RegisterCommands()
	admin.P = &Pipe
	utils.P = &Pipe
	fmt.Println("Starting bot..")
	bot, err := discordgo.New("Bot " + discordToken)
	Session = bot
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		support.ErrorLog(fmt.Errorf("%s: An error occurred when attempting to create the Discord session\nDetails: %s", time.Now(), err))
		return
	}

	err = bot.Open()

	if err != nil {
		fmt.Println("error opening connection,", err)
		support.ErrorLog(fmt.Errorf("%s: An error occurred when attempting to connect to Discord\nDetails: %s", time.Now(), err))
		return
	}

	bot.AddHandler(messageCreate)
	bot.AddHandlerOnce(support.Chat)
	bot.ChannelMessageSend(support.Config.FactorioChannelID, "The server is booting...")
	bot.UpdateStatus(0, support.Config.GameName)
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	bot.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Print("[" + m.Author.Username + "] " + m.Content)

	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.ChannelID == support.Config.FactorioChannelID {
		if strings.HasPrefix(m.Content, support.Config.Prefix) {
			command := strings.Split(m.Content[1:len(m.Content)], " ")
			name := strings.ToLower(command[0])
			commands.RunCommand(name, s, m)
			return
		}
		// Pipes normal chat allowing it to be seen ingame
		if *UpdateCmd == 0 {
			_, err := io.WriteString(Pipe, fmt.Sprintf("[Discord] <%s>: %s\r\n", m.Author.Username, m.ContentWithMentionsReplaced()))
			if err != nil {
				support.ErrorLog(fmt.Errorf("%s: An error occurred when attempting to pass Discord chat to in-game\nDetails: %s", time.Now(), err))
			}
		}
		return
	}
}
