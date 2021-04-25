/**
 * @Author: FB
 * @Description:
 * @File:  main.go
 * @Version: 1.0.0
 * @Date: 2019/9/7 14:01
 */
package main

import (
	"net/smtp"
	"strings"
)

func SendMail(mailTo []string, subject string, body string) error {
	// 邮箱地址
	userEmail := "535780809@qq.com"
	// 端口号，:25也行
	mailSmtpPort := ":587"
	//邮箱的授权码，去邮箱自己获取
	Mail_Password := "bckzfwqczeahwybjfead"
	// 此处填写SMTP服务器
	Mail_Smtp_Host := "smtp.qq.com"
	auth := smtp.PlainAuth("", userEmail, Mail_Password, Mail_Smtp_Host)
	nickname := "家庭安全小管家"
	user := userEmail
	contentType := "Content-Type: text/plain; charset=UTF-8"


	msg := []byte("To: " + strings.Join(mailTo, ",") + "\r\nFrom: " + nickname +
		"<" + user + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	err := smtp.SendMail(Mail_Smtp_Host+mailSmtpPort, auth, user, mailTo, msg)

	return err
}