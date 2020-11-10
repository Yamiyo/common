package routineutils

import (
	"time"

	"github.com/Yamiyo/common/timeutils"
)

func routineRunnerSample() error {
	st, _ := time.Parse(time.RFC3339, "2019-05-25T10:00:00Z")
	et, _ := time.Parse(time.RFC3339, "2020-07-05T11:00:00Z")

	routineRunner := &RoutineRunner{
		Start: st,
		End:   et,
		//Result: make(map[string]float64),
		Result: make([]float64, 0, 100),
		// SliceFunc 可以不傳, 但 Run 時請呼叫 RunWithSliceMonth 或 RunWithSliceInterval
		SliceFunc: func(start time.Time, end time.Time) []*timeutils.PickTime {
			return timeutils.ParseTimeRangeToPickTimes(start, end, 60)
		},
		Query: func(start time.Time, end time.Time) (interface{}, error) {
			return querySample(start, end)
		},
		MapReduceFunc: func(Result interface{}, QueryResult interface{}) (interface{}, error) {
			//result := Result.([]float64)
			result := Result.(map[string]float64)
			queryResult := QueryResult.(float64)
			if value, ok := result["key"]; !ok {
				result["key"] = queryResult
			} else {
				result["key"] = value + queryResult
			}

			return result, nil
		},
	}

	// 使用傳入的 Func
	if err := Run(routineRunner, 5); err != nil {
		return err
	}
	println(routineRunner.Result.(map[string]float64)["key"])

	// 使用 SliceMonth
	routineRunner.Result = make(map[string]float64)
	if err := RunWithSliceMonth(routineRunner, 5); err != nil {
		return err
	}
	println(routineRunner.Result.(map[string]float64)["key"])

	// 使用 SliceInterval
	routineRunner.Result = make(map[string]float64)
	if err := RunWithSliceInterval(routineRunner, 5, 60); err != nil {
		return err
	}
	println(routineRunner.Result.(map[string]float64)["key"])

	return nil
}

func querySample(s time.Time, e time.Time) (float64, error) {
	return 10.01, nil
}
