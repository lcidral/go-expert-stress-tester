package progress

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/progress"
	"os"
	"sync"
	"time"
)

// representa o progresso e pode rastrear multiplos workders
type Tracker struct {
	pw           progress.Writer
	totalTracker *progress.Tracker
	workers      map[int]*progress.Tracker
	mu           sync.Mutex
	done         bool
}

// cria uma nova instancia de Tracker para o monitoramenteo de multiplos workers
func NewProgressTracker(workers, totalRequests int) *Tracker {
	pw := progress.NewWriter()
	pw.SetAutoStop(false)
	pw.SetTrackerLength(40)
	pw.SetNumTrackersExpected(workers + 1)
	pw.SetStyle(progress.StyleBlocks)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Options.PercentFormat = "%4.1f%%"
	pw.SetOutputWriter(os.Stdout)

	totalTracker := &progress.Tracker{
		Message: "Total Progress",
		Total:   int64(totalRequests),
		Units:   progress.UnitsDefault,
	}
	pw.AppendTracker(totalTracker)

	workersMap := make(map[int]*progress.Tracker)
	requestsPerWorker := totalRequests / workers

	for i := 0; i < workers; i++ {
		tracker := &progress.Tracker{
			Message: fmt.Sprintf("Worker #%d", i+1),
			Total:   int64(requestsPerWorker),
			Units:   progress.UnitsDefault,
		}
		workersMap[i] = tracker
		pw.AppendTracker(tracker)
	}

	return &Tracker{
		pw:           pw,
		totalTracker: totalTracker,
		workers:      workersMap,
	}
}

func (pt *Tracker) Start() {
	go pt.pw.Render()
}

// usa mutex para controlar a sincronia do progresso entre diferentes instancias de Trackers
func (pt *Tracker) Stop() {
	pt.mu.Lock()
	if !pt.done {
		pt.done = true
		pt.pw.Stop()
	}
	pt.mu.Unlock()
}

// usa mutex para garantir e incrementar de forma segura o progresso de um worker
func (pt *Tracker) IncrementWorker(workerID int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if tracker, ok := pt.workers[workerID]; ok {
		tracker.Increment(1)
		pt.totalTracker.Increment(1)
	}
}
