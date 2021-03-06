package main

import (
	"flag"
	"log"
	"os"

	"github.com/altid/libs/config"
	"github.com/altid/libs/config/types"
	"github.com/altid/libs/fs"
	"github.com/bwmarrin/discordgo"
)

var (
	mtpt  = flag.String("p", "/tmp/altid", "Path for filesystem")
	srv   = flag.String("s", "discord", "Name of service")
	debug = flag.Bool("d", false, "enable debug logging")
	setup = flag.Bool("conf", false, "Set up config file")
)

func main() {
	flag.Parse()
	if flag.Lookup("h") != nil {
		flag.Usage()
		os.Exit(1)
	}

	conf := &struct {
		Address string              `altid:"address,no_prompt"`
		Auth    types.Auth          `altid:"auth,Authentication method to use"`
		User    string              `altid:"user,prompt:Discord login (email address)"`
		Logdir  types.Logdir        `altid:"logdir,no_prompt"`
		Listen  types.ListenAddress `altid:"listen_address,no_prompt"`
	}{"discordapp.com", "password", "", "", ""}

	if *setup {
		if e := config.Create(conf, *srv, "", *debug); e != nil {
			log.Fatal(e)
		}

		os.Exit(0)
	}

	if e := config.Marshal(conf, *srv, "", *debug); e != nil {
		log.Fatal(e)
	}

	dg, err := discordgo.New(conf.User, string(conf.Auth))
	if err != nil {
		log.Fatalf("Error initiating discord session %v", err)
	}

	s := &server{}
	dg.AddHandler(s.ready)
	dg.AddHandler(s.msgCreate)
	dg.AddHandler(s.msgUpdate)
	dg.AddHandler(s.msgDelete)
	dg.AddHandler(s.chanPins)
	dg.AddHandler(s.chanCreate)
	dg.AddHandler(s.chanUpdate)
	dg.AddHandler(s.chanDelete)
	dg.AddHandler(s.guildUpdate)
	dg.AddHandler(s.guildMemNew)
	dg.AddHandler(s.guildMemBye)
	dg.AddHandler(s.guildMemUpd)
	dg.AddHandler(s.userUpdate)

	ctrl, err := fs.New(s, string(conf.Logdir), *mtpt, *srv, "feed", *debug)
	if err != nil {
		log.Fatal(err)
	}

	defer ctrl.Cleanup()

	s.c = ctrl
	s.dg = dg

	ctrl.SetCommands(Commands...)
	ctrl.CreateBuffer("server", "feed")

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer dg.Close()
	ctrl.Listen()
}
