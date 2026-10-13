package main

import (
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
)

func main() {
	// جلب التوكين من متغيرات البيئة (Environment Variables)
	botToken := os.Getenv("8201691244:AAHAouLJOTPhXPa4huf4Qr_q34yb11E7PFY")
	if botToken == "" {
		log.Fatal("خطأ: لم يتم العثور على BOT_TOKEN في متغيرات البيئة!")
	}

	// إعدادات البوت
	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("فشل في بدء تشغيل البوت: %v", err)
		return
	}

	// أمر التشغيل الأولي /start
	b.Handle("/start", func(c tele.Context) error {
		return c.Send("🦅 أهلاً بك! بوت صقر الكردي يعمل الآن بأقصى سرعة وقوة باستخدام لغة Go 🚀")
	})

	// الرد التلقائي على الرسائل النصية
	b.Handle(tele.OnText, func(c tele.Context) error {
		text := c.Text()
		return c.Send("تم استلام رسالتك بنجاح: " + text)
	})

	log.Println("⚡ البوت يعمل بنجاح وبأعلى أداء...")
	b.Start()
}
