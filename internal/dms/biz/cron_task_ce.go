//go:build !enterprise

package biz

func (ctu *CronTaskUsecase) registerUserActivityCronTasks() error {
	return nil
}
