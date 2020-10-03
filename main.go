package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
	"github.com/theykk/telegram-gitlab/gitlab"
)

const (
	path = "/webhooks"
)

var (
	Version = "dev"
)

type Config struct {
	Telegram struct {
		Chat  int64  `yaml:"chat"`
		Token string `yaml:"token"`
	} `yaml:"telegram"`

	Server struct {
		Port   int    `yaml:"port"`
		Secret string `yaml:"secret"`
	} `yaml:"server"`
}

func main() {
	log.Printf("Gitlab telegram bot : %s", Version)
	viper.SetConfigName("config")                 // name of config file (without extension)
	viper.SetConfigType("yaml")                   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/telegram-gitlab/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.telegram-gitlab") // call multiple times to add many search paths
	viper.AddConfigPath(".")

	// optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	var conf Config
	erru := viper.Unmarshal(&conf)
	if erru != nil {
		log.Fatalf("Unable to decode into struct, %v", erru)
	}

	//hook, _ := gitlab.New(gitlab.Options.Secret(conf.Server.Secret))
	bot, err := tgbotapi.NewBotAPI(conf.Telegram.Token)
	if err != nil {
		log.Panicf("Bot can't connect , %v", err)
	}
	bot.Debug = true

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)

		token := r.Header.Get("X-Gitlab-Token")
		if token != conf.Server.Secret {
			log.Print("Secret mismatch")
		}
		var pay gitlab.Gitlab
		err := json.Unmarshal(body, &pay)
		if err != nil {
			fmt.Printf("Json unmarshal error, %v", err)
		}
		fmt.Printf("%v", pay)

		if pay.ObjectAttributes.Status == "pending" {

			rawMsgS := fmt.Sprintf("▶️ *%s* tarafından *%s/%s* projesi için *CI/CD* çalıştırıldı.", pay.User.Name, pay.Project.Namespace, pay.Project.Name)
			msg := tgbotapi.NewMessage(conf.Telegram.Chat, rawMsgS)
			msg.ParseMode = "markdown"

			_, berr := bot.Send(msg)
			if berr != nil {
				log.Printf("Message can't send, %v", berr)
			}
			_, _ = fmt.Fprint(w, "OK")
		} else {
			var rawMsg string
			for _, s := range pay.Builds {
				if s.Status == "success" {
					rawMsg = fmt.Sprintf("✅ *%s* tarafından *%s/%s* projesi için *%s* işi çalıştırıldı. İş başarıyla tamamlandı.", s.User.Name, pay.Project.Namespace, pay.Project.Name, s.Name)
				} else if s.Status == "failed" {
					rawMsg = fmt.Sprintf("❌ *%s* tarafından *%s/%s* projesi için *%s* işi çalıştırıldı. İş başarısız oldu.", s.User.Name, pay.Project.Namespace, pay.Project.Name, s.Name)
				} else if s.Status == "skipped" {
					rawMsg = fmt.Sprintf("⛔ *%s* tarafından *%s/%s* projesi için çalıştırılan *%s* işi, bir önceki iş başarısız olduğu için durduruldu.", s.User.Name, pay.Project.Namespace, pay.Project.Name, s.Name)
				} else {
					rawMsg = ""
				}
				if len(rawMsg) < 0 {
					return
				}
				msg := tgbotapi.NewMessage(conf.Telegram.Chat, rawMsg)
				msg.ParseMode = "markdown"
				_, berr := bot.Send(msg)
				if berr != nil {
					log.Printf("Message can't send, %v", berr)
				}
			}
		}

	})
	herr := http.ListenAndServe(":"+strconv.Itoa(conf.Server.Port), nil)
	if herr != nil {
		log.Fatalf("Server can't start, %v", herr)
	}
	log.Printf("Http serve port : %v", conf.Server.Port)
}
