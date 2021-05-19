package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const LOCK_DEVICE = 0
const (
	REGISTER_REQ = iota
	HB
	ERR_NOTIFY_REQ

	REGISTER_RESP
	UPDATE_PASSWD
	TEMP_PASSWD

	UPDATE_PASSWD_RESP
	TEMP_PASSWD_RESP
)

var (
	RecvMail = []string{"386344008@qq.com"}
	Tip      = "家庭智能管家安全提醒"
)

var connMap map[int]*Conn

func init() {
	connMap = make(map[int]*Conn, 8)
}

type Conn struct {
	net.Conn
	ctx              context.Context
	cancelFunc       context.CancelFunc
	done             chan error
	passwd           []byte
	writeCh          chan []byte
	errCnt           int
	errInputNotifyCh chan struct{}
	updateNotifyCh   chan struct{}
	tempPwdNotifyCh  chan struct{}
	lastLiveTime     int64
}

func ConnHandle(rowConn net.Conn) {
	defer rowConn.Close()

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	conn := &Conn{
		Conn:             rowConn,
		ctx:              ctx,
		cancelFunc:       cancel,
		done:             make(chan error, 3),
		writeCh:          make(chan []byte, 128),
		errInputNotifyCh: make(chan struct{}, 128),
		updateNotifyCh:   make(chan struct{}, 128),
		tempPwdNotifyCh:  make(chan struct{}, 128),
	}
	oldConn, ok := connMap[LOCK_DEVICE]
	if ok {
		log.Println("old connect is exist, close it.")
		oldConn.Close()
	}
	connMap[LOCK_DEVICE] = conn

	wg.Add(4)
	go readHandle(wg, conn)
	go writeHandle(wg, conn)
	go PwdErrListener(wg, conn)
	go UpdateDynamicPwd(wg, conn)

	select {
	case <-conn.done:
		cancel()
	case <-conn.ctx.Done():
	}
	wg.Wait()
	log.Println("conn close")
}

func readHandle(wg *sync.WaitGroup, conn *Conn) {
	defer wg.Done()

	var buff []byte = make([]byte, 4)
	for {
		select {
		case <-conn.ctx.Done():
			return
		default:
			_, err := io.ReadFull(conn, buff)
			if err != nil {
				log.Printf("read err: %v", err)
				conn.done <- err
				return
			}
			if buff[0] != start {
				log.Printf("start sign err")
				conn.done <- errors.New("start sign err")
				return
			}
			msgType := buff[2]
			size := buff[3]
			data := make([]byte, size+1)
			_, err = io.ReadFull(conn, data)
			if err != nil {
				log.Printf("read err: %v", err)
				conn.done <- err
				return
			}
			if data[size] != end {
				log.Printf("end sign err")
				conn.done <- errors.New("end sign err")
				return
			}
			data = data[:size]

			switch msgType {
			case REGISTER_REQ:
				if size == 6 {
					conn.passwd = data[0:6]
				}
				log.Println("register success: ", data)
				log.Println("register resp temppwd: ", dynamicPwd)
				sendRegisterResp(conn)
				sendUpdateDynamicPwd(conn)
			case HB:
				log.Println("hb")
				sendHbResp(conn)
				break
			case ERR_NOTIFY_REQ:
				conn.errInputNotifyCh <- struct{}{}
			case UPDATE_PASSWD_RESP:
				conn.updateNotifyCh <- struct{}{}
			case TEMP_PASSWD_RESP:
				conn.tempPwdNotifyCh <- struct{}{}
			}
		}
	}
}

func writeHandle(wg *sync.WaitGroup, conn *Conn) {
	defer wg.Done()
	for {
		select {
		case data := <-conn.writeCh:
			size := len(data)
			var n int = 0
			var err error
			for n < size {
				n, err = conn.Write(data)
				if err != nil {
					log.Printf("write err:%v", err)
					conn.done <- err
					return
				}
				data = data[n:]
			}
		case <-conn.ctx.Done():
			return
		}
	}
}

func PwdErrListener(wg *sync.WaitGroup, conn *Conn) {
	defer wg.Done()

	timerCh := time.After(60 * time.Second)
	for {
		select {
		case <-timerCh:
			conn.errCnt = 0
		case <-conn.errInputNotifyCh:
			conn.errCnt++
			timerCh = time.After(60 * time.Second)
			if conn.errCnt >= 3 {
				conn.errCnt = 0
				//定义收件人
				mailTo := RecvMail
				//邮件主题为"Hello"
				subject := Tip
				// 邮件正文
				body := "你的密码锁多次密码输入错误，如非本人操作请注意财产安全！！！"
				err := SendMail(mailTo, subject, body)
				if err != nil {
					log.Println("send fail", err)
				}
			}
		case <-conn.ctx.Done():
			return
		}
	}
}

func UpdateDynamicPwd(wg *sync.WaitGroup, conn *Conn) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			sendUpdateDynamicPwd(conn)
		case <-conn.ctx.Done():
			return
		}
	}

}

func sendRegisterResp(conn *Conn) {
	sendData := make([]byte, 0, 5)
	sendData = append(sendData, start, REGISTER_RESP, 8)
	sendData = append(sendData, dynamicPwd...)
	sendData = append(sendData, end)

	conn.writeCh <- sendData
}

func sendHbResp(conn *Conn) {
	sendData := make([]byte, 0, 5)
	sendData = append(sendData, start, HB, 1)
	sendData = append(sendData, 9)
	sendData = append(sendData, end)

	conn.writeCh <- sendData
}

func sendUpdateDynamicPwd(conn *Conn) {
	sendData := make([]byte, 0, 5)
	sendData = append(sendData, start, TEMP_PASSWD, 8)
	sendData = append(sendData, dynamicPwd...)
	sendData = append(sendData, end)

	conn.writeCh <- sendData
}

func sendUpdatePwd(conn *Conn, newPwd []byte) {
	sendData := make([]byte, 0, 5)
	sendData = append(sendData, start, UPDATE_PASSWD, 6)
	sendData = append(sendData, newPwd...)
	sendData = append(sendData, end)

	conn.writeCh <- sendData
}
