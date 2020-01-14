package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/antik9/social-net/internal/config"
	"github.com/antik9/social-net/pkg/models"
)

func chatMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if user := getUserBySession(r); user != nil {
		if other := getUserById(vars); other != nil {
			messages := models.GetMessagesForUsers(user, other)
			listOfMessages := ListOfMessages{Messages: messages, User: user, Other: other}
			marshalled, err := json.Marshal(listOfMessages)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Write(marshalled)
			return
		}
		http.NotFound(w, r)
		return
	}
	http.Error(w, "session does not exist", http.StatusForbidden)
}

func saveChatMessage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	requestData := struct {
		Message  string `json:"message"`
		UserId   int    `json:"userId"`
		FriendId int    `json:"friendId"`
	}{}
	data, _ := ioutil.ReadAll(r.Body)

	err = json.Unmarshal(data, &requestData)
	if err != nil {
		http.Error(w, "inproper data", http.StatusBadRequest)
		return
	}

	if requestData.Message != "" && requestData.UserId != 0 && requestData.FriendId != 0 {
		user := models.GetUserById(requestData.UserId)
		friend := models.GetUserById(requestData.FriendId)
		if user == nil || friend == nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		models.SaveMessage(requestData.Message, user, friend)
	}
	http.Error(w, "empty data", http.StatusBadRequest)
}

func ServeChatForever() {
	router := mux.NewRouter()
	router.HandleFunc("/chat/message", saveChatMessage)
	router.HandleFunc("/chat/{id}", chatMessages)

	log.Fatal(http.ListenAndServe(
		fmt.Sprintf(
			"%s:%s",
			config.Conf.ChatServer.Host, config.Conf.ChatServer.Port,
		), router,
	))
}
