package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
)

// type cmd func(in, out chan interface{})
func RunPipeline(cmds ...cmd) {
	in := make(chan interface{})

	wg := sync.WaitGroup{}
	for i, _ := range cmds {
		stageIn := in
		out := make(chan interface{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmds[i](stageIn, out)
			close(out)
		}()
		in = out
	}
	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	wg := sync.WaitGroup{}
	mu := sync.RWMutex{}
	users := make(map[uint64]string)
	for i := range in { // Передаём emails через канал
		wg.Add(1)
		go func() {
			defer wg.Done()

			email := i.(string)
			user := GetUser(email)
			mu.Lock()
			_, exists := users[user.ID] // Если нет idшника в мапе
			if !exists {
				users[user.ID] = user.Email
			}
			mu.Unlock()
			if !exists {
				out <- user
			}
		}()
	}
	wg.Wait()
	// 	in - string
	// 	out - User
}

func SelectMessages(in, out chan interface{}) {
	wg := sync.WaitGroup{}
	var users []User = []User{}
	for j := range in {
		user := j.(User) // Поставили соот-щий тип данных
		users = append(users, user)
		if len(users) < 2 {
			continue
		}
		batch := users
		users = nil
		if len(batch) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				messages, err := GetMessages(batch...)
				if err != nil {
					log.Println("Ошибка получения")
					return
				}
				for _, id := range messages {
					out <- id
				}
			}()
		}
	}
	if len(users) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			messages, err := GetMessages(users...)
			if err != nil {
				log.Println("Ошибка получения")
				return
			}
			for _, id := range messages {
				out <- id
			}
		}()
	}

	wg.Wait()
	// 	in - User
	// 	out - MsgID
}

const (
	goroutinesNum = 5
)

func CheckSpam(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for i := 0; i < goroutinesNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range in {
				msg := i.(MsgID)
				spam, err := HasSpam(msg)
				if err != nil {
					fmt.Println("Ошибка проверки на спам")
				}
				out <- MsgData{ID: msg, HasSpam: spam}
			}
		}()
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
	var spamTrue []MsgData = []MsgData{}
	var spamFalse []MsgData = []MsgData{}

	for val := range in {
		data := val.(MsgData)
		if data.HasSpam {
			spamTrue = append(spamTrue, data)
		} else {
			spamFalse = append(spamFalse, data)
		}
	}
	sort.Slice(spamTrue, func(i, j int) bool {
		return spamTrue[i].ID < spamTrue[j].ID
	})
	sort.Slice(spamFalse, func(i, j int) bool {
		return spamFalse[i].ID < spamFalse[j].ID
	})
	for _, data := range spamTrue {
		id := strconv.FormatUint(uint64(data.ID), 10)
		spam := strconv.FormatBool(data.HasSpam)
		out <- spam + " " + id
	}
	for _, data := range spamFalse {
		id := strconv.FormatUint(uint64(data.ID), 10)
		spam := strconv.FormatBool(data.HasSpam)
		out <- spam + " " + id
	}
}
