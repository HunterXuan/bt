package scheduler

import (
	"github.com/HunterXuan/bt/app/controller/job"
	"github.com/robfig/cron/v3"
	"log"
)

var Scheduler *cron.Cron

func InitScheduler() {
	log.Print("Scheduler Initializing...")

	Scheduler = cron.New()
	_, err := Scheduler.AddJob("@every 30m", &job.Cleaner{})
	if err != nil {
		log.Panicln("InitScheduler:", err)
	}

	_, err = Scheduler.AddJob("@every 30m", &job.Stats{})
	if err != nil {
		log.Panicln("InitScheduler:", err)
	}

	Scheduler.Start()

	log.Println("Scheduler Initialized!")
}
