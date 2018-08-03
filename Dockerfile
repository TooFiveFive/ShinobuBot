FROM golang:1.8-onbuild

ONBUILD RUN go get github.com/bwmarrin/discordgo
ONBUILD RUN go get github.com/SlyMarbo/rss
ONBUILD RUN go get github.com/jasonlvhit/gocron