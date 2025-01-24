package commands

import (
	"context"
	"os"

	"github.com/RobBrazier/readflow/internal/config"
	"github.com/adhocore/gronx/pkg/tasker"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

type ScheduleCommand struct {
	ctx      context.Context
	cfg      *config.Config
	cmd      *cli.Command
	schedule string
	timezone string
	sync     SyncCommand
}

func (c ScheduleCommand) Run() error {

	task := tasker.New(tasker.Option{
		Verbose: c.cmd.Bool("verbose"),
		Tz:      c.timezone,
	}).WithContext(c.ctx)

	task.Log = log.Default().StandardLog()

	task.Task(c.schedule, func(ctx context.Context) (int, error) {
		err := NewSyncCommand(ctx, c.cmd)
		if err != nil {
			return 1, err
		}
		return 0, nil
	})

	task.Run()
	return nil
}

func getFallback(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func NewScheduleCommand(ctx context.Context, cmd *cli.Command) error {
	cfg := config.GetFromContext(ctx)
	schedule := getFallback(cmd.String("schedule"), os.Getenv("CRON_SCHEDULE"), "@hourly")
	timezone := getFallback(os.Getenv("TZ"), "Local")
	log.Info("Starting with", "schedule", schedule, "timezone", timezone)

	command := ScheduleCommand{
		ctx:      ctx,
		cmd:      cmd,
		cfg:      cfg,
		schedule: schedule,
		timezone: timezone,
	}
	return command.Run()
}
