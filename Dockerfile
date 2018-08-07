FROM arm32v6/golang:1.10.3-alpine

ONBUILD RUN go get github.com/bwmarrin/discordgo
ONBUILD RUN go get github.com/SlyMarbo/rss
ONBUILD RUN go get github.com/jasonlvhit/gocron