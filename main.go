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
	"strconv"
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

var Guilds []string

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
	dg.UpdateStatus(0, "Say s!help")
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
		dg.ChannelMessageSend(channel.ID,  "-kitsu anime " + chName)
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

	gocron.Every(10).Seconds().Do(func() {

		fmt.Println("Checked. part of: " + strconv.Itoa(len(dg.State.Guilds)) + " Guilds")
		guilds := make([]string,len(dg.State.Guilds))
		for ind, guild := range dg.State.Guilds {
			guilds[ind] = guild.ID
		}
		Guilds = guilds
	})
	_, timev := gocron.NextRun()
	fmt.Println(timev)

	<- gocron.Start()

}

type Insult struct {
	Insults []string `json:"insult"`
}
type Profile struct {
	Bio string `json:"bio"`
	FavouriteAnime []string `json:"anime"`
	FavouriteManga []string `json:"manga"`
	Links []string `json:"links"`
}
var Timing = false
var CancelTimer = false

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

			num := rand.Intn(len(insults.Insults))
			message := insults.Insults[num]

			dg.ChannelMessageSend(m.ChannelID, "**" + message + " " + m.Mentions[0].Mention() + "**")
		}
		if strings.Contains(m.Content, "add") {
			for _, element := range adminChannels {
				if m.ChannelID == element && strings.Contains(m.Content, "> ") {
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
				} else {
					dg.ChannelMessageSend(m.ChannelID, "*You don't have permission to do that here*")
				}
			}
		}
		if strings.Contains(m.Content, "delete") {

			for _, element := range adminChannels {
				if m.ChannelID == element {
					if strings.Split(m.Content, " ")[1] == "insult" && strings.Contains(m.Content, "> ") {
						deleted := strings.Split(m.Content, "> ")[1]
						var index = -1
						for i,insult := range insults.Insults {
							if deleted == insult {
								index = i
							}
						}
						if index != -1 {
							arrayNew := remove(insults.Insults, index)
							var insultsNew Insult
							insultsNew.Insults = arrayNew

							jData,_ := json.Marshal(insultsNew)
							fmt.Println(string(jData))
							file, err := os.OpenFile("insults.json",os.O_CREATE, 0666)
							if err != nil {
								log.Fatal("Cannot create file", err)
							}
							defer file.Close()

							io.WriteString(file, string(jData))

							//if other method doesn't work
							ioutil.WriteFile("insults.json", jData, 0644)
							dg.ChannelMessageSend(m.ChannelID, "*" + deleted + "* removed from insults list")
						} else {
							dg.ChannelMessageSend(m.ChannelID, "*" + deleted + "* not found")
						}
					}
				} else {
					dg.ChannelMessageSend(m.ChannelID, "*You don't have permission to do that here*")
				}
			}

		}
		if strings.Contains(m.Content, "list") && !strings.Contains(m.Content, "add") && strings.Contains(m.Content, " "){
			if strings.Split(m.Content, " ")[1] == "insults" {
				var insultString string
				for _,insult := range insults.Insults {
					insultString += insult + "  👉  "
				}
				dg.ChannelMessageSend(m.ChannelID, "**List of insults:**" )
				dg.ChannelMessageSend(m.ChannelID, insultString)
			}
		}


		if strings.Contains(m.Content, "profile") && !strings.Contains(m.Content, "add") {
			if strings.Contains(m.Content, "edit") {
				profile, notExist := os.Open("./profiles/" + m.Author.ID + ".json")
				var index = 0
				var take = 0

				if notExist != nil {
					dg.ChannelMessageSend(m.ChannelID, "No profile found. Type *s!profile* to create it.")
				} else {
					var profileObj Profile
					profileByte, _ := ioutil.ReadAll(profile)

					json.Unmarshal(profileByte, &profileObj)
					if strings.Contains(m.Content, "> ") && len(strings.Split(m.Content, "> ")) >= 1 {
						if len(strings.Split(strings.Split(m.Content, "> ")[0], " ")) >= 4 {
							intex, err := strconv.Atoi(strings.Split(m.Content, " ")[3])
							if err != nil {
								index = 0
							} else {
								index = intex
								take = 1
							}
						}
						if strings.Contains(strings.Split(m.Content, " ")[2], "bio") {
							add := strings.Split(m.Content, "> ")[1]
							profileObj.Bio = add
							profileData, err := json.Marshal(profileObj)
							if err != nil {
								log.Fatal(err)
							}
							ioutil.WriteFile("./profiles/" + m.Author.ID + ".json", profileData, 0644)
							defer profile.Close()
							dg.ChannelMessageSend(m.ChannelID, "*Changes successful!*")
						}
						if strings.Contains(strings.Split(m.Content, " ")[2], "favAnime") {
							add := strings.Split(m.Content, "> ")[1]
							if index - 1 >= len(profileObj.FavouriteAnime) {
								dg.ChannelMessageSend(m.ChannelID, "*Index out of range*")
								return
							}
							if len(profileObj.FavouriteAnime) == 10 {
								dg.ChannelMessageSend(m.ChannelID, "*Max of 10 entries allowed*")
								return
							}

							addArray := make([]string, len(profileObj.FavouriteAnime) + 1 - take)
							for i := range profileObj.FavouriteAnime {
								addArray[i] = profileObj.FavouriteAnime[i]
							}
							if index == 0 {
								addArray[len(addArray) - 1] = add
							} else {
								addArray[index - 1] = add
							}

							profileObj.FavouriteAnime = addArray
							profileData, err := json.Marshal(profileObj)
							if err != nil {
								log.Fatal(err)
							}
							ioutil.WriteFile("./profiles/" + m.Author.ID + ".json", profileData, 0644)
							defer profile.Close()
							dg.ChannelMessageSend(m.ChannelID, "*Changes successful!*")
						}
						if strings.Contains(strings.Split(m.Content, " ")[2], "favManga") {
							add := strings.Split(m.Content, "> ")[1]
							if index - 1 >= len(profileObj.FavouriteManga) {
								dg.ChannelMessageSend(m.ChannelID, "*Index out of range*")
								return
							}
							if len(profileObj.FavouriteManga) == 10 {
								dg.ChannelMessageSend(m.ChannelID, "*Max of 10 entries allowed*")
								return
							}

							addArray := make([]string, len(profileObj.FavouriteManga) + 1 - take)
							for i := range profileObj.FavouriteManga {
								addArray[i] = profileObj.FavouriteManga[i]
							}
							if index == 0 {
								addArray[len(addArray) - 1] = add
							} else {
								addArray[index - 1] = add
							}

							profileObj.FavouriteManga = addArray
							profileData, err := json.Marshal(profileObj)
							if err != nil {
								log.Fatal(err)
							}
							ioutil.WriteFile("./profiles/" + m.Author.ID + ".json", profileData, 0644)
							defer profile.Close()
							dg.ChannelMessageSend(m.ChannelID, "*Changes successful!*")
						}
						if strings.Contains(strings.Split(m.Content, " ")[2], "link") {
							add := "<" + strings.Split(m.Content, "> ")[1] + ">"
							if index - 1 >= len(profileObj.Links) {
								dg.ChannelMessageSend(m.ChannelID, "*Index out of range*")
								return
							}
							if len(profileObj.Links) == 10 {
								dg.ChannelMessageSend(m.ChannelID, "*Max of 10 entries allowed*")
								return
							}

							addArray := make([]string, len(profileObj.Links) + 1 - take)
							for i := range profileObj.Links {
								addArray[i] = profileObj.Links[i]
							}
							if index == 0 {
								addArray[len(addArray) - 1] = add
							} else {
								addArray[index - 1] = add
							}

							profileObj.Links = addArray
							profileData, err := json.Marshal(profileObj)
							if err != nil {
								log.Fatal(err)
							}
							ioutil.WriteFile("./profiles/" + m.Author.ID + ".json", profileData, 0644)
							defer profile.Close()
							dg.ChannelMessageSend(m.ChannelID, "*Changes successful!*")
						}
					} else {
						dg.ChannelMessageSend(m.ChannelID, "**Type:** `s!profile edit | bio / favAnime / favManga / link | entry no / leave blank | > entry`")
						dg.ChannelMessageSend(m.ChannelID, "You can add more than one favAnime, favManga or link")
					}
				}
			} else {
				if len(m.Mentions) > 0 {
					profile, notExist := os.Open("./profiles/" + m.Mentions[0].ID + ".json")
					if notExist != nil {
						dg.ChannelMessageSend(m.ChannelID, "*Can't find that profile*")
					} else {
						defer profile.Close()

						profileValue, _ := ioutil.ReadAll(profile)

						var profileVar Profile
						json.Unmarshal(profileValue, &profileVar)
						dg.ChannelMessageSend(m.ChannelID, "***Profile of: ***" + m.Mentions[0].Mention() + ":" + `
----------
` +"**Bio: **" + profileVar.Bio + `
` + "**Favourite Anime:** " + strings.Join(profileVar.FavouriteAnime, " | ") + `
` + "**Favourite Manga:** " + strings.Join(profileVar.FavouriteManga, " | ") + `
` + "**Links:** " + strings.Join(profileVar.Links, " | "))
					}
				} else {
					profile, notExist := os.Open("./profiles/" + m.Author.ID + ".json")

					if notExist != nil {
						created, err := os.Create("./profiles/" + m.Author.ID + ".json")
						var empty Profile
						emptydata, err := json.Marshal(empty)
						if err != nil {
							log.Fatal(err)
						}
						defer created.Close()
						ioutil.WriteFile("./profiles/" + m.Author.ID + ".json", emptydata, 0644)

						fmt.Println("Created " +  m.Author.ID + ".json")
						dg.ChannelMessageSend(m.ChannelID, "Created your profile. Type *s!profile edit* to edit it")
					} else {
						defer profile.Close()

						profileValue, _ := ioutil.ReadAll(profile)

						var profileVar Profile
						json.Unmarshal(profileValue, &profileVar)
						dg.ChannelMessageSend(m.ChannelID, "***Profile of: ***" + m.Author.Mention() + ":" + `
----------
` +"**Bio: **" + profileVar.Bio + `
` + "**Favourite Anime:** " + strings.Join(profileVar.FavouriteAnime, " | ") + `
` + "**Favourite Manga:** " + strings.Join(profileVar.FavouriteManga, " | ") + `
` + "**Links:** " + strings.Join(profileVar.Links, " | "))
					}

				}

			}
		}


		if strings.Contains(m.Content, "username") && !strings.Contains(m.Content, "add") {
			if strings.Contains(m.Content, "random") {
				num := rand.Intn(len(UsernameRands))
				fmt.Println(num)
				username := UsernameRands[num]

				for ind := range Guilds {
					fmt.Println(Guilds[ind])
					dg.GuildMemberNickname(Guilds[ind], m.Author.ID, username)
				}

			} else {
				if !strings.Contains(m.Content, "⚡") && strings.Contains(m.Content, "> ") {
					for ind := range Guilds {
						fmt.Println(Guilds[ind])
						dg.GuildMemberNickname(Guilds[ind], m.Author.ID, strings.Split(m.Content, "> ")[1])
					}
				} else {
					dg.ChannelMessageSend(m.ChannelID, "*Parameters are missing*")
				}
			}

		}
		if strings.Contains(m.Content, "timer cancel") && !strings.Contains(m.Content, "add") {
			dg.ChannelMessageSend(m.ChannelID, "*Timer Cancelled!*")
			CancelTimer = true
		}

		if strings.Contains(m.Content, "timer") && !strings.Contains(m.Content, "add") && strings.Contains(m.Content, "> ") {
			time, err := strconv.Atoi(strings.Split(m.Content, "> ")[1])
			if  err != nil {
				dg.ChannelMessageSend(m.ChannelID, "*Invalid parameters*")
			} else {
				if time > 0 {
					if time > 1000 {
						dg.ChannelMessageSend(m.ChannelID, "*Time must be less than 1000 seconds*")
					} else {
						if !Timing {
							s := gocron.NewScheduler()
							Timing = true
							timeMessage,_ := dg.ChannelMessageSend(m.ChannelID, "```TIMER: " + strconv.Itoa(time) + "```")
							s.Every(1).Second().Do(func() {
								time -= 2
								if CancelTimer {
									CancelTimer = false
									time = -100
									s.Clear()
								}
								if time > 0 {
									dg.ChannelMessageEdit(m.ChannelID, timeMessage.ID, "```TIMER: "  + strconv.Itoa(time) + "```")
								} else {
									Timing = false
									if time == -100 {
										dg.ChannelMessageEdit(m.ChannelID, timeMessage.ID, "```TIMER CANCELLED!```")
									} else {
										dg.ChannelMessageEdit(m.ChannelID, timeMessage.ID, "```TIMER FINISHED!```")
									}

									s.Clear()
								}
							})

							<-s.Start()
						} else {
							dg.ChannelMessageSend(m.ChannelID, "*There's already a timer counting down.*")
						}
					}



				} else {
					dg.ChannelMessageSend(m.ChannelID, "Type `s!timer > ` *time in seconds*")
				}
			}
		}

		if strings.Contains(m.Content, "help") && !strings.Contains(m.Content, "add") {
			dg.ChannelMessageSend(m.ChannelID, `
**Type *s!* followed by:**
profile
profile @user
profile edit | bio / favAnime / favManga / link | entry no / leave blank | > entry
usename random OR > *custom name*
timer > *time in seconds* OR cancel
insult *@user*
list insults
**ADMINS:**
add OR delete > *newInsult*`)
		}


	}
}
func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}