package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/SlyMarbo/rss"
	"github.com/jasonlvhit/gocron"
	"strings"
	"io/ioutil"
	"encoding/json"
	"log"
	"io"
	"math/rand"
	"time"
)

// Variables used for command line parameters
var (
	Token string
)
const (
	PermissionCreateInstantInvite = 1 << iota
	PermissionKickMembers
	PermissionBanMembers
	PermissionAdministrator
	PermissionManageChannels
	PermissionManageServer
	PermissionAddReactions
	PermissionViewAuditLogs


	PermissionAllChannel =
		PermissionCreateInstantInvite |
		PermissionManageChannels |
		PermissionAddReactions |
		PermissionViewAuditLogs
	PermissionAll = PermissionAllChannel |
		PermissionKickMembers |
		PermissionBanMembers |
		PermissionManageServer |
		PermissionAdministrator
)

//json types
type Episodes struct {
	Episodes []Episode `json:"episode"`
}
type Episodes20 struct {
	Episodes [20]Episode `json:"episode"`
}
type Episode struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

func init() {
	flag.StringVar(&Token, "t", TokenBot, "Bot Token")
	flag.Parse()
}

func main() {

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	go mainCron(dg)
	dg.AddHandler(respondTo)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func editC(dg *discordgo.Session, chName string, chEp string) {

	jsonFile, err := os.Open("shows.json")

	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var episodes Episodes
	json.Unmarshal(byteValue, &episodes)

	if  strings.Compare(episodes.Episodes[0].Name, chName + " " + chEp) == 0 {
		fmt.Println("No Updates")
	} else {
		var create = new(discordgo.ChannelEdit)
		create.Name = chName + " " + chEp
		create.Position = -100
		create.ParentID = "471391284221837342"
		create.NSFW = false
		create.Topic = "Share your thoughts on " + chName + " Episode " + chEp + " here!"
		channel,err := dg.GuildChannelCreate("357256989853745152", chName + " " + chEp, "text")
		if err != nil {
			print(err)
		}
		dg.ChannelEditComplex(channel.ID, create)
		dg.ChannelMessageSend(channel.ID,  "**" + chName + " Episode " + chEp + "** has just been released. Share your thoughts on it here! ⚡⚡⚡")
		dg.ChannelMessageSend(channel.ID,  "*Source Code: https://github.com/TooFiveFive/ShinobuBot*")
		dg.ChannelMessageSend(channel.ID,  "/poll 'Did you Enjoy the Episode?'")
		fmt.Println(chName + " " + chEp)

		//delete 20th channel
		dg.ChannelDelete(episodes.Episodes[19].Id)


		var jsonWrite Episode
		jsonWrite.Name = chName + " " + chEp
		jsonWrite.Id = channel.ID

		var jAll [20]Episode
		jAll[0] = jsonWrite
		for i := 1; i < 20; i++ {
			jAll[i] = episodes.Episodes[i - 1]
		}
		var jsonAll Episodes20
		jsonAll.Episodes = jAll
		jData,_ := json.Marshal(jsonAll)
		fmt.Println(string(jData))
		file, err := os.OpenFile("shows.json",os.O_CREATE, 0666)
		if err != nil {
			log.Fatal("Cannot create file", err)
		}
		defer file.Close()

		io.WriteString(file, string(jData))

		//if other method doesn't work
		ioutil.WriteFile("shows.json", jData, 0644)
		dg.ChannelEditComplex(channel.ID, create)
	}

}

func mainCron(dg *discordgo.Session) {

	gocron.Every(30).Seconds().Do(func() {
		feed, err := rss.Fetch("http://horriblesubs.info/rss.php?res=1080")
		fmt.Println("Latest: " + feed.Items[0].Title)
		if err != nil {
			print("stop")
		}
		s := strings.Split(feed.Items[0].Title, "] ")
		ep := strings.Split(s[1], " - ")
		epn := ep[0]
		epi := strings.Split(ep[1], " [")[0]

		editC(dg, epn, epi)
	})
	_, timev := gocron.NextRun()
	fmt.Println(timev)


	<- gocron.Start()

}

type Insult struct {
	Insults []string `json:"insult"`
}

func respondTo(dg *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == dg.State.User.ID {
		return
	}

	var adminChannels = [...]string{"471445082600636428"}

	jsonFile, err := os.Open("insults.json")

	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var insults Insult
	json.Unmarshal(byteValue, &insults)


	if strings.HasPrefix(m.Content, "s!") {
		if strings.Contains(m.Content, "insult") && len(m.Mentions) == 1 {

			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)
			num := r1.Intn(len(insults.Insults))
			message := insults.Insults[num]

			dg.ChannelMessageSend(m.ChannelID, "**" + message + " " + m.Mentions[0].Mention() + "**")
		}
		if strings.Contains(m.Content, "add") {
			for _, element := range adminChannels {
				if m.ChannelID == element {
					if strings.Split(m.Content, " ")[1] == "insult" {

						add := strings.Split(m.Content, "> ")[1]
						insultsAdd := make([]string, len(insults.Insults) + 1)
						for ind := 0; ind < len(insults.Insults); ind++ {
							insultsAdd[ind] = insults.Insults[ind]
						}
						insultsAdd[len(insults.Insults)] = add
						insultsData := Insult{insultsAdd}

						jData,_ := json.Marshal(insultsData)
						fmt.Println(string(jData))
						file, err := os.OpenFile("insults.json",os.O_CREATE, 0666)
						if err != nil {
							log.Fatal("Cannot create file", err)
						}
						defer file.Close()

						io.WriteString(file, string(jData))

						//if other method doesn't work
						ioutil.WriteFile("insults.json", jData, 0644)

						dg.ChannelMessageSend(m.ChannelID, "I added *" + add + "* to my insult list you sick boi.")
					}
				}
			}
		}

		if strings.Contains(m.Content, "username") && !strings.Contains(m.Content, "add") {
			if strings.Contains(m.Content, "random") {
				s1 := rand.NewSource(time.Now().UnixNano())
				r1 := rand.New(s1)
				num := r1.Intn(len(UsernameRands))
				fmt.Println(num)
				username := UsernameRands[num]

				dg.GuildMemberNickname("357256989853745152", m.Author.ID, username)
			} else {
				if !strings.Contains(m.Content, "⚡") {
					dg.GuildMemberNickname("357256989853745152", m.Author.ID, strings.Split(m.Content, "> ")[1])
				}
			}

		}

		if strings.Contains(m.Content, "help") && !strings.Contains(m.Content, "add") {
			dg.ChannelMessageSend(m.ChannelID, "Type `s!` followed by:")
			dg.ChannelMessageSend(m.ChannelID, "- `usename` `random` OR `> custom name`")
			dg.ChannelMessageSend(m.ChannelID, "- `insult @user`")
		}


	}
}
