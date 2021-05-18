package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var dynamicPwd []byte

func init() {
	rand.Seed(time.Now().UnixNano())
	dynamicPwd = make([]byte, 0, 8)
	for i := 0; i < 8; i++ {
		r := rand.Uint64()%10 + '0'
		dynamicPwd = append(dynamicPwd, byte(r))
	}
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Unix()
	fmt.Println("dynamicPwd: ", string(dynamicPwd))
	timer := time.NewTimer(time.Duration(t-now.Unix()) * time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				rand.Seed(time.Now().UnixNano())
				var newPwd []byte = make([]byte, 0, 8)
				for i := 0; i < 8; i++ {
					r := rand.Uint64()%10 + '0'
					newPwd = append(newPwd, byte(r))
				}
				dynamicPwd = newPwd
				now := time.Now()
				fmt.Println("update dynamicPwd: ", string(dynamicPwd))
				t := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Unix()
				timer = time.NewTimer(time.Duration(t-now.Unix()) * time.Second)
				conn, ok := connMap[LOCK_DEVICE]
				if ok {
					sendUpdateDynamicPwd(conn)
				}
			}
		}
	}()
}

func UpdatePwd(w http.ResponseWriter, r *http.Request) {
	var err error
	fmt.Println("enter")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var data PasswordRequest
	err = decoder.Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	fmt.Println(data.Password)
	conn, ok := connMap[0]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sendUpdatePwd(conn, []byte(data.Password))
}

func DynamicPwd(w http.ResponseWriter, r *http.Request) {
	fmt.Println("dynamicPwd:", dynamicPwd)
	for {
		n, err := w.Write(dynamicPwd)
		if err != nil {
			fmt.Println(err)
		}
		if n >= 8 {
			return
		}
	}
}
