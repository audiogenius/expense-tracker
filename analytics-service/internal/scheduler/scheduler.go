package scheduler

import (
	"context"
	"fmt"
	"time"

	"analytics-service/internal/analytics"
	"analytics-service/internal/messaging"
	"analytics-service/internal/ollama"
	"analytics-service/internal/types"

	"github.com/jackc/pgx/v5/pgxpool"
	cron "github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// Scheduler represents cron scheduler
type Scheduler struct {
	cron      *cron.Cron
	analytics *analytics.Engine
	messaging *messaging.Generator
	ollama    *ollama.Client
	db        *pgxpool.Pool
	chatIDs   []int64
	timezone  *time.Location
}

// NewScheduler creates new scheduler
func NewScheduler(db *pgxpool.Pool, analytics *analytics.Engine, messaging *messaging.Generator, ollama *ollama.Client, chatIDs []int64) *Scheduler {
	// Set Moscow timezone
	moscow, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load Moscow timezone, using UTC")
		moscow = time.UTC
	}

	return &Scheduler{
		cron:      cron.New(cron.WithLocation(moscow)),
		analytics: analytics,
		messaging: messaging,
		ollama:    ollama,
		db:        db,
		chatIDs:   chatIDs,
		timezone:  moscow,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	log.Info().Msg("Starting analytics scheduler")

	// Daily report at 20:00 Moscow time
	_, err := s.cron.AddFunc("0 20 * * *", func() {
		s.runDailyReport(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to add daily report job: %w", err)
	}

	// Anomaly check every 6 hours
	_, err = s.cron.AddFunc("0 */6 * * *", func() {
		s.runAnomalyCheck(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to add anomaly check job: %w", err)
	}

	// Weekly trend analysis on Sundays at 21:00
	_, err = s.cron.AddFunc("0 21 * * 0", func() {
		s.runWeeklyAnalysis(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to add weekly analysis job: %w", err)
	}

	// Health check every hour
	_, err = s.cron.AddFunc("0 * * * *", func() {
		s.runHealthCheck(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to add health check job: %w", err)
	}

	s.cron.Start()
	log.Info().Msg("Analytics scheduler started successfully")

	// Run initial health check
	go s.runHealthCheck(ctx)

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	log.Info().Msg("Stopping analytics scheduler")
	s.cron.Stop()
}

// runDailyReport runs daily financial report
func (s *Scheduler) runDailyReport(ctx context.Context) {
	log.Info().Msg("Running daily report")

	now := time.Now().In(s.timezone)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.timezone)
	endOfDay := startOfDay.Add(24 * time.Hour)

	analysis, err := s.analytics.AnalyzePeriod(ctx, "day", startOfDay, endOfDay)
	if err != nil {
		log.Error().Err(err).Msg("Failed to analyze daily period")
		return
	}

	// Try to enhance with AI if available
	if s.ollama != nil {
		if isHealthy, _ := s.ollama.HealthCheck(ctx); isHealthy {
			aiMessage, err := s.ollama.GenerateDailyReport(ctx, *analysis)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to generate AI-enhanced daily report, using fallback")
			} else {
				// Use AI-generated message
				analysis.Insights = []string{aiMessage}
			}
		}
	}

	if err := s.messaging.GenerateDailyReport(ctx, analysis, s.chatIDs); err != nil {
		log.Error().Err(err).Msg("Failed to send daily report")
	}
}

// runAnomalyCheck runs anomaly detection
func (s *Scheduler) runAnomalyCheck(ctx context.Context) {
	log.Info().Msg("Running anomaly check")

	now := time.Now().In(s.timezone)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.timezone)
	endOfDay := startOfDay.Add(24 * time.Hour)

	analysis, err := s.analytics.AnalyzePeriod(ctx, "day", startOfDay, endOfDay)
	if err != nil {
		log.Error().Err(err).Msg("Failed to analyze period for anomaly check")
		return
	}

	// Only send alerts if there are high-severity anomalies
	hasHighSeverity := false
	for _, anomaly := range analysis.Anomalies {
		if anomaly.Severity == "high" {
			hasHighSeverity = true
			break
		}
	}

	if hasHighSeverity {
		if err := s.messaging.GenerateAnomalyAlert(ctx, analysis, s.chatIDs); err != nil {
			log.Error().Err(err).Msg("Failed to send anomaly alert")
		}
	}
}

// runWeeklyAnalysis runs weekly trend analysis
func (s *Scheduler) runWeeklyAnalysis(ctx context.Context) {
	log.Info().Msg("Running weekly analysis")

	now := time.Now().In(s.timezone)
	startOfWeek := now.AddDate(0, 0, -7)
	endOfWeek := now

	analysis, err := s.analytics.AnalyzePeriod(ctx, "week", startOfWeek, endOfWeek)
	if err != nil {
		log.Error().Err(err).Msg("Failed to analyze weekly period")
		return
	}

	// Try to enhance with AI if available
	if s.ollama != nil {
		if isHealthy, _ := s.ollama.HealthCheck(ctx); isHealthy {
			aiMessage, err := s.ollama.GenerateFinancialInsight(ctx, *analysis)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to generate AI-enhanced weekly analysis, using fallback")
			} else {
				// Use AI-generated insights
				analysis.Insights = []string{aiMessage}
			}
		}
	}

	if err := s.messaging.GenerateTrendNotification(ctx, analysis, s.chatIDs); err != nil {
		log.Error().Err(err).Msg("Failed to send weekly analysis")
	}
}

// runHealthCheck runs health check
func (s *Scheduler) runHealthCheck(ctx context.Context) {
	log.Debug().Msg("Running health check")

	// Check Ollama availability
	if s.ollama != nil {
		isHealthy, err := s.ollama.HealthCheck(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Ollama health check failed")
		} else if !isHealthy {
			log.Warn().Msg("Ollama is not available, using fallback analytics")
		} else {
			log.Debug().Msg("Ollama is healthy")
		}
	}

	// Check database connection
	if err := s.db.Ping(ctx); err != nil {
		log.Error().Err(err).Msg("Database health check failed")
	} else {
		log.Debug().Msg("Database is healthy")
	}
}

// GetScheduledJobs returns list of scheduled jobs
func (s *Scheduler) GetScheduledJobs() []types.ScheduledJob {
	entries := s.cron.Entries()
	jobs := make([]types.ScheduledJob, len(entries))

	for i, entry := range entries {
		jobs[i] = types.ScheduledJob{
			ID:       int(entry.ID),
			Next:     entry.Next,
			Prev:     entry.Prev,
			Schedule: fmt.Sprintf("%v", entry.Schedule),
		}
	}

	return jobs
}

// AddCustomJob adds a custom scheduled job
func (s *Scheduler) AddCustomJob(spec string, job func()) (cron.EntryID, error) {
	return s.cron.AddFunc(spec, job)
}

// RemoveJob removes a scheduled job
func (s *Scheduler) RemoveJob(id cron.EntryID) {
	s.cron.Remove(id)
}
