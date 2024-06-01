package services

import (
	adapters "github.com/WildEgor/e-shop-cdn/internal/adapters/storage"
	"github.com/WildEgor/e-shop-cdn/internal/repositories"
	"github.com/go-co-op/gocron/v2"
	"log/slog"
	"time"
)

type CronService struct {
	sc gocron.Scheduler
	fr repositories.IFilesRepository
	sa adapters.IFileStorage
}

func NewCronService(fr repositories.IFilesRepository, sa adapters.IFileStorage) *CronService {
	sc, _ := gocron.NewScheduler()
	return &CronService{
		sc,
		fr,
		sa,
	}
}

func (s *CronService) Run() error {
	_, err := s.sc.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(s.cleanupJob),
	)

	s.sc.Start()
	slog.Info("cron: start")

	return err
}

func (s *CronService) Stop() error {
	return s.sc.Shutdown()
}

func (s *CronService) cleanupJob() {
	slog.Info("cron: cleanup")

	for file := range s.fr.StreamDeletedFiles() {

		slog.Info("delete file")

		if err := s.sa.Delete(file.Name); err != nil {
			slog.Warn(err.Error())
			continue
		}

		if err := s.fr.DeleteById(file.Id.Hex()); err != nil {
			slog.Warn(err.Error())
			return
		}
	}
}
