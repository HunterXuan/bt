package job

import (
	"github.com/HunterXuan/bt/app/domain/service/stats"
	"log"
)

type Stats struct{}

func (s *Stats) Run() {
	log.Println("Stats start generating")

	if err := stats.UpdateStatsCache(); err != nil {
		log.Println("Stats (err):", err)
	}

	log.Println("Stats finish generating")
}
