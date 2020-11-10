package routineutils

import (
	"context"
	"errors"
	"time"

	"github.com/Yamiyo/common/timeutils"
)

const (
	Default     = 0
	MonthSlice  = 1
	MinInterval = 2
)

type RoutineRunnable interface {
	getPickTimes(status int, minInterval *int) ([]*timeutils.PickTime, error)
	goSliceFunc(start time.Time, end time.Time) []*timeutils.PickTime
	goQuery(start time.Time, end time.Time) (interface{}, error)
	goMapReduceFunc(QueryResult interface{}) error
}

type RoutineRunner struct {
	Start         time.Time
	End           time.Time
	Result        interface{}
	SliceFunc     func(start time.Time, end time.Time) []*timeutils.PickTime
	Query         func(start time.Time, end time.Time) (interface{}, error)
	MapReduceFunc func(Result interface{}, QueryResult interface{}) (interface{}, error)
}

func (rr *RoutineRunner) getPickTimes(status int, minInterval *int) ([]*timeutils.PickTime, error) {
	switch status {
	case Default:
		if nil == rr.SliceFunc {
			return nil, errors.New("plz set slice function")
		}
		return rr.SliceFunc(rr.Start, rr.End), nil
	case MonthSlice:
		return timeutils.GetMonthSlice(rr.Start, rr.End), nil
	case MinInterval:
		return timeutils.ParseTimeRangeToPickTimes(rr.Start, rr.End, *minInterval), nil
	default:
		return nil, errors.New("plz set slice function")
	}
}

func (rr *RoutineRunner) goSliceFunc(start time.Time, end time.Time) []*timeutils.PickTime {
	return rr.SliceFunc(start, end)
}

func (rr *RoutineRunner) goQuery(start time.Time, end time.Time) (interface{}, error) {
	if nil == rr.Query {
		return nil, errors.New("plz set query function")
	}
	return rr.Query(start, end)
}

func (rr *RoutineRunner) goMapReduceFunc(QueryResult interface{}) error {
	if nil == rr.MapReduceFunc {
		return errors.New("plz set query function")
	}
	result, err := rr.MapReduceFunc(rr.Result, QueryResult)
	rr.Result = result
	return err
}

func Run(routineRunnable RoutineRunnable, poolSize int) error {
	pickTimes, err := routineRunnable.getPickTimes(Default, nil)
	if err != nil {
		return err
	}
	return work(routineRunnable, pickTimes, poolSize)
}

func RunWithSliceMonth(routineRunnable RoutineRunnable, poolSize int) error {
	pickTimes, err := routineRunnable.getPickTimes(MonthSlice, nil)
	if err != nil {
		return err
	}
	return work(routineRunnable, pickTimes, poolSize)
}

func RunWithSliceInterval(routineRunnable RoutineRunnable, poolSize int, minInterval int) error {
	pickTimes, err := routineRunnable.getPickTimes(MinInterval, &minInterval)
	if err != nil {
		return err
	}
	return work(routineRunnable, pickTimes, poolSize)
}

func work(routineRunnable RoutineRunnable, pickTimes []*timeutils.PickTime, poolSize int) error {
	if poolSize <= 0 {
		return errors.New("pool size not less than 0")
	}
	if len(pickTimes) <= 0 {
		return errors.New("time slice size not less than 0")
	}
	if poolSize > len(pickTimes) {
		poolSize = len(pickTimes)
	}

	resultChan := make(chan interface{}, len(pickTimes))
	errChan := make(chan error, len(pickTimes))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workChan := make(chan *timeutils.PickTime, len(pickTimes))
	for _, pickTime := range pickTimes {
		workChan <- pickTime
	}

	for i := 0; i < poolSize; i++ {
		go query(routineRunnable, ctx, resultChan, errChan, workChan)
	}

	for range pickTimes {
		select {
		case err := <-errChan:
			cancel()
			return err
		case mapReduceItems := <-resultChan:
			if err := routineRunnable.goMapReduceFunc(mapReduceItems); err != nil {
				return err
			}
		}
	}

	return nil
}

func query(routineRunnable RoutineRunnable, ctx context.Context, retChan chan interface{}, errChan chan error, workChan chan *timeutils.PickTime) {
	for {
		select {
		case pickTime := <-workChan:
			item, err := routineRunnable.goQuery(pickTime.StartTime, pickTime.EndTime)
			if err != nil {
				errChan <- err
				return
			}
			retChan <- item
		case <-ctx.Done():
			return
		}
	}
}
